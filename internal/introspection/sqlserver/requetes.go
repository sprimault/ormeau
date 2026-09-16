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

// Requêtes de l'extraction. Chacune porte sur un schéma, passé en paramètre :
// l'extraction les répète schéma par schéma plutôt que de tout lire pour
// filtrer ensuite, la description des types coûtant une compilation par table.
const (
	// requeteTablesExtraction collecte les tables du schéma. Les vues sont
	// lues à part : elles n'ont ni clé ni défaut à compléter.
	requeteTablesExtraction = `
SELECT t.name
FROM sys.tables t
JOIN sys.schemas s ON s.schema_id = t.schema_id
WHERE t.is_ms_shipped = 0 AND s.name = @p1
ORDER BY t.name`

	// requeteColonnesExtraction lit les colonnes dans sys.columns, qui les
	// porte toutes, y compris celles qu'un SELECT * ne rend pas.
	//
	// type_systeme est le type de base, qui décide de la normalisation et de
	// la lecture de max_length : celui d'un type alias est le type dont il
	// dérive, que TYPE_NAME rend depuis system_type_id. Les types CLR partagent
	// tous le system_type_id 240, et gardent leur propre nom.
	//
	// Une colonne a au plus une contrainte DEFAULT : la jointure ne duplique
	// aucune ligne.
	requeteColonnesExtraction = `
SELECT t.name,
       c.name,
       c.column_id,
       CASE WHEN ty.system_type_id = 240 THEN ty.name ELSE TYPE_NAME(ty.system_type_id) END AS type_systeme,
       CONVERT(int, c.max_length),
       CONVERT(int, c.precision),
       CONVERT(int, c.scale),
       c.is_nullable,
       c.is_identity,
       dc.definition
FROM sys.columns c
JOIN sys.tables t ON t.object_id = c.object_id
JOIN sys.schemas s ON s.schema_id = t.schema_id
JOIN sys.types ty ON ty.user_type_id = c.user_type_id
LEFT JOIN sys.default_constraints dc
       ON dc.parent_object_id = c.object_id AND dc.parent_column_id = c.column_id
WHERE t.is_ms_shipped = 0 AND s.name = @p1
ORDER BY t.name, c.column_id`

	// requeteTypesDecrits rend le type de chaque colonne tel que le serveur
	// l'écrit : SQL Server n'a pas d'équivalent au format_type de PostgreSQL,
	// et reconstruire la chaîne depuis max_length, precision et scale
	// oublierait ce qu'on n'a pas pensé à lister — l'échelle d'un
	// datetime2(3), par exemple. type_brut est verbatim, pas reconstruit.
	//
	// La fonction décrit une requête sans l'exécuter. Le texte décrit est
	// composé côté serveur, depuis des noms du catalogue passés par
	// QUOTENAME : rien de ce qui entre dans la chaîne ne vient de l'appelant.
	//
	// Une table qu'elle ne sait pas décrire ne lève pas d'erreur : elle rend
	// des lignes dont error_number est renseigné. L'extraction les lit et
	// s'arrête, plutôt que de laisser des types vides.
	requeteTypesDecrits = `
SELECT t.name,
       r.name,
       r.system_type_name,
       r.user_type_name,
       r.error_number,
       r.error_message
FROM sys.tables t
JOIN sys.schemas s ON s.schema_id = t.schema_id
CROSS APPLY sys.dm_exec_describe_first_result_set(
    N'SELECT * FROM ' + QUOTENAME(s.name) + N'.' + QUOTENAME(t.name), NULL, 0) r
WHERE t.is_ms_shipped = 0 AND s.name = @p1
ORDER BY t.name, r.column_ordinal`

	// requeteClesPrimaires rend les colonnes de chaque clé primaire dans
	// l'ordre de la clé, qui n'est pas celui de la table : (b, a) n'est pas
	// (a, b).
	requeteClesPrimaires = `
SELECT t.name, k.name, c.name
FROM sys.key_constraints k
JOIN sys.tables t ON t.object_id = k.parent_object_id
JOIN sys.schemas s ON s.schema_id = t.schema_id
JOIN sys.index_columns ic ON ic.object_id = k.parent_object_id AND ic.index_id = k.unique_index_id
JOIN sys.columns c ON c.object_id = ic.object_id AND c.column_id = ic.column_id
WHERE k.type = 'PK' AND t.is_ms_shipped = 0 AND s.name = @p1
ORDER BY t.name, ic.key_ordinal`

	// requeteClesEtrangeres rend une ligne par couple de colonnes, dans
	// l'ordre de la contrainte. Les actions sont lues sous leur forme
	// textuelle, plus sûre que les codes numériques à traduire.
	requeteClesEtrangeres = `
SELECT t.name,
       fk.name,
       sc.name,
       tc.name,
       fk.delete_referential_action_desc,
       fk.update_referential_action_desc,
       cp.name,
       cr.name
FROM sys.foreign_keys fk
JOIN sys.tables t ON t.object_id = fk.parent_object_id
JOIN sys.schemas s ON s.schema_id = t.schema_id
JOIN sys.objects tc ON tc.object_id = fk.referenced_object_id
JOIN sys.schemas sc ON sc.schema_id = tc.schema_id
JOIN sys.foreign_key_columns fkc ON fkc.constraint_object_id = fk.object_id
JOIN sys.columns cp ON cp.object_id = fkc.parent_object_id AND cp.column_id = fkc.parent_column_id
JOIN sys.columns cr ON cr.object_id = fkc.referenced_object_id AND cr.column_id = fkc.referenced_column_id
WHERE t.is_ms_shipped = 0 AND s.name = @p1
ORDER BY t.name, fk.name, fkc.constraint_column_id`
)
