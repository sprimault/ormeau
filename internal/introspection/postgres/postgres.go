// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Package postgres introspecte PostgreSQL via pg_catalog.
//
// information_schema serait portable mais perd trop : IDENTITY contre séquence,
// colonnes générées, index partiels, méthodes d'index, commentaires, ordre réel
// des colonnes d'un index. C'est plus de code, et c'est là que se joue la
// qualité de l'outil.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/sprimault/ormeau/internal/introspection"
)

// Importer ce paquet suffit à rendre « postgres » utilisable.
func init() {
	introspection.Enregistrer("postgres", Ouvrir)
}

// delaiRequete plafonne chaque requête de catalogue. Une introspection lancée
// contre une base de production ne doit pas pouvoir s'y installer : mieux vaut
// échouer que rester accroché.
const delaiRequete = 30 * time.Second

// pilote détient la connexion. Non exporté : Ouvrir rend l'interface, et un
// pilote sans connexion ne doit pas pouvoir exister.
type pilote struct {
	conn *pgx.Conn
}

// Ouvrir établit la connexion et la bascule en lecture seule pour toute sa
// durée. L'absence d'écriture est un invariant : la faire tenir par le serveur
// plutôt que par la discipline du code est ce qui la rend vérifiable.
//
// Le DSN ne ressort jamais tel quel d'une erreur, seulement masqué.
func Ouvrir(ctx context.Context, dsn string) (introspection.Introspecteur, error) {
	config, err := pgx.ParseConfig(introspection.NettoyerDSN(dsn))
	if err != nil {
		return nil, fmt.Errorf("dsn illisible (%s): %w", introspection.Masquer(dsn), err)
	}

	// Posé à la connexion plutôt que par un SET ensuite : il n'existe alors
	// aucune fenêtre pendant laquelle la session serait inscriptible. Vaut pour
	// toute sa durée, échantillonnage compris.
	if config.RuntimeParams == nil {
		config.RuntimeParams = map[string]string{}
	}
	config.RuntimeParams["default_transaction_read_only"] = "on"
	config.RuntimeParams["application_name"] = "ormeau"

	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("connexion a %s: %w", introspection.Masquer(dsn), err)
	}
	return &pilote{conn: conn}, nil
}

// Inventorier alimente l'arbre de sélection sans introspecter : une requête,
// aucune lecture de données.
func (p *pilote) Inventorier(ctx context.Context, schemas []string) ([]introspection.TableSommaire, error) {
	if len(schemas) == 0 {
		schemas = []string{"public"}
	}

	ctx, annuler := context.WithTimeout(ctx, delaiRequete)
	defer annuler()

	lignes, err := p.conn.Query(ctx, requeteInventaire, schemas)
	if err != nil {
		return nil, fmt.Errorf("inventaire des tables: %w", err)
	}
	defer lignes.Close()

	// Jamais nil : une portée sans table est un résultat vide, pas une absence
	// de résultat.
	sommaires := []introspection.TableSommaire{}
	for lignes.Next() {
		var s introspection.TableSommaire
		var commentaire *string

		if err := lignes.Scan(&s.Schema, &s.Nom, &commentaire, &s.NbColonnes,
			&s.LignesEstimees, &s.ClePrimaire, &s.ReferenceVers); err != nil {
			return nil, fmt.Errorf("lecture de l'inventaire: %w", err)
		}
		if commentaire != nil {
			s.Commentaire = *commentaire
		}
		// reltuples vaut -1 tant que la table n'a jamais été analysée.
		if s.LignesEstimees < 0 {
			s.LignesEstimees = 0
		}
		sommaires = append(sommaires, s)
	}
	if err := lignes.Err(); err != nil {
		return nil, fmt.Errorf("parcours de l'inventaire: %w", err)
	}
	return sommaires, nil
}

// Fermer libère la connexion.
func (p *pilote) Fermer() error {
	if p.conn == nil {
		return nil
	}
	// Contexte propre : celui de l'appelant est souvent déjà annulé au moment
	// de fermer, ce qui laisserait la connexion ouverte côté serveur.
	ctx, annuler := context.WithTimeout(context.Background(), 5*time.Second)
	defer annuler()
	return p.conn.Close(ctx)
}
