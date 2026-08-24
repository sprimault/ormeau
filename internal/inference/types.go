// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
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
func typerColonne(c *calque.Colonne) (correspondance, bool) {
	corr, connu := parTypeNormalise[c.TypeNormalise]
	if !connu {
		return correspondance{"string", "string"}, false
	}
	return affiner(corr, c), true
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
		if strings.Contains(brut, "with time zone") || strings.Contains(brut, "timestamptz") {
			return correspondance{corr.php, "datetimetz_immutable"}
		}
	case calque.TypeHeure:
		if strings.Contains(brut, "with time zone") || strings.Contains(brut, "timetz") {
			return correspondance{corr.php, "time_immutable"}
		}
	case calque.TypeEntier:
		// Un bigint dépasse l'int 32 bits ; Doctrine le rend en string sur les
		// plateformes 32 bits, d'où le type PHP élargi.
		if strings.Contains(brut, "bigint") || strings.Contains(brut, "int8") {
			return correspondance{"int", "bigint"}
		}
		if strings.Contains(brut, "smallint") || strings.Contains(brut, "int2") {
			return correspondance{"int", "smallint"}
		}
	case calque.TypeTexte:
		// text n'a pas de longueur : le distinguer de varchar évite un
		// VARCHAR(255) arbitraire à la régénération du schéma.
		if c.Longueur == nil && (strings.Contains(brut, "text") || strings.Contains(brut, "clob")) {
			return correspondance{"string", "text"}
		}
	}
	return corr
}

// Type PHP de chaque type Doctrine, pour les cas où le type Doctrine est donné
// et le type PHP à retrouver. Ce n'est pas l'inverse exact de la table
// ci-dessus : plusieurs types Doctrine rendent la même chaîne PHP, et une
// inversion mécanique en aurait perdu la moitié.
var phpParTypeDoctrine = map[string]string{
	"integer":  "int",
	"smallint": "int",
	"bigint":   "int",
	"float":    "float",
	"boolean":  "bool",

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
	"simple_array": "array",
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
