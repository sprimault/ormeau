// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Package sqlserver introspecte Microsoft SQL Server via sys.*.
//
// Les vues normalisées d'INFORMATION_SCHEMA perdent ce qui fait l'intérêt d'une
// reprise : la distinction IDENTITY contre séquence, les colonnes calculées et
// leur persistance, les index filtrés, et les commentaires, que SQL Server range
// dans des extended properties plutôt que dans le catalogue des colonnes.
package sqlserver

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/microsoft/go-mssqldb"

	"github.com/sprimault/ormeau/internal/introspection"
)

// Importer ce paquet suffit à rendre « sqlserver » utilisable.
func init() {
	introspection.Enregistrer("sqlserver", Ouvrir)
}

// delaiRequete plafonne chaque requête de catalogue, comme pour les autres
// pilotes : une introspection lancée contre une base de production ne doit pas
// pouvoir s'y installer.
const delaiRequete = 30 * time.Second

// pilote détient la connexion. Non exporté : Ouvrir rend l'interface, et un
// pilote sans connexion ne doit pas pouvoir exister.
type pilote struct {
	db *sql.DB
}

// Ouvrir établit la connexion.
//
// Contrairement à PostgreSQL, la lecture seule ne se pose pas sur la session :
// SQL Server n'a pas d'équivalent à default_transaction_read_only, et
// ApplicationIntent=ReadOnly ne concerne que les réplicas d'un groupe de
// disponibilité — sur un serveur ordinaire, il est accepté puis ignoré. Poser
// un réglage qui ne garantit rien donnerait une fausse assurance.
//
// L'invariant tient donc par le code : toutes les requêtes sont dans
// requetes.go, et un test les parcourt pour refuser tout verbe d'écriture. La
// vraie protection reste, côté utilisateur, un compte en lecture seule, ce que
// la documentation recommande.
//
// Le DSN ne ressort jamais tel quel d'une erreur, seulement masqué.
func Ouvrir(ctx context.Context, dsn string) (introspection.Introspecteur, error) {
	db, err := sql.Open("sqlserver", introspection.NettoyerDSN(dsn))
	if err != nil {
		// L'erreur du pilote peut recopier la chaîne de connexion : elle n'est
		// ni enveloppée ni citée, et le DSN masqué montre la valeur fautive.
		return nil, fmt.Errorf("dsn illisible (%s)", introspection.Masquer(dsn))
	}

	// Une seule connexion : l'introspection est séquentielle, et un pool
	// ouvrirait des sessions supplémentaires sur une base de production sans
	// rien accélérer.
	db.SetMaxOpenConns(1)

	ctxPing, annuler := context.WithTimeout(ctx, delaiRequete)
	defer annuler()
	if err := db.PingContext(ctxPing); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("connexion a %s: %w", introspection.Masquer(dsn), err)
	}
	return &pilote{db: db}, nil
}

// Fermer libère la connexion.
func (p *pilote) Fermer() error {
	return p.db.Close()
}

// Decrire rend ce qu'on apprend du serveur en s'y connectant : sa version, le
// catalogue courant et les schémas parmi lesquels choisir.
func (p *pilote) Decrire(ctx context.Context) (introspection.Serveur, error) {
	ctx, annuler := context.WithTimeout(ctx, delaiRequete)
	defer annuler()

	var version, edition, catalogue string
	if err := p.db.QueryRowContext(ctx, requeteServeur).Scan(&version, &edition, &catalogue); err != nil {
		return introspection.Serveur{}, fmt.Errorf("lecture des proprietes du serveur: %w", err)
	}

	schemas, err := p.schemas(ctx)
	if err != nil {
		return introspection.Serveur{}, err
	}

	return introspection.Serveur{
		SGBD: "sqlserver",
		// L'édition figure à côté de la version : Express plafonne la taille des
		// bases et Developer interdit la production, deux choses qu'on veut
		// lire avant de lancer une extraction chez un client.
		Version:   strings.TrimSpace(version + " " + edition),
		Catalogue: catalogue,
		Schemas:   schemas,
	}, nil
}

