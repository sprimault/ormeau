// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sprimault/ormeau/internal/calque"
	"github.com/sprimault/ormeau/internal/inference"
)

// Suffixes des trois fichiers d'une base. Le préfixe est commun, ce qui permet
// de déduire deux chemins du troisième et de n'en taper qu'un.
const (
	suffixePhysique  = ".calque.json"
	suffixeLogique   = ".logique.json"
	suffixeDecisions = ".decisions.yaml"
)

// inferer applique heuristiques et décisions à un calque physique.
//
// Aucun accès à la base : l'inférence est une fonction pure, et se rejoue donc
// autant de fois qu'on veut sur un calque rapporté de chez un client. C'est ce
// qui permet de corriger un arbitrage sans redemander les identifiants.
func inferer(args []string) error {
	jeu := flag.NewFlagSet("inferer", flag.ExitOnError)
	decisions := jeu.String("decisions", "", "fichier de décisions ; à défaut <base>"+suffixeDecisions)
	sortie := jeu.String("sortie", "", "calque logique à écrire ; à défaut <base>"+suffixeLogique)
	espaceDeNoms := jeu.String("espace-de-noms", "", "espace de noms PHP ; une décision du fichier prime")
	positionnels, err := analyser(jeu, args)
	if err != nil {
		return err
	}
	if len(positionnels) != 1 {
		return errors.New("un calque physique est attendu en argument")
	}
	chemin := positionnels[0]

	physique, err := calque.LirePhysique(chemin)
	if err != nil {
		return err
	}

	base := baseDuChemin(chemin)
	cheminDecisions := defaut(*decisions, base+suffixeDecisions)

	d, err := chargerOuPreremplir(cheminDecisions, physique, *decisions != "")
	if err != nil {
		return err
	}
	// Le drapeau ne sert qu'à défaut de décision : celui qui a écrit son espace
	// de noms dans le fichier ne veut pas qu'un oubli de ligne de commande le
	// remplace.
	if d.EspaceDeNoms == "" && *espaceDeNoms != "" {
		d.EspaceDeNoms = *espaceDeNoms
	}

	logique, avertissements := inference.Inferer(physique, d)

	cheminSortie := defaut(*sortie, base+suffixeLogique)
	if err := logique.Ecrire(cheminSortie); err != nil {
		return err
	}

	rendreCompte(cheminSortie, logique, avertissements)
	return nil
}

// analyser lit les drapeaux quel que soit leur rang, et rend les arguments qui
// n'en sont pas.
//
// flag s'arrête au premier argument positionnel : sans ce tour, la forme
// naturelle — ormeau inferer gescom.calque.json --decisions f.yaml — verrait le
// drapeau traité comme un second fichier. C'est celle que la ligne d'usage
// montre, et celle que tout le monde tape.
func analyser(jeu *flag.FlagSet, args []string) ([]string, error) {
	var positionnels []string

	reste := args
	for {
		if err := jeu.Parse(reste); err != nil {
			return nil, err
		}
		if jeu.NArg() == 0 {
			return positionnels, nil
		}
		positionnels = append(positionnels, jeu.Arg(0))
		reste = jeu.Args()[1:]
	}
}

