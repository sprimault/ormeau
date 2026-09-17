// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"sort"
	"strings"

	"github.com/sprimault/ormeau/internal/calque"
)

// verifieeJSON dit si une vérification de la table garantit que la colonne
// contient un objet ou un tableau JSON, sous la forme que rend le catalogue du
// dialecte.
//
// Conservatrice, comme la lecture des énumérations : l'expression entière doit
// être la validation, et rien d'autre. Un CHECK qui ajoute une borne ou une
// règle dit autre chose, que le type json ne rendrait pas.
//
// Sous SQL Server, sys.check_constraints réécrit ISJSON(c) = 1 en
// (isjson([c])=(1)). Sans contrainte de type, ISJSON n'accepte qu'un objet ou
// un tableau, comme OBJECT et ARRAY. VALUE et SCALAR admettent un scalaire,
// que Doctrine relirait en nombre ou en chaîne là où la propriété attend un
// tableau : ils ne sont pas reconnus.
func verifieeJSON(sgbd string, t *calque.Table, colonne string) bool {
	if sgbd != "sqlserver" {
		return false
	}
	nom := "[" + strings.ReplaceAll(colonne, "]", "]]") + "]"
	for _, v := range t.Verifications {
		for _, appel := range []string{"isjson(" + nom + ")", "isjson(" + nom + ",OBJECT)", "isjson(" + nom + ",ARRAY)"} {
			switch v.Expression {
			case "(" + appel + "=(1))", "((1)=" + appel + ")", "(" + appel + ">(0))", "(" + nom + " IS NULL OR " + appel + "=(1))":
				return true
			}
		}
	}
	return false
}

// textesJSON rend, par colonne qualifiée, les textes illimités qu'une
// vérification déclare JSON. Unicode compris : l'inférence les garde en
// chaîne, mais le fichier de décisions les propose.
func textesJSON(p *calque.Physique) map[string]bool {
	jsons := map[string]bool{}
	for i := range p.Tables {
		t := &p.Tables[i]
		for j := range t.Colonnes {
			c := &t.Colonnes[j]
			if c.TypeNormalise == calque.TypeTexte && c.Longueur == nil && !estTableau(c) && verifieeJSON(p.Source.SGBD, t, c.Nom) {
				jsons[t.Schema+"."+t.Nom+"."+c.Nom] = true
			}
		}
	}
	return jsons
}

// jsonsUnicodeAProposer rend, triées, les colonnes qualifiées à proposer en
// json : textes Unicode illimités déclarés JSON, qu'aucune décision ne type
// encore.
func jsonsUnicodeAProposer(p *calque.Physique, d *Decisions) []string {
	var proposes []string
	jsons := textesJSON(p)
	for i := range p.Tables {
		t := &p.Tables[i]
		for j := range t.Colonnes {
			cible := t.Schema + "." + t.Nom + "." + t.Colonnes[j].Nom
			if _, decide := d.TypesForces[cible]; jsons[cible] && texteUnicodeSansLongueur(&t.Colonnes[j]) && !decide {
				proposes = append(proposes, cible)
			}
		}
	}
	sort.Strings(proposes)
	return proposes
}
