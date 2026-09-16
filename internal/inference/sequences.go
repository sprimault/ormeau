// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"strings"

	"github.com/sprimault/ormeau/internal/calque"
)

// nomDeSequence rend le nom de séquence que le calque logique porte, lu dans un
// défaut de genre sequence par nomEcritDeSequence.
//
// Un préfixe public écrit sans guillemets est retiré. public est le schéma par
// défaut de PostgreSQL, et le nom nu désigne la même séquence sous le chemin
// par défaut. Surtout, DBAL 3 relit une séquence de public sans son schéma, par
// comparaison littérale au nom public et non au schéma courant
// (PostgreSQLSchemaManager::_getPortableSequenceDefinition) : qualifiée, elle
// fait proposer à schema:update et migrations:diff un CREATE SEQUENCE d'une
// séquence qui existe. La règle ne sert que la cible ORM 2 : ORM 3 rend la clé
// en IDENTITY et DBAL 4 résout un nom nu contre le schéma courant. Quand le
// plancher passera à ORM 3, elle pourra partir. "public" entre guillemets reste :
// PostgreSQL ne rend pas cette forme, et "Public" est un autre schéma.
func nomDeSequence(expression string) (string, bool) {
	nom, ok := nomEcritDeSequence(expression)
	if !ok {
		return "", false
	}
	nom = strings.TrimPrefix(nom, "public.")
	if nom == "" {
		return "", false
	}
	return nom, true
}

// nomEcritDeSequence lit le nom de la séquence dans un défaut de genre
// sequence, tel que PostgreSQL le rend : nextval('facture_id_seq'::regclass).
//
// Le physique garde l'expression verbatim, et c'est ici qu'elle se traduit.
// Le nom sort tel que le catalogue l'a écrit, entre guillemets doubles quand il
// en exige, seules les apostrophes doublées du littéral SQL redevenant simples.
// L'extraction se fait sous un search_path vide : le nom est donc qualifié par
// son schéma. Un calque plus ancien le qualifie ou non selon la session qui
// l'a produit.
//
// Deux formes sont reconnues : celle d'aujourd'hui, et celle qu'une base
// restaurée depuis PostgreSQL 8.0 garde, nextval(('x'::text)::regclass).
// Toute autre, une séquence choisie à l'exécution par exemple, rend faux :
// aucun nom n'y est écrit, et en fabriquer un serait inventer.
func nomEcritDeSequence(expression string) (string, bool) {
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

// rattacherSequence retrouve dans le physique la séquence que désigne un nom
// écrit, ou rend nil.
//
// Le nom est celui du défaut, avant tout retrait de public : c'est lui qui dit
// le schéma. Qualifié, il désigne une seule séquence. Nu, il vient d'un calque
// extrait avant que le search_path soit fixé, et désignait ce que la session
// voyait ; il n'est rattaché que si une seule séquence porte ce nom. Choisir
// entre deux schémas donnerait l'incrément d'une autre séquence.
func rattacherSequence(nom string, sequences []calque.Sequence) *calque.Sequence {
	parties := partiesIdentifiant(nom)
	var trouvee *calque.Sequence
	for i := range sequences {
		s := &sequences[i]
		switch {
		case len(parties) == 2 && s.Schema == parties[0] && s.Nom == parties[1]:
			return s
		case len(parties) == 1 && s.Nom == parties[0]:
			if trouvee != nil {
				return nil
			}
			trouvee = s
		}
	}
	return trouvee
}

// partiesIdentifiant découpe un nom qualifié à la manière de PostgreSQL : un
// point hors guillemets sépare deux parties, les guillemets d'une partie sont
// retirés et un guillemet doublé redevient simple. Une partie non citée reste
// telle quelle : le catalogue ne l'écrit sans guillemets que lorsqu'elle est
// déjà en minuscules. Rend nil pour un nom mal formé.
func partiesIdentifiant(nom string) []string {
	var parties []string
	for reste := nom; ; {
		var partie string
		if corps, cite := strings.CutPrefix(reste, `"`); cite {
			var b strings.Builder
			for {
				i := strings.IndexByte(corps, '"')
				if i < 0 {
					return nil
				}
				b.WriteString(corps[:i])
				corps = corps[i+1:]
				if !strings.HasPrefix(corps, `"`) {
					break
				}
				b.WriteByte('"')
				corps = corps[1:]
			}
			partie, reste = b.String(), corps
		} else {
			fin := strings.IndexAny(reste, `."`)
			if fin < 0 {
				fin = len(reste)
			}
			if reste[:fin] == "" {
				return nil
			}
			partie, reste = reste[:fin], reste[fin:]
		}
		parties = append(parties, partie)

		if reste == "" {
			return parties
		}
		suite, ok := strings.CutPrefix(reste, ".")
		if !ok || suite == "" {
			return nil
		}
		reste = suite
	}
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
