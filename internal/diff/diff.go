// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Package diff compare deux calques physiques.
//
// C'est le mode le plus utile au quotidien : l'inverse de doctrine:schema:update,
// pour du legacy où le schéma bouge sans passer par les migrations. Il n'écrit
// jamais rien.
//
// La comparaison elle-même reste à écrire (phase 8). Elle supposera les deux
// calques triés : c'est le déterminisme de la sérialisation qui rend ce diff
// exploitable plutôt que bruyant.
//
// Elle n'est pas déclarée tant qu'elle n'est pas écrite : ne pouvant pas
// signaler son inachèvement par son type de retour, une ébauche rendrait une
// tranche vide — soit « aucun écart », soit exactement le contraire de ce que
// le mode diff existe pour dire.
package diff

// Genre est la nature d'un écart.
type Genre string

// Sérialisés tels quels : les renommer casserait les pipelines qui filtrent.
const (
	Ajout        Genre = "ajout"
	Suppression  Genre = "suppression"
	Modification Genre = "modification"
)

// Ecart localise une divergence par un chemin qualifié — schema.table.colonne —
// plutôt que par un couple d'objets : lisible en texte comme en JSON.
type Ecart struct {
	Genre  Genre  `json:"genre"`
	Chemin string `json:"chemin"`
	Avant  string `json:"avant,omitempty"`
	Apres  string `json:"apres,omitempty"`
}

// Divergent permet à un pipeline d'échouer sur un écart de schéma.
func Divergent(ecarts []Ecart) bool {
	return len(ecarts) > 0
}
