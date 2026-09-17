// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"slices"
	"strings"

	"github.com/sprimault/ormeau/internal/calque"
)

// Les associations viennent des clés étrangères déclarées, et de rien d'autre.
//
// C'est ce qui les distingue du nommage : une clé étrangère est un constat du
// catalogue, pas un jugement. Ce qui reste à décider est étroit — de quel côté
// est le propriétaire, quelle cardinalité, comment nommer la propriété — et
// chaque réponse se lit dans le physique.
//
// Les clés étrangères jamais déclarées, cas majoritaire sur du legacy, ne se
// devinent pas ici : elles demandent d'échantillonner les données, ou une
// décision. Une décision passe par le même chemin qu'une clé déclarée, et
// l'emporte sur elle — voir relations.go.

// schemaLogique porte ce qu'une table doit savoir des autres pour se traduire.
//
// Construit une fois par inférence : une association a besoin du nom de classe
// de sa cible, et le chercher table par table serait quadratique sur une base
// à quatre cents tables.
type schemaLogique struct {
	// nomsParTable donne le nom de classe d'une table qualifiée.
	nomsParTable map[string]string

	// jointures marque les tables de jointure pure, qui ne produisent pas
	// d'entité mais une association sur chacune de leurs deux cibles.
	jointures map[string]*jointurePure

	// parents donne, pour une table dont la clé primaire est aussi une clé
	// étrangère, cette clé. Elle porte un héritage si une décision le déclare,
	// un un-vers-un sinon.
	parents map[string]*calque.CleEtrangere

	// tables donne la table physique de chaque table non écartée, par nom
	// qualifié.
	tables map[string]*calque.Table

	// heritages donne la place de chaque table dans une hiérarchie décidée.
	heritages map[string]heritageRetenu

	// enumerations donne, par colonne qualifiée, le type énuméré reconnu.
	enumerations map[string]enumeree

	// relations donne, par table source qualifiée, les relations forcées qui
	// s'y appliquent.
	relations map[string][]relationForcee

	// sequences sont celles du physique, où une clé retrouve la séquence que
	// son défaut nomme.
	sequences []calque.Sequence

	// sgbd est celui du physique : un défaut calculé ne se lit que dans la
	// forme de son dialecte.
	sgbd string

	// jsons marque, par colonne qualifiée, les textes illimités qu'une
	// vérification déclare JSON.
	jsons map[string]bool
}

// jointurePure décrit une table qui n'existe que pour relier deux autres.
type jointurePure struct {
	table  *calque.Table
	gauche *calque.CleEtrangere
	droite *calque.CleEtrangere
}

// analyser repère ce qui se décide à l'échelle du schéma et non de la table :
// tables de jointure et héritages.
//
// Les deux ne peuvent pas se déduire d'une table isolée — il faut savoir ce que
// les clés étrangères désignent, et si la cible existe dans le calque.
func analyser(p *calque.Physique, d *Decisions, prefixes []string) *schemaLogique {
	s := &schemaLogique{
		nomsParTable: make(map[string]string, len(p.Tables)),
		jointures:    map[string]*jointurePure{},
		parents:      map[string]*calque.CleEtrangere{},
		tables:       map[string]*calque.Table{},
		heritages:    map[string]heritageRetenu{},
		sequences:    p.Sequences,
		sgbd:         p.Source.SGBD,
		jsons:        textesJSON(p),
	}

	ignorees := ensemble(d.TablesIgnorees)
	for i := range p.Tables {
		t := &p.Tables[i]
		cible := t.Schema + "." + t.Nom
		if ignorees[cible] {
			continue
		}

		nom, _ := nomEntite(t, d, prefixes)
		s.nomsParTable[cible] = nom
		s.tables[cible] = t

		if j := reconnaitreJointure(t); j != nil {
			s.jointures[cible] = j
			continue
		}
		if fk := reconnaitreHeritage(t); fk != nil {
			s.parents[cible] = fk
		}
	}

	// Une jointure ou un héritage dont une clé vise autre chose que
	// l'identifiant de sa cible ne se mappe pas : la table redevient une entité
	// ordinaire, et inferrerAssociations signale la clé. Après la boucle, les
	// tables visées n'étant pas toutes connues avant.
	for cible, j := range s.jointures {
		if !s.viseIdentifiant(j.gauche) || !s.viseIdentifiant(j.droite) {
			delete(s.jointures, cible)
		}
	}
	for cible, fk := range s.parents {
		if !s.viseIdentifiant(fk) {
			delete(s.parents, cible)
		}
	}
	return s
}

