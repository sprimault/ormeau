// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
)

// TestTyperColonne couvre le vocabulaire fermé des types normalisés.
//
// Chaque cas vaut pour son couple : un type PHP juste avec un type Doctrine
// faux — ou l'inverse — donne une entité que Doctrine accepte et qui échoue à
// l'hydratation, sans que rien ne le signale avant l'exécution.
func TestTyperColonne(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom       string
		colonne   calque.Colonne
		php       string
		doctrine  string
		certitude bool
	}{
		{
			"entier ordinaire",
			calque.Colonne{TypeBrut: "integer", TypeNormalise: calque.TypeEntier},
			"int", "integer", true,
		},
		{
			"booleen",
			calque.Colonne{TypeBrut: "boolean", TypeNormalise: calque.TypeBooleen},
			"bool", "boolean", true,
		},
		{
			"decimal rendu en chaine, mapping officiel de Doctrine",
			calque.Colonne{TypeBrut: "numeric(12,2)", TypeNormalise: calque.TypeDecimal},
			"string", "decimal", true,
		},
		{
			"uuid",
			calque.Colonne{TypeBrut: "uuid", TypeNormalise: calque.TypeUUID},
			"string", "guid", true,
		},
		{
			"json",
			calque.Colonne{TypeBrut: "jsonb", TypeNormalise: calque.TypeJSON},
			"array", "json", true,
		},
		{
			"intervalle",
			calque.Colonne{TypeBrut: "interval", TypeNormalise: calque.TypeIntervalle},
			"\\DateInterval", "dateinterval", true,
		},
		{
			"type inconnu, rendu en chaine et signale",
			calque.Colonne{TypeBrut: "hierarchyid", TypeNormalise: calque.TypeInconnu},
			"string", "string", false,
		},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			corr, sur := typerColonne(&c.colonne)
			if corr.php != c.php || corr.doctrine != c.doctrine {
				t.Errorf("typerColonne(%s) = %s/%s, attendu %s/%s",
					c.colonne.TypeBrut, corr.php, corr.doctrine, c.php, c.doctrine)
			}
			if sur != c.certitude {
				t.Errorf("certitude = %v, attendue %v", sur, c.certitude)
			}
		})
	}
}

// TestAffiner couvre ce que le type normalisé ne distingue pas.
//
// Le vocabulaire fermé dit « horodatage » sans dire s'il porte un fuseau, et
// « entier » sans dire s'il tient sur 64 bits. C'est type_brut qui le sait, et
// c'est pour ça qu'il est verbatim.
func TestAffiner(t *testing.T) {
	t.Parallel()

	longueur := 255

	cas := []struct {
		nom      string
		colonne  calque.Colonne
		doctrine string
	}{
		{
			"horodatage sans fuseau",
			calque.Colonne{TypeBrut: "timestamp without time zone", TypeNormalise: calque.TypeHorodatage},
			"datetime_immutable",
		},
		{
			"horodatage avec fuseau, forme longue",
			calque.Colonne{TypeBrut: "timestamp with time zone", TypeNormalise: calque.TypeHorodatage},
			"datetimetz_immutable",
		},
		{
			"horodatage avec fuseau, forme courte",
			calque.Colonne{TypeBrut: "timestamptz", TypeNormalise: calque.TypeHorodatage},
			"datetimetz_immutable",
		},
		{
			"bigint",
			calque.Colonne{TypeBrut: "bigint", TypeNormalise: calque.TypeEntier},
			"bigint",
		},
		{
			"int8, alias PostgreSQL de bigint",
			calque.Colonne{TypeBrut: "int8", TypeNormalise: calque.TypeEntier},
			"bigint",
		},
		{
			"smallint",
			calque.Colonne{TypeBrut: "smallint", TypeNormalise: calque.TypeEntier},
			"smallint",
		},
		{
			"text sans longueur",
			calque.Colonne{TypeBrut: "text", TypeNormalise: calque.TypeTexte},
			"text",
		},
		{
			"varchar borne reste string",
			calque.Colonne{TypeBrut: "character varying(255)", TypeNormalise: calque.TypeTexte, Longueur: &longueur},
			"string",
		},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			corr, _ := typerColonne(&c.colonne)
			if corr.doctrine != c.doctrine {
				t.Errorf("typerColonne(%s) rend %s, attendu %s", c.colonne.TypeBrut, corr.doctrine, c.doctrine)
			}
		})
	}
}

// TestForcerAccordeLeTypePHP vérifie qu'un type Doctrine décidé entraîne son
// type PHP.
//
// C'est le cas du char(1) valant O/N d'une base reprise : l'utilisateur force
// boolean, et laisser string en type PHP donnerait une entité qui ne compile
// même pas.
func TestForcerAccordeLeTypePHP(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom      string
		depart   correspondance
		force    string
		php      string
		doctrine string
	}{
		{"chaine vers booleen", correspondance{"string", "string"}, "boolean", "bool", "boolean"},
		{"chaine vers entier", correspondance{"string", "string"}, "integer", "int", "integer"},
		{"entier vers bigint", correspondance{"int", "integer"}, "bigint", "int", "bigint"},
		{"chaine vers text, meme famille", correspondance{"string", "string"}, "text", "string", "text"},
		{"chaine vers horodatage", correspondance{"string", "string"}, "datetime_immutable", "\\DateTimeImmutable", "datetime_immutable"},
		{"type personnalise, le PHP est conserve", correspondance{"string", "string"}, "carbon_immutable", "string", "carbon_immutable"},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			corr := forcer(c.depart, c.force)
			if corr.php != c.php || corr.doctrine != c.doctrine {
				t.Errorf("forcer(%v, %q) = %s/%s, attendu %s/%s",
					c.depart, c.force, corr.php, corr.doctrine, c.php, c.doctrine)
			}
		})
	}
}

// TestTypeNullable vérifie le point d'interrogation des propriétés facultatives.
func TestTypeNullable(t *testing.T) {
	t.Parallel()

	if obtenu := typeNullable("string", true); obtenu != "?string" {
		t.Errorf("typeNullable(string, true) = %q, attendu ?string", obtenu)
	}
	if obtenu := typeNullable("string", false); obtenu != "string" {
		t.Errorf("typeNullable(string, false) = %q, attendu string", obtenu)
	}
	if obtenu := typeNullable("\\DateTimeImmutable", true); obtenu != "?\\DateTimeImmutable" {
		t.Errorf("typeNullable sur une classe = %q, attendu ?\\DateTimeImmutable", obtenu)
	}
}
