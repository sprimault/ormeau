// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
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
// Les clés étrangères jamais déclarées, cas majoritaire sur du legacy, ne sont
// pas ici : elles demandent d'échantillonner les données, ou une décision.

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

	// parents donne, pour une table qui hérite, la clé étrangère qui porte
	// l'héritage.
	parents map[string]*calque.CleEtrangere
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

		if j := reconnaitreJointure(t); j != nil {
			s.jointures[cible] = j
			continue
		}
		if fk := reconnaitreHeritage(t); fk != nil {
			s.parents[cible] = fk
		}
	}
	return s
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
// partage l'identifiant de la mère, et sa clé primaire pointe dessus. Doctrine
// l'exprime en JOINED.
//
// Il faut que la clé étrangère couvre exactement la clé primaire. Une clé
// primaire composite dont une seule colonne est étrangère décrit une relation
// ordinaire, pas une hiérarchie — le cas d'une table d'affectation datée.
func reconnaitreHeritage(t *calque.Table) *calque.CleEtrangere {
	if t.ClePrimaire == nil || len(t.ClePrimaire.Colonnes) == 0 {
		return nil
	}

	primaires := ensemble(t.ClePrimaire.Colonnes)
	for i := range t.ClesEtrangeres {
		fk := &t.ClesEtrangeres[i]
		if len(fk.Colonnes) != len(primaires) {
			continue
		}

		couvre := true
		for _, c := range fk.Colonnes {
			if !primaires[c] {
				couvre = false
				break
			}
		}
		if couvre && fk.SchemaCible+"."+fk.TableCible != t.Schema+"."+t.Nom {
			return fk
		}
	}
	return nil
}

// inferrerAssociations rend les associations d'une entité : celles qu'elle
// porte, et celles dont elle est la cible.
func inferrerAssociations(t *calque.Table, s *schemaLogique, parColonne map[string]*calque.Propriete) ([]calque.Association, []calque.Avertissement) {
	cible := t.Schema + "." + t.Nom

	var associations []calque.Association
	var avertissements []calque.Avertissement

	// Côté propriétaire : une par clé étrangère déclarée, sauf celle qui porte
	// l'héritage — elle devient la hiérarchie, pas une propriété.
	heritage := s.parents[cible]
	for i := range t.ClesEtrangeres {
		fk := &t.ClesEtrangeres[i]
		if fk == heritage {
			continue
		}

		nomCible, connue := s.nomsParTable[fk.SchemaCible+"."+fk.TableCible]
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

		associations = append(associations, associationPortee(t, fk, nomCible, parColonne))
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
	for _, u := range t.Unicites {
		if len(u.Colonnes) != len(colonnes) {
			continue
		}
		cherchees := ensemble(colonnes)
		couvre := true
		for _, c := range u.Colonnes {
			if !cherchees[c] {
				couvre = false
				break
			}
		}
		if couvre {
			return true
		}
	}
	return false
}
