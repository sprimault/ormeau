// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package sqlserver

// Requêtes de catalogue, toutes en lecture seule.
//
// Elles sont rassemblées ici pour qu'un test puisse les parcourir et refuser
// tout verbe d'écriture : SQL Server n'a pas d'équivalent au
// default_transaction_read_only de PostgreSQL, et l'invariant « aucune écriture
// dans la base introspectée » ne peut donc pas être tenu par le serveur. Le
// tenir par le code demande de pouvoir le vérifier.
//
// sys.* plutôt qu'INFORMATION_SCHEMA : les vues normalisées perdent la
// distinction IDENTITY contre séquence, les colonnes calculées, les index
// filtrés et les commentaires d'extended property.
const (
	// requeteServeur décrit le serveur atteint. SERVERPROPERTY rend sql_variant,
	// converti pour que le pilote reçoive des chaînes.
	requeteServeur = `
SELECT CONVERT(nvarchar(128), SERVERPROPERTY('ProductVersion')),
       CONVERT(nvarchar(128), SERVERPROPERTY('Edition')),
       DB_NAME()`

	// requeteBases énumère les bases exploitables. Les quatre bases système
	// sont écartées : leur calque n'aurait aucun intérêt. state_desc filtre
	// celles qui ne sont pas en ligne, qu'une lecture ferait échouer.
	requeteBases = `
SELECT name
FROM sys.databases
WHERE database_id > 4 AND state_desc = 'ONLINE' AND HAS_DBACCESS(name) = 1
ORDER BY name`

	// requeteSchemas énumère les schémas qui portent au moins une table.
	// Les schémas système et ceux créés pour les rôles fixes n'en portent
	// aucune, et disparaissent d'eux-mêmes.
	requeteSchemas = `
SELECT DISTINCT s.name
FROM sys.tables t
JOIN sys.schemas s ON s.schema_id = t.schema_id
WHERE t.is_ms_shipped = 0
ORDER BY s.name`

	// requeteInventaire alimente l'arbre de sélection. Une seule requête, aucune
	// lecture de données : le nombre de lignes vient des statistiques de
	// partition, et non d'un COUNT qui parcourrait chaque table d'une base de
	// production pour afficher un ordre de grandeur.
	//
	// Le commentaire est une extended property nommée MS_Description : SQL
	// Server n'a pas de COMMENT ON, et c'est la convention que suivent SSMS et
	// les outils de modélisation.
	//
	// is_ms_shipped écarte ce que Microsoft installe lui-même : master porte
	// MSreplication_options et les spt_fallback_*, qui sont des tables
	// ordinaires du catalogue et rempliraient l'arbre de lignes qu'on ne mappe
	// jamais.
	requeteInventaire = `
SELECT s.name AS [schema],
       t.name AS nom,
       CONVERT(nvarchar(max), ep.value) AS commentaire,
       (SELECT COUNT(*) FROM sys.columns c WHERE c.object_id = t.object_id) AS nb_colonnes,
       CONVERT(bigint, ISNULL((SELECT SUM(p.rows) FROM sys.partitions p
                               WHERE p.object_id = t.object_id AND p.index_id IN (0, 1)), 0)) AS lignes_estimees,
       CONVERT(bit, CASE WHEN EXISTS (SELECT 1 FROM sys.key_constraints k
                                      WHERE k.parent_object_id = t.object_id AND k.type = 'PK')
                         THEN 1 ELSE 0 END) AS cle_primaire,
       ISNULL(STUFF((SELECT DISTINCT ',' + sf.name + '.' + tf.name
                     FROM sys.foreign_keys fk
                     JOIN sys.tables tf ON tf.object_id = fk.referenced_object_id
                     JOIN sys.schemas sf ON sf.schema_id = tf.schema_id
                     WHERE fk.parent_object_id = t.object_id
                     FOR XML PATH(''), TYPE).value('.', 'nvarchar(max)'), 1, 1, ''), '') AS reference_vers
FROM sys.tables t
JOIN sys.schemas s ON s.schema_id = t.schema_id
LEFT JOIN sys.extended_properties ep
       ON ep.major_id = t.object_id AND ep.minor_id = 0 AND ep.class = 1 AND ep.name = 'MS_Description'
WHERE t.is_ms_shipped = 0
ORDER BY s.name, t.name`

	// requeteColonnes décrit une table au dépliement, sans l'introspecter
	// entièrement. max_length est en octets : pour nchar et nvarchar il vaut
	// deux fois la longueur déclarée, et -1 pour (max) — le type brut se
	// reconstruit donc ici, où l'on sait de quel type il s'agit.
	requeteColonnes = `
SELECT c.name,
       c.column_id,
       ty.name + CASE
           WHEN ty.name IN ('nchar', 'nvarchar') AND c.max_length = -1 THEN '(max)'
           WHEN ty.name IN ('nchar', 'nvarchar') THEN '(' + CONVERT(varchar(11), c.max_length / 2) + ')'
           WHEN ty.name IN ('char', 'varchar', 'binary', 'varbinary') AND c.max_length = -1 THEN '(max)'
           WHEN ty.name IN ('char', 'varchar', 'binary', 'varbinary') THEN '(' + CONVERT(varchar(11), c.max_length) + ')'
           WHEN ty.name IN ('decimal', 'numeric') THEN '(' + CONVERT(varchar(11), c.precision) + ',' + CONVERT(varchar(11), c.scale) + ')'
           ELSE ''
       END AS type_brut,
       c.is_nullable,
       CONVERT(bit, CASE WHEN EXISTS (
           SELECT 1 FROM sys.key_constraints k
           JOIN sys.index_columns ic ON ic.object_id = k.parent_object_id AND ic.index_id = k.unique_index_id
           WHERE k.parent_object_id = c.object_id AND k.type = 'PK' AND ic.column_id = c.column_id)
           THEN 1 ELSE 0 END) AS cle_primaire,
       CONVERT(nvarchar(max), ep.value) AS commentaire
FROM sys.columns c
JOIN sys.tables t ON t.object_id = c.object_id
JOIN sys.schemas s ON s.schema_id = t.schema_id
JOIN sys.types ty ON ty.user_type_id = c.user_type_id
LEFT JOIN sys.extended_properties ep
       ON ep.major_id = c.object_id AND ep.minor_id = c.column_id AND ep.class = 1 AND ep.name = 'MS_Description'
WHERE s.name = @p1 AND t.name = @p2
ORDER BY c.column_id`
)
