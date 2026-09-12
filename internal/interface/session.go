// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"errors"
	"io/fs"
	"net/http"
	"os"

	"github.com/sprimault/ormeau/internal/config"
)

// gererSession lit, enregistre ou efface le brouillon d'arbitrage d'une base.
//
// « Session » désigne ici le travail en cours sur un écran, pas la connexion à
// une base : celle-ci vit en mémoire et meurt avec le processus.
func (s *serveur) gererSession(w http.ResponseWriter, r *http.Request) {
	if s.emplacements == nil {
		repondreErreur(w, http.StatusServiceUnavailable,
			"aucun répertoire de configuration : les brouillons ne peuvent pas être enregistrés")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.lireSession(w, r)
	case http.MethodPost:
		s.ecrireSession(w, r)
	case http.MethodDelete:
		s.effacerSession(w, r)
	default:
		repondreErreur(w, http.StatusMethodNotAllowed, "méthode non acceptée")
	}
}

// lireSession rend le brouillon d'une base, ou dit pourquoi il a été écarté.
//
// Les deux empreintes sont vérifiées ici et non dans la page : le serveur a le
// calque et le fichier de décisions sous la main, le navigateur n'a que ce
// qu'on lui a dit. Un brouillon caduc est effacé dans la foulée — le garder
// pour le rejeter à chaque ouverture n'aide personne.
func (s *serveur) lireSession(w http.ResponseWriter, r *http.Request) {
	base := r.URL.Query().Get("base")
	if !baseValide(w, base) {
		return
	}

	travail := s.repertoireCourant()
	session, err := s.emplacements.LireSession(base, travail)
	if err != nil {
		repondreErreur(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if session == nil {
		repondreJSON(w, http.StatusOK, ReponseSession{})
		return
	}

	raison := s.sessionCaduque(*session, base, travail)
	if raison != "" {
		if err := s.emplacements.SupprimerSession(base, travail); err != nil {
			repondreErreur(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		repondreJSON(w, http.StatusOK, ReponseSession{Ecartee: raison})
		return
	}
	repondreJSON(w, http.StatusOK, ReponseSession{Session: session})
}

// sessionCaduque dit pourquoi un brouillon ne s'applique plus, ou rien quand il
// s'applique encore.
//
// Deux états doivent correspondre. Le calque, parce qu'une extraction l'a
// peut-être réécrit et que des décisions prises sur l'ancien porteraient sur un
// schéma qui a bougé. Le fichier de décisions, parce qu'il fait foi : un
// brouillon rejoué par-dessus une version retouchée à la main réintroduirait en
// silence ce qu'on venait d'en retirer.
func (s *serveur) sessionCaduque(session config.Session, base, travail string) string {
	physique, err := s.calques.lire(cheminCalque(travail, base))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return "le calque de cette base n'est plus dans le répertoire de travail"
	case err != nil:
		return "le calque de cette base n'est plus lisible"
	case physique.Source.Empreinte != session.EmpreinteCalque:
		return "le calque a été réextrait depuis ce brouillon"
	}

	contenu, err := os.ReadFile(s.cheminDecisions(base)) // #nosec G304
	courante := ""
	if err == nil {
		courante = empreinteFichier(contenu)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return "le fichier de décisions n'est plus lisible"
	}
	if courante != session.EmpreinteDecisions {
		return "le fichier de décisions a changé depuis ce brouillon"
	}
	return ""
}

// ecrireSession enregistre le brouillon d'une base.
func (s *serveur) ecrireSession(w http.ResponseWriter, r *http.Request) {
	var requete config.Session
	if !decoder(w, r, &requete) {
		return
	}
	if !baseValide(w, requete.Base) {
		return
	}

	if err := s.emplacements.EcrireSession(requete, s.repertoireCourant()); err != nil {
		repondreErreur(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	repondreJSON(w, http.StatusOK, ReponseSession{Session: &requete})
}

// effacerSession retire le brouillon d'une base, ce que fait l'écran après un
// enregistrement réussi : ce qui est dans le fichier n'a plus à être retenu
// ailleurs.
func (s *serveur) effacerSession(w http.ResponseWriter, r *http.Request) {
	var requete ReferenceSession
	if !decoder(w, r, &requete) {
		return
	}
	if !baseValide(w, requete.Base) {
		return
	}

	if err := s.emplacements.SupprimerSession(requete.Base, s.repertoireCourant()); err != nil {
		repondreErreur(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	repondreJSON(w, http.StatusOK, ReponseSession{})
}
