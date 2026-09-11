// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/sprimault/ormeau/internal/calque"
	"github.com/sprimault/ormeau/internal/introspection"
)

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

	base := introspection.BaseDuDSN(c.dsn)
	if !motifBase.MatchString(base) {
		repondreErreur(w, http.StatusBadRequest,
			fmt.Sprintf("la base %q ne peut pas nommer un fichier : lettres, chiffres, tiret et souligné seulement", base))
		return
	}

	fichier := base + ".calque.json"
	physique, err := calque.LirePhysique(filepath.Join(s.repertoire, fichier))
	switch {
	case errors.Is(err, fs.ErrNotExist):
		repondreErreur(w, http.StatusNotFound, fmt.Sprintf("aucun %s dans le répertoire de travail", fichier))
		return
	case err != nil:
		repondreErreur(w, http.StatusUnprocessableEntity, fmt.Sprintf("%s illisible : %v", fichier, err))
		return
	}

	retirees := len(physique.Statistiques) > 0
	physique.Statistiques = nil

	contenu, err := calque.Serialiser(physique)
	if err != nil {
		repondreErreur(w, http.StatusInternalServerError, err.Error())
		return
	}
	repondreJSON(w, http.StatusOK, ReponseCalque{
		Fichier:              fichier,
		ExtraitLe:            physique.Source.ExtraitLe,
		Empreinte:            physique.Source.Empreinte,
		Contenu:              string(contenu),
		StatistiquesRetirees: retirees,
	})
}
