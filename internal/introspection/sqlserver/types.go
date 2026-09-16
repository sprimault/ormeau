// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package sqlserver

import (
	"strings"

	"github.com/sprimault/ormeau/internal/calque"
)

// typesSysteme réduit le nom d'un type système au vocabulaire fermé. La clé
// est le type de base : un type alias se normalise comme le type dont il
// dérive, et garde son propre nom dans type_brut.
//
// timestamp (rowversion), hierarchyid et sql_variant restent absents, donc
// inconnus : le premier est un compteur que le serveur écrit seul, le deuxième
// un type CLR, le troisième n'a pas de type du tout. Les ranger dans binaire
// ou texte dirait une chose fausse sur ce que la colonne accepte.
var typesSysteme = map[string]calque.TypeNorm{
	// bit est le booléen de SQL Server, faute d'autre (décision de format de
	// la phase 7).
	"bit": calque.TypeBooleen,

	"tinyint":  calque.TypeEntier,
	"smallint": calque.TypeEntier,
	"int":      calque.TypeEntier,
	"bigint":   calque.TypeEntier,

	"decimal":    calque.TypeDecimal,
	"numeric":    calque.TypeDecimal,
	"money":      calque.TypeDecimal,
	"smallmoney": calque.TypeDecimal,

	"real":  calque.TypeFlottant,
	"float": calque.TypeFlottant,

	"char":     calque.TypeTexte,
	"varchar":  calque.TypeTexte,
	"nchar":    calque.TypeTexte,
	"nvarchar": calque.TypeTexte,
	"text":     calque.TypeTexte,
	"ntext":    calque.TypeTexte,

	"binary":    calque.TypeBinaire,
	"varbinary": calque.TypeBinaire,
	"image":     calque.TypeBinaire,

	"date":           calque.TypeDate,
	"time":           calque.TypeHeure,
	"datetime":       calque.TypeHorodatage,
	"datetime2":      calque.TypeHorodatage,
	"smalldatetime":  calque.TypeHorodatage,
	"datetimeoffset": calque.TypeHorodatage,

	"uniqueidentifier": calque.TypeUUID,

	// json est natif depuis SQL Server 2025. Avant, c'est un nvarchar(max) que
	// seul un CHECK ISJSON qualifie, et c'est à l'inférence d'en juger.
	"json": calque.TypeJSON,
	"xml":  calque.TypeXML,

	"geometry":  calque.TypeGeometrie,
	"geography": calque.TypeGeometrie,
}

// normaliserType réduit un type système au vocabulaire fermé.
func normaliserType(typeSysteme string) calque.TypeNorm {
	if norme, connu := typesSysteme[typeSysteme]; connu {
		return norme
	}
	return calque.TypeInconnu
}

// longueur rend la longueur déclarée d'une chaîne ou d'un binaire, dans le
// même sens que le pilote PostgreSQL : en caractères, et absente pour ce qui
// n'en déclare pas.
//
// max_length est en octets. Les types n… stockent deux octets par caractère,
// d'où la division ; -1 signale (max), qui n'est pas une longueur — le rendre
// comme tel ferait d'un nvarchar(max) un nvarchar(-1). text, ntext et image ne
// déclarent rien non plus : leur max_length est un pointeur de seize octets.
func longueur(typeSysteme string, maxLength int) *int {
	if maxLength < 0 {
		return nil
	}
	switch typeSysteme {
	case "char", "varchar", "binary", "varbinary":
		return &maxLength
	case "nchar", "nvarchar":
		n := maxLength / 2
		return &n
	}
	return nil
}

// precisionEchelle rend précision et échelle d'un décimal, et rien pour les
// autres types : sys.columns en renseigne aussi pour un int ou un datetime2,
// mais ce sont des caractéristiques du type, pas des déclarations, et
// PostgreSQL ne les porte pas davantage.
func precisionEchelle(typeSysteme string, precision, echelle int) (*int, *int) {
	if typeSysteme != "decimal" && typeSysteme != "numeric" {
		return nil, nil
	}
	return &precision, &echelle
}

