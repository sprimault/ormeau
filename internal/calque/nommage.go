// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package calque

import "regexp"

// motifBase est ce qu'un nom de base doit respecter pour nommer les fichiers
// d'une base : <base>.calque.json, <base>.decisions.yaml, <base>.logique.json.
//
// Refusé plutôt que nettoyé : PostgreSQL accepte « ../x » comme nom de base
// entre guillemets, et un nettoyage se contourne là où un refus tient.
var motifBase = regexp.MustCompile(`^[\p{L}\p{N}_-]+$`)

// NomDeBaseValide dit si un nom de base peut préfixer les trois fichiers d'une
// base sans sortir du répertoire où ils s'écrivent. La ligne de commande et
// l'interface appliquent la même règle : une base extraite par l'une doit
// pouvoir se rouvrir dans l'autre.
func NomDeBaseValide(nom string) bool {
	return motifBase.MatchString(nom)
}
