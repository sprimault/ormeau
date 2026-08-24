// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"sort"

	"github.com/sprimault/ormeau/internal/calque"
)

// Le côté inverse d'une association, et les relations que porte une table de
// jointure.
//
// Les deux se traitent après coup, quand toutes les entités existent : le côté
// inverse d'une relation vit sur une autre entité que celle qui la déclare, et
// une jointure pure en alimente deux à la fois.

// ajouterCotesInverses complète chaque association propriétaire par son côté
// inverse sur l'entité cible.
//
// Doctrine n'en a pas besoin pour écrire en base — le propriétaire suffit —,
// mais sans lui, $client->getCommandes() n'existe pas, et c'est le premier
// geste de quiconque utilise les entités générées.
//
// Les entités sont parcourues dans l'ordre du calque, déjà trié, et les
// associations ajoutées dans celui des propriétaires : deux inférences du même
// physique rendent les mêmes octets.
func ajouterCotesInverses(logique *calque.Logique) {
	parNom := make(map[string]*calque.Entite, len(logique.Entites))
	for i := range logique.Entites {
		parNom[logique.Entites[i].Nom] = &logique.Entites[i]
	}

	// Les noms déjà pris sur chaque entité, propriétés comprises : une
	// association nommée comme une propriété produirait deux membres de même
	// nom dans la classe générée.
	pris := map[string]map[string]bool{}
	for i := range logique.Entites {
		e := &logique.Entites[i]
		noms := make(map[string]bool, len(e.Proprietes)+len(e.Associations))
		for _, p := range e.Proprietes {
			noms[p.Nom] = true
		}
		for _, a := range e.Associations {
			noms[a.Nom] = true
		}
		pris[e.Nom] = noms
	}

	for i := range logique.Entites {
		source := &logique.Entites[i]

		for j := range source.Associations {
			a := &source.Associations[j]
			if !a.Proprietaire {
				continue
			}

			destination, existe := parNom[a.Cible]
			if !existe {
				continue
			}

			inverse := calque.Association{
				Nom:          nomLibre(pris[destination.Nom], camelCase(source.Nom), a.Nom),
				Genre:        genreInverse(a.Genre),
				Cible:        source.Nom,
				Proprietaire: false,
				MappeePar:    a.Nom,
				Origine:      a.Origine,
			}
			pris[destination.Nom][inverse.Nom] = true

			a.InverseePar = inverse.Nom
			destination.Associations = append(destination.Associations, inverse)
		}
	}
}

// genreInverse retourne la cardinalité.
//
// Un-vers-un est symétrique ; plusieurs-vers-un devient une collection de
// l'autre côté. Le plusieurs-vers-plusieurs n'apparaît pas ici : les tables de
// jointure posent leurs deux côtés d'un coup.
func genreInverse(genre calque.GenreAssociation) calque.GenreAssociation {
	if genre == calque.UnVersUn {
		return calque.UnVersUn
	}
	return calque.UnVersPlusieurs
}

// nomLibre rend le nom souhaité, ou une variante libre.
//
// Deux clés étrangères d'une même table vers la même cible — un client
// facturé et un client livré — donneraient deux côtés inverses de même nom. Le
// nom de l'association propriétaire les départage, parce qu'il vient de la
// colonne et qu'elles diffèrent forcément.
func nomLibre(pris map[string]bool, souhaite, discriminant string) string {
	if !pris[souhaite] {
		return souhaite
	}

	// pascalCase et non capitaliser : le discriminant vient d'un nom
	// d'association déjà en casse chameau, et capitaliser en écraserait les
	// majuscules internes — clientLivre donnerait Clientlivre.
	variante := souhaite + pascalCase(discriminant)
	if !pris[variante] {
		return variante
	}

	// Trois clés étrangères vers la même table avec des noms de colonnes qui
	// se réduisent au même : rare au point d'être suspect, mais un nom en
	// double ne compile pas, alors qu'un nom numéroté se corrige par une
	// décision.
	for n := 2; ; n++ {
		numerote := variante + string(rune('0'+n))
		if !pris[numerote] {
			return numerote
		}
	}
}

