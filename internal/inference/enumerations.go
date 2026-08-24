// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"sort"
	"strings"

	"github.com/sprimault/ormeau/internal/calque"
)

// Une colonne à valeurs fermées devient un enum PHP.
//
// Deux sources, de fiabilité très différente. Le type énuméré natif est un
// constat : le catalogue le déclare, ses valeurs sont exactes, rien à deviner.
// La contrainte de vérification demande de lire une expression écrite par le
// serveur, et toutes ne décrivent pas une énumération — la plupart n'en
// décrivent pas.
//
// Ce qui reste incertain dans les deux cas est le nom des cas PHP : ACTIF donne
// Actif sans hésitation, O et N ne donnent rien de lisible. Le nom se propose
// alors, il ne s'invente pas.

// enumeree décrit ce qu'on a reconnu sur une colonne.
type enumeree struct {
	nom     string
	valeurs []string
	origine calque.Origine
}

// enumerationsDuSchema rend, par colonne qualifiée, l'énumération reconnue.
//
// Les décisions passent avant tout : une énumération déclarée à la main gagne,
// y compris contre un type natif — c'est le moyen de nommer les cas d'un O/N
// que rien ne peut deviner.
func enumerationsDuSchema(p *calque.Physique, d *Decisions) map[string]enumeree {
	reconnues := map[string]enumeree{}

	natifs := make(map[string][]string, len(p.TypesEnumeres))
	for i := range p.TypesEnumeres {
		te := &p.TypesEnumeres[i]
		natifs[te.Nom] = te.Valeurs
	}

	for i := range p.Tables {
		t := &p.Tables[i]

		for j := range t.Colonnes {
			c := &t.Colonnes[j]
			cible := t.Schema + "." + t.Nom + "." + c.Nom

			if valeurs, natif := natifs[c.TypeEnumere]; natif && c.TypeEnumere != "" {
				reconnues[cible] = enumeree{
					nom:     pascalCase(c.TypeEnumere),
					valeurs: valeurs,
					origine: calque.OrigineContrainte,
				}
			}
		}

		// La vérification n'écrase pas un type natif : elle le redit au mieux,
		// et le contredit au pire.
		for _, v := range t.Verifications {
			colonne, valeurs := valeursDUneVerification(v.Expression, t)
			if colonne == "" {
				continue
			}

			cible := t.Schema + "." + t.Nom + "." + colonne
			if _, deja := reconnues[cible]; deja {
				continue
			}
			reconnues[cible] = enumeree{
				nom:     pascalCase(t.Nom + "_" + colonne),
				valeurs: valeurs,
				origine: calque.OrigineVerification,
			}
		}
	}

	appliquerEnumerationsDecidees(reconnues, d)
	return reconnues
}

// appliquerEnumerationsDecidees remplace ce qui a été arbitré à la main.
//
// Une décision porte le nom du type et l'appariement des cas : c'est le seul
// endroit où un O peut devenir Oui, parce que seul un humain sait ce que la
// lettre abrège.
func appliquerEnumerationsDecidees(reconnues map[string]enumeree, d *Decisions) {
	for _, forcee := range d.Enumerations {
		if forcee.Colonne == "" || forcee.Nom == "" {
			continue
		}

		e := enumeree{nom: forcee.Nom, origine: calque.OrigineDecision}
		if existante, connue := reconnues[forcee.Colonne]; connue {
			e.valeurs = existante.valeurs
		}
		for valeur := range forcee.Cas {
			if !contientValeur(e.valeurs, valeur) {
				e.valeurs = append(e.valeurs, valeur)
			}
		}
		sort.Strings(e.valeurs)

		reconnues[forcee.Colonne] = e
	}
}

// valeursDUneVerification extrait la colonne et les valeurs d'un CHECK qui
// décrit une énumération, ou rend une colonne vide.
//
// Conservateur, et il doit l'être : la plupart des contraintes de vérification
// ne sont pas des énumérations, et en transformer une en enum PHP produirait un
// type que la base ne garantit pas. Trois conditions, toutes nécessaires.
//
// Une seule colonne mentionnée. Un CHECK qui compare deux colonnes exprime une
// règle métier — « un motif est requis quand l'état vaut ANNULEE » —, pas un
// domaine de valeurs.
//
// Au moins deux littéraux chaîne. Un seul décrit une valeur imposée, ce qu'un
// enum à un cas n'apporte pas.
//
// Aucun opérateur de comparaison ordonnée. `solde >= 0` mentionne une colonne
// et un littéral sans rien énumérer.
func valeursDUneVerification(expression string, t *calque.Table) (string, []string) {
	// Un seul test couvre les deux familles : >= et <= comparent un ordre, <>
	// exclut une valeur, et aucune des trois n'énumère un domaine.
	if strings.ContainsAny(expression, "<>") {
		return "", nil
	}

	valeurs := litterauxChaine(expression)
	if len(valeurs) < 2 {
		return "", nil
	}

	var trouvee string
	for i := range t.Colonnes {
		nom := t.Colonnes[i].Nom
		if !mentionne(expression, nom) {
			continue
		}
		if trouvee != "" {
			return "", nil
		}
		trouvee = nom
	}
	return trouvee, valeurs
}

// litterauxChaine rend les littéraux entre apostrophes, dédoublonnés et dans
// l'ordre de l'expression.
//
// L'ordre est celui du CHECK et non l'alphabétique : il porte souvent une
// progression métier — prospect, actif, suspendu, radié — que trier détruirait.
//
// Les transtypages que PostgreSQL colle derrière chaque littéral sont ignorés :
// ils suivent l'apostrophe fermante, jamais l'ouvrante.
func litterauxChaine(expression string) []string {
	var valeurs []string
	vues := map[string]bool{}

	reste := expression
	for {
		debut := strings.Index(reste, "'")
		if debut < 0 {
			return valeurs
		}
		reste = reste[debut+1:]

		fin := strings.Index(reste, "'")
		if fin < 0 {
			return valeurs
		}

		if valeur := reste[:fin]; valeur != "" && !vues[valeur] {
			vues[valeur] = true
			valeurs = append(valeurs, valeur)
		}
		reste = reste[fin+1:]
	}
}

