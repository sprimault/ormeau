// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package postgres

// Requêtes de catalogue. Littérales, jamais assemblées à partir d'une entrée :
// c'est ce qui rend l'injection structurellement impossible à l'introspection.
// Un ORDER BY partout — l'ordre des lignes ne doit dépendre ni du
// planificateur ni du parallélisme.

// requeteSource identifie la base sans jamais toucher au DSN.
const requeteSource = `
SELECT current_setting('server_version') AS version,
       current_database()                AS catalogue
`

// requeteSchemas sert quand la portée n'en nomme aucun. Retenir « public »
// seul rendrait un calque vide pour toute base qui n'y range rien — et une
// base legacy range rarement dans public.
const requeteSchemas = `
SELECT nspname
FROM pg_namespace
WHERE nspname NOT LIKE 'pg\_%'
  AND nspname <> 'information_schema'
ORDER BY nspname
`

// datallowconn écarte template0, sur laquelle aucune connexion n'est possible.
// datistemplate écarte template1 et les autres modèles : ils ne portent que le
// squelette d'une base, et en produire un calque n'apprendrait rien.
const requeteBases = `
SELECT datname
FROM pg_database
WHERE datallowconn
  AND NOT datistemplate
  AND datname <> 'postgres'
ORDER BY datname
`

// Les séquences autonomes comme celles qui appartiennent à une colonne : la
// distinction se fait plus loin, sur le défaut de la colonne. Une séquence
// possédée reste un objet du catalogue, et l'omettre rendrait le DDL
// incomplet.
const requeteSequences = `
SELECT n.nspname       AS schema,
       c.relname       AS nom,
       s.seqincrement  AS increment,
       s.seqmin        AS minimum,
       s.seqmax        AS maximum,
       s.seqcycle      AS cyclique
FROM pg_sequence s
         JOIN pg_class c ON c.oid = s.seqrelid
         JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = ANY ($1)
ORDER BY n.nspname, c.relname
`

// pg_get_viewdef rend la définition verbatim. Les vues ne produisent pas
// d'entité, mais un calque amputé ne permettrait plus de reconstruire un DDL
// équivalent.
const requeteVues = `
SELECT n.nspname                        AS schema,
       c.relname                        AS nom,
       pg_get_viewdef(c.oid, true)      AS definition,
       c.relkind = 'm'                  AS materialisee
FROM pg_class c
         JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE c.relkind IN ('v', 'm')
  AND n.nspname = ANY ($1)
ORDER BY n.nspname, c.relname
`

// requeteTables porte ce qui appartient à la table elle-même : son
// commentaire. Séparée des colonnes parce qu'une table sans colonne visible
// doit quand même apparaître dans le calque.
const requeteTables = `
SELECT n.nspname               AS schema,
       c.relname               AS nom,
       obj_description(c.oid)  AS commentaire
FROM pg_class c
         JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE c.relkind IN ('r', 'p')
  AND n.nspname = ANY ($1)
ORDER BY n.nspname, c.relname
`

// requeteColonnesIndex rend les colonnes d'un index dans l'ordre de l'index,
// qui n'est pas celui de la table. pg_index.indkey est un tableau d'attnum ;
// l'ordinalité de unnest préserve l'ordre déclaré.
//
// Les index d'expression ont un attnum à zéro : la colonne est alors nulle, et
// l'index n'est pas exploitable comme liste de colonnes.
// La classe d'opérateurs vient de indclass, apparié par ordinalité au même
// rang que la colonne. opcdefault dit si elle est implicite : ne remonter que
// les classes explicites évite de charger le calque d'un int4_ops par index
// trivial, et c'est ce que fait pg_get_indexdef lui-même.
const requeteColonnesIndex = `
SELECT n.nspname     AS schema,
       c.relname     AS table_nom,
       i.relname     AS index_nom,
       a.attname     AS colonne,
       o.opcname     AS classe_operateurs,
       o.opcdefault  AS classe_par_defaut,
       k.ordinalite
FROM pg_index ix
         JOIN pg_class c ON c.oid = ix.indrelid
         JOIN pg_class i ON i.oid = ix.indexrelid
         JOIN pg_namespace n ON n.oid = c.relnamespace
         CROSS JOIN LATERAL unnest(ix.indkey) WITH ORDINALITY AS k(attnum, ordinalite)
         LEFT JOIN pg_attribute a ON a.attrelid = ix.indrelid AND a.attnum = k.attnum
         LEFT JOIN pg_opclass o ON o.oid = ix.indclass[k.ordinalite - 1]
WHERE NOT ix.indisprimary
  AND n.nspname = ANY ($1)
ORDER BY n.nspname, c.relname, i.relname, k.ordinalite
`

