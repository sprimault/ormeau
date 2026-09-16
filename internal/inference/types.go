// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"maps"
	"slices"
	"strings"

	"github.com/sprimault/ormeau/internal/calque"
)

// C'est ici que le calque cesse d'être neutre : le type Doctrine suppose la
// destination, et c'est pour cette raison qu'il vit dans le logique.

// typePHP et typeDoctrine sont appariés, jamais choisis séparément : un
// type_php qui ne correspond pas à son type_doctrine produit une entité que
// Doctrine accepte et qui échoue à l'hydratation.
type correspondance struct {
	php      string
	doctrine string
}

// Correspondances par défaut du vocabulaire fermé. Les cas qui dépendent du
// type verbatim — fuseau horaire, entier long — sont traités par affiner.
var parTypeNormalise = map[calque.TypeNorm]correspondance{
	calque.TypeEntier:   {"int", "integer"},
	calque.TypeFlottant: {"float", "float"},
	calque.TypeBooleen:  {"bool", "boolean"},
	calque.TypeTexte:    {"string", "string"},
	calque.TypeBinaire:  {"string", "blob"},

	// Doctrine rend decimal en string, et c'est volontaire : un float perdrait
	// de la précision sur un montant. Contre-intuitif, mais c'est le mapping
	// officiel, et le contredire produirait des écarts silencieux.
	calque.TypeDecimal: {"string", "decimal"},

	calque.TypeDate:       {"\\DateTimeImmutable", "date_immutable"},
	calque.TypeHeure:      {"\\DateTimeImmutable", "time_immutable"},
	calque.TypeHorodatage: {"\\DateTimeImmutable", "datetime_immutable"},
	calque.TypeIntervalle: {"\\DateInterval", "dateinterval"},

	calque.TypeUUID: {"string", "guid"},
	calque.TypeJSON: {"array", "json"},
	calque.TypeXML:  {"string", "text"},

	// Aucun type Doctrine natif : la chaîne conserve la valeur telle quelle,
	// et l'utilisateur reste libre d'un type personnalisé.
	calque.TypeGeometrie: {"string", "string"},
	calque.TypeReseau:    {"string", "string"},

	// Une énumération devient un type PHP dédié quand l'heuristique la
	// reconnaît ; à défaut, elle reste la chaîne qu'elle est en base.
	calque.TypeEnumereNorm: {"string", "string"},
}

// typerColonne rend le couple type PHP / type Doctrine d'une colonne.
//
// Le second retour dit si la correspondance est sûre. Un type non reconnu rend
// une chaîne — la valeur reste lisible — mais l'appelant doit le signaler
// plutôt que de laisser croire à une traduction fidèle.
//
// Un tableau PostgreSQL porte le type normalisé de son élément, et seuls les
// crochets de type_brut disent le reste. Aucun type Doctrine ne lit son
// littéral {1,2} : typé comme l'élément, il serait hydraté par transtypage en
// une valeur fausse, sans erreur. La chaîne le garde tel quel, et simple_array
// ne conviendrait pas, qui sépare par virgules sans accolades ni échappement.
func typerColonne(c *calque.Colonne) (correspondance, bool) {
	corr, connu := parTypeNormalise[c.TypeNormalise]
	if !connu || estTableau(c) {
		return correspondance{"string", "string"}, false
	}
	return affiner(corr, c), true
}

// estTableau dit si la colonne est un tableau : format_type écrit toujours ses
// crochets en fin de type, character varying(20)[] comme integer[][].
func estTableau(c *calque.Colonne) bool {
	return strings.HasSuffix(strings.TrimSpace(c.TypeBrut), "[]")
}

// affiner départage ce que le vocabulaire fermé ne distingue pas. Le type
// normalisé dit « horodatage » sans dire s'il porte un fuseau, et « entier »
// sans dire s'il tient sur 64 bits : c'est type_brut qui le sait.
func affiner(corr correspondance, c *calque.Colonne) correspondance {
	brut := strings.ToLower(c.TypeBrut)

	switch c.TypeNormalise {
	case calque.TypeHorodatage:
		// Un horodatage avec fuseau perd son décalage s'il est mappé en
		// datetime_immutable : la valeur relue n'est plus la même instant.
		if avecFuseau(c) {
			return correspondance{corr.php, "datetimetz_immutable"}
		}
	case calque.TypeHeure:
		if avecFuseau(c) {
			return correspondance{corr.php, "time_immutable"}
		}
	case calque.TypeEntier:
		// Le type PHP d'un bigint dépend de la cible — DBAL 3 le rend en
		// string, DBAL 4 en int — et c'est le générateur qui la connaît. Le
		// int écrit ici n'est pas lu pour un type de sa table.
		if strings.Contains(brut, "bigint") || strings.Contains(brut, "int8") {
			return correspondance{"int", "bigint"}
		}
		// tinyint, de SQL Server comme de MySQL, n'a pas de type Doctrine :
		// smallint est le plus proche. En integer, migrations:diff proposerait
		// d'élargir la colonne.
		if strings.Contains(brut, "smallint") || strings.Contains(brut, "int2") || strings.Contains(brut, "tinyint") {
			return correspondance{"int", "smallint"}
		}
	case calque.TypeBinaire:
		// Un binaire de longueur déclarée — binary(n), varbinary(n) — se
		// recrée à sa longueur, fixe ou non. Sans longueur (bytea,
		// varbinary(max), image), il reste un blob.
		if c.Longueur != nil {
			return correspondance{corr.php, "binary"}
		}
	case calque.TypeTexte:
		// Sans longueur déclarée, un texte est illimité, quel que soit le nom
		// de son type : text, varchar de PostgreSQL, varchar(max) de SQL
		// Server. Rendu en chaîne, il deviendrait un VARCHAR(255) que
		// migrations:diff proposerait d'appliquer à la base d'origine.
		//
		// Sauf un texte Unicode de SQL Server : DBAL écrit text en
		// VARCHAR(MAX), et migrations:diff proposerait d'y convertir la colonne
		// — conversion qui réussit en remplaçant par « ? » tout caractère hors
		// de la page de code (essai du 2026-09-16). La chaîne reste, avec
		// texte_unicode_sans_equivalent : sa troncature échoue au moins
		// bruyamment.
		if c.Longueur == nil && !texteUnicodeSansLongueur(c) {
			return correspondance{"string", "text"}
		}
	case calque.TypeFlottant:
		// Doctrine distingue la simple précision depuis DBAL 4.1 ; c'est au
		// générateur de replier sur float quand la cible ne la connaît pas.
		if brut == "real" {
			return correspondance{corr.php, "smallfloat"}
		}
	case calque.TypeJSON:
		// jsonb dédoublonne et réordonne les clés à l'écriture : recréé en
		// json, il garderait tout ce qu'il écrasait. Type Doctrine depuis
		// DBAL 4.3, option de colonne avant, repli au générateur.
		if brut == "jsonb" {
			return correspondance{corr.php, "jsonb"}
		}
	}
	return corr
}