// mentionne dit si l'expression cite ce nom de colonne comme un mot entier.
//
// La recherche est bornée par des caractères qui ne peuvent pas appartenir à un
// identifiant : sans ça, la colonne actif serait trouvée dans inactif, et deux
// colonnes sembleraient citées là où une seule l'est.
func mentionne(expression, colonne string) bool {
	reste := expression
	decalage := 0

	for {
		i := strings.Index(reste, colonne)
		if i < 0 {
			return false
		}

		avant := byte(' ')
		if decalage+i > 0 {
			avant = expression[decalage+i-1]
		}
		apres := byte(' ')
		if fin := decalage + i + len(colonne); fin < len(expression) {
			apres = expression[fin]
		}

		if !partieDIdentifiant(avant) && !partieDIdentifiant(apres) {
			return true
		}
		reste = reste[i+len(colonne):]
		decalage += i + len(colonne)
	}
}

// partieDIdentifiant dit si l'octet peut appartenir à un identifiant SQL.
func partieDIdentifiant(b byte) bool {
	switch {
	case b >= 'a' && b <= 'z', b >= 'A' && b <= 'Z', b >= '0' && b <= '9':
		return true
	case b == '_':
		return true
	}
	return false
}

// collecterEnumerations rend les types énumérés à générer, triés par nom.
//
// Deux colonnes peuvent partager le même type — c'est le cas normal d'un type
// natif référencé par plusieurs tables. Il ne se génère alors qu'une fois, et
// les avertissements sur ses cas ne se répètent pas non plus.
func collecterEnumerations(s *schemaLogique, d *Decisions) ([]calque.Enumeration, []calque.Avertissement) {
	casDecides := map[string]map[string]string{}
	for _, forcee := range d.Enumerations {
		if forcee.Nom != "" && len(forcee.Cas) > 0 {
			casDecides[forcee.Nom] = forcee.Cas
		}
	}

	var enumerations []calque.Enumeration
	var avertissements []calque.Avertissement
	produits := map[string]bool{}

	for _, cible := range clesTriees(s.enumerations) {
		e := s.enumerations[cible]
		if produits[e.nom] {
			continue
		}
		produits[e.nom] = true

		enumeration, avs := enumerationLogique(cible, e, casDecides[e.nom])
		enumerations = append(enumerations, enumeration)
		avertissements = append(avertissements, avs...)
	}

	sort.SliceStable(enumerations, func(i, j int) bool {
		return enumerations[i].Nom < enumerations[j].Nom
	})
	return enumerations, avertissements
}

// enumerationLogique traduit une énumération reconnue en type PHP à générer.
//
// Le second retour porte les avertissements sur le nom des cas : c'est la seule
// part de l'énumération qui relève du jugement. La valeur stockée est un
// constat, le nom du cas est ce qu'un développeur lira dans son code.
func enumerationLogique(cible string, e enumeree, decides map[string]string) (calque.Enumeration, []calque.Avertissement) {
	enumeration := calque.Enumeration{
		Nom:         e.nom,
		TypeSupport: "string",
		Origine:     e.origine,
	}

	var avertissements []calque.Avertissement
	for _, valeur := range e.valeurs {
		nom, lisible := nomDeCas(valeur, decides)
		if !lisible {
			avertissements = append(avertissements, calque.Avertissement{
				Code:       calque.CodeCasEnumerationOpaque,
				Cible:      cible,
				Message:    "la valeur " + valeur + " donne un cas nommé " + nom + " ; enumerations permet de l'apparier à un nom lisible",
				Resolution: calque.ResolutionParDefaut,
				Confiance:  0.4,
			})
		}
		enumeration.Cas = append(enumeration.Cas, calque.CasEnumeration{Nom: nom, Valeur: valeur})
	}

	return enumeration, avertissements
}

// nomDeCas rend le nom PHP d'une valeur stockée.
//
// Le second retour dit si le nom apprend quelque chose. ACTIF donne Actif, et
// personne n'a besoin de plus ; O donne O, ce qui ne dit ni Oui ni Ouvert ni
// Optionnel. Ce n'est pas une erreur — la base ne stocke que ça — mais l'appelant
// doit pouvoir le signaler.
//
// Une valeur qui ne commence pas par une lettre reçoit le préfixe Cas : un
// identifiant PHP ne démarre pas par un chiffre, et une énumération de codes
// numériques est courante sur du legacy.
func nomDeCas(valeur string, decides map[string]string) (string, bool) {
	if nom, decide := decides[valeur]; decide {
		return nom, true
	}

	nom := pascalCase(valeur)
	if nom == "" {
		return "Vide", false
	}
	if premier := []rune(nom)[0]; premier == '_' {
		nom = "Cas" + nom[1:]
	}

	// La lisibilité se juge sur la valeur stockée, pas sur le nom produit : 01
	// donne Cas01, cinq caractères qui n'apprennent rien de plus que les deux
	// d'origine. Deux caractères ou moins ne portent pas de sens par eux-mêmes
	// — O, N, A, 01 —, et c'est ce que le fichier de décisions sert à nommer.
	return nom, len([]rune(valeur)) > 2
}

// contientValeur dit si la liste porte la valeur.
func contientValeur(valeurs []string, valeur string) bool {
	for _, v := range valeurs {
		if v == valeur {
			return true
		}
	}
	return false
}
