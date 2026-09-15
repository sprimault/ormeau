// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Package diff compare deux calques physiques.
//
// C'est le mode le plus utile au quotidien : l'inverse de doctrine:schema:update,
// pour du legacy où le schéma bouge sans passer par les migrations. Il n'écrit
// jamais rien.
//
// La comparaison rapproche les objets par clé, jamais par position : le tri du
// calque range contraintes et index par nom, et un nom qui change décalerait
// tout le reste. Les contraintes et les index se rapprochent par ce qu'ils
// portent — colonnes, cible —, leur nom n'est qu'une propriété comparée : une
// contrainte renommée sort en modification, pas en suppression suivie d'un
// ajout. C'est aussi ce qui permet de comparer une base à celle que Doctrine
// recrée, où les clés étrangères s'appellent FK_….
//
// Ce paquet ne sait rien de la destination. Les écarts qu'un ORM ne peut pas
// éviter se tolèrent chez celui qui le sait, pas ici.
package diff

import (
	"cmp"
	"slices"
	"strings"

	"github.com/sprimault/ormeau/internal/calque"
)

// Genre est la nature d'un écart.
type Genre string

// Les trois genres d'écart.
const (
	Ajout        Genre = "ajout"
	Suppression  Genre = "suppression"
	Modification Genre = "modification"
)

// Objet est la sorte d'élément du calque que l'écart désigne.
type Objet string

// Objets comparés, dans l'ordre où les écarts d'une même table se rangent.
const (
	ObjetTable        Objet = "table"
	ObjetColonne      Objet = "colonne"
	ObjetClePrimaire  Objet = "cle_primaire"
	ObjetCleEtrangere Objet = "cle_etrangere"
	ObjetUnicite      Objet = "unicite"
	ObjetIndex        Objet = "index"
	ObjetVerification Objet = "verification"
	ObjetVue          Objet = "vue"
	ObjetSequence     Objet = "sequence"
	ObjetTypeEnumere  Objet = "type_enumere"
)

// rangObjet ordonne les objets d'une même table dans la sortie.
var rangObjet = map[Objet]int{
	ObjetTable: 0, ObjetColonne: 1, ObjetClePrimaire: 2, ObjetCleEtrangere: 3,
	ObjetUnicite: 4, ObjetIndex: 5, ObjetVerification: 6,
	ObjetVue: 7, ObjetSequence: 8, ObjetTypeEnumere: 9,
}

// Ecart localise une divergence par des champs plutôt que par un chemin en
// chaîne : un filtre porte sur l'objet, le nom ou la propriété sans avoir à
// relire ce qu'on vient de mettre en forme, et un nom qui contient un point
// reste sans ambiguïté.
//
// Schema est celui de b : les schémas de a y sont traduits par Options. Table
// est vide pour une vue, une séquence ou un type, et Nom pour un écart sur la
// table elle-même. Propriete n'est renseignée que pour une modification ;
// Avant et Apres le sont alors, vides pour un ajout ou une suppression.
type Ecart struct {
	Genre     Genre
	Objet     Objet
	Schema    string
	Table     string
	Nom       string
	Propriete string
	Avant     string
	Apres     string
}

// Chemin met l'écart en forme pour un humain. Il ne sert pas à filtrer : les
// champs sont là pour ça.
func (e Ecart) Chemin() string {
	var b strings.Builder
	b.WriteString(e.Schema)
	if e.Table != "" {
		b.WriteString("." + e.Table)
	}
	b.WriteString(" " + string(e.Objet))
	if e.Nom != "" {
		b.WriteString(" " + e.Nom)
	}
	if e.Propriete != "" {
		b.WriteString(" : " + e.Propriete)
	}
	return b.String()
}

// Divergent permet à un pipeline d'échouer sur un écart de schéma.
func Divergent(ecarts []Ecart) bool {
	return len(ecarts) > 0
}

// Options règle ce que la comparaison tient pour équivalent.
type Options struct {
	// Schemas traduit un schéma de a en son équivalent dans b : une base
	// recréée ailleurs que dans son schéma d'origine se compare quand même.
	// Il s'applique aux noms d'objets et aux cibles des clés étrangères, jamais
	// aux expressions (défauts, vérifications, vues), qui restent verbatim.
	Schemas map[string]string
}