// longueurFixe dit si une chaîne ou un binaire est de longueur fixe, ce que le
// type normalisé ne dit pas.
//
// Le pilote le lit dans son catalogue et le porte au calque physique. Un calque
// extrait avant ce champ ne le porte pas : la forme character(n) de
// PostgreSQL, seule reconnue jusque-là, reste lue, pour qu'une régénération
// depuis un ancien calque ne change rien.
func longueurFixe(c *calque.Colonne) bool {
	if (c.TypeNormalise != calque.TypeTexte && c.TypeNormalise != calque.TypeBinaire) || estTableau(c) {
		return false
	}
	return c.LongueurFixe ||
		(c.TypeNormalise == calque.TypeTexte && strings.HasPrefix(strings.ToLower(c.TypeBrut), "character("))
}

// texteUnicodeSansLongueur dit si la colonne est un texte Unicode illimité de
// SQL Server, que Doctrine ne sait pas recréer sans perte. Le calque ne porte
// pas le caractère Unicode d'un texte : c'est type_brut qui le dit.
func texteUnicodeSansLongueur(c *calque.Colonne) bool {
	brut := strings.ToLower(strings.TrimSpace(c.TypeBrut))
	return c.TypeNormalise == calque.TypeTexte && c.Longueur == nil && (brut == "nvarchar(max)" || brut == "ntext")
}

// avecFuseau dit si un horodatage ou une heure porte un fuseau, ce que le type
// normalisé ne dit pas.
func avecFuseau(c *calque.Colonne) bool {
	brut := strings.ToLower(c.TypeBrut)
	return strings.Contains(brut, "with time zone") || strings.Contains(brut, "timestamptz") || strings.Contains(brut, "timetz")
}

// Type PHP de chaque type Doctrine, pour les cas où le type Doctrine est donné
// et le type PHP à retrouver. Ce n'est pas l'inverse exact de la table
// ci-dessus : plusieurs types Doctrine rendent la même chaîne PHP, et une
// inversion mécanique en aurait perdu la moitié.
var phpParTypeDoctrine = map[string]string{
	"integer":    "int",
	"smallint":   "int",
	"bigint":     "int",
	"float":      "float",
	"smallfloat": "float",
	"boolean":    "bool",

	"decimal": "string",
	"string":  "string",
	"text":    "string",
	"guid":    "string",
	"blob":    "string",
	"binary":  "string",

	"date_immutable":       "\\DateTimeImmutable",
	"time_immutable":       "\\DateTimeImmutable",
	"datetime_immutable":   "\\DateTimeImmutable",
	"datetimetz_immutable": "\\DateTimeImmutable",
	"dateinterval":         "\\DateInterval",

	"json":         "array",
	"jsonb":        "array",
	"simple_array": "array",
}

// TypesDoctrine rend, triés, les types Doctrine dont l'inférence connaît le
// type PHP. Ce sont ceux que l'interface suggère : un type absent de la liste
// se force quand même, et garde le type PHP que la colonne avait produit.
func TypesDoctrine() []string {
	return slices.Sorted(maps.Keys(phpParTypeDoctrine))
}

// forcer applique un type Doctrine décidé, et met le type PHP en accord.
//
// Les deux ne se choisissent jamais séparément : forcer decimal en laissant un
// type PHP de \DateTimeImmutable donne une entité que Doctrine accepte et qui
// échoue à l'hydratation, sans que rien ne l'ait signalé avant l'exécution.
//
// Un type personnalisé — inconnu de la table — garde le type PHP que la colonne
// avait produit. C'est presque toujours juste, et l'inventer serait pire :
// personne d'autre que l'auteur du type ne sait ce qu'il hydrate.
func forcer(corr correspondance, doctrine string) correspondance {
	php := corr.php
	if connu, trouve := phpParTypeDoctrine[doctrine]; trouve {
		php = connu
	}
	return correspondance{php, doctrine}
}

// typeNullable rend la déclaration PHP d'une propriété facultative.
func typeNullable(php string, nullable bool) string {
	if !nullable {
		return php
	}
	return "?" + php
}
