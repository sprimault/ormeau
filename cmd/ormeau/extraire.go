// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sprimault/ormeau/internal/calque"
	"github.com/sprimault/ormeau/internal/introspection"

	// Chaque pilote s'enregistre par son init : l'importer ici est ce qui le
	// rend disponible, sans que le code commun ait à le connaître.
	_ "github.com/sprimault/ormeau/internal/introspection/postgres"
)

// delaiExtraction plafonne l'opération entière. Généreux : quatre cents tables
// sur un serveur chargé prennent du temps, et abandonner à mi-parcours ne rend
// service à personne.
const delaiExtraction = 10 * time.Minute

// extraire lit le catalogue d'une base et écrit un calque physique.
//
// La validation des drapeaux est ici : aucun paquet interne ne lit flag ni
// os.Getenv, les dépendances y sont injectées.
func extraire(args []string) error {
	jeu := flag.NewFlagSet("extraire", flag.ExitOnError)
	dsn := jeu.String("dsn", "", "chaîne de connexion, préfixée par le SGBD (ou ORMEAU_DSN)")
	sortie := jeu.String("sortie", "", "fichier de sortie")
	schemas := jeu.String("schemas", "", "schémas à introspecter, séparés par des virgules")
	echantillonner := jeu.Bool("echantillonner", false, "lire des données pour détecter énumérations et clés étrangères implicites")
	cardinalite := jeu.Int("cardinalite-max", 64, "plafond au-delà duquel une colonne ne produit plus d'échantillon")
	if err := jeu.Parse(args); err != nil {
		return err
	}

	// La variable d'environnement évite de laisser le mot de passe dans
	// l'historique du shell et dans ps.
	chaine := *dsn
	if chaine == "" {
		chaine = os.Getenv("ORMEAU_DSN")
	}
	if chaine == "" {
		return errors.New("--dsn ou ORMEAU_DSN est requis")
	}
	if *sortie == "" {
		return errors.New("--sortie est requis")
	}

	sgbd, err := introspection.SGBDDepuisDSN(chaine)
	if err != nil {
		return err
	}

	ctx, annuler := context.WithTimeout(context.Background(), delaiExtraction)
	defer annuler()

	pilote, err := introspection.Ouvrir(ctx, sgbd, chaine)
	if err != nil {
		return err
	}
	defer func() {
		if err := pilote.Fermer(); err != nil {
			fmt.Fprintf(os.Stderr, "fermeture de la connexion : %v\n", err)
		}
	}()

	physique, err := pilote.Extraire(ctx, introspection.Portee{
		Schemas:        decouper(*schemas),
		Echantillonner: *echantillonner,
		CardinaliteMax: *cardinalite,
	})
	if err != nil {
		return err
	}

	// Les anomalies vont sur la sortie d'erreur : la sortie standard peut
	// porter autre chose, et elles ne sont pas fatales — un calque partiel
	// mais annoncé vaut mieux qu'un échec.
	anomalies := physique.Valider()
	for _, a := range anomalies {
		fmt.Fprintf(os.Stderr, "%s  %s : %s\n", a.Code, a.Cible, a.Message)
	}

	// Posé au dernier moment, et exclu de l'empreinte : deux extractions de la
	// même base restent identiques.
	physique.Source.ExtraitLe = time.Now().UTC().Format(time.RFC3339)
	if err := physique.Ecrire(*sortie); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "%s : %d table(s), %d colonne(s), %d anomalie(s)\n",
		*sortie, len(physique.Tables), compterColonnes(physique), len(anomalies))
	fmt.Fprintf(os.Stderr, "empreinte %s\n", physique.Source.Empreinte)
	return nil
}

// decouper rend les schémas d'une liste séparée par des virgules. Une entrée
// vide n'est pas un schéma : « public, » ne demande pas deux schémas.
func decouper(liste string) []string {
	if strings.TrimSpace(liste) == "" {
		return nil
	}

	var schemas []string
	for _, brut := range strings.Split(liste, ",") {
		if nom := strings.TrimSpace(brut); nom != "" {
			schemas = append(schemas, nom)
		}
	}
	return schemas
}

func compterColonnes(p *calque.Physique) int {
	var n int
	for i := range p.Tables {
		n += len(p.Tables[i].Colonnes)
	}
	return n
}
