// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

//go:build integration

package postgres

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/sprimault/ormeau/internal/introspection"
	"github.com/sprimault/ormeau/internal/introspection/ddltest"
)

// Le conteneur est monté par `make containers`, depuis tests/ddl/postgres.sql.
//
// Le DSN se surcharge : le conteneur peut tourner ailleurs que sur le poste —
// sur la machine de construction, ou comme service de CI. Un test qui code son
// hôte en dur ne tourne que chez celui qui l'a écrit.
const dsnParDefaut = "postgres://postgres:ormeau@127.0.0.1:35432/gescom"

// dsnDeTest rend le DSN du conteneur, qu'ORMEAU_TEST_DSN permet de remplacer :
// la même suite doit pouvoir viser un serveur distant sans être recompilée.
func dsnDeTest() string {
	if dsn := os.Getenv("ORMEAU_TEST_DSN"); dsn != "" {
		return dsn
	}
	return dsnParDefaut
}

// cheminDDL est le DDL dont la base de test doit être issue.
const cheminDDL = "../../../tests/ddl/postgres.sql"

// TestMain refuse de lancer la suite contre une base qui ne vient pas du DDL
// du dépôt. Deux cas passeraient sinon sans rien signaler : un conteneur dont
// le volume a survécu à une modification du DDL, qui fait tester l'ancien
// schéma, et un ORMEAU_TEST_DSN qui vise une autre base que celle du
// conteneur.
func TestMain(m *testing.M) {
	if err := controlerBaseDeTest(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

// controlerBaseDeTest compare l'empreinte du DDL du dépôt à celle que
// tests/ddl/20-empreinte.sh a posée en commentaire de la base à sa création.
func controlerBaseDeTest() error {
	ctx, annuler := context.WithTimeout(context.Background(), 10*time.Second)
	defer annuler()

	i, err := Ouvrir(ctx, dsnDeTest())
	if err != nil {
		return fmt.Errorf("connexion a la base de test (make containers ?): %w", err)
	}
	defer func() { _ = i.Fermer() }()

	var commentaire *string
	err = i.(*pilote).conn.QueryRow(ctx, `
SELECT shobj_description(oid, 'pg_database')
FROM pg_database
WHERE datname = current_database()`).Scan(&commentaire)
	if err != nil {
		return fmt.Errorf("lecture de l'empreinte du DDL: %w", err)
	}
	return ddltest.Controler(commentaire, cheminDDL)
}

// ouvrirOuEchouer ouvre une connexion et l'inscrit au nettoyage du test.
func ouvrirOuEchouer(t *testing.T) introspection.Introspecteur {
	t.Helper()
	return ouvrirDepuis(t, dsnDeTest())
}

// ouvrirDepuis ouvre une connexion sur un DSN donné, pour les tests qui
// comparent deux sessions.
func ouvrirDepuis(t *testing.T, dsn string) introspection.Introspecteur {
	t.Helper()

	ctx, annuler := context.WithTimeout(context.Background(), 10*time.Second)
	defer annuler()

	p, err := introspection.Ouvrir(ctx, "postgres", dsn)
	if err != nil {
		t.Fatalf("connexion (make containers ?) : %v", err)
	}
	t.Cleanup(func() {
		if err := p.Fermer(); err != nil {
			t.Errorf("fermeture : %v", err)
		}
	})
	return p
}

// TestOuvrirEtFermer vérifie le cycle de vie de la connexion, nettoyage compris.
func TestOuvrirEtFermer(t *testing.T) {
	ouvrirOuEchouer(t)
}

// L'invariant le plus important du projet : aucune écriture dans la base
// introspectée. Il tient par le serveur, pas par la discipline du code, et
// c'est ce test qui le prouve.
func TestConnexionEnLectureSeule(t *testing.T) {
	p := ouvrirOuEchouer(t)

	ctx, annuler := context.WithTimeout(context.Background(), 10*time.Second)
	defer annuler()

	// Passe par l'interface : le test n'a pas à connaître le type concret, mais
	// il doit atteindre la connexion. On force donc une écriture par le seul
	// chemin disponible, une requête d'inventaire sur un schéma inexistant ne
	// suffirait pas à prouver quoi que ce soit.
	concret, ok := p.(*pilote)
	if !ok {
		t.Fatalf("type de pilote inattendu : %T", p)
	}

	_, err := concret.conn.Exec(ctx, "CREATE TABLE ormeau_ne_doit_pas_exister (id int)")
	if err == nil {
		// Nettoyage de principe : si l'invariant est cassé, ne pas laisser la
		// table derrière soi.
		_, _ = concret.conn.Exec(ctx, "DROP TABLE IF EXISTS ormeau_ne_doit_pas_exister")
		t.Fatal("une écriture a réussi : la session n'est pas en lecture seule")
	}
}

// TestInventorier vérifie l'arbre servant à l'interface : les tables du schéma
// de test y figurent, celles des schémas système non.
func TestInventorier(t *testing.T) {
	p := ouvrirOuEchouer(t)

	ctx, annuler := context.WithTimeout(context.Background(), 15*time.Second)
	defer annuler()

	sommaires, err := p.Inventorier(ctx, []string{"gescom"})
	if err != nil {
		t.Fatalf("inventaire : %v", err)
	}
	if len(sommaires) == 0 {
		t.Fatal("aucune table inventoriée")
	}

	parNom := make(map[string]introspection.TableSommaire, len(sommaires))
	for _, s := range sommaires {
		if s.Schema != "gescom" {
			t.Errorf("table hors du schéma demandé : %s.%s", s.Schema, s.Nom)
		}
		parNom[s.Nom] = s
	}

	// Les cas du DDL de référence qui décident du contenu de l'inventaire.
	cas := []struct {
		table       string
		clePrimaire bool
		reference   string
	}{
		{"t_commercial", true, ""},
		{"t_client", true, "gescom.t_commercial"},
		{"t_client_tag", true, "gescom.t_client"},
		{"t_client_tag", true, "gescom.t_tag"}, // jointure pure
		{"t_log_import", false, ""},            // aucune clé primaire, cas courant sur du legacy
		{"t_facture", true, ""},                // clé étrangère implicite : rien de déclaré
		{"t_référence", true, ""},              // identifiants accentués et réservés
	}

	for _, c := range cas {
		s, present := parNom[c.table]
		if !present {
			t.Errorf("%s absente de l'inventaire", c.table)
			continue
		}
		if s.ClePrimaire != c.clePrimaire {
			t.Errorf("%s : clé primaire %v, attendu %v", c.table, s.ClePrimaire, c.clePrimaire)
		}
		if s.NbColonnes < 1 {
			t.Errorf("%s : %d colonnes", c.table, s.NbColonnes)
		}
		if c.reference != "" && !contient(s.ReferenceVers, c.reference) {
			t.Errorf("%s ne référence pas %s : %v", c.table, c.reference, s.ReferenceVers)
		}
	}

	// La vue n'est pas une table : elle n'a rien à faire dans l'inventaire.
	if _, present := parNom["v_client_actif"]; present {
		t.Error("la vue v_client_actif apparaît dans l'inventaire des tables")
	}
	// Le commentaire du catalogue doit remonter.
	if parNom["t_commercial"].Commentaire != "Force de vente" {
		t.Errorf("commentaire de t_commercial : %q", parNom["t_commercial"].Commentaire)
	}
}

// L'ordre ne doit dépendre ni du planificateur ni de l'ordonnancement : deux
// inventaires successifs rendent la même liste.
func TestInventorierEstDeterministe(t *testing.T) {
	p := ouvrirOuEchouer(t)

	ctx, annuler := context.WithTimeout(context.Background(), 20*time.Second)
	defer annuler()

	premier, err := p.Inventorier(ctx, []string{"gescom"})
	if err != nil {
		t.Fatalf("premier inventaire : %v", err)
	}
	for i := 0; i < 3; i++ {
		second, err := p.Inventorier(ctx, []string{"gescom"})
		if err != nil {
			t.Fatalf("inventaire %d : %v", i, err)
		}
		if len(second) != len(premier) {
			t.Fatalf("%d tables puis %d", len(premier), len(second))
		}
		for j := range premier {
			if second[j].Schema != premier[j].Schema || second[j].Nom != premier[j].Nom {
				t.Fatalf("ordre instable au rang %d : %s.%s puis %s.%s", j,
					premier[j].Schema, premier[j].Nom, second[j].Schema, second[j].Nom)
			}
		}
	}
}

// Énumérer les bases est ce qui permet d'extraire un serveur entier sans
// redemander les identifiants pour chacune.
func TestListerBases(t *testing.T) {
	p := ouvrirOuEchouer(t)

	listeur, ok := p.(introspection.ListeurDeBases)
	if !ok {
		t.Fatal("le pilote postgres n'implemente pas ListeurDeBases")
	}

	ctx, annuler := context.WithTimeout(context.Background(), 10*time.Second)
	defer annuler()

	bases, err := listeur.ListerBases(ctx)
	if err != nil {
		t.Fatalf("liste des bases : %v", err)
	}
	if !contient(bases, "gescom") {
		t.Errorf("gescom absente : %v", bases)
	}

	// Les bases système ne produiraient que des calques sans intérêt, et
	// template0 refuse même la connexion.
	for _, systeme := range []string{"postgres", "template0", "template1"} {
		if contient(bases, systeme) {
			t.Errorf("la base systeme %s est listee", systeme)
		}
	}
}

// Decrire alimente l'écran de connexion : le serveur atteint, sa version, et
// les schémas parmi lesquels choisir.
func TestDecrire(t *testing.T) {
	p := ouvrirOuEchouer(t)

	descripteur, ok := p.(introspection.DescripteurServeur)
	if !ok {
		t.Fatal("le pilote postgres n'implemente pas DescripteurServeur")
	}

	ctx, annuler := context.WithTimeout(context.Background(), 10*time.Second)
	defer annuler()

	serveur, err := descripteur.Decrire(ctx)
	if err != nil {
		t.Fatalf("description : %v", err)
	}

	if serveur.SGBD != "postgres" {
		t.Errorf("sgbd %q, attendu \"postgres\"", serveur.SGBD)
	}
	if serveur.Catalogue != "gescom" {
		t.Errorf("catalogue %q, attendu \"gescom\"", serveur.Catalogue)
	}
	if serveur.Version == "" {
		t.Error("version vide : l'ecran de connexion n'a rien a afficher")
	}
	if !contient(serveur.Schemas, "public") {
		t.Errorf("public absent des schemas : %v", serveur.Schemas)
	}

	// Les schémas système n'ont rien à faire dans une liste où l'on choisit ce
	// qu'on va introspecter.
	for _, systeme := range []string{"pg_catalog", "information_schema", "pg_toast"} {
		if contient(serveur.Schemas, systeme) {
			t.Errorf("le schema systeme %s est propose", systeme)
		}
	}
}

// Colonnes décrit une table pour l'écran de sélection : assez pour décider si
// on mappe une colonne, et rien de plus.
func TestColonnes(t *testing.T) {
	p := ouvrirOuEchouer(t)

	listeur, ok := p.(introspection.ListeurDeColonnes)
	if !ok {
		t.Fatal("le pilote postgres n'implemente pas ListeurDeColonnes")
	}

	ctx, annuler := context.WithTimeout(context.Background(), 10*time.Second)
	defer annuler()

	colonnes, err := listeur.Colonnes(ctx, "gescom", "t_client")
	if err != nil {
		t.Fatalf("colonnes : %v", err)
	}
	if len(colonnes) == 0 {
		t.Fatal("aucune colonne rendue")
	}

	// L'ordre est celui du catalogue : le front affiche la table comme elle est
	// déclarée, et non par ordre alphabétique.
	for i, c := range colonnes {
		if c.Position != i+1 {
			t.Errorf("colonne %s en position %d, attendue %d", c.Nom, c.Position, i+1)
		}
		if c.TypeBrut == "" {
			t.Errorf("colonne %s sans type brut : c'est ce qui distingue un citext d'un text", c.Nom)
		}
	}

	var identifiantes int
	for _, c := range colonnes {
		if c.ClePrimaire {
			identifiantes++
		}
	}
	if identifiantes == 0 {
		t.Error("aucune clé primaire signalée : le front les laisserait décocher")
	}
}

// Une table inconnue rend une liste vide, pas une erreur : c'est une table
// disparue depuis l'inventaire, pas une panne.
func TestColonnesTableInconnue(t *testing.T) {
	p := ouvrirOuEchouer(t)

	listeur, ok := p.(introspection.ListeurDeColonnes)
	if !ok {
		t.Fatal("le pilote postgres n'implemente pas ListeurDeColonnes")
	}

	ctx, annuler := context.WithTimeout(context.Background(), 10*time.Second)
	defer annuler()

	colonnes, err := listeur.Colonnes(ctx, "gescom", "table_qui_n_existe_pas")
	if err != nil {
		t.Fatalf("colonnes : %v", err)
	}
	if len(colonnes) != 0 {
		t.Errorf("%d colonne(s) rendues pour une table inconnue", len(colonnes))
	}
}

// Le DSN ne doit ressortir d'aucune erreur du pilote, pas même masqué.
//
// L'absence du secret dans l'erreur ne prouve rien s'il n'a pas été soumis :
// un serveur injoignable refuse la connexion avant tout échange, et le test
// passerait. D'où les deux préconditions — le vrai DSN se connecte, et le
// refus sous le faux mot de passe est bien celui de l'authentification.
func TestConnexionRefuseeSansFuiteDuSecret(t *testing.T) {
	_ = ouvrirOuEchouer(t)

	const secret = "Mot-De-Passe-Qui-Ne-Doit-Pas-Fuiter"
	u, err := url.Parse(dsnDeTest())
	if err != nil || u.User == nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
		t.Fatal("ce test exige un DSN de test en URL postgres:// qui nomme un utilisateur")
	}
	u.User = url.UserPassword(u.User.Username(), secret)

	ctx, annuler := context.WithTimeout(context.Background(), 10*time.Second)
	defer annuler()

	_, err = introspection.Ouvrir(ctx, "postgres", u.String())
	if err == nil {
		t.Fatal("connexion acceptee avec un mot de passe faux")
	}
	var refus *pgconn.PgError
	if !errors.As(err, &refus) || refus.Code != "28P01" {
		t.Fatalf("refus d'authentification (28P01) attendu, le mot de passe n'a pas ete soumis : %v", err)
	}
	if strings.Contains(err.Error(), secret) {
		t.Error("le mot de passe apparait dans l'erreur de connexion")
	}
}

// Un schéma qui n'existe pas rend une liste vide, pas une erreur : c'est une
// portée sans table, pas une panne.
func TestInventorierSchemaInexistant(t *testing.T) {
	p := ouvrirOuEchouer(t)

	ctx, annuler := context.WithTimeout(context.Background(), 10*time.Second)
	defer annuler()

	sommaires, err := p.Inventorier(ctx, []string{"schema_qui_n_existe_pas"})
	if err != nil {
		t.Fatalf("inventaire : %v", err)
	}
	if len(sommaires) != 0 {
		t.Errorf("%d tables rendues pour un schéma inexistant", len(sommaires))
	}
}

// contient dit si la liste porte la valeur.
func contient(liste []string, valeur string) bool {
	for _, v := range liste {
		if v == valeur {
			return true
		}
	}
	return false
}
