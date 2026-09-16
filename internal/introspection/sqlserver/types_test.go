// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package sqlserver

import (
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
)

// TestNormaliserType couvre les types de la base de référence et les trois que
// le pilote laisse inconnus exprès.
func TestNormaliserType(t *testing.T) {
	t.Parallel()

	cas := map[string]calque.TypeNorm{
		"bit":              calque.TypeBooleen,
		"smallint":         calque.TypeEntier,
		"money":            calque.TypeDecimal,
		"real":             calque.TypeFlottant,
		"nchar":            calque.TypeTexte,
		"binary":           calque.TypeBinaire,
		"time":             calque.TypeHeure,
		"datetimeoffset":   calque.TypeHorodatage,
		"uniqueidentifier": calque.TypeUUID,
		"json":             calque.TypeJSON,
		"geography":        calque.TypeGeometrie,
		"timestamp":        calque.TypeInconnu,
		"hierarchyid":      calque.TypeInconnu,
		"sql_variant":      calque.TypeInconnu,
	}
	for typeSysteme, attendu := range cas {
		if obtenu := normaliserType(typeSysteme); obtenu != attendu {
			t.Errorf("%s : %q, attendu %q", typeSysteme, obtenu, attendu)
		}
	}
}

// TestLongueur : max_length est en octets, et -1 n'est pas une longueur.
func TestLongueur(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom         string
		typeSysteme string
		maxLength   int
		attendu     int
		absente     bool
	}{
		{"nvarchar en caracteres", "nvarchar", 240, 120, false},
		{"nchar en caracteres", "nchar", 28, 14, false},
		{"binary en octets", "binary", 64, 64, false},
		{"varchar en octets", "varchar", 255, 255, false},
		{"nvarchar(max)", "nvarchar", -1, 0, true},
		{"varbinary(max)", "varbinary", -1, 0, true},
		{"entier sans longueur", "int", 4, 0, true},
		{"ntext sans longueur", "ntext", 16, 0, true},
	}
	for _, c := range cas {
		obtenu := longueur(c.typeSysteme, c.maxLength)
		switch {
		case c.absente && obtenu != nil:
			t.Errorf("%s : longueur %d, attendue absente", c.nom, *obtenu)
		case !c.absente && (obtenu == nil || *obtenu != c.attendu):
			t.Errorf("%s : longueur %v, attendue %d", c.nom, obtenu, c.attendu)
		}
	}
}

// TestLongueurFixe : les formes var… et (max) sont variables.
func TestLongueurFixe(t *testing.T) {
	t.Parallel()

	cas := map[string]bool{
		"char": true, "nchar": true, "binary": true,
		"varchar": false, "nvarchar": false, "varbinary": false, "ntext": false, "int": false,
	}
	for typeSysteme, attendu := range cas {
		if obtenu := longueurFixe(typeSysteme); obtenu != attendu {
			t.Errorf("%s : longueur fixe %v, attendue %v", typeSysteme, obtenu, attendu)
		}
	}
}

// TestPrecisionEchelle : seul un décimal les déclare.
func TestPrecisionEchelle(t *testing.T) {
	t.Parallel()

	if p, e := precisionEchelle("decimal", 12, 2); p == nil || e == nil || *p != 12 || *e != 2 {
		t.Errorf("decimal(12,2) : %v, %v", p, e)
	}
	// decimal(10,0) n'est pas int : une échelle nulle reste présente.
	if _, e := precisionEchelle("numeric", 10, 0); e == nil || *e != 0 {
		t.Errorf("numeric(10,0) : echelle %v", e)
	}
	if p, e := precisionEchelle("money", 19, 4); p == nil || e == nil || *p != 19 || *e != 4 {
		t.Errorf("money : %v, %v, attendu 19, 4", p, e)
	}
	if p, e := precisionEchelle("datetime2", 27, 7); p != nil || e != nil {
		t.Errorf("datetime2 : %v, %v, attendues absentes", p, e)
	}
}

// TestClasserDefaut part des définitions relevées sur la base de référence.
func TestClasserDefaut(t *testing.T) {
	t.Parallel()

	cas := []struct {
		definition string
		genre      calque.GenreDefaut
		valeur     string
	}{
		{"((0))", calque.DefautLitteral, "0"},
		{"((-1.5E+2))", calque.DefautLitteral, "-1.5E+2"},
		{"('ACTIF')", calque.DefautLitteral, "ACTIF"},
		{"(N'l''été')", calque.DefautLitteral, "l'été"},
		{"('(a)')", calque.DefautLitteral, "(a)"},
		{"(getdate())", calque.DefautExpression, "(getdate())"},
		{"(newid())", calque.DefautExpression, "(newid())"},
		{"(NULL)", calque.DefautExpression, "(NULL)"},
		{"('a'+'b')", calque.DefautExpression, "('a'+'b')"},
		{"((1)+(2))", calque.DefautExpression, "((1)+(2))"},
		{"(NEXT VALUE FOR [ventes].[sq_avoir])", calque.DefautSequence, "(NEXT VALUE FOR [ventes].[sq_avoir])"},
	}
	for _, c := range cas {
		d := classerDefaut(c.definition)
		if d == nil {
			t.Errorf("%s : aucun defaut", c.definition)
			continue
		}
		if d.Genre != c.genre || d.Valeur != c.valeur {
			t.Errorf("%s : %s %q, attendu %s %q", c.definition, d.Genre, d.Valeur, c.genre, c.valeur)
		}
	}
	if d := classerDefaut("  "); d != nil {
		t.Errorf("definition vide : %+v", d)
	}
}

// TestActionReferentielle : NO_ACTION est l'absence de clause, et une valeur
// inconnue se refuse au lieu de se confondre avec elle.
func TestActionReferentielle(t *testing.T) {
	t.Parallel()

	cas := map[string]calque.Action{
		"NO_ACTION":   "",
		"CASCADE":     calque.ActionCascade,
		"SET_NULL":    calque.ActionSetNull,
		"SET_DEFAULT": calque.ActionSetDefault,
	}
	for desc, attendu := range cas {
		if obtenu, connue := actionReferentielle(desc); !connue || obtenu != attendu {
			t.Errorf("%s : %q (connue %v), attendu %q", desc, obtenu, connue, attendu)
		}
	}
	if _, connue := actionReferentielle("RESTRICT"); connue {
		t.Error("RESTRICT n'existe pas sous SQL Server")
	}
}