// viseIdentifiant dit si la clé étrangère désigne exactement la clé primaire
// de sa table cible, ou si la cible est hors du calque et n'en décide pas.
//
// Doctrine n'associe que vers l'identifiant : chaque colonne référencée doit
// en être une, et toutes doivent l'être (SchemaValidator). Une clé déclarée
// vers une colonne unique, fréquente sur une base reprise, donnerait un mapping
// refusé.
func (s *schemaLogique) viseIdentifiant(fk *calque.CleEtrangere) bool {
	t, connue := s.tables[fk.SchemaCible+"."+fk.TableCible]
	if !connue {
		return true
	}
	return t.ClePrimaire != nil && memesColonnes(fk.ColonnesCibles, t.ClePrimaire.Colonnes)
}

// tableGeneree rend la table physique d'une table qui produit une entité, ou
// nil : une table écartée ou de jointure n'a pas de classe.
func (s *schemaLogique) tableGeneree(cible string) *calque.Table {
	if _, jointure := s.jointures[cible]; jointure {
		return nil
	}
	return s.tables[cible]
}

// reconnaitreJointure dit si la table n'existe que pour relier deux autres.
//
// Quatre conditions, toutes nécessaires : exactement deux clés étrangères, une
// clé primaire composite, cette clé couvrant exactement les colonnes des deux
// étrangères, et aucune autre colonne.
//
// La dernière est celle qui compte. Une table de liaison qui porte une quantité
// ou une date d'effet est une entité à part entière : la transformer en
// association ferait disparaître ses données du modèle, sans que rien ne le
// signale avant la première requête qui les cherche.
func reconnaitreJointure(t *calque.Table) *jointurePure {
	if len(t.ClesEtrangeres) != 2 || t.ClePrimaire == nil || len(t.ClePrimaire.Colonnes) < 2 {
		return nil
	}

	gauche, droite := &t.ClesEtrangeres[0], &t.ClesEtrangeres[1]

	portantes := map[string]bool{}
	for _, fk := range t.ClesEtrangeres {
		for _, c := range fk.Colonnes {
			portantes[c] = true
		}
	}
	if len(portantes) != len(t.Colonnes) {
		return nil
	}
	for i := range t.Colonnes {
		if !portantes[t.Colonnes[i].Nom] {
			return nil
		}
	}

	// La clé primaire doit couvrir les mêmes colonnes : deux clés étrangères
	// sans unicité laissent passer des doublons, ce qui n'est pas une relation
	// plusieurs-vers-plusieurs mais une table de faits.
	if len(t.ClePrimaire.Colonnes) != len(portantes) {
		return nil
	}
	for _, c := range t.ClePrimaire.Colonnes {
		if !portantes[c] {
			return nil
		}
	}

	return &jointurePure{table: t, gauche: gauche, droite: droite}
}

// reconnaitreHeritage dit si la clé primaire est aussi une clé étrangère.
//
// C'est la forme que prend l'héritage par jointure en base : la table fille
// partage l'identifiant de la mère, et sa clé primaire pointe dessus. Mais
// c'est aussi celle d'une extension un-vers-un, et le schéma ne distingue pas
// les deux : la clé rendue ici porte un héritage si une décision le déclare,
// un un-vers-un sinon (heritages.go).
//
// Il faut que la clé étrangère couvre exactement la clé primaire. Une clé
// primaire composite dont une seule colonne est étrangère décrit une relation
// ordinaire, pas une hiérarchie — le cas d'une table d'affectation datée.
func reconnaitreHeritage(t *calque.Table) *calque.CleEtrangere {
	if t.ClePrimaire == nil || len(t.ClePrimaire.Colonnes) == 0 {
		return nil
	}

	for i := range t.ClesEtrangeres {
		fk := &t.ClesEtrangeres[i]
		if memesColonnes(fk.Colonnes, t.ClePrimaire.Colonnes) && fk.SchemaCible+"."+fk.TableCible != t.Schema+"."+t.Nom {
			return fk
		}
	}
	return nil
}

