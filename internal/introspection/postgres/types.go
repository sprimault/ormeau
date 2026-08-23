// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"strings"

	"github.com/sprimault/ormeau/internal/calque"
)

// La normalisation part de typname, le nom interne du type, et non de
// format_type : celui-ci rend « character varying(120) », dont l'analyse
// dépendrait de la locale du serveur. typname rend « varchar », stable.
//
// TypeBrut, lui, garde format_type verbatim — c'est ce qui sauve la mise devant
// un domaine maison ou une extension.
var typesInternes = map[string]calque.TypeNorm{
	"bool": calque.TypeBooleen,

	"int2": calque.TypeEntier,
	"int4": calque.TypeEntier,
	"int8": calque.TypeEntier,

	"numeric": calque.TypeDecimal,
	"money":   calque.TypeDecimal,

	"float4": calque.TypeFlottant,
	"float8": calque.TypeFlottant,

	"char":    calque.TypeTexte,
	"bpchar":  calque.TypeTexte,
	"varchar": calque.TypeTexte,
	"text":    calque.TypeTexte,
	"name":    calque.TypeTexte,
	"citext":  calque.TypeTexte,

	"bytea": calque.TypeBinaire,

	"date":        calque.TypeDate,
	"time":        calque.TypeHeure,
	"timetz":      calque.TypeHeure,
	"timestamp":   calque.TypeHorodatage,
	"timestamptz": calque.TypeHorodatage,
	"interval":    calque.TypeIntervalle,

	"uuid": calque.TypeUUID,

	"json":  calque.TypeJSON,
	"jsonb": calque.TypeJSON,
	"xml":   calque.TypeXML,

	"point":    calque.TypeGeometrie,
	"line":     calque.TypeGeometrie,
	"lseg":     calque.TypeGeometrie,
	"box":      calque.TypeGeometrie,
	"path":     calque.TypeGeometrie,
	"polygon":  calque.TypeGeometrie,
	"circle":   calque.TypeGeometrie,
	"geometry": calque.TypeGeometrie,

	"inet":     calque.TypeReseau,
	"cidr":     calque.TypeReseau,
	"macaddr":  calque.TypeReseau,
	"macaddr8": calque.TypeReseau,
}

// normaliserType réduit un type PostgreSQL au vocabulaire fermé du calque.
//
// estEnumere l'emporte : un type énuméré natif porte un typname propre à la
// base, absent de toute table de correspondance.
func normaliserType(typeInterne string, estEnumere bool) calque.TypeNorm {
	if estEnumere {
		return calque.TypeEnumereNorm
	}
	// Les tableaux portent le type de leur élément préfixé d'un souligné. Le
	// calque n'a pas de notion de tableau : on rend le type de l'élément, et
	// TypeBrut garde le « [] » qui dit le reste.
	nom := strings.TrimPrefix(typeInterne, "_")

	if norme, connu := typesInternes[nom]; connu {
		return norme
	}
	return calque.TypeInconnu
}

// classerDefaut range une expression de défaut dans le vocabulaire fermé.
//
// La distinction porte tout l'intérêt du champ : DEFAULT 'now()' est un
// littéral qui vaudra toujours la chaîne « now() », DEFAULT now() est une
// expression évaluée à chaque insertion. Les confondre produit des entités
// fausses.
func classerDefaut(expression string) *calque.Defaut {
	if expression == "" {
		return nil
	}

	nettoyee := strings.TrimSpace(expression)

	// nextval('...') désigne une séquence, et c'est ce qui distingue une
	// colonne sérielle d'une colonne IDENTITY.
	if strings.HasPrefix(strings.ToLower(nettoyee), "nextval(") {
		return &calque.Defaut{Genre: calque.DefautSequence, Valeur: nettoyee}
	}

	// Un littéral est rendu par le catalogue entre apostrophes, éventuellement
	// suivi d'un transtypage : 'ACTIF'::character varying. On garde la valeur
	// nue, le type de la colonne dit le reste.
	if valeur, ok := litteral(nettoyee); ok {
		return &calque.Defaut{Genre: calque.DefautLitteral, Valeur: valeur}
	}

	// Les nombres et les mots-clés booléens arrivent sans apostrophe.
	if estNombre(nettoyee) || strings.EqualFold(nettoyee, "true") || strings.EqualFold(nettoyee, "false") {
		return &calque.Defaut{Genre: calque.DefautLitteral, Valeur: nettoyee}
	}

	return &calque.Defaut{Genre: calque.DefautExpression, Valeur: nettoyee}
}

// litteral extrait la valeur d'une constante entre apostrophes, transtypage
// éventuel retiré. Rend false si l'expression n'est pas une constante simple —
// une concaténation, par exemple, reste une expression.
func litteral(expression string) (string, bool) {
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
		// Fin de la constante : ce qui suit ne peut être qu'un transtypage.
		suite := strings.TrimSpace(reste[i+1:])
		if suite == "" || strings.HasPrefix(suite, "::") {
			return valeur.String(), true
		}
		return "", false
	}
	return "", false
}

func estNombre(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		switch {
		case r >= '0' && r <= '9', r == '.':
		case (r == '-' || r == '+') && i == 0:
		default:
			return false
		}
	}
	return true
}

// actionReferentielle traduit les codes d'un caractère de pg_constraint.
//
// Le catalogue ne distingue pas « aucune action » d'une absence de clause :
// confdeltype vaut 'a' dans les deux cas, ce qui est la valeur par défaut de
// SQL et ne mérite pas d'apparaître dans le calque.
func actionReferentielle(code string) calque.Action {
	switch code {
	case "c":
		return calque.ActionCascade
	case "n":
		return calque.ActionSetNull
	case "d":
		return calque.ActionSetDefault
	case "r":
		return calque.ActionRestrict
	default:
		return ""
	}
}
