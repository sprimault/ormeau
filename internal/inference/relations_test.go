// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
)

// TestColonneDuCalque couvre la lecture d'une référence schema.table.colonne,
// y compris quand un nom de table contient un point.
func TestColonneDuCalque(t *testing.T) {
	t.Parallel()

	colonnes := func(noms ...string) []calque.Colonne {
		jeu := make([]calque.Colonne, 0, len(noms))
		for i, nom := range noms {
			jeu = append(jeu, calque.Colonne{Nom: nom, Position: i + 1})
		}
		return jeu
	}
	tables := []*calque.Table{
		{Schema: "public", Nom: "client", Colonnes: colonnes("id", "adresse.rue")},
		{Schema: "public", Nom: "client.adresse", Colonnes: colonnes("rue")},
	}

	cas := []struct {
		nom, reference string
		table, colonne string
		trouvee        bool
	}{
		{"colonne simple", "public.client.id", "client", "id", true},
		{"table la plus longue", "public.client.adresse.rue", "client.adresse", "rue", true},
		{"colonne absente", "public.client.nom", "", "", false},
		{"table absente", "public.fournisseur.id", "", "", false},
		{"table sans colonne", "public.client", "", "", false},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			table, colonne := colonneDuCalque(tables, c.reference)
			if (table != nil) != c.trouvee {
				t.Fatalf("table trouvée : %v, attendu %v", table != nil, c.trouvee)
			}
			if table != nil && (table.Nom != c.table || colonne != c.colonne) {
				t.Errorf("rendu %s.%s, attendu %s.%s", table.Nom, colonne, c.table, c.colonne)
			}
		})
	}
}
