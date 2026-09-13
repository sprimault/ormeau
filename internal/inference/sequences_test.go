// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import "testing"

// TestNomDeSequence couvre les formes que rend pg_get_expr, relevées sous
// PostgreSQL 17, et ce qui doit rester sans nom plutôt que d'en recevoir un
// faux.
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
		{`nextval('public."séq''bizarre"'::regclass)`, `public."séq'bizarre"`, true},
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