// classerDefaut range la définition d'une contrainte DEFAULT dans le
// vocabulaire fermé.
//
// SQL Server l'enveloppe toujours de parenthèses, doublées autour d'un nombre :
// ((0)), ('ACTIF'), (getdate()). Elles sont retirées pour reconnaître un
// littéral, dont seule la valeur est gardée, comme sous PostgreSQL. Une
// expression ou une séquence garde la définition telle que le catalogue la
// rend, parenthèses comprises.
func classerDefaut(definition string) *calque.Defaut {
	nettoyee := strings.TrimSpace(definition)
	if nettoyee == "" {
		return nil
	}
	nue := sansParentheses(nettoyee)

	if strings.HasPrefix(strings.ToUpper(nue), "NEXT VALUE FOR ") {
		return &calque.Defaut{Genre: calque.DefautSequence, Valeur: nettoyee}
	}
	if valeur, ok := litteral(nue); ok {
		return &calque.Defaut{Genre: calque.DefautLitteral, Valeur: valeur}
	}
	if estNombre(nue) {
		return &calque.Defaut{Genre: calque.DefautLitteral, Valeur: nue}
	}
	return &calque.Defaut{Genre: calque.DefautExpression, Valeur: nettoyee}
}

// sansParentheses retire les parenthèses qui enveloppent toute l'expression,
// autant de fois qu'il y en a. « (a) + (b) » commence et finit par une
// parenthèse sans en être enveloppé : la profondeur retombe à zéro avant la
// fin, et l'expression est rendue telle quelle.
func sansParentheses(expression string) string {
	for strings.HasPrefix(expression, "(") && fermeEnDernier(expression) {
		expression = strings.TrimSpace(expression[1 : len(expression)-1])
	}
	return expression
}

// fermeEnDernier dit si la parenthèse ouvrante du début se referme sur le
// dernier caractère. Une parenthèse dans une chaîne citée ne compte pas.
func fermeEnDernier(expression string) bool {
	profondeur := 0
	citee := false
	for i := 0; i < len(expression); i++ {
		switch c := expression[i]; {
		case c == '\'':
			citee = !citee
		case citee:
		case c == '(':
			profondeur++
		case c == ')':
			profondeur--
			if profondeur == 0 {
				return i == len(expression)-1
			}
		}
	}
	return false
}

// litteral extrait la valeur d'une chaîne citée, préfixée ou non de N. Rend
// false si quoi que ce soit suit l'apostrophe fermante : 'a' + 'b' reste une
// expression.
func litteral(expression string) (string, bool) {
	if strings.HasPrefix(expression, "N'") {
		expression = expression[1:]
	}
	if !strings.HasPrefix(expression, "'") {
		return "", false
	}

	reste := expression[1:]
	var valeur strings.Builder
	for i := 0; i < len(reste); i++ {
		if reste[i] != '\'' {
			valeur.WriteByte(reste[i])
			continue
		}
		// Apostrophe doublée : elle fait partie de la valeur.
		if i+1 < len(reste) && reste[i+1] == '\'' {
			valeur.WriteByte('\'')
			i++
			continue
		}
		return valeur.String(), i == len(reste)-1
	}
	return "", false
}

// estNombre reconnaît un littéral numérique, notation scientifique comprise :
// SQL Server rend un défaut float en ((1.5E+2)). Laxiste comme son pendant
// PostgreSQL, puisque le catalogue ne rend que des littéraux valides.
func estNombre(s string) bool {
	if s == "" {
		return false
	}
	chiffre := false
	for i, r := range s {
		switch {
		case r >= '0' && r <= '9':
			chiffre = true
		case r == '.':
		case (r == 'e' || r == 'E') && chiffre:
		case (r == '-' || r == '+') && (i == 0 || s[i-1] == 'e' || s[i-1] == 'E'):
		default:
			return false
		}
	}
	return chiffre
}

// actionReferentielle traduit delete_referential_action_desc et son pendant
// de mise à jour.
//
// NO_ACTION est l'absence de clause, rendue vide comme sous PostgreSQL. SQL
// Server n'a pas de RESTRICT. Une valeur inconnue rend une chaîne vide qu'on
// ne peut pas distinguer de NO_ACTION : l'appelant la refuse avant.
func actionReferentielle(desc string) (calque.Action, bool) {
	switch desc {
	case "NO_ACTION":
		return "", true
	case "CASCADE":
		return calque.ActionCascade, true
	case "SET_NULL":
		return calque.ActionSetNull, true
	case "SET_DEFAULT":
		return calque.ActionSetDefault, true
	}
	return "", false
}
