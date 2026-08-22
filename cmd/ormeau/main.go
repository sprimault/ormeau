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

const usage = `ormeau — rétro-ingénierie de bases de données vers des entités ORM

Usage :
  ormeau extraire --dsn <dsn> --sortie <fichier.calque.json>
  ormeau inferer  <fichier.calque.json> [--decisions <fichier.yaml>] --sortie <fichier.logique.json>
  ormeau diff     <fichier.calque.json> [--dsn <dsn>]

Commandes :
  extraire   lit le catalogue et écrit un calque physique
  inferer    applique les heuristiques et écrit un calque logique
  diff       compare un calque enregistré à l'état actuel de la base
`

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
	default:
		fmt.Fprintf(os.Stderr, "commande inconnue : %s\n\n%s", os.Args[1], usage)
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func extraire(args []string) error {
	jeu := flag.NewFlagSet("extraire", flag.ExitOnError)
	dsn := jeu.String("dsn", "", "chaîne de connexion, préfixée par le SGBD")
	sortie := jeu.String("sortie", "", "fichier de sortie")
	schemas := jeu.String("schemas", "", "schémas à introspecter, séparés par des virgules")
	echantillonner := jeu.Bool("echantillonner", false, "lire des données pour détecter énumérations et clés étrangères implicites")
	cardinalite := jeu.Int("cardinalite-max", 64, "plafond au-delà duquel une colonne ne produit plus d'échantillon")
	if err := jeu.Parse(args); err != nil {
		return err
	}

	if *dsn == "" || *sortie == "" {
		return fmt.Errorf("--dsn et --sortie sont requis")
	}
	_ = schemas
	_ = echantillonner
	_ = cardinalite

	return fmt.Errorf("à implémenter : phase 2 de la feuille de route")
}

func inferer(args []string) error {
	jeu := flag.NewFlagSet("inferer", flag.ExitOnError)
	decisions := jeu.String("decisions", "", "fichier de décisions YAML")
	sortie := jeu.String("sortie", "", "fichier de sortie")
	espaceDeNoms := jeu.String("espace-de-noms", "App\\Entity", "espace de noms des entités générées")
	if err := jeu.Parse(args); err != nil {
		return err
	}

	if jeu.NArg() != 1 {
		return fmt.Errorf("un calque physique est attendu en argument")
	}
	_ = decisions
	_ = sortie
	_ = espaceDeNoms

	return fmt.Errorf("à implémenter : phase 3 de la feuille de route")
}

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
