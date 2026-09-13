// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import "strings"

// nomDeSequence lit le nom de la séquence dans un défaut de genre sequence,
// tel que PostgreSQL le rend : nextval('facture_id_seq'::regclass).
//
// Le physique garde l'expression verbatim, et c'est ici qu'elle se traduit.
// Le nom sort tel que le catalogue l'a écrit : qualifié par son schéma quand
// la séquence n'était pas visible depuis la session d'extraction, entre
// guillemets doubles quand il en exige. Seules les apostrophes doublées du
// littéral SQL redeviennent simples.
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
	if !ok || nom == "" || reste != suffixe {
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
