// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Commande ormeau : introspecte une base, produit un calque, et le compare.
//
// La génération d'entités n'est pas ici : elle vit dans le paquet PHP, parce que
// la régénération non destructive demande un AST PHP.
package main

import (
	"flag"
	"fmt"
	"os"
)

// version est posée à l'édition de liens par -ldflags -X. Sa valeur par défaut
// vaut pour un binaire construit sans, par go run ou go build direct.
//
// Elle doit exister : le linker ne signale pas une cible -X absente, et la
// version serait silencieusement perdue sur tous les binaires publiés.
var version = "dev"

// usage est ce que voit quelqu'un qui lance le binaire sans argument.
const usage = `ormeau — rétro-ingénierie de bases de données vers des entités ORM

Usage :
  ormeau extraire --dsn <dsn> --sortie <fichier.calque.json>
  ormeau extraire --sgbd <sgbd> --hote <hote> --utilisateur <nom> [--base <base>] --sortie <chemin>
  ormeau inferer  <fichier.calque.json> [--decisions <fichier.yaml>] [--sortie <fichier.logique.json>]
  ormeau diff     <fichier.calque.json> [--dsn <dsn>]

Commandes :
  extraire   lit le catalogue et écrit un calque physique
  inferer    applique les heuristiques et écrit un calque logique
  diff       compare un calque enregistré à l'état actuel de la base
  version    affiche la version du binaire

Sans --base, toutes les bases du serveur sont extraites et --sortie désigne
alors un répertoire.

Les trois fichiers d'une base partagent leur préfixe — gescom.calque.json,
gescom.decisions.yaml, gescom.logique.json — et inferer déduit les deux autres
du premier. Au premier passage, il écrit le fichier de décisions prérempli,
entièrement en commentaire.

Variables d'environnement :
  ORMEAU_DSN   chaîne de connexion, à défaut de --dsn
  ORMEAU_MDP   mot de passe ; il n'existe pas de drapeau, qui l'exposerait dans ps
`

// versionAffichee rend la ligne que voit l'utilisateur. Ce qu'il recopiera
// dans une issue : le nom seul ne dirait pas de quel binaire il parle.
func versionAffichee() string {
	return "ormeau " + version
}

// main route la sous-commande et traduit l'erreur en code de retour. Seul
// endroit du projet où os.Exit est acceptable.
func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "extraire":
		err = extraire(os.Args[2:])
	case "inferer":
		err = inferer(os.Args[2:])
	case "diff":
		err = diffuser(os.Args[2:])
	case "-h", "--help", "aide":
		fmt.Print(usage)
		return
	case "version", "--version":
		fmt.Println(versionAffichee())
		return
	default:
		fmt.Fprintf(os.Stderr, "commande inconnue : %s\n\n%s", os.Args[1], usage)
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// diffuser compare un calque enregistré à la base, ou à un second calque.
// N'écrit rien, et rend un code non nul en cas de divergence.
func diffuser(args []string) error {
	jeu := flag.NewFlagSet("diff", flag.ExitOnError)
	dsn := jeu.String("dsn", "", "base à comparer ; à défaut, un second calque est attendu en argument")
	format := jeu.String("format", "texte", "texte ou json")
	if err := jeu.Parse(args); err != nil {
		return err
	}
	_ = dsn
	_ = format

	return fmt.Errorf("à implémenter : phase 8 de la feuille de route")
}
