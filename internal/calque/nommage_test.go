// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package calque

import "testing"

// TestNomDeBaseValide couvre ce qu'un catalogue peut rendre comme nom de base
// et qui ne doit pas devenir un chemin : un nom accentué ou à tiret nomme un
// fichier, un séparateur, un point ou un caractère de contrôle, non.
func TestNomDeBaseValide(t *testing.T) {
	t.Parallel()

	cas := map[string]bool{
		"gescom":         true,
		"paie-2026":      true,
		"t_référentiel":  true,
		"":               false,
		"../x":           false,
		"a/b":            false,
		`a\b`:            false,
		"gescom.calque":  false,
		"ma base":        false,
		"gescom\x1b[31m": false,
	}

	for nom, attendu := range cas {
		if obtenu := NomDeBaseValide(nom); obtenu != attendu {
			t.Errorf("NomDeBaseValide(%q) = %v, attendu %v", nom, obtenu, attendu)
		}
	}
}
