// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"strings"
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
)

// TestVerifieeJSON couvre les formes que rend sys.check_constraints, celles
// qui admettent un scalaire ou disent autre chose, et le dialecte : la même
// expression n'a de sens que dans le catalogue qui l'écrit.
func TestVerifieeJSON(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom, sgbd, colonne, expression string
		attendu                        bool
	}{
		{"égal à un", "sqlserver", "c", "(isjson([c])=(1))", true},
		{"un égal", "sqlserver", "c", "((1)=isjson([c]))", true},
		{"positif", "sqlserver", "c", "(isjson([c])>(0))", true},
		{"nul permis", "sqlserver", "c", "([c] IS NULL OR isjson([c])=(1))", true},
		{"objet", "sqlserver", "c", "(isjson([c],OBJECT)=(1))", true},
		{"tableau", "sqlserver", "c", "(isjson([c],ARRAY)=(1))", true},
		{"crochet échappé", "sqlserver", "a]b", "(isjson([a]]b])=(1))", true},
		{"valeur scalaire admise", "sqlserver", "c", "(isjson([c],VALUE)=(1))", false},
		{"scalaire", "sqlserver", "c", "(isjson([c],SCALAR)=(1))", false},
		{"condition en plus", "sqlserver", "c", "(isjson([c])=(1) AND len([c])<(4000))", false},
		{"autre colonne", "sqlserver", "c", "(isjson([d])=(1))", false},
		{"invalide exigé", "sqlserver", "c", "(isjson([c])=(0))", false},
		{"autre dialecte", "postgres", "c", "(isjson([c])=(1))", false},
	}
	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			table := &calque.Table{Verifications: []calque.Verification{{Expression: c.expression}}}
			if obtenu := verifieeJSON(c.sgbd, table, c.colonne); obtenu != c.attendu {
				t.Errorf("verifieeJSON(%s, %q) = %v, attendu %v", c.sgbd, c.expression, obtenu, c.attendu)
			}
		})
	}
}

// TestLeFichierProposeJSONPourUnTexteUnicode vérifie que le prérempli propose
// json, en commentaire, pour le seul texte Unicode déclaré JSON qu'aucune
// décision ne type : ni pour un texte déjà rendu en json, ni pour une colonne
// forcée.
func TestLeFichierProposeJSONPourUnTexteUnicode(t *testing.T) {
	t.Parallel()

	p := physiqueDeReference(t, "json-verifie")
	fichier := string(EcrireDecisions(p, &Decisions{}, "gescom"))
	if !strings.Contains(fichier, "#     dbo.document.unicode: json\n#     dbo.document.unicode_force: json\n") {
		t.Errorf("proposition absente ou mal formée :\n%s", fichier)
	}
	if strings.Contains(fichier, "dbo.document.contenu: json") {
		t.Errorf("un texte déjà rendu en json ne se propose pas :\n%s", fichier)
	}

	decide := string(EcrireDecisions(p, &Decisions{TypesForces: map[string]string{"dbo.document.unicode_force": "json"}}, "gescom"))
	if strings.Contains(decide, "#     dbo.document.unicode_force: json") {
		t.Errorf("une colonne forcée ne se propose plus :\n%s", decide)
	}
	if !strings.Contains(decide, "#     dbo.document.unicode: json\n") {
		t.Errorf("la colonne non forcée reste proposée :\n%s", decide)
	}
}