// inferrerAssociations rend les associations que porte une entité : une par
// clé étrangère déclarée, et une par relation forcée. Leurs côtés inverses
// viennent après, quand toutes les entités existent.
//
// Une clé étrangère dont une colonne est écartée ne donne rien, et le côté
// inverse ne se déduit donc pas non plus : elle est notée dans parties pour
// chacune de ses colonnes écartées. Une relation forcée sur une telle colonne
// n'arrive pas jusqu'ici, verifierRelationsForcees l'a refusée.
func inferrerAssociations(t *calque.Table, s *schemaLogique, parColonne map[string]*calque.Propriete, ecartees map[string]bool, parties map[string][]string) ([]calque.Association, []calque.Avertissement) {
	cible := t.Schema + "." + t.Nom

	var associations []calque.Association
	var avertissements []calque.Avertissement

	forcees := s.relations[cible]
	decidees := make(map[string]bool, len(forcees))
	for _, r := range forcees {
		decidees[r.colonne] = true
	}

	// Côté propriétaire : une par clé étrangère déclarée, sauf celle qui porte
	// un héritage décidé — elle devient la hiérarchie, pas une association — et
	// celles dont une colonne porte une relation forcée : la décision gagne, et
	// deux associations sur la même colonne écriraient deux fois la même valeur.
	// Sans décision, la clé primaire étrangère donne un un-vers-un comme les
	// autres.
	var heritage *calque.CleEtrangere
	if retenu, decide := s.heritages[cible]; decide && !retenu.racine {
		heritage = s.parents[cible]
	}
	for i := range t.ClesEtrangeres {
		fk := &t.ClesEtrangeres[i]
		if fk == heritage || slices.ContainsFunc(fk.Colonnes, func(c string) bool { return decidees[c] }) {
			continue
		}

		nomCible, connue := s.nomsParTable[fk.SchemaCible+"."+fk.TableCible]
		if citees := slices.DeleteFunc(slices.Clone(fk.Colonnes), func(c string) bool { return !ecartees[c] }); len(citees) > 0 {
			partie := "clé étrangère vers " + fk.SchemaCible + "." + fk.TableCible
			if connue {
				partie = "association " + nomAssociation(fk, nomCible)
			}
			for _, colonne := range citees {
				parties[colonne] = append(parties[colonne], partie)
			}
			continue
		}
		if !connue {
			// La cible est hors du calque — portée restreinte, ou table
			// ignorée. L'association ne se génère pas, mais la colonne reste
			// une propriété ordinaire : les données ne disparaissent pas.
			avertissements = append(avertissements, calque.Avertissement{
				Code:       calque.CodeCibleHorsPortee,
				Cible:      cible + "." + strings.Join(fk.Colonnes, ","),
				Message:    "la clé étrangère désigne " + fk.SchemaCible + "." + fk.TableCible + ", absente du calque ; colonne laissée en propriété",
				Resolution: calque.ResolutionAucune,
				Confiance:  1,
			})
			continue
		}
		if !s.viseIdentifiant(fk) {
			avertissements = append(avertissements, calque.Avertissement{
				Code:  calque.CodeReferenceHorsIdentifiant,
				Cible: cible + "." + strings.Join(fk.Colonnes, ","),
				Message: "la clé étrangère désigne " + fk.SchemaCible + "." + fk.TableCible + " (" + strings.Join(fk.ColonnesCibles, ", ") +
					"), qui n'est pas sa clé primaire : Doctrine n'associe que vers l'identifiant ; colonne laissée en propriété",
				Resolution: calque.ResolutionAucune,
				Confiance:  1,
			})
			continue
		}

		associations = append(associations, associationPortee(t, fk, nomCible, parColonne))
	}

	for _, r := range forcees {
		fk := r.fk
		// Une clé déclarée sur la même colonne vers la même table garde son
		// comportement à la suppression : la décision renomme ou requalifie la
		// relation, elle ne change pas ce que fait la base.
		for i := range t.ClesEtrangeres {
			declaree := &t.ClesEtrangeres[i]
			if slices.Equal(declaree.Colonnes, fk.Colonnes) &&
				declaree.SchemaCible == fk.SchemaCible && declaree.TableCible == fk.TableCible {
				fk.ALaSuppression = declaree.ALaSuppression
			}
		}

		a := associationPortee(t, &fk, s.nomsParTable[fk.SchemaCible+"."+fk.TableCible], parColonne)
		if r.genre != "" {
			a.Genre = r.genre
		}
		if r.nom != "" {
			a.Nom = r.nom
		}
		a.Origine = calque.OrigineDecision
		associations = append(associations, a)
	}

	return associations, avertissements
}

