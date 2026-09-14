// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import "strings"

// nomDeSequence lit le nom de la séquence dans un défaut de genre sequence,
// tel que PostgreSQL le rend : nextval('facture_id_seq'::regclass).
//
// Le physique garde l'expression verbatim, et c'est ici qu'elle se traduit.
// Le nom sort tel que le catalogue l'a écrit, entre guillemets doubles quand il
// en exige, seules les apostrophes doublées du littéral SQL redevenant simples.
// L'extraction se fait sous un search_path vide : le nom est donc qualifié par
// son schéma. Un calque plus ancien le qualifie ou non selon la session qui
// l'a produit.
//
// Un préfixe public écrit sans guillemets est retiré. public est le schéma par
// défaut de PostgreSQL, et le nom nu désigne la même séquence sous le chemin
// par défaut. Surtout, DBAL 3 relit une séquence de public sans son schéma, par
// comparaison littérale au nom public et non au schéma courant
// (PostgreSQLSchemaManager::_getPortableSequenceDefinition) : qualifiée, elle
// fait proposer à schema:update et migrations:diff un CREATE SEQUENCE d'une
// séquence qui existe. Relevé le 2026-09-14 sous ORM 2.14.3 et DBAL 3.10.6. La
// règle ne sert que la cible ORM 2 : ORM 3 rend la clé en IDENTITY et DBAL 4
// résout un nom nu contre le schéma courant. Quand le plancher passera à ORM 3,
// elle pourra partir sans refaire l'essai. "public" entre guillemets reste :
// PostgreSQL ne rend pas cette forme, et "Public" est un autre schéma.
//
// Deux formes sont reconnues : celle d'aujourd'hui, et celle qu'une base
// restaurée depuis PostgreSQL 8.0 garde, nextval(('x'::text)::regclass).
// Toute autre, une séquence choisie à l'exécution par exemple, rend faux :
// aucun nom n'y est écrit, et en fabriquer un serait inventer.
func nomDeSequence(expression string) (string, bool) {
	reste, ok := strings.CutPrefix(expression, "nextval(")
	if !ok {
		return "", false
	}
	suffixe := "::regclass)"
	if heritee, ok := strings.CutPrefix(reste, "("); ok {
		reste, suffixe = heritee, "::text)::regclass)"
	}

	nom, reste, ok := litteralSQL(reste)
	if !ok || reste != suffixe {
		return "", false
	}
	nom = strings.TrimPrefix(nom, "public.")
	if nom == "" {
		return "", false
	}
	return nom, true
}

// litteralSQL lit un littéral entre apostrophes en tête de texte, et rend sa
// valeur et ce qui le suit. Une apostrophe doublée est une apostrophe du
// littéral, pas sa fin.
func litteralSQL(texte string) (valeur, reste string, ok bool) {
	corps, ok := strings.CutPrefix(texte, "'")
	if !ok {
		return "", "", false
	}

	var b strings.Builder
	for {
		i := strings.IndexByte(corps, '\'')
		if i < 0 {
			return "", "", false
		}
		b.WriteString(corps[:i])
		corps = corps[i+1:]
		if !strings.HasPrefix(corps, "'") {
			return b.String(), corps, true
		}
		b.WriteByte('\'')
		corps = corps[1:]
	}
}