// traduire rend le schéma de a dans l'espace de b.
func (o Options) traduire(schema string) string {
	if traduit, ok := o.Schemas[schema]; ok {
		return traduit
	}
	return schema
}

// Comparer rend les écarts qui mènent de a à b, dans un ordre total : schéma,
// table, objet, nom, propriété. Les calques ne sont pas modifiés.
func Comparer(a, b *calque.Physique, o Options) []Ecart {
	var ecarts []Ecart

	paires, seulsA, seulsB := apparier(a.Tables, b.Tables,
		func(t *calque.Table) string { return cle(o.traduire(t.Schema), t.Nom) },
		func(t *calque.Table) string { return cle(t.Schema, t.Nom) },
		nil, proprietesTable)
	for _, t := range seulsA {
		ecarts = append(ecarts, Ecart{Genre: Suppression, Objet: ObjetTable, Schema: o.traduire(t.Schema), Table: t.Nom})
	}
	for _, t := range seulsB {
		ecarts = append(ecarts, Ecart{Genre: Ajout, Objet: ObjetTable, Schema: t.Schema, Table: t.Nom})
	}
	for _, p := range paires {
		ecarts = append(ecarts, comparerTables(p[0], p[1], o)...)
	}

	ecarts = append(ecarts, comparerHorsTable(a.Vues, b.Vues, ObjetVue, o,
		func(v *calque.Vue) (string, string) { return v.Schema, v.Nom }, proprietesVue)...)
	ecarts = append(ecarts, comparerHorsTable(a.Sequences, b.Sequences, ObjetSequence, o,
		func(s *calque.Sequence) (string, string) { return s.Schema, s.Nom }, proprietesSequence)...)
	ecarts = append(ecarts, comparerHorsTable(a.TypesEnumeres, b.TypesEnumeres, ObjetTypeEnumere, o,
		func(t *calque.TypeEnumere) (string, string) { return t.Schema, t.Nom }, proprietesTypeEnumere)...)

	slices.SortStableFunc(ecarts, func(x, y Ecart) int {
		return cmp.Or(
			cmp.Compare(x.Schema, y.Schema),
			cmp.Compare(x.Table, y.Table),
			cmp.Compare(rangObjet[x.Objet], rangObjet[y.Objet]),
			cmp.Compare(x.Nom, y.Nom),
			cmp.Compare(x.Propriete, y.Propriete),
			cmp.Compare(x.Genre, y.Genre),
		)
	})
	return ecarts
}

// comparerTables rend les écarts entre deux tables rapprochées : les siens,
// puis ceux de chaque sorte d'objet qu'elle porte.
func comparerTables(a, b *calque.Table, o Options) []Ecart {
	lieu := Ecart{Schema: b.Schema, Table: b.Nom}

	ecarts := modifications(a, b, proprietesTable, lieu, ObjetTable, "")

	ecarts = append(ecarts, comparerDansTable(a.Colonnes, b.Colonnes, lieu, ObjetColonne,
		func(c *calque.Colonne) string { return c.Nom },
		func(c *calque.Colonne) string { return c.Nom },
		func(c *calque.Colonne) string { return c.Nom },
		proprietesColonne)...)

	ecarts = append(ecarts, comparerDansTable(presente(a.ClePrimaire), presente(b.ClePrimaire), lieu, ObjetClePrimaire,
		func(*calque.ClePrimaire) string { return "" },
		func(*calque.ClePrimaire) string { return "" },
		func(c *calque.ClePrimaire) string { return c.Nom },
		proprietesClePrimaire)...)

	ecarts = append(ecarts, comparerDansTable(a.ClesEtrangeres, b.ClesEtrangeres, lieu, ObjetCleEtrangere,
		func(c *calque.CleEtrangere) string {
			return cle(liste(c.Colonnes), o.traduire(c.SchemaCible), c.TableCible, liste(c.ColonnesCibles))
		},
		func(c *calque.CleEtrangere) string {
			return cle(liste(c.Colonnes), c.SchemaCible, c.TableCible, liste(c.ColonnesCibles))
		},
		func(c *calque.CleEtrangere) string { return c.Nom },
		proprietesCleEtrangere)...)

	ecarts = append(ecarts, comparerDansTable(a.Unicites, b.Unicites, lieu, ObjetUnicite,
		func(c *calque.Contrainte) string { return liste(c.Colonnes) },
		func(c *calque.Contrainte) string { return liste(c.Colonnes) },
		func(c *calque.Contrainte) string { return c.Nom },
		proprietesUnicite)...)

	// Seules les colonnes identifient un index : un prédicat ou une méthode
	// qui change est une modification, que la phase qui les porte doit voir.
	ecarts = append(ecarts, comparerDansTable(a.Index, b.Index, lieu, ObjetIndex,
		func(i *calque.Index) string { return liste(i.Colonnes) },
		func(i *calque.Index) string { return liste(i.Colonnes) },
		func(i *calque.Index) string { return i.Nom },
		proprietesIndex)...)

	// Une vérification n'a pas d'autre identité que son expression : les noms
	// des CHECK du legacy sont souvent générés. Une expression modifiée sort
	// donc en suppression suivie d'un ajout.
	ecarts = append(ecarts, comparerDansTable(a.Verifications, b.Verifications, lieu, ObjetVerification,
		func(v *calque.Verification) string { return v.Expression },
		func(v *calque.Verification) string { return v.Expression },
		func(v *calque.Verification) string { return v.Nom },
		proprietesVerification)...)

	return ecarts
}

