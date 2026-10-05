// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package main

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

// exclusions nomme les avis que la validation laisse passer.
//
// Liste fermée et courte par construction : le contrôle échoue aussi quand une
// entrée d'ici ne correspond plus à rien au rapport, ce qui force à la retirer
// au lieu de la laisser dormir.
//
// N'y entre qu'un avis dont l'exposition est nulle **et** qu'aucune montée de
// version ne peut régler. Un avis corrigé en amont se corrige, il ne s'exclut
// pas.
//
// Un avis critique n'y entre jamais : npm est appelé avec
// `--audit-level=critical` et échoue seul dans ce cas, avant que ce contrôle
// ne parle. Une ligne d'ici n'y changerait rien, et c'est voulu.
var exclusions = map[string]Exclusion{
	"GHSA-vfj7-8cjw-p6xm": {
		Raison: "braces n'est atteint que par steiger, qui contrôle le découpage FSD : " +
			"rien de cette chaîne n'entre dans le bundle embarqué, et les motifs glob " +
			"viennent du dépôt, jamais d'une entrée non fiable. Aucune version corrigée " +
			"n'existe — braces 3.0.3 est la dernière publiée — et le steiger que npm " +
			"propose en --force est une rétrogradation qui traîne la même chaîne.",
		Retrait: "à la publication d'un braces corrigé, au retrait de steiger, " +
			"ou au retrait de l'avis, que son mainteneur conteste (micromatch/braces#70).",
	},
}
