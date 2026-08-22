// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package diff

import "testing"

// Divergent fait échouer un pipeline. Le faux négatif est le pire cas : CI
// verte alors que la base a bougé sous les entités.
func TestDivergent(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom     string
		ecarts  []Ecart
		attendu bool
	}{
		{"aucun écart", nil, false},
		{"tranche vide", []Ecart{}, false},
		{"un ajout", []Ecart{{Genre: Ajout, Chemin: "public.client.email"}}, true},
		{"une suppression", []Ecart{{Genre: Suppression, Chemin: "public.facture"}}, true},
		{
			"une modification",
			[]Ecart{{Genre: Modification, Chemin: "public.client.id", Avant: "integer", Apres: "bigint"}},
			true,
		},
		{
			"plusieurs écarts",
			[]Ecart{{Genre: Ajout, Chemin: "a"}, {Genre: Suppression, Chemin: "b"}},
			true,
		},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			if obtenu := Divergent(c.ecarts); obtenu != c.attendu {
				t.Errorf("Divergent(%v) = %v, attendu %v", c.ecarts, obtenu, c.attendu)
			}
		})
	}
}

// Sérialisés tels quels : les renommer casserait les pipelines qui filtrent.
func TestGenresStables(t *testing.T) {
	t.Parallel()

	cas := map[Genre]string{
		Ajout:        "ajout",
		Suppression:  "suppression",
		Modification: "modification",
	}

	for genre, attendu := range cas {
		if string(genre) != attendu {
			t.Errorf("genre %q, attendu %q", genre, attendu)
		}
	}
}
