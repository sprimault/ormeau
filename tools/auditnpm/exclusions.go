// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package main

import "strings"

// Exclusion dit pourquoi un avis ne fait pas échouer la validation, et à quelle
// condition sa ligne se retire. Les deux champs sont obligatoires : une
// exclusion sans date de péremption se reconduit indéfiniment, et une exclusion
// sans raison se relit six mois plus tard sans que personne n'ose y toucher.
type Exclusion struct {
	// Raison dit pourquoi l'avis ne menace pas ce projet. Elle porte sur
	// l'exposition réelle, pas sur la gêne que l'avis occasionne.
	Raison string
	// Retrait dit ce qui doit arriver pour que la ligne disparaisse.
	Retrait string
}

// motivee rend vrai quand l'exclusion porte ses deux champs.
//
// Isolée du test qui l'applique pour que la règle reste éprouvée alors que la
// liste est vide : une boucle sur zéro entrée passe toujours, et ne prouverait
// plus qu'une ligne muette est refusée.
func (e Exclusion) motivee() bool {
	return strings.TrimSpace(e.Raison) != "" && strings.TrimSpace(e.Retrait) != ""
}

// exclusions nomme les avis que la validation laisse passer.
//
// Vide, et c'est son état normal, pas un oubli : son unique entrée couvrait
// GHSA-vfj7-8cjw-p6xm sur braces, dont le mainteneur contestait l'existence
// (micromatch/braces#70), et elle est partie au retrait de l'avis en amont le
// 2026-10-23 — exactement la condition que sa ligne annonçait.
//
// Liste fermée et courte par construction : le contrôle échoue aussi quand une
// entrée d'ici ne correspond plus à rien au rapport, ce qui force à la retirer
// au lieu de la laisser dormir.
//
// N'y entre qu'un avis dont l'exposition est nulle **et** qu'aucune montée de
// version ne peut régler. Un avis corrigé en amont se corrige, il ne s'exclut
// pas, même quand le correctif n'est atteignable que par une montée du paquet
// qui le traîne.
//
// Un avis critique n'y entre jamais : npm est appelé avec
// `--audit-level=critical` et échoue seul dans ce cas, avant que ce contrôle
// ne parle. Une ligne d'ici n'y changerait rien, et c'est voulu.
var exclusions = map[string]Exclusion{}
