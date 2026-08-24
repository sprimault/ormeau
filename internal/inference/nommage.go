// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"strings"
	"unicode"
)

// Conversion de casse seulement. Le retrait des préfixes et la singularisation
// viennent avec les heuristiques de nommage : ce sont des jugements, pas des
// transformations mécaniques, et ils appellent leurs cas de référence.

// pascalCase rend un nom de classe à partir d'un identifiant de base.
//
//	t_client       -> TClient
//	CLIENT         -> Client
//	t_référence    -> TRéférence
//	N° Commande    -> NCommande
func pascalCase(nom string) string {
	var b strings.Builder
	for _, mot := range decouperIdentifiant(nom) {
		b.WriteString(capitaliser(mot))
	}
	return valideEnPHP(b.String())
}

// camelCase rend un nom de propriété.
//
//	cli_nom        -> cliNom
//	CLI_CA_TTC     -> cliCaTtc
//	Date de vente  -> dateDeVente
func camelCase(nom string) string {
	mots := decouperIdentifiant(nom)
	if len(mots) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(strings.ToLower(mots[0]))
	for _, mot := range mots[1:] {
		b.WriteString(capitaliser(mot))
	}
	return valideEnPHP(b.String())
}

// valideEnPHP garantit que l'identifiant produit en est un.
//
// Un identifiant PHP ne commence pas par un chiffre, et une table nommée
// 2024_ventes ou une colonne 3EME_TRIMESTRE en produiraient un. Le fichier
// généré ne compilerait pas, et l'erreur tomberait chez l'utilisateur, loin de
// l'inférence qui l'a causée.
//
// Le souligné en tête est la convention la moins surprenante : elle garde le
// nom lisible et reste valide comme classe comme propriété.
func valideEnPHP(nom string) string {
	if nom == "" {
		return ""
	}
	if r := []rune(nom)[0]; unicode.IsDigit(r) {
		return "_" + nom
	}
	return nom
}

// decouperIdentifiant sépare sur tout ce qui n'est ni lettre ni chiffre, et sur
// les changements de casse. Un nom de colonne legacy mélange les deux — cli_nom,
// CliNom, cli-nom désignent la même chose et doivent donner la même propriété.
//
// Tout le reste sépare, et pas seulement le souligné, le tiret et l'espace :
// une base réelle produit N° Commande, Prix (HT), % remise. Ces caractères
// passeraient dans l'identifiant PHP — les octets au-delà de l'ASCII y sont
// même acceptés —, mais donneraient un $n°Commande qu'aucun développeur n'écrit
// et qu'aucun outil ne saura relire.
//
// Les lettres accentuées sont des lettres, et restent : é n'est pas un
// séparateur, sinon Libellé donnerait Libell.
func decouperIdentifiant(nom string) []string {
	var mots []string
	var courant []rune

	for _, r := range nom {
		switch {
		case !unicode.IsLetter(r) && !unicode.IsDigit(r):
			if len(courant) > 0 {
				mots = append(mots, string(courant))
				courant = nil
			}
		case unicode.IsUpper(r) && len(courant) > 0 && !unicode.IsUpper(courant[len(courant)-1]):
			// Coupure sur une majuscule qui suit une minuscule : cliNom donne
			// deux mots, CLI n'en donne qu'un.
			mots = append(mots, string(courant))
			courant = []rune{r}
		default:
			courant = append(courant, r)
		}
	}
	if len(courant) > 0 {
		mots = append(mots, string(courant))
	}
	return mots
}

// capitaliser met la première lettre en majuscule et le reste en minuscules.
// Les accents sont préservés : un identifiant peut en porter, et les remplacer
// produirait une classe qui ne correspond plus à sa table.
func capitaliser(mot string) string {
	if mot == "" {
		return ""
	}

	runes := []rune(strings.ToLower(mot))
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
