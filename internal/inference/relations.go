// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"cmp"
	"slices"
	"strings"

	"github.com/sprimault/ormeau/internal/calque"
)

// Les relations forcées sont les clés étrangères que personne n'a déclarées :
// l'utilisateur les écrit, l'inférence les traite comme si le catalogue les
// portait. Une relation forcée gagne sur une clé déclarée sur la même colonne,
// comme toute décision gagne sur l'heuristique.

// relationForcee est une relation décidée, vérifiée contre le calque.
type relationForcee struct {
	// colonne est celle de la table source qui porte la relation.
	colonne string
	// fk décrit la relation comme une clé étrangère d'une colonne, pour passer
	// par le même chemin qu'une clé déclarée.
	fk calque.CleEtrangere
	// genre est vide quand la décision laisse l'unicité trancher.
	genre calque.GenreAssociation
	nom   string
}

// verifierRelationsForcees rattache chaque relation forcée à sa table source,
// et signale celles qui ne s'appliquent pas.
//
// Une relation se refuse plutôt que de s'inventer : une colonne absente, une
// table hors du calque ou écartée, un genre qu'une colonne ne peut pas porter.
// Ce refus est le signal que la base a bougé sous le fichier, ou que la
// décision a été mal écrite — dans les deux cas, l'utilisateur doit le voir.
//
// Une colonne ne porte qu'une relation. Les décisions sont triées avant d'être
// lues, comme le fichier les réécrit : c'est la première dans cet ordre qui
// l'emporte, quelle que soit la façon dont le fichier a été rédigé.
func verifierRelationsForcees(p *calque.Physique, d *Decisions, s *schemaLogique) []calque.Avertissement {
	if len(d.RelationsForcees) == 0 {
		return nil
	}

	// Seules les tables qui produisent une entité peuvent porter ou recevoir
	// une association : une table écartée ou de jointure n'a pas de classe.
	var generees []*calque.Table
	for i := range p.Tables {
		t := &p.Tables[i]
		cible := t.Schema + "." + t.Nom
		if _, nommee := s.nomsParTable[cible]; !nommee {
			continue
		}
		if _, jointure := s.jointures[cible]; jointure {
			continue
		}
		generees = append(generees, t)
	}

	relations := slices.Clone(d.RelationsForcees)
	slices.SortStableFunc(relations, func(x, y RelationForcee) int {
		return cmp.Or(
			strings.Compare(x.Source, y.Source),
			strings.Compare(x.Cible, y.Cible),
			strings.Compare(x.Nom, y.Nom),
			strings.Compare(x.Genre, y.Genre),
		)
	})

	s.relations = map[string][]relationForcee{}
	reliees := map[string]bool{}
	var avertissements []calque.Avertissement
	refuser := func(r RelationForcee, raison string) {
		avertissements = append(avertissements, calque.Avertissement{
			Code:       calque.CodeDecisionOrpheline,
			Cible:      r.Source,
			Message:    "relation forcée vers " + r.Cible + " non appliquée : " + raison,
			Resolution: calque.ResolutionAucune,
			Confiance:  1,
		})
	}

	for _, r := range relations {
		source, colonneSource := colonneDuCalque(generees, r.Source)
		if source == nil {
			refuser(r, "la source n'est pas une colonne d'une table générée")
			continue
		}
		// Deux décisions qui se contredisent : la colonne ne serait plus dans
		// l'entité, et l'association l'écrirait quand même.
		if ecarteeParDecision(source, d, colonneSource) {
			refuser(r, "la colonne source est retirée de l'entité par colonnes_ignorees")
			continue
		}
		cible, colonneCible := colonneDuCalque(generees, r.Cible)
		if cible == nil {
			refuser(r, "la cible n'est pas une colonne d'une table générée")
			continue
		}
		// Doctrine n'associe que vers l'identifiant : une colonne unique, ou
		// une colonne d'une clé composite, donnerait un mapping refusé.
		if cible.ClePrimaire == nil || !slices.Equal(cible.ClePrimaire.Colonnes, []string{colonneCible}) {
			refuser(r, "la cible n'est pas l'identifiant de "+cible.Schema+"."+cible.Nom+" : Doctrine n'associe que vers la clé primaire")
			continue
		}

		// Une colonne désigne une ligne : elle porte un objet, jamais une
		// collection. Le côté collection existe, mais sur l'autre entité, et il
		// se déduit.
		genre := calque.GenreAssociation(r.Genre)
		if genre != "" && genre != calque.PlusieursVersUn && genre != calque.UnVersUn {
			refuser(r, "le genre "+r.Genre+" ne part pas d'une colonne ; plusieurs_vers_un ou un_vers_un, l'autre côté se déduit")
			continue
		}
		if reliees[r.Source] {
			refuser(r, "la colonne porte déjà une relation forcée")
			continue
		}
		reliees[r.Source] = true

		qualifiee := source.Schema + "." + source.Nom
		s.relations[qualifiee] = append(s.relations[qualifiee], relationForcee{
			colonne: colonneSource,
			fk: calque.CleEtrangere{
				Colonnes:       []string{colonneSource},
				SchemaCible:    cible.Schema,
				TableCible:     cible.Nom,
				ColonnesCibles: []string{colonneCible},
			},
			genre: genre,
			nom:   r.Nom,
		})
	}
	return avertissements
}

// colonneDuCalque retrouve la table et la colonne que désigne une référence
// schema.table.colonne, ou rend une table nulle.
//
// La table retenue est la plus longue qui préfixe la référence : un point dans
// un nom de table ne coupe pas la référence au mauvais endroit.
func colonneDuCalque(tables []*calque.Table, reference string) (*calque.Table, string) {
	var trouvee *calque.Table
	var colonne string
	longueur := 0
	for _, t := range tables {
		qualifiee := t.Schema + "." + t.Nom
		if strings.HasPrefix(reference, qualifiee+".") && len(qualifiee) > longueur {
			trouvee, colonne, longueur = t, reference[len(qualifiee)+1:], len(qualifiee)
		}
	}
	if trouvee == nil {
		return nil, ""
	}
	for i := range trouvee.Colonnes {
		if trouvee.Colonnes[i].Nom == colonne {
			return trouvee, colonne
		}
	}
	return nil, ""
}