// chargerOuPreremplir lit le fichier de décisions, ou l'écrit s'il n'existe pas.
//
// Écrit une fois et jamais réécrit : le fichier porte des arbitrages que
// personne ne veut voir disparaître parce qu'une commande a été relancée. Un
// utilisateur qui veut repartir des propositions supprime le sien, ce qui est
// un geste conscient.
//
// Un chemin donné explicitement et introuvable est une erreur, pas une
// invitation à en créer un : c'est une faute de frappe bien plus souvent qu'une
// intention.
func chargerOuPreremplir(chemin string, physique *calque.Physique, explicite bool) (*inference.Decisions, error) {
	// Le chemin vient de la ligne de commande : lire et écrire le fichier que
	// l'utilisateur désigne est la fonction même de l'outil, et il a déjà les
	// droits du compte qui lance le binaire.
	//
	// Cette justification ne vaudra pas pour l'interface : là, le chemin
	// viendrait du navigateur, et un ../../.ssh/id_rsa ferait sauter le
	// confinement. C'est pourquoi l'API n'accepte qu'un nom logique.
	if _, err := os.Stat(chemin); err == nil { // #nosec G703
		return inference.LireDecisions(chemin)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	if explicite {
		return nil, fmt.Errorf("fichier de décisions introuvable : %s", chemin)
	}

	prerempli := inference.EcrireDecisions(physique, &inference.Decisions{})
	if err := os.WriteFile(chemin, prerempli, 0o600); err != nil { // #nosec G703
		return nil, err
	}

	fmt.Fprintf(os.Stderr, "%s écrit, entièrement en commentaire : rien n'est appliqué tant que\n"+
		"vous n'avez pas décommenté. Il porte les renommages que l'outil propose.\n\n", chemin)

	return &inference.Decisions{}, nil
}

// baseDuChemin rend le préfixe commun aux trois fichiers d'une base.
//
// gescom.calque.json donne gescom, et les deux autres chemins s'en déduisent.
// Un nom qui ne suit pas la convention garde son extension retirée : mieux vaut
// un calque.logique.json inattendu qu'un refus sur un nom que l'utilisateur a
// choisi.
func baseDuChemin(chemin string) string {
	if strings.HasSuffix(chemin, suffixePhysique) {
		return strings.TrimSuffix(chemin, suffixePhysique)
	}
	return strings.TrimSuffix(chemin, filepath.Ext(chemin))
}

// defaut rend la valeur donnée, ou la valeur de repli.
func defaut(valeur, repli string) string {
	if valeur != "" {
		return valeur
	}
	return repli
}

// rendreCompte affiche ce qui a été produit et ce qui reste à trancher.
//
// Les avertissements vont sur la sortie d'erreur, comme pour l'extraction : la
// sortie standard reste libre, et ils ne sont pas des échecs. Un calque logique
// partiel accompagné de vingt avertissements précis vaut mieux qu'une erreur
// fatale.
func rendreCompte(chemin string, logique *calque.Logique, avertissements []calque.Avertissement) {
	fmt.Fprintf(os.Stderr, "%s : %d entité(s), %d association(s), %d énumération(s), %d trait(s)\n",
		chemin, len(logique.Entites), compterAssociations(logique),
		len(logique.Enumerations), len(logique.Traits))

	if len(avertissements) == 0 {
		return
	}

	fmt.Fprintf(os.Stderr, "\n%d avertissement(s) :\n", len(avertissements))
	for _, a := range avertissements {
		fmt.Fprintf(os.Stderr, "  %-24s %-32s %s\n", a.Code, a.Cible, a.Message)
	}

	// Le décompte par code sert de filtre : c'est sur lui qu'on décide s'il y a
	// matière à ouvrir le fichier de décisions, sans relire deux cents lignes.
	fmt.Fprintln(os.Stderr, "\nPar code :")
	for _, code := range codesTries(avertissements) {
		fmt.Fprintf(os.Stderr, "  %-24s %d\n", code.nom, code.nombre)
	}
}

// compterAssociations totalise les associations de toutes les entités, les deux
// côtés compris.
func compterAssociations(logique *calque.Logique) int {
	var n int
	for i := range logique.Entites {
		n += len(logique.Entites[i].Associations)
	}
	return n
}

// decompte apparie un code d'avertissement à son nombre d'occurrences.
type decompte struct {
	nom    string
	nombre int
}

// codesTries rend les codes d'avertissement du plus fréquent au moins fréquent,
// puis par ordre alphabétique à égalité.
//
// L'ordre alphabétique départage pour que deux exécutions rendent la même
// liste : le décompte se lit dans un journal de CI, où une ligne qui bouge sans
// raison se remarque.
func codesTries(avertissements []calque.Avertissement) []decompte {
	parCode := map[string]int{}
	for _, a := range avertissements {
		parCode[a.Code]++
	}

	codes := make([]decompte, 0, len(parCode))
	for nom, nombre := range parCode {
		codes = append(codes, decompte{nom, nombre})
	}

	sort.SliceStable(codes, func(i, j int) bool {
		if codes[i].nombre != codes[j].nombre {
			return codes[i].nombre > codes[j].nombre
		}
		return codes[i].nom < codes[j].nom
	})
	return codes
}
