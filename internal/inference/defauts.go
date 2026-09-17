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
// Les formes sont celles que rend le catalogue de chaque dialecte. Sous
// PostgreSQL 17, pg_get_expr : CURRENT_TIMESTAMP garde sa casse, now() ses
// parenthèses. Sous SQL Server, sys.default_constraints : toute définition est
// enveloppée de parenthèses, et CURRENT_TIMESTAMP y devient getdate().

// instantsDeTransaction rendent, sous PostgreSQL, l'instant de début de la
// transaction, avec fuseau.
var instantsDeTransaction = []string{"now()", "CURRENT_TIMESTAMP", "transaction_timestamp()"}

// Formes SQL Server. getdate() rend l'instant de l'instruction, à l'heure
// locale du serveur et sans décalage ; DBAL 3 écrit la date et l'heure du jour
// en le convertissant, et une base qu'il a créée se relit comme celle qu'il
// reprend.
const (
	instantSQLServer        = "(getdate())"
	dateConvertieSQLServer  = "(CONVERT([date],getdate()))"
	heureConvertieSQLServer = "(CONVERT([time],getdate()))"
)

// sensDuDefaut rend le sens d'un défaut calculé, lu sur l'expression, sur le
// type de la colonne et sur le dialecte qui l'a écrite : une forme n'a de sens
// que dans le catalogue qui la rend.
func sensDuDefaut(sgbd string, c *calque.Colonne) (calque.ExpressionDefaut, bool) {
	switch sgbd {
	case "postgres":
		return sensPostgres(c)
	case "sqlserver":
		return sensSQLServer(c)
	}
	return "", false
}

// sensPostgres : now() vaut la date du jour sur une date, l'instant sur un
// horodatage, et rien de certain sur une heure.
//
// LOCALTIMESTAMP n'est reconnu que sans fuseau : sur une colonne qui en porte
// un, la valeur stockée dépend du fuseau de la session qui insère.
// clock_timestamp() et les formes à précision, CURRENT_TIMESTAMP(0), ne le sont
// pas : Doctrine n'a d'expression que pour l'instant courant, à la précision de
// la colonne.
func sensPostgres(c *calque.Colonne) (calque.ExpressionDefaut, bool) {
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

// sensSQLServer : getdate() vaut l'instant sur un horodatage, la date ou
// l'heure du jour sur une date ou une heure. Sans décalage, il n'est pas
// reconnu sur une colonne qui en porte un : l'heure locale y est stockée
// étiquetée +00:00, un instant faux dès que le serveur n'est pas en UTC.
//
// sysdatetimeoffset() et sysdatetime(), plus précis, ne sont pas reconnus,
// pour la même raison que clock_timestamp() sous PostgreSQL : Doctrine n'écrit
// que getdate(). getutcdate() rend l'heure UTC, pas l'heure locale.
func sensSQLServer(c *calque.Colonne) (calque.ExpressionDefaut, bool) {
	expression := c.Defaut.Valeur

	switch c.TypeNormalise {
	case calque.TypeHorodatage:
		if expression == instantSQLServer && !avecFuseau(c) {
			return calque.DefautHorodatageCourant, true
		}
	case calque.TypeDate:
		if expression == instantSQLServer || expression == dateConvertieSQLServer {
			return calque.DefautDateCourante, true
		}
	case calque.TypeHeure:
		if expression == instantSQLServer || expression == heureConvertieSQLServer {
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
