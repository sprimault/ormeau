// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/sprimault/ormeau/internal/config"
)

// ormeauConfigDir court-circuite la résolution par la plateforme.
//
// Indispensable aux tests, qui écriraient sinon dans la configuration réelle du
// poste, et c'est aussi ce qui rend un usage portable possible — le binaire et
// sa configuration sur la même clé.
const ormeauConfigDir = "ORMEAU_CONFIG_DIR"

// ouvrirConfiguration résout la racine de configuration et pose son
// arborescence.
//
// La variable d'environnement se lit ici et pas dans internal/config, qui reste
// étanche : un paquet interne reçoit ses dépendances, il ne va pas les chercher.
//
// L'échec est fatal plutôt que silencieux. Un répertoire de configuration
// inaccessible veut dire que ni les profils ni les brouillons ne pourront être
// enregistrés, et le découvrir au moment d'enregistrer coûte le travail en
// cours.
func ouvrirConfiguration() (*config.Emplacements, error) {
	racine := os.Getenv(ormeauConfigDir)
	if racine == "" {
		resolue, err := config.RacineParDefaut()
		if err != nil {
			return nil, err
		}
		racine = resolue
	}
	return config.Ouvrir(racine)
}
