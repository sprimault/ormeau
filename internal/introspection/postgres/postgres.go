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
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/sprimault/ormeau/internal/calque"
	"github.com/sprimault/ormeau/internal/introspection"
)

// Importer ce paquet suffit à rendre « postgres » utilisable.
func init() {
	introspection.Enregistrer("postgres", Ouvrir)
}

// errNonImplemente couvre ce qui reste à écrire du pilote.
var errNonImplemente = errors.New("extraction postgres non implementee, phase 2 de la feuille de route")

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
	config, err := pgx.ParseConfig(dsn)
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

// Extraire rend un calque trié. Ce qu'elle ne capture pas est perdu : aucune
// couche en aval ne peut le retrouver.
func (p *pilote) Extraire(ctx context.Context, portee introspection.Portee) (*calque.Physique, error) {
	return nil, errNonImplemente
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

// Requêtes de catalogue. Gardées ici, littérales, plutôt que construites : une
// requête lisible telle qu'elle sera envoyée vaut mieux qu'un constructeur.

// pg_attribute donne l'identité, les colonnes générées, la collation et
// l'ordre réel des colonnes. information_schema, non.
const requeteColonnes = `
SELECT c.relname                                            AS table_nom,
       n.nspname                                            AS table_schema,
       a.attname                                            AS colonne,
       a.attnum                                             AS position,
       format_type(a.atttypid, a.atttypmod)                 AS type_brut,
       t.typname                                            AS type_interne,
       NOT a.attnotnull                                     AS nullable,
       a.attidentity <> ''                                  AS identite,
       pg_get_expr(d.adbin, d.adrelid)                       AS defaut,
       a.attgenerated                                       AS generee,
       col_description(c.oid, a.attnum)                     AS commentaire,
       co.collname                                          AS collation
FROM pg_attribute a
         JOIN pg_class c ON c.oid = a.attrelid
         JOIN pg_namespace n ON n.oid = c.relnamespace
         JOIN pg_type t ON t.oid = a.atttypid
         LEFT JOIN pg_attrdef d ON d.adrelid = c.oid AND d.adnum = a.attnum
         LEFT JOIN pg_collation co ON co.oid = a.attcollation
WHERE c.relkind IN ('r', 'p')
  AND a.attnum > 0
  AND NOT a.attisdropped
  AND n.nspname = ANY ($1)
ORDER BY n.nspname, c.relname, a.attnum
`

// Clés primaires, étrangères, unicités et CHECK en une passe.
// pg_get_constraintdef rend la définition verbatim.
const requeteContraintes = `
SELECT n.nspname                       AS table_schema,
       c.relname                       AS table_nom,
       con.conname                     AS contrainte,
       con.contype                     AS genre,
       pg_get_constraintdef(con.oid)   AS definition,
       con.conkey                      AS colonnes,
       cf.relname                      AS table_cible,
       nf.nspname                      AS schema_cible,
       con.confkey                     AS colonnes_cibles,
       con.confdeltype                 AS a_la_suppression,
       con.confupdtype                 AS a_la_mise_a_jour
FROM pg_constraint con
         JOIN pg_class c ON c.oid = con.conrelid
         JOIN pg_namespace n ON n.oid = c.relnamespace
         LEFT JOIN pg_class cf ON cf.oid = con.confrelid
         LEFT JOIN pg_namespace nf ON nf.oid = cf.relnamespace
WHERE n.nspname = ANY ($1)
ORDER BY n.nspname, c.relname, con.conname
`

// L'index de clé primaire est exclu, les contraintes le rendent déjà. Prédicat
// et méthode d'accès n'existent que dans le catalogue natif.
const requeteIndex = `
SELECT n.nspname                                   AS table_schema,
       c.relname                                   AS table_nom,
       i.relname                                   AS index_nom,
       ix.indisunique                              AS unique,
       am.amname                                   AS methode,
       pg_get_expr(ix.indpred, ix.indrelid)         AS predicat,
       pg_get_indexdef(ix.indexrelid)              AS definition
FROM pg_index ix
         JOIN pg_class c ON c.oid = ix.indrelid
         JOIN pg_class i ON i.oid = ix.indexrelid
         JOIN pg_namespace n ON n.oid = c.relnamespace
         JOIN pg_am am ON am.oid = i.relam
WHERE NOT ix.indisprimary
  AND n.nspname = ANY ($1)
ORDER BY n.nspname, c.relname, i.relname
`

// enumsortorder, pas l'ordre alphabétique : c'est l'ordre de déclaration.
const requeteTypesEnumeres = `
SELECT n.nspname AS schema,
       t.typname AS nom,
       e.enumlabel AS valeur
FROM pg_type t
         JOIN pg_namespace n ON n.oid = t.typnamespace
         JOIN pg_enum e ON e.enumtypid = t.oid
WHERE n.nspname = ANY ($1)
ORDER BY n.nspname, t.typname, e.enumsortorder
`

// requeteInventaire alimente l'arbre de sélection. Une seule requête, aucune
// lecture de données : reltuples est une estimation du planificateur.
const requeteInventaire = `
SELECT n.nspname                                  AS schema,
       c.relname                                  AS nom,
       obj_description(c.oid)                     AS commentaire,
       c.relnatts                                 AS nb_colonnes,
       c.reltuples::bigint                        AS lignes_estimees,
       EXISTS (SELECT 1 FROM pg_constraint pk
               WHERE pk.conrelid = c.oid AND pk.contype = 'p') AS cle_primaire,
       COALESCE(
           (SELECT array_agg(DISTINCT nf.nspname || '.' || cf.relname)
            FROM pg_constraint fk
                     JOIN pg_class cf ON cf.oid = fk.confrelid
                     JOIN pg_namespace nf ON nf.oid = cf.relnamespace
            WHERE fk.conrelid = c.oid AND fk.contype = 'f'),
           '{}'
       )                                          AS reference_vers
FROM pg_class c
         JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE c.relkind IN ('r', 'p')
  AND n.nspname = ANY ($1)
ORDER BY n.nspname, c.relname
`
