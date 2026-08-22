// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package calque

import "testing"

// Deux schémas peuvent porter une table homonyme : la recherche n'a de sens
// que qualifiée.
func TestTableParNomQualifieParLeSchema(t *testing.T) {
	t.Parallel()

	p := &Physique{Tables: []Table{
		{Nom: "client", Schema: "public"},
		{Nom: "client", Schema: "archive"},
	}}

	cas := []struct {
		nom     string
		schema  string
		table   string
		trouvee bool
	}{
		{"schéma courant", "public", "client", true},
		{"schéma secondaire", "archive", "client", true},
		{"schéma inconnu", "temp", "client", false},
		{"table inconnue", "public", "facture", false},
		{"casse différente", "PUBLIC", "client", false},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			obtenue := p.TableParNom(c.schema, c.table)
			if (obtenue != nil) != c.trouvee {
				t.Fatalf("%s.%s : trouvée=%v, attendu %v", c.schema, c.table, obtenue != nil, c.trouvee)
			}
			if obtenue != nil && obtenue.Schema != c.schema {
				t.Errorf("table du schéma %q rendue pour une recherche dans %q", obtenue.Schema, c.schema)
			}
		})
	}
}

// Le pointeur désigne la table du calque, pas une copie.
func TestTableParNomRendUnPointeurSurLeCalque(t *testing.T) {
	t.Parallel()

	p := &Physique{Tables: []Table{{Nom: "client", Schema: "public"}}}

	p.TableParNom("public", "client").Commentaire = "clients actifs"

	if p.Tables[0].Commentaire != "clients actifs" {
		t.Error("la modification n'a pas atteint le calque")
	}
}

// Une colonne absente rend nil, pas une valeur zéro.
func TestColonneParNom(t *testing.T) {
	t.Parallel()

	tbl := &Table{Colonnes: []Colonne{
		{Nom: "id", Position: 1},
		{Nom: "client_id", Position: 2},
	}}

	if c := tbl.ColonneParNom("client_id"); c == nil || c.Position != 2 {
		t.Errorf("colonne client_id : %+v", c)
	}
	if c := tbl.ColonneParNom("absente"); c != nil {
		t.Errorf("colonne inexistante rendue : %+v", c)
	}
	if c := tbl.ColonneParNom("ID"); c != nil {
		t.Error("la recherche doit être sensible à la casse : c'est le catalogue qui fait foi")
	}
}

// Une table sans colonne existe pendant la construction du calque.
func TestColonneParNomSurTableSansColonne(t *testing.T) {
	t.Parallel()

	if c := (&Table{}).ColonneParNom("id"); c != nil {
		t.Errorf("colonne rendue pour une table sans colonne : %+v", c)
	}
}
