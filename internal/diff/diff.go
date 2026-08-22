// Package diff compare deux calques physiques.
//
// C'est le mode le plus utile au quotidien : l'inverse de doctrine:schema:update,
// pour du legacy où le schéma bouge sans passer par les migrations. Il n'écrit
// jamais rien.
package diff

import (
	"github.com/sprimault/ormeau/internal/calque"
)

type Genre string

const (
	Ajout       Genre = "ajout"
	Suppression Genre = "suppression"
	Modification Genre = "modification"
)

type Ecart struct {
	Genre  Genre  `json:"genre"`
	Chemin string `json:"chemin"`
	Avant  string `json:"avant,omitempty"`
	Apres  string `json:"apres,omitempty"`
}

// Comparer suppose les deux calques triés : c'est le déterminisme de la
// sérialisation qui rend ce diff exploitable plutôt que bruyant.
func Comparer(avant, apres *calque.Physique) []Ecart {
	panic("à implémenter")
}

// Divergent permet à un pipeline d'échouer sur un écart de schéma.
func Divergent(ecarts []Ecart) bool {
	return len(ecarts) > 0
}