// requeteColonnesContrainte rend les colonnes d'une contrainte dans l'ordre
// déclaré. conkey vaut pour les clés primaires, uniques et étrangères ;
// confkey pour le côté référencé d'une clé étrangère.
const requeteColonnesContrainte = `
SELECT n.nspname     AS schema,
       c.relname     AS table_nom,
       con.conname   AS contrainte,
       con.contype   AS genre,
       'source'      AS cote,
       a.attname     AS colonne,
       k.ordinalite
FROM pg_constraint con
         JOIN pg_class c ON c.oid = con.conrelid
         JOIN pg_namespace n ON n.oid = c.relnamespace
         CROSS JOIN LATERAL unnest(con.conkey) WITH ORDINALITY AS k(attnum, ordinalite)
         JOIN pg_attribute a ON a.attrelid = con.conrelid AND a.attnum = k.attnum
WHERE n.nspname = ANY ($1)
  AND con.contype IN ('p', 'u', 'f')

UNION ALL

SELECT n.nspname     AS schema,
       c.relname     AS table_nom,
       con.conname   AS contrainte,
       con.contype   AS genre,
       'cible'       AS cote,
       a.attname     AS colonne,
       k.ordinalite
FROM pg_constraint con
         JOIN pg_class c ON c.oid = con.conrelid
         JOIN pg_namespace n ON n.oid = c.relnamespace
         CROSS JOIN LATERAL unnest(con.confkey) WITH ORDINALITY AS k(attnum, ordinalite)
         JOIN pg_attribute a ON a.attrelid = con.confrelid AND a.attnum = k.attnum
WHERE n.nspname = ANY ($1)
  AND con.contype = 'f'

ORDER BY schema, table_nom, contrainte, cote, ordinalite
`

// pg_attribute donne l'identité, les colonnes générées, la collation et
// l'ordre réel des colonnes. information_schema, non.
const requeteColonnes = `
SELECT c.relname                                            AS table_nom,
       n.nspname                                            AS table_schema,
       a.attname                                            AS colonne,
       a.attnum                                             AS position,
       format_type(a.atttypid, a.atttypmod)                 AS type_brut,
       t.typname                                            AS type_interne,
       t.typtype = 'e'                                      AS est_enumere,
       CASE WHEN t.typname IN ('varchar', 'bpchar', 'char') AND a.atttypmod > 4
            THEN a.atttypmod - 4 END                        AS longueur,
       CASE WHEN t.typname = 'numeric' AND a.atttypmod > 4
            THEN ((a.atttypmod - 4) >> 16) & 65535 END      AS precision,
       CASE WHEN t.typname = 'numeric' AND a.atttypmod > 4
            THEN (a.atttypmod - 4) & 65535 END              AS echelle,
       NOT a.attnotnull                                     AS nullable,
       a.attidentity <> ''                                  AS identite,
       pg_get_expr(d.adbin, d.adrelid)                       AS defaut,
       a.attgenerated::text                                 AS generee,
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
       cf.relname                      AS table_cible,
       nf.nspname                      AS schema_cible,
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
