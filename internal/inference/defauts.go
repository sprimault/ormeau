// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"slices"
	"strings"

	"github.com/sprimault/ormeau/internal/calque"
)

// Un défaut calculé ne se recopie pas : son texte appartient au dialecte, et
// now() n'est pas une valeur qu'un générateur sache écrire. Ce qui se reporte
// est son sens, et seulement quand il est certain.
//
// Les formes sont celles que rend pg_get_expr sous PostgreSQL 17 :
// CURRENT_TIMESTAMP garde sa casse, now() ses parenthèses.

// instantsDeTransaction rendent l'instant de début de la transaction, avec
// fuseau.
var instantsDeTransaction = []string{"now()", "CURRENT_TIMESTAMP", "transaction_timestamp()"}

// sensDuDefaut rend le sens d'un défaut calculé, lu sur l'expression et sur le
// type de la colonne : now() vaut la date du jour sur une date, l'instant sur
// un horodatage, et rien de certain sur une heure.
//
// LOCALTIMESTAMP n'est reconnu que sans fuseau : sur une colonne qui en porte
// un, la valeur stockée dépend du fuseau de la session qui insère.
// clock_timestamp() et les formes à précision, CURRENT_TIMESTAMP(0), ne le sont
// pas : Doctrine n'a d'expression que pour l'instant de la transaction, à la
// précision de la colonne.
func sensDuDefaut(c *calque.Colonne) (calque.ExpressionDefaut, bool) {
	expression := c.Defaut.Valeur
	instant := slices.Contains(instantsDeTransaction, expression) ||
		(expression == "LOCALTIMESTAMP" && !avecFuseau(c))

	switch c.TypeNormalise {
	case calque.TypeHorodatage:
		if instant {
			return calque.DefautHorodatageCourant, true
		}
	case calque.TypeDate:
		if instant || expression == "CURRENT_DATE" {
			return calque.DefautDateCourante, true
		}
	case calque.TypeHeure:
		if expression == "CURRENT_TIME" || expression == "LOCALTIME" {
			return calque.DefautHeureCourante, true
		}
	}
	return "", false
}

// estDefautNul dit si une expression ne fait que transtyper NULL : DEFAULT NULL
// est l'absence de défaut, que le catalogue rend NULL::character varying.
func estDefautNul(expression string) bool {
	return strings.HasPrefix(expression, "NULL::")
}
