# Feuille de route

Ordre choisi pour qu'une chaîne bout en bout existe le plus tôt possible : mieux
vaut un PostgreSQL vers Doctrine qui marche vraiment que cinq pilotes sans
générateur.

## Phase 1 — Le calque physique

Structures Go, sérialisation déterministe, empreinte, JSON Schema v1 publié.
Chargement et validation. Aucun accès base : le format d'abord, écrit à la main
sur un exemple réel.

## Phase 2 — Introspection PostgreSQL

Premier pilote, via `pg_catalog`. Tables, colonnes, PK, FK, unicités, index,
`CHECK`, séquences, types énumérés, commentaires, colonnes générées.
Conteneur de test et DDL de référence couvrant les cas tordus.

## Phase 3 — Inférence et calque logique

Entités, propriétés, associations. Heuristiques de base : table de jointure pure,
suffixe `_id`, préfixes, singularisation. Avertissements avec code, cible,
confiance et origine. Fichier de décisions lu et prioritaire.

## Phase 4 — Génération Doctrine

Paquet PHP, commande `ormeau:generer`, mode classe de base séparée. Attributs
PHP 8, énumérations, traits d'horodatage.

## Phase 5 — Aller-retour

`calque -> entités -> schema:create -> diff structurel`. C'est la phase qui
valide tout ce qui précède, et probablement celle qui fera remonter le plus de
manques dans le format.

## Phase 6 — Introspection MySQL, MariaDB et SQL Server

Deuxième et troisième dialectes. C'est là qu'on découvre ce que le calque v1 ne
capture pas ; incrémenter `version_ri` si nécessaire, une seule fois de
préférence.

MariaDB partage le protocole de MySQL et se traite dans le même paquet, la
variante étant détectée à la connexion. Elle en diverge assez pour compter comme
un SGBD à part entière dans le calque : elle a de vraies séquences là où MySQL
n'a qu'`AUTO_INCREMENT`, et son type `JSON` n'est qu'un alias de `LONGTEXT`.

## Phase 7 — Échantillonnage

Statistiques, détection d'énumérations par cardinalité, détection des clés
étrangères implicites. Optionnel, plafonné, lecture seule.

## Phase 8 — Diff

`ormeau diff` entre deux calques physiques, et `ormeau:synchroniser` entre calque
et entités existantes. Sortie lisible, sortie JSON, code de retour exploitable en
CI.

## Phase 9 — Interface de sélection et d'arbitrage

`ormeau interface` : front local embarqué. Saisie de connexion, arbre des tables
alimenté par `Inventorier`, sélection avec propagation des dépendances par clé
étrangère, puis écran d'arbitrage des avertissements produisant un
`decisions.yaml`.

Placée ici et pas avant : tant que l'inférence ne produit pas de vrais
avertissements sur plusieurs dialectes, l'écran d'arbitrage n'aurait rien à
afficher.

## Phase 10 — Régénération par AST

Mode avancé : réécriture ciblée avec `nikic/php-parser`, préservation des
méthodes métier et du formatage. Tests de survie des modifications manuelles.

## Phase 11 — Publication

README bilingue, documentation d'installation, avertissement d'usage, subtree
split vers Packagist, image Docker multi-arch, binaires de version.

## Hors périmètre v1

SQLite et Oracle (le format doit d'abord se stabiliser sur trois dialectes),
générateurs autres que Doctrine, migrations. Les données des tables, les droits,
les procédures stockées et les triggers ne sont pas dans le périmètre du calque,
jamais.

L'interface n'est ni un client SQL ni un explorateur de données : pas d'éditeur
de requêtes, pas d'affichage de lignes, pas de modification.
