// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"strconv"
	"strings"

	"github.com/sprimault/ormeau/internal/calque"
)

// Ce qui se compare de chaque objet, déclaré comme des données plutôt qu'écrit
// dans la boucle : TestChaqueChampEstCompareOuExclu confronte ces listes aux
// champs réels du calque, et un champ ajouté au format sans qu'on ait décidé
// de sa comparaison fait échouer le test au lieu de passer en silence.

// propriete est un champ comparé : son nom JSON, qui nomme aussi l'écart, et sa
// valeur rendue en texte.
//
// Le rendu sert à la fois à comparer et à afficher : il doit donc être sans
// perte. Un pointeur absent et un zéro restent distincts, et une liste garde
// ses éléments entre guillemets pour qu'un nom contenant une virgule ne se
// confonde pas avec deux noms.
type propriete[T any] struct {
	nom    string
	valeur func(*T) string
}

// proprietesTable ne porte que ce qui est propre à la table ; ses colonnes,
// clés et index sont des objets à part.
var proprietesTable = []propriete[calque.Table]{
	{"commentaire", func(t *calque.Table) string { return t.Commentaire }},
	{"options.moteur", func(t *calque.Table) string {
		if t.Options == nil {
			return ""
		}
		return t.Options.Moteur
	}},
	{"options.collation", func(t *calque.Table) string {
		if t.Options == nil {
			return ""
		}
		return t.Options.Collation
	}},
}

// proprietesColonne compare tout ce que le catalogue dit d'une colonne,
// position comprise : un ordre qui change est un écart de DDL.
var proprietesColonne = []propriete[calque.Colonne]{
	{"position", func(c *calque.Colonne) string { return strconv.Itoa(c.Position) }},
	{"type_brut", func(c *calque.Colonne) string { return c.TypeBrut }},
	{"type_normalise", func(c *calque.Colonne) string { return string(c.TypeNormalise) }},
	{"longueur", func(c *calque.Colonne) string { return entier(c.Longueur) }},
	{"precision", func(c *calque.Colonne) string { return entier(c.Precision) }},
	{"echelle", func(c *calque.Colonne) string { return entier(c.Echelle) }},
	{"nullable", func(c *calque.Colonne) string { return strconv.FormatBool(c.Nullable) }},
	{"auto_increment", func(c *calque.Colonne) string { return strconv.FormatBool(c.AutoIncrement) }},
	{"identite", func(c *calque.Colonne) string { return string(c.Identite) }},
	{"defaut", func(c *calque.Colonne) string {
		if c.Defaut == nil {
			return ""
		}
		return string(c.Defaut.Genre) + " " + c.Defaut.Valeur
	}},
	{"generee", func(c *calque.Colonne) string {
		if c.Generee == nil {
			return ""
		}
		if c.Generee.Stockee {
			return "stockee " + c.Generee.Expression
		}
		return "virtuelle " + c.Generee.Expression
	}},
	{"collation", func(c *calque.Colonne) string { return c.Collation }},
	{"collation_schema", func(c *calque.Colonne) string { return c.CollationSchema }},
	{"type_enumere", func(c *calque.Colonne) string { return c.TypeEnumere }},
	{"commentaire", func(c *calque.Colonne) string { return c.Commentaire }},
}

// proprietesClePrimaire garde l'ordre des colonnes : (a, b) et (b, a) ne sont
// pas le même index.
var proprietesClePrimaire = []propriete[calque.ClePrimaire]{
	{"nom", func(c *calque.ClePrimaire) string { return c.Nom }},
	{"colonnes", func(c *calque.ClePrimaire) string { return liste(c.Colonnes) }},
}

// proprietesCleEtrangere : colonnes et cible font l'identité, le reste se
// compare.
var proprietesCleEtrangere = []propriete[calque.CleEtrangere]{
	{"nom", func(c *calque.CleEtrangere) string { return c.Nom }},
	{"a_la_suppression", func(c *calque.CleEtrangere) string { return string(c.ALaSuppression) }},
	{"a_la_mise_a_jour", func(c *calque.CleEtrangere) string { return string(c.ALaMiseAJour) }},
}

// proprietesUnicite : les colonnes font l'identité, seul le nom se compare.
var proprietesUnicite = []propriete[calque.Contrainte]{
	{"nom", func(c *calque.Contrainte) string { return c.Nom }},
}

// proprietesIndex : les colonnes font l'identité.
var proprietesIndex = []propriete[calque.Index]{
	{"nom", func(i *calque.Index) string { return i.Nom }},
	{"unique", func(i *calque.Index) string { return strconv.FormatBool(i.Unique) }},
	{"predicat", func(i *calque.Index) string { return i.Predicat }},
	{"methode", func(i *calque.Index) string { return i.Methode }},
	{"operateurs", func(i *calque.Index) string { return liste(i.Operateurs) }},
	{"ordres", func(i *calque.Index) string {
		ordres := make([]string, len(i.Ordres))
		for rang, o := range i.Ordres {
			ordres[rang] = string(o)
		}
		return liste(ordres)
	}},
}

// proprietesVerification : l'expression fait l'identité.
var proprietesVerification = []propriete[calque.Verification]{
	{"nom", func(v *calque.Verification) string { return v.Nom }},
}

// proprietesVue : les colonnes d'une vue ne sont pas comparées, aucun pilote
// ne les produit.
var proprietesVue = []propriete[calque.Vue]{
	{"definition", func(v *calque.Vue) string { return v.Definition }},
	{"materialisee", func(v *calque.Vue) string { return strconv.FormatBool(v.Materialisee) }},
}

// proprietesSequence compare les bornes : un maximum qui change est un écart
// de DDL même quand aucune entité n'en dépend.
var proprietesSequence = []propriete[calque.Sequence]{
	{"increment", func(s *calque.Sequence) string { return strconv.FormatInt(s.Increment, 10) }},
	{"depart", func(s *calque.Sequence) string { return entier64(s.Depart) }},
	{"minimum", func(s *calque.Sequence) string { return entier64(s.Minimum) }},
	{"maximum", func(s *calque.Sequence) string { return entier64(s.Maximum) }},
	{"cyclique", func(s *calque.Sequence) string { return strconv.FormatBool(s.Cyclique) }},
}

// proprietesTypeEnumere garde l'ordre des valeurs : PostgreSQL ordonne un type
// énuméré par déclaration.
var proprietesTypeEnumere = []propriete[calque.TypeEnumere]{
	{"valeurs", func(t *calque.TypeEnumere) string { return liste(t.Valeurs) }},
}

// entier rend un pointeur d'entier, vide quand il est absent.
func entier(p *int) string {
	if p == nil {
		return ""
	}
	return strconv.Itoa(*p)
}

// entier64 rend un pointeur d'entier 64 bits, vide quand il est absent.
func entier64(p *int64) string {
	if p == nil {
		return ""
	}
	return strconv.FormatInt(*p, 10)
}

// liste rend une liste de noms entre guillemets, dans son ordre.
func liste(valeurs []string) string {
	cites := make([]string, len(valeurs))
	for i, v := range valeurs {
		cites[i] = strconv.Quote(v)
	}
	return strings.Join(cites, ", ")
}