// posterJointures ajoute les relations plusieurs-vers-plusieurs portées par les
// tables de jointure pure.
//
// La table elle-même ne produit pas d'entité : elle n'a que ses deux clés, et
// une classe CommandeArticle sans propriété n'apprendrait rien à personne. Ses
// deux extrémités reçoivent chacune une collection.
//
// Le propriétaire est arbitraire — les deux côtés sont interchangeables en
// base — mais il doit être stable : c'est la première clé étrangère du
// catalogue, que l'extraction trie par nom de contrainte.
func posterJointures(logique *calque.Logique, s *schemaLogique) []calque.Avertissement {
	parNom := make(map[string]*calque.Entite, len(logique.Entites))
	for i := range logique.Entites {
		parNom[logique.Entites[i].Nom] = &logique.Entites[i]
	}

	var avertissements []calque.Avertissement

	for _, cible := range clesJointures(s.jointures) {
		j := s.jointures[cible]

		nomGauche, gaucheConnue := s.nomsParTable[j.gauche.SchemaCible+"."+j.gauche.TableCible]
		nomDroite, droiteConnue := s.nomsParTable[j.droite.SchemaCible+"."+j.droite.TableCible]
		if !gaucheConnue || !droiteConnue {
			avertissements = append(avertissements, calque.Avertissement{
				Code:       calque.CodeCibleHorsPortee,
				Cible:      cible,
				Message:    "table de jointure dont une extrémité est absente du calque, association non générée",
				Resolution: calque.ResolutionAucune,
				Confiance:  1,
			})
			continue
		}

		gauche, droite := parNom[nomGauche], parNom[nomDroite]
		if gauche == nil || droite == nil {
			continue
		}

		avertissements = append(avertissements, calque.Avertissement{
			Code:       calque.CodeJointurePure,
			Cible:      cible,
			Message:    "deux clés étrangères et rien d'autre : rendue en association " + nomGauche + " vers " + nomDroite + ", sans entité",
			Resolution: calque.ResolutionParDefaut,
			Confiance:  0.9,
		})

		table := &calque.TableJointure{
			Nom:             j.table.Nom,
			Schema:          j.table.Schema,
			Jointure:        colonnesJointure(j.gauche),
			JointureInverse: colonnesJointure(j.droite),
		}

		// Le nom se vérifie libre sur l'entité qui le portera, pas sur celle
		// qu'il désigne : c'est la classe générée qui ne peut pas avoir deux
		// membres du même nom.
		proprietaire := calque.Association{
			Nom:           nomLibreSur(gauche, camelCase(nomDroite)),
			Genre:         calque.PlusieursVersPlusieurs,
			Cible:         nomDroite,
			Proprietaire:  true,
			TableJointure: table,
			Origine:       calque.OrigineContrainte,
		}
		inverse := calque.Association{
			Nom:          nomLibreSur(droite, camelCase(nomGauche)),
			Genre:        calque.PlusieursVersPlusieurs,
			Cible:        nomGauche,
			Proprietaire: false,
			MappeePar:    proprietaire.Nom,
			Origine:      calque.OrigineContrainte,
		}
		proprietaire.InverseePar = inverse.Nom

		gauche.Associations = append(gauche.Associations, proprietaire)
		droite.Associations = append(droite.Associations, inverse)
	}

	return avertissements
}

// nomLibreSur rend un nom d'association que l'entité ne porte pas déjà.
func nomLibreSur(e *calque.Entite, souhaite string) string {
	pris := make(map[string]bool, len(e.Proprietes)+len(e.Associations))
	for _, p := range e.Proprietes {
		pris[p.Nom] = true
	}
	for _, a := range e.Associations {
		pris[a.Nom] = true
	}
	return nomLibre(pris, souhaite, "Liees")
}

// colonnesJointure traduit les colonnes d'une clé étrangère de jointure.
func colonnesJointure(fk *calque.CleEtrangere) []calque.ColonneJointure {
	colonnes := make([]calque.ColonneJointure, 0, len(fk.Colonnes))
	for rang, colonne := range fk.Colonnes {
		jointure := calque.ColonneJointure{
			Colonne:        colonne,
			ALaSuppression: fk.ALaSuppression,
		}
		if rang < len(fk.ColonnesCibles) {
			jointure.ColonneReferencee = fk.ColonnesCibles[rang]
		}
		colonnes = append(colonnes, jointure)
	}
	return colonnes
}

// clesJointures rend les tables de jointure dans un ordre stable.
func clesJointures(m map[string]*jointurePure) []string {
	cles := make([]string, 0, len(m))
	for cle := range m {
		cles = append(cles, cle)
	}
	sort.Strings(cles)
	return cles
}
