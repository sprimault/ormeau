// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"sort"
	"strconv"
	"strings"

	"github.com/sprimault/ormeau/internal/calque"
)

// Les colonnes d'horodatage se répètent d'une table à l'autre, et se sortent
// dans un trait plutôt que d'être recopiées dans chaque entité.
//
// Contrairement au nommage, le trait ne change le sens de rien : il déplace des
// propriétés identiques dans un fichier partagé. C'est une réorganisation, pas
// un jugement, et elle s'applique — mais seulement quand elle sert vraiment.

// Noms de colonnes reconnus comme des horodatages techniques, en français et en
// anglais.
//
// La liste est fermée et le restera : elle ne cherche pas à deviner qu'une
// colonne est un horodatage — le type le dit déjà — mais qu'elle relève de la
// plomberie plutôt que du métier. date_facture est un horodatage, et n'a rien à
// faire dans un trait.
var horodatagesTechniques = map[string]bool{
	"created_at": true, "updated_at": true, "deleted_at": true,
	"date_creation": true, "date_modification": true, "date_suppression": true,
	"cree_le": true, "modifie_le": true, "supprime_le": true,
	"date_creat": true, "date_maj": true,
}

// Nombre d'entités en dessous duquel un trait ne se justifie pas.
//
// Un trait utilisé une seule fois est un fichier de plus pour rien : les
// propriétés sont aussi bien dans l'entité, où on les lit sans ouvrir autre
// chose.
const minimumEntitesParTrait = 2

// extraireTraits sort les propriétés d'horodatage des entités et rend les
// traits correspondants.
//
// Le regroupement se fait par signature — mêmes noms, mêmes types, mêmes
// nullabilités. Deux tables horodatées différemment ne partagent pas de trait :
// un trait dont les colonnes ne correspondent pas produirait un mapping faux,
// et Doctrine ne s'en plaindrait qu'à l'exécution.
func extraireTraits(logique *calque.Logique) []calque.Avertissement {
	parSignature := map[string][]int{}
	proprietesParSignature := map[string][]calque.Propriete{}

	for i := range logique.Entites {
		candidates := horodatagesDe(&logique.Entites[i])
		if len(candidates) == 0 {
			continue
		}

		signature := signatureDe(candidates)
		parSignature[signature] = append(parSignature[signature], i)
		proprietesParSignature[signature] = candidates
	}

	var traits []calque.Trait
	var avertissements []calque.Avertissement

	for _, signature := range clesTriees(parSignature) {
		entites := parSignature[signature]
		if len(entites) < minimumEntitesParTrait {
			continue
		}

		proprietes := proprietesParSignature[signature]
		nom := nomDeTrait(proprietes, traits)

		for _, rang := range entites {
			e := &logique.Entites[rang]
			e.Traits = append(e.Traits, nom)
			e.Proprietes = sansLesHorodatages(e.Proprietes, proprietes)

			avertissements = append(avertissements, calque.Avertissement{
				Code:       calque.CodeTraitDeduit,
				Cible:      e.Table.Schema + "." + e.Table.Nom,
				Message:    "colonnes d'horodatage sorties dans le trait " + nom + ", partagé par " + strings.Join(nomsDesEntites(logique, entites), ", "),
				Resolution: calque.ResolutionParDefaut,
				Confiance:  0.8,
			})
		}

		traits = append(traits, calque.Trait{Nom: nom, Proprietes: proprietes})
	}

	logique.Traits = traits
	return avertissements
}

// horodatagesDe rend les propriétés d'horodatage technique d'une entité, dans
// l'ordre du calque.
//
// Une propriété qui porte une association ou un identifiant n'y figure jamais :
// sortir une clé d'une entité la rendrait ingénérable.
func horodatagesDe(e *calque.Entite) []calque.Propriete {
	dansLaCle := map[string]bool{}
	if e.Identifiant != nil {
		for _, nom := range e.Identifiant.Proprietes {
			dansLaCle[nom] = true
		}
	}

	var candidates []calque.Propriete
	for _, p := range e.Proprietes {
		if dansLaCle[p.Nom] || !horodatagesTechniques[strings.ToLower(p.Colonne)] {
			continue
		}
		candidates = append(candidates, p)
	}
	return candidates
}

// signatureDe rend une clé qui identifie un jeu de propriétés interchangeable.
//
// Le nom, le type et la nullabilité en font partie : deux tables dont l'une a
// updated_at facultative et l'autre obligatoire ne peuvent pas partager le même
// trait.
func signatureDe(proprietes []calque.Propriete) string {
	parties := make([]string, 0, len(proprietes))
	for _, p := range proprietes {
		partie := p.Nom + ":" + p.TypePHP + ":" + p.TypeDoctrine
		if p.Nullable {
			partie += ":nullable"
		}
		parties = append(parties, partie)
	}
	sort.Strings(parties)
	return strings.Join(parties, "|")
}

// nomDeTrait nomme un trait d'après ce qu'il contient.
//
// Horodatage tant qu'il est libre, puis le nom des propriétés : deux
// signatures différentes dans le même calque doivent donner deux fichiers
// distincts, et un Horodatage2 n'apprendrait rien sur ce qui les sépare.
func nomDeTrait(proprietes []calque.Propriete, deja []calque.Trait) string {
	pris := make(map[string]bool, len(deja))
	for _, t := range deja {
		pris[t.Nom] = true
	}

	if !pris["Horodatage"] {
		return "Horodatage"
	}

	var b strings.Builder
	for _, p := range proprietes {
		b.WriteString(pascalCase(p.Nom))
	}
	nom := b.String()
	if nom == "" || pris[nom] {
		// strconv et non une arithmétique sur rune : au-delà de neuf traits,
		// '0'+n sort de la plage des chiffres et produit un deux-points.
		return "Horodatage" + strconv.Itoa(len(deja)+1)
	}
	return nom
}

// sansLesHorodatages rend les propriétés qui restent sur l'entité.
func sansLesHorodatages(proprietes, sorties []calque.Propriete) []calque.Propriete {
	aRetirer := make(map[string]bool, len(sorties))
	for _, p := range sorties {
		aRetirer[p.Nom] = true
	}

	restantes := make([]calque.Propriete, 0, len(proprietes))
	for _, p := range proprietes {
		if !aRetirer[p.Nom] {
			restantes = append(restantes, p)
		}
	}
	return restantes
}

// nomsDesEntites rend les noms de classes désignés par leurs rangs.
func nomsDesEntites(logique *calque.Logique, rangs []int) []string {
	noms := make([]string, 0, len(rangs))
	for _, rang := range rangs {
		noms = append(noms, logique.Entites[rang].Nom)
	}
	return noms
}
