// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"github.com/sprimault/ormeau/internal/calque"
	"github.com/sprimault/ormeau/internal/inference"
)

// L'arbitrage se joue hors ligne : aucun point d'entrée de ce fichier ne
// touche une base ni ne demande de session. Il désigne la base par son nom,
// relit son calque dans le répertoire de travail et rejoue l'inférence, qui est
// une fonction pure. C'est ce qui permet d'arbitrer un calque rapporté de chez
// un client, sans jamais rouvrir sa base.

// decisions lit ou enregistre le fichier de décisions d'une base.
func (s *serveur) decisions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.lireDecisions(w, r)
	case http.MethodPost:
		s.ecrireDecisions(w, r)
	default:
		repondreErreur(w, http.StatusMethodNotAllowed, "méthode non acceptée")
	}
}

// lireDecisions rend le fichier de décisions d'une base, et ce qu'il faut pour
// le réécrire sans rien perdre.
func (s *serveur) lireDecisions(w http.ResponseWriter, r *http.Request) {
	base := r.URL.Query().Get("base")
	if !baseValide(w, base) {
		return
	}

	chemin := s.cheminDecisions(base)
	contenu, err := os.ReadFile(chemin) // #nosec G304
	switch {
	case errors.Is(err, fs.ErrNotExist):
		repondreJSON(w, http.StatusOK, ReponseDecisions{})
		return
	case err != nil:
		repondreErreur(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	d, err := inference.DecisionsDepuis(contenu)
	if err != nil {
		repondreErreur(w, http.StatusUnprocessableEntity,
			fmt.Sprintf("%s illisible : %v", filepath.Base(chemin), err))
		return
	}
	repondreJSON(w, http.StatusOK, ReponseDecisions{
		Existe:           true,
		Decisions:        *d,
		EmpreinteFichier: empreinteFichier(contenu),
		Manuel:           inference.ContenuManuel(inference.BaseDesDecisions(chemin), contenu),
	})
}

// ecrireDecisions enregistre le fichier de décisions d'une base.
//
// Trois refus, que leur code distingue : le calque a changé depuis le début de
// l'arbitrage ; le fichier a changé sur disque depuis sa lecture, retouché dans
// un éditeur pendant que l'écran était ouvert ; ou il porte un travail humain
// que la réécriture perdrait et que l'écran n'a pas confirmé vouloir écraser.
//
// Les écritures sont sérialisées : deux onglets qui enregistrent ensemble ne
// peuvent pas passer tous deux la garde d'empreinte avant que l'un écrive.
func (s *serveur) ecrireDecisions(w http.ResponseWriter, r *http.Request) {
	var requete RequeteEcritureDecisions
	if !decoder(w, r, &requete) {
		return
	}
	if requete.EmpreintePhysique == "" {
		repondreErreur(w, http.StatusBadRequest, "l'empreinte du calque est requise : l'écriture porte sur un calque lu")
		return
	}
	physique, ok := s.calqueDeBase(w, requete.Base, requete.EmpreintePhysique)
	if !ok {
		return
	}

	s.ecriture.Lock()
	defer s.ecriture.Unlock()

	chemin := s.cheminDecisions(requete.Base)
	actuel, err := os.ReadFile(chemin) // #nosec G304
	existe := err == nil
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		repondreErreur(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	var empreinteActuelle string
	if existe {
		empreinteActuelle = empreinteFichier(actuel)
	}
	if requete.EmpreinteFichier != empreinteActuelle {
		repondreRefus(w, CodeDecisionsModifiees, "le fichier de décisions a changé sur disque depuis sa lecture")
		return
	}

	base := inference.BaseDesDecisions(chemin)
	if existe && !requete.EcraserManuel && inference.ContenuManuel(base, actuel) {
		repondreRefus(w, CodeContenuManuel, "le fichier de décisions porte des modifications faites à la main, que la réécriture perdrait")
		return
	}

	contenu := inference.EcrireDecisions(physique, &requete.Decisions, base)
	if err := ecrireEnPlace(chemin, contenu); err != nil {
		repondreErreur(w, http.StatusInternalServerError, err.Error())
		return
	}
	repondreJSON(w, http.StatusOK, ReponseEcritureDecisions{EmpreinteFichier: empreinteFichier(contenu)})
}

// inferer rejoue l'inférence sur le calque d'une base, avec le brouillon de
// décisions de l'écran.
//
// La réponse ne porte pas le calque logique. Sur quatre cents tables, il pèse
// plusieurs mégaoctets, et l'écran le redemande à chaque modification : un
// résumé par entité suffit à la liste, le détail se demande à l'ouverture.
func (s *serveur) inferer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		repondreErreur(w, http.StatusMethodNotAllowed, "méthode non acceptée")
		return
	}

	var requete RequeteInference
	if !decoder(w, r, &requete) {
		return
	}
	physique, ok := s.calqueDeBase(w, requete.Base, requete.EmpreintePhysique)
	if !ok {
		return
	}

	logique, avertissements := inference.Inferer(physique, &requete.Decisions)
	repondreJSON(w, http.StatusOK, ReponseInference{
		EmpreintePhysique: physique.Source.Empreinte,
		Avertissements:    jamaisNil(avertissements),
		Propositions:      jamaisNil(inference.Proposer(physique, &requete.Decisions)),
		Enumerations:      enumerationsInferees(logique),
		Entites:           resumes(logique),
		TypesDoctrine:     inference.TypesDoctrine(),
	})
}

// entite rend le détail d'une entité inférée et la table physique dont elle
// vient, pour les voir côte à côte.
func (s *serveur) entite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		repondreErreur(w, http.StatusMethodNotAllowed, "méthode non acceptée")
		return
	}

	var requete RequeteEntite
	if !decoder(w, r, &requete) {
		return
	}
	physique, ok := s.calqueDeBase(w, requete.Base, requete.EmpreintePhysique)
	if !ok {
		return
	}

	table := physique.TableParNom(requete.Schema, requete.Table)
	if table == nil {
		repondreErreur(w, http.StatusNotFound,
			fmt.Sprintf("table %s.%s absente du calque", requete.Schema, requete.Table))
		return
	}

	logique, _ := inference.Inferer(physique, &requete.Decisions)
	reponse := ReponseEntite{TablePhysique: *table}
	for i := range logique.Entites {
		if e := &logique.Entites[i]; e.Table.Schema == requete.Schema && e.Table.Nom == requete.Table {
			reponse.Entite = e
			break
		}
	}
	repondreJSON(w, http.StatusOK, reponse)
}

