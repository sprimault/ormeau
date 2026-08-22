// Package postgres introspecte PostgreSQL via pg_catalog.
//
// information_schema serait portable mais perd trop : IDENTITY contre séquence,
// colonnes générées, index partiels, méthodes d'index, commentaires, ordre réel
// des colonnes d'un index. C'est plus de code, et c'est là que se joue la
// qualité de l'outil.
package postgres

import (
	"context"

	"github.com/sprimault/ormeau/internal/calque"
	"github.com/sprimault/ormeau/internal/introspection"
)

func init() {
	introspection.Enregistrer("postgres", Ouvrir)
}

type Pilote struct {
	// conn *pgx.Conn
}

func Ouvrir(ctx context.Context, dsn string) (introspection.Introspecteur, error) {
	panic("à implémenter : connexion en lecture seule, délai maximal par requête")
}

func (p *Pilote) Inventorier(ctx context.Context, schemas []string) ([]introspection.TableSommaire, error) {
	panic("à implémenter")
}

func (p *Pilote) Extraire(ctx context.Context, portee introspection.Portee) (*calque.Physique, error) {
	panic("à implémenter")
}

func (p *Pilote) Fermer() error {
	panic("à implémenter")
}

// Requêtes de catalogue. Gardées ici, littérales, plutôt que construites : une
// requête lisible telle qu'elle sera envoyée vaut mieux qu'un constructeur.

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
