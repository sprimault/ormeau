// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
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
	// ContinueOnError et non ExitOnError : la fonction rend déjà une erreur,
	// que main traduit en code de retour. Un os.Exit caché ici court-circuiterait
	// la fermeture de la connexion.
	jeu := flag.NewFlagSet("extraire", flag.ContinueOnError)
	dsn := jeu.String("dsn", "", "chaîne de connexion complète (ou ORMEAU_DSN)")
	sgbd := jeu.String("sgbd", "", "postgres, mysql, mariadb, sqlserver, oracle")
	hote := jeu.String("hote", "", "nom ou adresse du serveur")
	port := jeu.Int("port", 0, "port du serveur ; celui du SGBD par défaut")
	utilisateur := jeu.String("utilisateur", "", "identifiant de connexion")
	base := jeu.String("base", "", "base à introspecter ; toutes celles du serveur si omise")
	sortie := jeu.String("sortie", "", "fichier de sortie, ou répertoire quand plusieurs bases sont extraites")
	schemas := jeu.String("schemas", "", "schémas à introspecter, séparés par des virgules")
	echantillonner := jeu.Bool("echantillonner", false, "lire des données pour détecter énumérations et clés étrangères implicites")
	cardinalite := jeu.Int("cardinalite-max", 64, "plafond au-delà duquel une colonne ne produit plus d'échantillon")
	if err := jeu.Parse(args); err != nil {
		return err
	}

	chaine, err := resoudreDSN(*dsn, introspection.Connexion{
		SGBD:        *sgbd,
		Hote:        *hote,
		Port:        *port,
		Utilisateur: *utilisateur,
		// Jamais un drapeau : il serait visible dans ps et dans l'historique.
		MotDePasse: os.Getenv("ORMEAU_MDP"),
		Base:       *base,
	})
	if err != nil {
		return err
	}
	if *sortie == "" {
		return errors.New("--sortie est requis")
	}

	nomSGBD, err := introspection.SGBDDepuisDSN(chaine)
	if err != nil {
		return err
	}

	ctx, annuler := context.WithTimeout(context.Background(), delaiExtraction)
	defer annuler()

	portee := introspection.Portee{
		Schemas:        decouper(*schemas),
		Echantillonner: *echantillonner,
		CardinaliteMax: *cardinalite,
	}

	if introspection.BaseDuDSN(chaine) != "" {
		return extraireUneBase(ctx, nomSGBD, chaine, *sortie, portee)
	}
	return extraireLeServeur(ctx, nomSGBD, chaine, *sortie, portee)
}

// resoudreDSN choisit entre le DSN complet et les drapeaux séparés.
//
// Le DSN l'emporte quand les deux sont donnés : c'est la forme la plus
// précise, et refuser la combinaison obligerait à défaire un ORMEAU_DSN
// d'environnement pour une seule exécution.
func resoudreDSN(dsn string, c introspection.Connexion) (string, error) {
	if dsn != "" {
		return dsn, nil
	}
	if c.SGBD != "" || c.Hote != "" {
		return c.DSN()
	}
	if depuisEnv := os.Getenv("ORMEAU_DSN"); depuisEnv != "" {
		return depuisEnv, nil
	}
	return "", errors.New("--dsn, ORMEAU_DSN ou les drapeaux --sgbd et --hote sont requis")
}

func extraireUneBase(ctx context.Context, sgbd, dsn, sortie string, portee introspection.Portee) error {
	pilote, err := introspection.Ouvrir(ctx, sgbd, dsn)
	if err != nil {
		return err
	}
	defer fermer(pilote)

	physique, err := pilote.Extraire(ctx, portee)
	if err != nil {
		return err
	}
	return ecrire(physique, sortie)
}

// extraireLeServeur produit un calque par base. Un calque décrit une base et
// une seule — source.catalogue est une chaîne — donc --sortie désigne ici un
// répertoire et non un fichier.
//
// Une base qui échoue n'interrompt pas les autres : sur un serveur partagé,
// une seule base sans droit de lecture rendrait la commande inutilisable.
func extraireLeServeur(ctx context.Context, sgbd, dsn, repertoire string, portee introspection.Portee) error {
	// L'énumération passe par une base d'administration : on ne peut pas se
	// connecter à un serveur sans nommer une base.
	pilote, err := introspection.Ouvrir(ctx, sgbd, introspection.AvecBase(dsn, baseAdministration(sgbd)))
	if err != nil {
		return err
	}

	listeur, ok := pilote.(introspection.ListeurDeBases)
	if !ok {
		fermer(pilote)
		return fmt.Errorf("le pilote %s ne sait pas enumerer les bases, nommer une base avec --base", sgbd)
	}

	bases, err := listeur.ListerBases(ctx)
	fermer(pilote)
	if err != nil {
		return err
	}
	if len(bases) == 0 {
		return errors.New("aucune base exploitable sur ce serveur")
	}

	if err := os.MkdirAll(repertoire, 0o750); err != nil {
		return err
	}

	var echecs int
	for _, nom := range bases {
		chemin := filepath.Join(repertoire, nom+".calque.json")
		if err := extraireUneBase(ctx, sgbd, introspection.AvecBase(dsn, nom), chemin, portee); err != nil {
			fmt.Fprintf(os.Stderr, "%s : %v\n", nom, err)
			echecs++
		}
	}
	if echecs == len(bases) {
		return fmt.Errorf("aucune des %d bases n'a pu etre extraite", len(bases))
	}
	if echecs > 0 {
		return fmt.Errorf("%d base(s) sur %d en echec", echecs, len(bases))
	}
	return nil
}

func ecrire(physique *calque.Physique, chemin string) error {
	// Les anomalies vont sur la sortie d'erreur : la sortie standard peut
	// porter autre chose, et elles ne sont pas fatales.
	anomalies := physique.Valider()
	for _, a := range anomalies {
		fmt.Fprintf(os.Stderr, "%s  %s : %s\n", a.Code, a.Cible, a.Message)
	}

	// Posé au dernier moment, et exclu de l'empreinte : deux extractions de la
	// même base restent identiques.
	physique.Source.ExtraitLe = time.Now().UTC().Format(time.RFC3339)
	if err := physique.Ecrire(chemin); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "%s : %d table(s), %d colonne(s), %d anomalie(s)\n",
		chemin, len(physique.Tables), compterColonnes(physique), len(anomalies))
	fmt.Fprintf(os.Stderr, "empreinte %s\n", physique.Source.Empreinte)
	return nil
}

func fermer(pilote introspection.Introspecteur) {
	if err := pilote.Fermer(); err != nil {
		fmt.Fprintf(os.Stderr, "fermeture de la connexion : %v\n", err)
	}
}

// baseAdministration est la base à laquelle se connecter pour en énumérer
// d'autres. Elle existe sur toute installation.
func baseAdministration(sgbd string) string {
	switch sgbd {
	case "postgres":
		return "postgres"
	case "sqlserver":
		return "master"
	default:
		return ""
	}
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
