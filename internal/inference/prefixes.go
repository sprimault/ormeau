// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"sort"
	"strings"

	"github.com/sprimault/ormeau/internal/calque"
)

// Le préfixe se trouve, il ne se demande pas : une base dont toutes les tables
// s'appellent T_QUELQUE_CHOSE porte une convention de nommage, pas un mot de
// trois lettres qui appartiendrait au domaine.
//
// Le coût d'une erreur est faible et réversible — un nom de classe discutable,
// que prefixes_a_retirer corrige —, ce qui autorise la détection. Elle reste
// conservatrice : tout ou rien, et jamais sur un préfixe long.

// Longueur maximale d'un préfixe détecté, séparateur compris.
//
// T_, tbl_, dbo_ tiennent dedans ; commande_ non, et c'est voulu : au-delà, un
// début commun est un domaine et non une convention. Trois tables nommées
// commande_ligne, commande_entete et commande_taxe ne demandent pas à devenir
// Ligne, Entete et Taxe.
const longueurMaxPrefixe = 5

// Nombre de tables en dessous duquel aucune détection n'a lieu.
//
// Sur deux tables, un début commun est une coïncidence aussi souvent qu'une
// convention, et rien ne permet de trancher.
const minimumTablesPourDetecter = 3

// prefixeCommun rend le préfixe que toutes les tables partagent, ou la chaîne
// vide.
//
// Toutes, sans exception ni seuil : une seule table hors convention suffit à
// rendre la détection incertaine, et un seuil à quatre-vingts pour cent
// produirait des noms de classes incohérents entre eux dans le même schéma —
// la moitié amputée, l'autre non.
//
// Le préfixe doit se terminer par un séparateur. Sans cette contrainte, quatre
// tables commençant par c donneraient un préfixe d'une lettre.
func prefixeCommun(tables []calque.Table) string {
	if len(tables) < minimumTablesPourDetecter {
		return ""
	}

	// Le plus court nom borne le préfixe possible, et il faut qu'il reste
	// quelque chose après : une table qui s'appelle exactement T_ n'en est pas
	// une, mais T_ suivi de rien produirait une classe sans nom.
	commun := strings.ToUpper(tables[0].Nom)
	for i := range tables {
		commun = debutCommun(commun, strings.ToUpper(tables[i].Nom))
		if commun == "" {
			return ""
		}
	}

	if len(commun) > longueurMaxPrefixe {
		commun = commun[:longueurMaxPrefixe]
	}

	// On recule jusqu'au dernier séparateur : T_CL commun à T_CLIENTS et
	// T_CLOTURES doit rendre T_, pas T_CL.
	coupe := strings.LastIndexAny(commun, "_-")
	if coupe < 0 {
		return ""
	}

	// Un préfixe qui consommerait un nom entier n'en est pas un.
	for i := range tables {
		if len(tables[i].Nom) <= coupe+1 {
			return ""
		}
	}

	// Rendu dans la casse de la première table, et non dans celle de la
	// comparaison : le préfixe s'affiche dans un avertissement, et un tbl_
	// annoncé TBL_ ne correspond à rien de ce que l'utilisateur voit en base.
	return tables[0].Nom[:coupe+1]
}

// debutCommun rend le plus long début partagé par deux chaînes.
func debutCommun(a, b string) string {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return a[:i]
		}
	}
	return a[:n]
}

// retirerPrefixe enlève le premier préfixe qui correspond, sans distinction de
// casse — une base SQL Server mélange T_CLIENTS et t_clients sans que cela
// désigne deux conventions.
//
// Rend le nom inchangé quand rien ne correspond, ou quand le retrait ne
// laisserait rien.
func retirerPrefixe(nom string, prefixes []string) string {
	bas := strings.ToLower(nom)
	for _, p := range prefixes {
		if p == "" {
			continue
		}
		if pb := strings.ToLower(p); strings.HasPrefix(bas, pb) && len(nom) > len(pb) {
			return nom[len(pb):]
		}
	}
	return nom
}

// prefixesRetenus rend les préfixes à retirer, et celui que l'outil a repéré
// sans qu'on le lui demande.
//
// Les deux ne se mélangent jamais : seule la décision agit. Un nom de table est
// un constat, et l'amputer d'un préfixe change ce que l'utilisateur lira dans
// son code pendant des années. La détection ne fait que signaler — c'est ce qui
// permet à quelqu'un qui découvre l'outil de savoir que l'option existe, sans
// avoir à lire la documentation pour s'en douter.
//
// Le second retour est vide quand une décision existe déjà : ce qui a été
// arbitré n'a pas à être proposé.
func prefixesRetenus(tables []calque.Table, d *Decisions) ([]string, string) {
	if len(d.PrefixesARetirer) > 0 {
		retenus := append([]string(nil), d.PrefixesARetirer...)
		sort.Strings(retenus)
		return retenus, ""
	}
	return nil, prefixeCommun(tables)
}
