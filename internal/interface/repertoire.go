// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// changerRepertoire règle le répertoire de travail depuis l'écran.
//
// C'est la seule exception à « aucun chemin de fichier ne vient du navigateur »,
// et elle est délibérée : sans elle, il faudrait relancer le binaire avec
// --repertoire pour viser un autre projet, et l'interface ne saurait pas
// mémoriser où l'on travaille. Elle tient parce que l'API est déjà gardée par
// jeton, cookie SameSite=Strict et contrôle d'Origin — quelqu'un qui l'atteint
// a déjà la main sur le poste.
//
// Ce que le chemin reçu ne décide pas : les noms de fichiers, toujours composés
// ici à partir d'un nom de base validé. Un chemin ne désigne qu'un répertoire
// existant, jamais un fichier.
func (s *serveur) changerRepertoire(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		repondreErreur(w, http.StatusMethodNotAllowed, "méthode non acceptée")
		return
	}

	var requete RequeteRepertoire
	if !decoder(w, r, &requete) {
		return
	}

	resolu, err := repertoireUtilisable(requete.Repertoire)
	if err != nil {
		repondreErreur(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	s.travail.Lock()
	s.repertoire = resolu
	s.travail.Unlock()

	// Le chemin résolu, et non celui qui a été saisi : un lien symbolique ou un
	// « .. » aboutit ailleurs que là où on croit, et l'écran doit montrer où les
	// fichiers iront vraiment avant qu'on en écrive un.
	repondreJSON(w, http.StatusOK, ReponseContexte{Repertoire: resolu, Version: s.version})
}

// repertoireUtilisable rend le chemin résolu d'un répertoire où l'on peut
// écrire, ou dit ce qui l'en empêche.
//
// Rien n'est créé : demander un répertoire qui n'existe pas est presque
// toujours une faute de frappe, et en créer un à l'aveugle sème des dossiers
// vides dont personne ne saura d'où ils viennent.
func repertoireUtilisable(demande string) (string, error) {
	if demande == "" {
		return "", fmt.Errorf("indiquer un répertoire")
	}
	if !filepath.IsAbs(demande) {
		return "", fmt.Errorf("indiquer un chemin absolu, pas %s", demande)
	}

	// Les liens sont suivis avant toute vérification : sans cela, on
	// contrôlerait les droits d'un chemin et on écrirait dans un autre.
	resolu, err := filepath.EvalSymlinks(filepath.Clean(demande))
	if err != nil {
		return "", fmt.Errorf("%s est introuvable", demande)
	}
	resolu, err = filepath.Abs(resolu)
	if err != nil {
		return "", fmt.Errorf("%s est introuvable", demande)
	}

	infos, err := os.Stat(resolu)
	if err != nil {
		return "", fmt.Errorf("%s est introuvable", demande)
	}
	if !infos.IsDir() {
		return "", fmt.Errorf("%s n'est pas un répertoire", resolu)
	}

	// Par une écriture réelle : sous Windows les bits de permission ne décrivent
	// pas les droits effectifs, et sous Linux ils ignorent les ACL comme le
	// montage en lecture seule.
	essai, err := os.CreateTemp(resolu, ".ormeau-*")
	if err != nil {
		return "", fmt.Errorf("%s n'accepte pas d'écriture", resolu)
	}
	nom := essai.Name()
	_ = essai.Close()
	_ = os.Remove(nom)

	return resolu, nil
}