// comparerDansTable rapproche une sorte d'objet d'une table et rend ses
// écarts. nom désigne l'objet dans l'écart : celui de a quand il existe, la
// base d'origine étant ce qu'on lit d'abord ; à défaut son identité.
func comparerDansTable[T any](a, b []T, lieu Ecart, objet Objet, idA, idB, nom func(*T) string, proprietes []propriete[T]) []Ecart {
	designer := func(x *T, id func(*T) string) string {
		if n := nom(x); n != "" {
			return n
		}
		if objet == ObjetClePrimaire {
			return ""
		}
		return "(" + id(x) + ")"
	}

	paires, seulsA, seulsB := apparier(a, b, idA, idB, nom, proprietes)
	var ecarts []Ecart
	for _, x := range seulsA {
		ecarts = append(ecarts, Ecart{Genre: Suppression, Objet: objet, Schema: lieu.Schema, Table: lieu.Table, Nom: designer(x, idA)})
	}
	for _, x := range seulsB {
		ecarts = append(ecarts, Ecart{Genre: Ajout, Objet: objet, Schema: lieu.Schema, Table: lieu.Table, Nom: designer(x, idB)})
	}
	for _, p := range paires {
		x, id := p[0], idA
		if nom(x) == "" {
			x, id = p[1], idB
		}
		ecarts = append(ecarts, modifications(p[0], p[1], proprietes, lieu, objet, designer(x, id))...)
	}
	return ecarts
}

// comparerHorsTable rapproche les objets qui ne vivent pas dans une table —
// vues, séquences, types — par schéma traduit et nom.
func comparerHorsTable[T any](a, b []T, objet Objet, o Options, lieu func(*T) (string, string), proprietes []propriete[T]) []Ecart {
	paires, seulsA, seulsB := apparier(a, b,
		func(x *T) string { s, n := lieu(x); return cle(o.traduire(s), n) },
		func(x *T) string { s, n := lieu(x); return cle(s, n) },
		nil, proprietes)

	var ecarts []Ecart
	for _, x := range seulsA {
		s, n := lieu(x)
		ecarts = append(ecarts, Ecart{Genre: Suppression, Objet: objet, Schema: o.traduire(s), Nom: n})
	}
	for _, x := range seulsB {
		s, n := lieu(x)
		ecarts = append(ecarts, Ecart{Genre: Ajout, Objet: objet, Schema: s, Nom: n})
	}
	for _, p := range paires {
		s, n := lieu(p[1])
		ecarts = append(ecarts, modifications(p[0], p[1], proprietes, Ecart{Schema: s}, objet, n)...)
	}
	return ecarts
}

