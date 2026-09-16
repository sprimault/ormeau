// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

//go:build integration

package sqlserver

import (
	"context"
	"os"
	"testing"

	"github.com/sprimault/ormeau/internal/introspection"
)

// Le conteneur est monté par `make containers`, depuis tests/ddl/sqlserver.sql.
//
// TrustServerCertificate parce que l'image se sert d'un certificat auto-signé :
// sans lui, le pilote refuse la connexion sur une erreur de certificat qui ne
// dit rien du vrai problème. C'est aussi ce qu'un utilisateur devra poser face
// à un serveur d'entreprise sans autorité reconnue.
const dsnParDefaut = "sqlserver://sa:Ormeau!2026@127.0.0.1:31433?database=gescom&TrustServerCertificate=true"

// dsnDeTest rend le DSN du conteneur, qu'ORMEAU_TEST_DSN_SQLSERVER permet de
// remplacer : la même suite doit pouvoir viser un serveur distant sans être
// recompilée, le conteneur tournant souvent sur la machine de construction.
func dsnDeTest() string {
	if dsn := os.Getenv("ORMEAU_TEST_DSN_SQLSERVER"); dsn != "" {
		return dsn
	}
	return dsnParDefaut
}

// ouvrir rend un pilote connecté, ou saute le test quand le conteneur manque.
func ouvrir(t *testing.T) introspection.Introspecteur {
	t.Helper()

	i, err := Ouvrir(context.Background(), dsnDeTest())
	if err != nil {
		t.Fatalf("connexion au conteneur : %v", err)
	}
	t.Cleanup(func() { _ = i.Fermer() })
	return i
}

// TestDecrireRendLeServeurAtteint : la version, le catalogue et les schémas
// sont ce que l'écran de connexion affiche en retour.
func TestDecrireRendLeServeurAtteint(t *testing.T) {
	d, ok := ouvrir(t).(introspection.DescripteurServeur)
	if !ok {
		t.Fatal("le pilote ne sait pas se décrire")
	}

	s, err := d.Decrire(context.Background())
	if err != nil {
		t.Fatalf("Decrire : %v", err)
	}
	if s.SGBD != "sqlserver" {
		t.Errorf("sgbd = %q, attendu sqlserver", s.SGBD)
	}
	if s.Catalogue != "gescom" {
		t.Errorf("catalogue = %q, attendu gescom", s.Catalogue)
	}
	if s.Version == "" {
		t.Error("version vide")
	}
	// Le DDL de test ne crée qu'un schéma, et dbo n'y porte aucune table.
	if len(s.Schemas) != 1 || s.Schemas[0] != "ventes" {
		t.Errorf("schemas = %v, attendu [ventes]", s.Schemas)
	}
}

// TestInventorierRendLesTablesEtLeursLiens : l'arbre de sélection vit de cette
// passe, commentaires et dépendances compris.
func TestInventorierRendLesTablesEtLeursLiens(t *testing.T) {
	i := ouvrir(t)

	tables, err := i.Inventorier(context.Background(), []string{"ventes"})
	if err != nil {
		t.Fatalf("Inventorier : %v", err)
	}
	if len(tables) == 0 {
		t.Fatal("aucune table")
	}

	par := map[string]introspection.TableSommaire{}
	for _, table := range tables {
		par[table.Nom] = table
	}

	// Commentaire d'extended property : SQL Server n'a pas de COMMENT ON, et
	// une lecture qui les raterait laisserait les entités sans documentation.
	if c := par["t_commercial"].Commentaire; c != "Force de vente" {
		t.Errorf("commentaire de t_commercial = %q", c)
	}
	// Table sans clé primaire : le cas doit remonter tel quel, pas disparaître.
	if par["t_log_import"].ClePrimaire {
		t.Error("t_log_import n'a pas de clé primaire")
	}
	if !par["t_client"].ClePrimaire {
		t.Error("t_client a une clé primaire")
	}
	// Auto-référence : la table se cite elle-même dans ses dépendances.
	if refs := par["t_categorie"].ReferenceVers; len(refs) != 1 || refs[0] != "ventes.t_categorie" {
		t.Errorf("references de t_categorie = %v", refs)
	}
	// Deux clés étrangères vers deux tables : l'interface propose les deux.
	if refs := par["t_client_tag"].ReferenceVers; len(refs) != 2 {
		t.Errorf("references de t_client_tag = %v, attendu deux", refs)
	}
}

// TestInventorierFiltreParSchema : le filtre se fait dans le pilote, faute de
// tableau paramétrable sous SQL Server.
func TestInventorierFiltreParSchema(t *testing.T) {
	i := ouvrir(t)

	tables, err := i.Inventorier(context.Background(), []string{"schema_absent"})
	if err != nil {
		t.Fatalf("Inventorier : %v", err)
	}
	if len(tables) != 0 {
		t.Errorf("%d tables pour un schéma qui n'existe pas", len(tables))
	}
}

// TestColonnesRendLesTypesReconstruits : max_length est en octets, et une
// longueur lue telle quelle donnerait nvarchar(60) pour un nvarchar(30).
func TestColonnesRendLesTypesReconstruits(t *testing.T) {
	l, ok := ouvrir(t).(introspection.ListeurDeColonnes)
	if !ok {
		t.Fatal("le pilote ne sait pas lister les colonnes")
	}

	colonnes, err := l.Colonnes(context.Background(), "ventes", "users")
	if err != nil {
		t.Fatalf("Colonnes : %v", err)
	}

	par := map[string]introspection.ColonneSommaire{}
	for _, c := range colonnes {
		par[c.Nom] = c
	}

	attendus := map[string]string{
		"nom":          "nvarchar(30)",
		"session_id":   "varchar(255)",
		"PasswordHash": "binary(64)",
		"Salt":         "uniqueidentifier",
		"mem_montant":  "bit",
		"nivhab":       "smallint",
	}
	for nom, attendu := range attendus {
		if obtenu := par[nom].TypeBrut; obtenu != attendu {
			t.Errorf("type de %s = %q, attendu %q", nom, obtenu, attendu)
		}
	}

	if !par["id_user"].ClePrimaire {
		t.Error("id_user est la clé primaire")
	}
	if par["nom"].ClePrimaire {
		t.Error("nom n'est pas dans la clé primaire")
	}
	if !par["nom"].Nullable || par["id_user"].Nullable {
		t.Error("nullabilité mal lue")
	}
	// L'ordre est celui de la table, pas alphabétique.
	if colonnes[0].Nom != "id_user" || colonnes[0].Position != 1 {
		t.Errorf("première colonne = %+v", colonnes[0])
	}
}

// TestColonnesCiteLesNomsPenibles : un nom à espaces ou à accents passe par un
// paramètre, jamais par une concaténation.
func TestColonnesCiteLesNomsPenibles(t *testing.T) {
	l, ok := ouvrir(t).(introspection.ListeurDeColonnes)
	if !ok {
		t.Fatal("le pilote ne sait pas lister les colonnes")
	}

	colonnes, err := l.Colonnes(context.Background(), "ventes", "t_référence")
	if err != nil {
		t.Fatalf("Colonnes : %v", err)
	}

	noms := []string{}
	for _, c := range colonnes {
		noms = append(noms, c.Nom)
	}
	for _, attendu := range []string{"order", "select", "Libellé", "reprise an"} {
		if !contient(noms, attendu) {
			t.Errorf("colonne %q absente de %v", attendu, noms)
		}
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
