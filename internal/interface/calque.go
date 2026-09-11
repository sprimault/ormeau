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
	"sync"

	"github.com/sprimault/ormeau/internal/calque"
	"github.com/sprimault/ormeau/internal/introspection"
)

// calquesLus garde le dernier calque analysé de chaque fichier.
//
// Le fichier est relu à chaque appel et son contenu comparé : seul un calque
// dont les octets ont changé est réanalysé. Rien ne se fie à la date de
// modification ni à la taille. Une réextraction réécrit souvent le calque dans
// la même seconde et à la même taille, et la garde d'empreinte comparerait
// alors l'ancien calque au lieu du fichier — précisément dans le cas qu'elle
// vise.
type calquesLus struct {
	mu        sync.Mutex
	parChemin map[string]calqueLu
}

// calqueLu apparie un calque analysé à l'empreinte des octets dont il vient.
type calqueLu struct {
	octets   [sha256.Size]byte
	physique *calque.Physique
}

// lire rend le calque d'un fichier, réanalysé seulement si son contenu a
// changé.
//
// Le calque rendu est partagé entre les appels : personne ne le modifie en
// place.
func (c *calquesLus) lire(chemin string) (*calque.Physique, error) {
	// Chemin composé par l'appelant à partir d'un nom de base validé.
	donnees, err := os.ReadFile(chemin) // #nosec G304
	if err != nil {
		return nil, err
	}
	somme := sha256.Sum256(donnees)

	c.mu.Lock()
	lu, connu := c.parChemin[chemin]
	c.mu.Unlock()
	if connu && lu.octets == somme {
		return lu.physique, nil
	}

	// L'analyse revient au paquet calque, seul à lire un calque. Si le fichier
	// change entre les deux lectures, l'analyse porte sur la plus récente ; son
	// contenu différera de l'empreinte retenue, et l'appel suivant la refera.
	physique, err := calque.LirePhysique(chemin)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	if c.parChemin == nil {
		c.parChemin = map[string]calqueLu{}
	}
	c.parChemin[chemin] = calqueLu{octets: somme, physique: physique}
	c.mu.Unlock()
	return physique, nil
}

// calqueDeBase lit le calque d'une base dans le répertoire de travail, et
// vérifie qu'il est celui que l'écran croit juger. Rend false quand il a déjà
// répondu.
//
// Une empreinte vide est celle d'un premier chargement, qui apprend l'empreinte
// courante. Une empreinte différente veut dire qu'une extraction a réécrit le
// calque entre-temps : des décisions prises sur l'ancien portent sur un schéma
// qui a bougé, et l'écran doit le savoir avant d'aller plus loin.
//
// Le refus d'un calque absent nomme le répertoire où il a été cherché.
// L'interface se lance souvent ailleurs que dans le projet, et le fichier
// existe alors bel et bien, à côté : sans le chemin, le message envoie chercher
// un calque qui n'a jamais manqué.
func (s *serveur) calqueDeBase(w http.ResponseWriter, base, empreinte string) (*calque.Physique, bool) {
	if !baseValide(w, base) {
		return nil, false
	}

	fichier := base + ".calque.json"
	physique, err := s.calques.lire(filepath.Join(s.repertoire, fichier))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		repondreErreur(w, http.StatusNotFound, fmt.Sprintf(
			"aucun %s dans %s : extraire la base, ou relancer l'interface depuis le projet qui porte ce calque",
			fichier, s.repertoire))
		return nil, false
	case err != nil:
		repondreErreur(w, http.StatusUnprocessableEntity, fmt.Sprintf("%s illisible : %v", fichier, err))
		return nil, false
	case empreinte != "" && empreinte != physique.Source.Empreinte:
		repondreRefus(w, CodeCalqueModifie, "le calque a changé depuis le début de l'arbitrage : une extraction l'a réécrit")
		return nil, false
	}
	return physique, true
}

// calqueDeSession rend le calque de la base d'une session, tel que le
// répertoire de travail le contient — celui d'une extraction qui vient de
// finir, ou celui qu'une session précédente a laissé.
//
// Le navigateur ne désigne pas le fichier : il nomme une session, et le chemin
// se compose ici à partir de sa base, validée comme à l'extraction.
//
// Les statistiques sont retirées avant l'envoi. Avec l'échantillonnage, elles
// portent des valeurs réelles, et l'interface n'en affiche aucune : elles
// alimentent l'inférence, elles ne se consultent pas.
func (s *serveur) calqueDeSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		repondreErreur(w, http.StatusMethodNotAllowed, "méthode non acceptée")
		return
	}

	c, ok := s.registre.trouver(r.URL.Query().Get("session"))
	if !ok {
		repondreErreur(w, http.StatusNotFound, "connexion inconnue ou déjà fermée")
		return
	}

	physique, ok := s.calqueDeBase(w, introspection.BaseDuDSN(c.dsn), "")
	if !ok {
		return
	}

	// Une copie, et non le calque lu : il est partagé avec l'arbitrage, qui a
	// besoin de ses statistiques.
	sansStatistiques := *physique
	sansStatistiques.Statistiques = nil

	contenu, err := calque.Serialiser(&sansStatistiques)
	if err != nil {
		repondreErreur(w, http.StatusInternalServerError, err.Error())
		return
	}
	repondreJSON(w, http.StatusOK, ReponseCalque{
		Fichier:              introspection.BaseDuDSN(c.dsn) + ".calque.json",
		ExtraitLe:            physique.Source.ExtraitLe,
		Empreinte:            physique.Source.Empreinte,
		Contenu:              string(contenu),
		StatistiquesRetirees: len(physique.Statistiques) > 0,
	})
}
