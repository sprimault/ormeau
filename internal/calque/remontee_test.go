// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package calque

import (
	"slices"
	"testing"
)

// tableAUnicite construit une table dont la contrainte d'unicité uq est
// soutenue par un index, comme une version 1 le reportait, à côté d'un index
// ordinaire et d'un index de même nom sur d'autres colonnes.
func tableAUnicite() Table {
	return Table{
		Nom: "client", Schema: "public",
		Unicites: []Contrainte{{Nom: "uq", Colonnes: []string{"siret"}}},
		Index: []Index{
			{Nom: "ix_nom", Colonnes: []string{"nom"}},
			{Nom: "uq", Colonnes: []string{"siret"}, Unique: true},
			{Nom: "uq_autre", Colonnes: []string{"code"}, Unique: true},
		},
	}
}

// TestRemonterRetireLIndexDUniciteDUneVersionUn vérifie qu'un physique v1 perd
// l'index reporté en double, et lui seul.
func TestRemonterRetireLIndexDUniciteDUneVersionUn(t *testing.T) {
	t.Parallel()

	p := &Physique{VersionRI: 1, Tables: []Table{tableAUnicite()}}
	p.remonter()

	noms := make([]string, 0, len(p.Tables[0].Index))
	for _, idx := range p.Tables[0].Index {
		noms = append(noms, idx.Nom)
	}
	if !slices.Equal(noms, []string{"ix_nom", "uq_autre"}) {
		t.Errorf("index après remontée %v, attendu [ix_nom uq_autre]", noms)
	}
	if p.VersionRI != 1 {
		t.Errorf("version %d, la version lue doit rester", p.VersionRI)
	}
}

// TestRemonterLaisseUneVersionDeuxIntacte vérifie qu'un physique v2 ne perd
// rien : un index de même nom qu'une unicité y est un vrai index.
func TestRemonterLaisseUneVersionDeuxIntacte(t *testing.T) {
	t.Parallel()

	p := &Physique{VersionRI: 2, Tables: []Table{tableAUnicite()}}
	p.remonter()

	if len(p.Tables[0].Index) != 3 {
		t.Errorf("index %+v, attendu les trois", p.Tables[0].Index)
	}
}

// TestRemonterGardeUnIndexQuiNePartageQueLeNom vérifie qu'un index de même nom
// mais d'autres colonnes, ou partiel, n'est pas pris pour celui de l'unicité.
func TestRemonterGardeUnIndexQuiNePartageQueLeNom(t *testing.T) {
	t.Parallel()

	table := tableAUnicite()
	table.Index = []Index{
		{Nom: "uq", Colonnes: []string{"siret", "pays"}, Unique: true},
		{Nom: "uq", Colonnes: []string{"siret"}, Unique: true, Predicat: "(actif)"},
	}
	p := &Physique{VersionRI: 1, Tables: []Table{table}}
	p.remonter()

	if len(p.Tables[0].Index) != 2 {
		t.Errorf("index %+v, attendu les deux", p.Tables[0].Index)
	}
}