// resumes rend ce que la liste montre de chaque entité.
func resumes(logique *calque.Logique) []ResumeEntite {
	resumes := make([]ResumeEntite, 0, len(logique.Entites))
	for i := range logique.Entites {
		e := &logique.Entites[i]
		resumes = append(resumes, ResumeEntite{
			Nom:            e.Nom,
			Table:          e.Table,
			NbProprietes:   len(e.Proprietes),
			NbAssociations: len(e.Associations),
			Heritage:       e.Heritage,
			Traits:         e.Traits,
			Origine:        e.Origine,
		})
	}
	return resumes
}

// enumerationsInferees rend les énumérations du calque logique, chacune avec les
// colonnes qui la portent.
//
// Le calque logique ne rattache pas une énumération à sa colonne : on la
// retrouve par les propriétés qui la nomment, traits compris. Or c'est la
// colonne que le fichier de décisions attend pour nommer des cas.
func enumerationsInferees(logique *calque.Logique) []EnumerationInferee {
	traits := map[string][]calque.Propriete{}
	for _, t := range logique.Traits {
		traits[t.Nom] = t.Proprietes
	}

	colonnes := map[string][]string{}
	for _, e := range logique.Entites {
		proprietes := slices.Clone(e.Proprietes)
		for _, nom := range e.Traits {
			proprietes = append(proprietes, traits[nom]...)
		}
		for _, p := range proprietes {
			if p.Enumeration != "" {
				colonnes[p.Enumeration] = append(colonnes[p.Enumeration], e.Table.Schema+"."+e.Table.Nom+"."+p.Colonne)
			}
		}
	}

	inferees := make([]EnumerationInferee, 0, len(logique.Enumerations))
	for _, en := range logique.Enumerations {
		inferees = append(inferees, EnumerationInferee{
			Nom:         en.Nom,
			TypeSupport: en.TypeSupport,
			Cas:         en.Cas,
			Origine:     en.Origine,
			Colonnes:    jamaisNil(slices.Sorted(slices.Values(colonnes[en.Nom]))),
		})
	}
	return inferees
}

// cheminDecisions compose le fichier de décisions d'une base validée.
func (s *serveur) cheminDecisions(base string) string {
	return filepath.Join(s.repertoireCourant(), base+".decisions.yaml")
}

// empreinteFichier identifie les octets d'un fichier, pour détecter qu'il a
// changé entre sa lecture et son écriture.
func empreinteFichier(contenu []byte) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(contenu))
}

// ecrireEnPlace écrit dans un fichier voisin puis le substitue au fichier visé.
//
// Un arrêt en cours d'écriture ne laisse pas un fichier tronqué à la place
// d'arbitrages commités : c'est l'ancien, ou le nouveau entier.
func ecrireEnPlace(chemin string, contenu []byte) error {
	temporaire, err := os.CreateTemp(filepath.Dir(chemin), "."+filepath.Base(chemin)+".*")
	if err != nil {
		return err
	}
	nom := temporaire.Name()

	_, errEcriture := temporaire.Write(contenu)
	if err := errors.Join(errEcriture, temporaire.Close()); err != nil {
		_ = os.Remove(nom)
		return err
	}
	if err := os.Rename(nom, chemin); err != nil {
		_ = os.Remove(nom)
		return err
	}
	return nil
}

// jamaisNil rend une liste vide plutôt qu'absente : le front la parcourt sans
// distinguer les deux.
func jamaisNil[T any](liste []T) []T {
	if liste == nil {
		return []T{}
	}
	return liste
}
