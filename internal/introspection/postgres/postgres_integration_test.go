// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

//go:build integration

package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/sprimault/ormeau/internal/introspection"
)

// Le conteneur est monté par `make containers`, depuis tests/ddl/postgres.sql.
//
// Le DSN se surcharge : le conteneur peut tourner ailleurs que sur le poste —
// sur la machine de construction, ou comme service de CI. Un test qui code son
// hôte en dur ne tourne que chez celui qui l'a écrit.
const dsnParDefaut = "postgres://postgres:ormeau@127.0.0.1:35432/gescom"

func dsnDeTest() string {
	if dsn := os.Getenv("ORMEAU_TEST_DSN"); dsn != "" {
		return dsn
	}
	return dsnParDefaut
}

func ouvrirOuEchouer(t *testing.T) introspection.Introspecteur {
	t.Helper()

	ctx, annuler := context.WithTimeout(context.Background(), 10*time.Second)
	defer annuler()

	p, err := introspection.Ouvrir(ctx, "postgres", dsnDeTest())
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
		{"t_log_import", false, ""}, // aucune clé primaire, cas courant sur du legacy
		{"t_facture", true, ""},     // clé étrangère implicite : rien de déclaré
		{"t_référence", true, ""},   // identifiants accentués et réservés
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

func contient(liste []string, valeur string) bool {
	for _, v := range liste {
		if v == valeur {
			return true
		}
	}
	return false
}
