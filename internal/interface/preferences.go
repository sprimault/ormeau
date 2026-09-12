// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"net/http"
	"sync"

	"github.com/sprimault/ormeau/internal/config"
)

// preferences garde en mémoire ce que le fichier porte, et sérialise ses
// réécritures.
//
// En mémoire parce que la page d'entrée les relit à chaque chargement, et
// qu'une lecture de disque par rafraîchissement n'apporte rien : ce serveur est
// le seul à écrire le fichier. Un avertissement de lecture est retenu avec
// elles pour être annoncé une fois, au démarrage, plutôt qu'à chaque requête.
type preferences struct {
	mu            sync.RWMutex
	valeurs       config.Preferences
	emplacements  *config.Emplacements
	avertissement string
}

// nouvellesPreferences lit le fichier et rend l'état de départ.
//
// Une erreur de lecture n'empêche pas de démarrer : l'interface s'ouvre en
// thème système et en français, ce qui reste utilisable. Refuser de servir
// l'écran de connexion parce qu'un fichier d'affichage est illisible serait
// hors de proportion.
func nouvellesPreferences(e *config.Emplacements) *preferences {
	p := &preferences{emplacements: e, valeurs: config.PreferencesParDefaut()}
	if e == nil {
		return p
	}

	valeurs, avertissement, err := e.LirePreferences()
	p.valeurs = valeurs
	switch {
	case err != nil:
		p.avertissement = err.Error()
	default:
		p.avertissement = avertissement
	}
	return p
}

// lire rend une copie des préférences courantes.
func (p *preferences) lire() config.Preferences {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.valeurs
}

// ecrire valide, enregistre, et ne retient en mémoire que ce qui a été écrit.
//
// Dans cet ordre : garder en mémoire une valeur que le disque n'a pas reçue
// ferait diverger la page servie du fichier, et la divergence ne se verrait
// qu'au prochain lancement.
func (p *preferences) ecrire(valeurs config.Preferences) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.emplacements != nil {
		if err := p.emplacements.EcrirePreferences(valeurs); err != nil {
			return err
		}
	}
	p.valeurs = valeurs
	return nil
}

// pagePreferences rend ce que le template injecte dans la page d'entrée.
func (s *serveur) pagePreferences() pageHTML {
	v := s.preferences.lire()

	page := pageHTML{
		Theme:          v.Theme,
		Langue:         v.Langue,
		LargeurArbre:   v.LargeurArbre,
		HauteurApercu:  v.HauteurApercu,
		LargeurEntites: v.LargeurEntites,
	}
	if v.ApercuOuvert != nil {
		page.ApercuOuvert = "0"
		if *v.ApercuOuvert {
			page.ApercuOuvert = "1"
		}
	}
	return page
}

// gererPreferences lit ou enregistre les préférences d'affichage.
//
// La lecture sert aux relectures ultérieures, jamais au premier rendu : celui-
// ci reçoit ses valeurs dans la page elle-même, et un aller-retour les ferait
// arriver après coup, c'est-à-dire après la bascule qu'on veut supprimer.
func (s *serveur) gererPreferences(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		repondreJSON(w, http.StatusOK, s.preferences.lire())
	case http.MethodPost:
		var recues config.Preferences
		if !decoder(w, r, &recues) {
			return
		}
		if err := s.preferences.ecrire(recues); err != nil {
			repondreErreur(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		repondreJSON(w, http.StatusOK, s.preferences.lire())
	default:
		repondreErreur(w, http.StatusMethodNotAllowed, "méthode non acceptée")
	}
}