// modifications rend un écart par propriété qui diffère entre deux objets
// rapprochés.
func modifications[T any](a, b *T, proprietes []propriete[T], lieu Ecart, objet Objet, nom string) []Ecart {
	var ecarts []Ecart
	for _, p := range proprietes {
		avant, apres := p.valeur(a), p.valeur(b)
		if avant == apres {
			continue
		}
		ecarts = append(ecarts, Ecart{
			Genre: Modification, Objet: objet, Schema: lieu.Schema, Table: lieu.Table,
			Nom: nom, Propriete: p.nom, Avant: avant, Apres: apres,
		})
	}
	return ecarts
}

// apparier rapproche les éléments de a et de b qui partagent une identité.
//
// Plusieurs éléments peuvent la partager — deux index sur les mêmes colonnes,
// de classes d'opérateurs différentes. Ils se rapprochent alors par priorité :
// d'abord ceux qui sont identiques en tout, puis ceux de même nom, puis le
// reste dans l'ordre des noms. Le résultat ne dépend donc ni de l'ordre des
// listes ni de celui des maps.
func apparier[T any](a, b []T, idA, idB, nom func(*T) string, proprietes []propriete[T]) (paires [][2]*T, seulsA, seulsB []*T) {
	groupesA := grouper(a, idA, nom)
	groupesB := grouper(b, idB, nom)

	identiques := func(x, y *T) bool {
		for _, p := range proprietes {
			if p.valeur(x) != p.valeur(y) {
				return false
			}
		}
		return true
	}
	memeNom := func(x, y *T) bool { return nom != nil && nom(x) == nom(y) }
	quelconques := func(*T, *T) bool { return true }

	for _, id := range clesTriees(groupesA, groupesB) {
		restantsA, restantsB := groupesA[id], groupesB[id]
		for _, critere := range []func(x, y *T) bool{identiques, memeNom, quelconques} {
			restantsA, restantsB = rapprocher(restantsA, restantsB, critere, &paires)
		}
		seulsA = append(seulsA, restantsA...)
		seulsB = append(seulsB, restantsB...)
	}
	return paires, seulsA, seulsB
}

// rapprocher associe chaque élément de a au premier de b qui satisfait le
// critère, et rend ceux qui restent de chaque côté.
func rapprocher[T any](a, b []*T, critere func(x, y *T) bool, paires *[][2]*T) ([]*T, []*T) {
	var restantsA []*T
	pris := make([]bool, len(b))
	for _, x := range a {
		trouve := false
		for j, y := range b {
			if !pris[j] && critere(x, y) {
				pris[j] = true
				*paires = append(*paires, [2]*T{x, y})
				trouve = true
				break
			}
		}
		if !trouve {
			restantsA = append(restantsA, x)
		}
	}
	var restantsB []*T
	for j, y := range b {
		if !pris[j] {
			restantsB = append(restantsB, y)
		}
	}
	return restantsA, restantsB
}

// grouper range les éléments par identité, chaque groupe trié par nom puis par
// position d'origine.
func grouper[T any](elements []T, id, nom func(*T) string) map[string][]*T {
	groupes := map[string][]*T{}
	for i := range elements {
		x := &elements[i]
		groupes[id(x)] = append(groupes[id(x)], x)
	}
	if nom != nil {
		for _, g := range groupes {
			slices.SortStableFunc(g, func(x, y *T) int { return cmp.Compare(nom(x), nom(y)) })
		}
	}
	return groupes
}

// clesTriees rend l'union des identités des deux côtés, triée.
func clesTriees[T any](a, b map[string][]*T) []string {
	cles := make([]string, 0, len(a)+len(b))
	for k := range a {
		cles = append(cles, k)
	}
	for k := range b {
		if _, dejaVue := a[k]; !dejaVue {
			cles = append(cles, k)
		}
	}
	slices.Sort(cles)
	return cles
}

// presente rend la clé primaire comme une liste d'au plus un élément, pour la
// rapprocher comme les autres objets d'une table.
func presente(c *calque.ClePrimaire) []calque.ClePrimaire {
	if c == nil {
		return nil
	}
	return []calque.ClePrimaire{*c}
}

// cle assemble une identité. Le séparateur est l'octet nul, qu'aucun nom
// d'objet ne peut contenir.
func cle(parties ...string) string {
	return strings.Join(parties, "\x00")
}
