// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"slices"
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
)

// TestNomDeSequence couvre les formes que rend pg_get_expr, relevées sous
// PostgreSQL 17, et ce qui doit rester sans nom plutôt que d'en recevoir un
// faux. Seul un préfixe public écrit sans guillemets se retire : "Public" est
// un autre schéma, et "public" une forme que PostgreSQL ne rend pas.
func TestNomDeSequence(t *testing.T) {
	t.Parallel()

	cas := []struct {
		expression string
		nom        string
		lu         bool
	}{
		{"nextval('facture_id_seq'::regclass)", "facture_id_seq", true},
		{"nextval('gescom.avoir_id_seq'::regclass)", "gescom.avoir_id_seq", true},
		{`nextval('"Compta"."Bon_Livraison_Id_seq"'::regclass)`, `"Compta"."Bon_Livraison_Id_seq"`, true},
		{`nextval('public."séq''bizarre"'::regclass)`, `"séq'bizarre"`, true},
		{"nextval('public.facture_id_seq'::regclass)", "facture_id_seq", true},
		{`nextval('public."public.x"'::regclass)`, `"public.x"`, true},
		{`nextval('"public.x"'::regclass)`, `"public.x"`, true},
		{`nextval('"Public".ecriture_id_seq'::regclass)`, `"Public".ecriture_id_seq`, true},
		{`nextval('"public"."Journal_Id_seq"'::regclass)`, `"public"."Journal_Id_seq"`, true},
		{"nextval('publicite_id_seq'::regclass)", "publicite_id_seq", true},
		{"nextval('public.'::regclass)", "", false},
		{"nextval(('ancienne_id_seq'::text)::regclass)", "ancienne_id_seq", true},
		{"nextval((current_setting('app.seq'::text))::regclass)", "", false},
		{"nextval(''::regclass)", "", false},
		{"nextval('facture_id_seq'::regclass) + 1", "", false},
		{"nextval('facture_id_seq)", "", false},
		{"nextval(('ancienne_id_seq'::text))", "", false},
		{"now()", "", false},
	}

	for _, c := range cas {
		nom, lu := nomDeSequence(c.expression)
		if nom != c.nom || lu != c.lu {
			t.Errorf("nomDeSequence(%q) = %q, %v ; attendu %q, %v", c.expression, nom, lu, c.nom, c.lu)
		}
	}
}

// TestPartiesIdentifiant couvre le découpage d'un nom qualifié tel que
// PostgreSQL l'écrit : un point hors guillemets sépare, un guillemet doublé
// appartient au nom.
func TestPartiesIdentifiant(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom     string
		parties []string
	}{
		{"facture_id_seq", []string{"facture_id_seq"}},
		{"public.bloc_id_seq", []string{"public", "bloc_id_seq"}},
		{`"Compta"."Bon_Livraison_Id_seq"`, []string{"Compta", "Bon_Livraison_Id_seq"}},
		{`public."séq'bizarre"`, []string{"public", "séq'bizarre"}},
		{`"a.b"`, []string{"a.b"}},
		{`"dit ""x"""`, []string{`dit "x"`}},
		{"a.b.c", []string{"a", "b", "c"}},
		{`"non fermé`, nil},
		{"a.", nil},
		{".a", nil},
		{`"a"b`, nil},
		{"", nil},
	}

	for _, c := range cas {
		if parties := partiesIdentifiant(c.nom); !slices.Equal(parties, c.parties) {
			t.Errorf("partiesIdentifiant(%q) = %q ; attendu %q", c.nom, parties, c.parties)
		}
	}
}

// TestRattacherSequence vérifie qu'un nom ne désigne une séquence du physique
// que sans ambiguïté : qualifié, par son schéma ; nu, s'il est seul à le
// porter. Deviner entre deux schémas donnerait l'incrément d'une autre.
func TestRattacherSequence(t *testing.T) {
	t.Parallel()

	sequences := []calque.Sequence{
		{Schema: "Compta", Nom: "Bon_Livraison_Id_seq"},
		{Schema: "gescom", Nom: "doublon_id_seq"},
		{Schema: "public", Nom: "bloc_id_seq"},
		{Schema: "public", Nom: "doublon_id_seq"},
		{Schema: "public", Nom: "facture_id_seq"},
		{Schema: "public", Nom: "séq'bizarre"},
	}

	cas := []struct {
		nom    string
		schema string
		trouve string
	}{
		{"public.bloc_id_seq", "public", "bloc_id_seq"},
		{`"Compta"."Bon_Livraison_Id_seq"`, "Compta", "Bon_Livraison_Id_seq"},
		{`public."séq'bizarre"`, "public", "séq'bizarre"},
		{"facture_id_seq", "public", "facture_id_seq"},
		{"gescom.doublon_id_seq", "gescom", "doublon_id_seq"},
		{"doublon_id_seq", "", ""},
		{"gescom.bloc_id_seq", "", ""},
		{"absente_seq", "", ""},
		{"base.public.bloc_id_seq", "", ""},
		{`"non fermé`, "", ""},
	}

	for _, c := range cas {
		s := rattacherSequence(c.nom, sequences)
		switch {
		case c.trouve == "" && s != nil:
			t.Errorf("rattacherSequence(%q) = %s.%s ; attendu aucune", c.nom, s.Schema, s.Nom)
		case c.trouve != "" && (s == nil || s.Schema != c.schema || s.Nom != c.trouve):
			t.Errorf("rattacherSequence(%q) = %v ; attendu %s.%s", c.nom, s, c.schema, c.trouve)
		}
	}
}