// associationPortee construit le côté qui porte la colonne de jointure.
//
// Se tromper de côté produit un mapping que Doctrine accepte et qui n'écrit
// rien en base : le propriétaire est celui dont la table porte la clé
// étrangère, et il n'y a rien à deviner.
func associationPortee(t *calque.Table, fk *calque.CleEtrangere, nomCible string, parColonne map[string]*calque.Propriete) calque.Association {
	a := calque.Association{
		Nom:          nomAssociation(fk, nomCible),
		Genre:        calque.PlusieursVersUn,
		Cible:        nomCible,
		Proprietaire: true,
		Origine:      calque.OrigineContrainte,
	}

	// Une unicité sur les colonnes portantes fait de la relation un
	// un-vers-un. C'est la seule chose qui distingue les deux, et elle se lit
	// dans le physique : le nommage n'y suffirait pas.
	if unicitePorte(t, fk.Colonnes) {
		a.Genre = calque.UnVersUn
	}

	for rang, colonne := range fk.Colonnes {
		jointure := calque.ColonneJointure{
			Colonne:        colonne,
			ALaSuppression: fk.ALaSuppression,
		}
		if rang < len(fk.ColonnesCibles) {
			jointure.ColonneReferencee = fk.ColonnesCibles[rang]
		}
		if propriete, connue := parColonne[colonne]; connue {
			jointure.Nullable = propriete.Nullable
		}
		a.Jointure = append(a.Jointure, jointure)
	}

	return a
}

// nomAssociation nomme la propriété qui porte la relation.
//
// La convention client_id vers client couvre la quasi-totalité des cas, et
// c'est du retrait de suffixe, pas de la morphologie : rien n'est deviné sur la
// langue ou le nombre.
//
// Une clé composite ou une colonne sans suffixe reconnaissable retombe sur le
// nom de la classe cible, qui est toujours juste même s'il est parfois moins
// naturel.
func nomAssociation(fk *calque.CleEtrangere, nomCible string) string {
	if len(fk.Colonnes) != 1 {
		return camelCase(nomCible)
	}

	nu := retirerSuffixeIdentifiant(fk.Colonnes[0])
	if nu == "" {
		return camelCase(nomCible)
	}
	return camelCase(nu)
}

// Suffixes qui désignent la clé d'une autre table, du plus long au plus court.
//
// L'ordre compte : _ident se terminant par _id, tester le court d'abord
// laisserait un ent orphelin.
var suffixesIdentifiant = []string{"_ident", "_code", "_num", "_no", "_fk", "_id"}

// retirerSuffixeIdentifiant enlève la marque de clé étrangère d'un nom de
// colonne, ou rend la chaîne vide quand il n'y en a pas.
//
// Rend aussi la chaîne vide quand le retrait ne laisserait rien : une colonne
// nommée id désigne la table elle-même, pas une relation.
func retirerSuffixeIdentifiant(colonne string) string {
	bas := strings.ToLower(colonne)
	for _, suffixe := range suffixesIdentifiant {
		if strings.HasSuffix(bas, suffixe) && len(colonne) > len(suffixe) {
			return colonne[:len(colonne)-len(suffixe)]
		}
	}
	return ""
}

// unicitePorte dit si une contrainte d'unicité couvre exactement ces colonnes.
//
// Exactement : une unicité sur (client_id, date) n'interdit pas deux lignes
// pour le même client, et la relation reste un plusieurs-vers-un.
func unicitePorte(t *calque.Table, colonnes []string) bool {
	// La clé primaire est une unicité. Une clé étrangère qui la couvre
	// exactement — la table enfant d'un héritage non déclaré — porte donc un
	// un-vers-un. Une clé qui n'en couvre qu'une partie, comme chacune des deux
	// d'une ligne de commande, n'est pas unique à elle seule.
	if t.ClePrimaire != nil && memesColonnes(t.ClePrimaire.Colonnes, colonnes) {
		return true
	}
	for _, u := range t.Unicites {
		if memesColonnes(u.Colonnes, colonnes) {
			return true
		}
	}
	return false
}

// memesColonnes dit si deux listes désignent le même ensemble de colonnes,
// dans n'importe quel ordre.
func memesColonnes(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	cherchees := ensemble(b)
	for _, c := range a {
		if !cherchees[c] {
			return false
		}
	}
	return true
}