// schemas rend les schémas qui portent au moins une table.
func (p *pilote) schemas(ctx context.Context) ([]string, error) {
	lignes, err := p.db.QueryContext(ctx, requeteSchemas)
	if err != nil {
		return nil, fmt.Errorf("lecture des schemas: %w", err)
	}
	defer func() { _ = lignes.Close() }()

	schemas := []string{}
	for lignes.Next() {
		var nom string
		if err := lignes.Scan(&nom); err != nil {
			return nil, fmt.Errorf("lecture d'un schema: %w", err)
		}
		schemas = append(schemas, nom)
	}
	return schemas, lignes.Err()
}

// ListerBases rend les bases exploitables du serveur, celles du système
// exclues.
func (p *pilote) ListerBases(ctx context.Context) ([]string, error) {
	ctx, annuler := context.WithTimeout(ctx, delaiRequete)
	defer annuler()

	lignes, err := p.db.QueryContext(ctx, requeteBases)
	if err != nil {
		return nil, fmt.Errorf("lecture des bases: %w", err)
	}
	defer func() { _ = lignes.Close() }()

	bases := []string{}
	for lignes.Next() {
		var nom string
		if err := lignes.Scan(&nom); err != nil {
			return nil, fmt.Errorf("lecture d'une base: %w", err)
		}
		bases = append(bases, nom)
	}
	return bases, lignes.Err()
}

// Inventorier alimente l'arbre de sélection sans introspecter : une requête,
// aucune lecture de données.
//
// Le filtre par schéma se fait ici plutôt que dans la requête : SQL Server n'a
// pas de tableau paramétrable, et composer une liste IN à partir de noms venus
// de l'appelant reviendrait à écrire du SQL avec des identifiants qu'on ne
// contrôle pas. La requête rend tout, le pilote retient ce qui est demandé.
func (p *pilote) Inventorier(ctx context.Context, schemas []string) ([]introspection.TableSommaire, error) {
	ctx, annuler := context.WithTimeout(ctx, delaiRequete)
	defer annuler()

	lignes, err := p.db.QueryContext(ctx, requeteInventaire)
	if err != nil {
		return nil, fmt.Errorf("inventaire des tables: %w", err)
	}
	defer func() { _ = lignes.Close() }()

	retenu := ensemble(schemas)
	tables := []introspection.TableSommaire{}
	for lignes.Next() {
		var t introspection.TableSommaire
		var commentaire sql.NullString
		var references string
		if err := lignes.Scan(&t.Schema, &t.Nom, &commentaire, &t.NbColonnes, &t.LignesEstimees, &t.ClePrimaire, &references); err != nil {
			return nil, fmt.Errorf("lecture d'une table: %w", err)
		}
		if retenu != nil && !retenu[t.Schema] {
			continue
		}
		t.Commentaire = commentaire.String
		if references != "" {
			t.ReferenceVers = strings.Split(references, ",")
		}
		tables = append(tables, t)
	}
	return tables, lignes.Err()
}

// Colonnes décrit une table au dépliement de l'arbre, sans l'introspecter.
func (p *pilote) Colonnes(ctx context.Context, schema, table string) ([]introspection.ColonneSommaire, error) {
	ctx, annuler := context.WithTimeout(ctx, delaiRequete)
	defer annuler()

	lignes, err := p.db.QueryContext(ctx, requeteColonnes, schema, table)
	if err != nil {
		return nil, fmt.Errorf("lecture des colonnes de %s.%s: %w", schema, table, err)
	}
	defer func() { _ = lignes.Close() }()

	colonnes := []introspection.ColonneSommaire{}
	for lignes.Next() {
		var c introspection.ColonneSommaire
		var commentaire sql.NullString
		if err := lignes.Scan(&c.Nom, &c.Position, &c.TypeBrut, &c.Nullable, &c.ClePrimaire, &commentaire); err != nil {
			return nil, fmt.Errorf("lecture d'une colonne: %w", err)
		}
		c.Commentaire = commentaire.String
		colonnes = append(colonnes, c)
	}
	return colonnes, lignes.Err()
}

// ensemble rend les schémas demandés sous forme d'ensemble, ou nil quand
// l'appelant n'en demande aucun en particulier — auquel cas tout est retenu.
func ensemble(schemas []string) map[string]bool {
	if len(schemas) == 0 {
		return nil
	}
	m := make(map[string]bool, len(schemas))
	for _, s := range schemas {
		m[s] = true
	}
	return m
}
