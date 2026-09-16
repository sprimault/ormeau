# Feuille de route

Ordre choisi pour qu'une chaîne bout en bout existe le plus tôt possible : mieux
vaut un PostgreSQL vers Doctrine qui marche vraiment que cinq pilotes sans
générateur.

| Phase | État |
|---|---|
| 1 — Calque physique | terminée |
| 2 — Introspection PostgreSQL | terminée |
| 3 — Inférence et calque logique | terminée, hors clés étrangères implicites |
| 4 — Interface | terminée |
| 5 — Génération Doctrine | terminée |
| 6 — Aller-retour | terminée |
| 7 — Introspection SQL Server | en cours |
| 8 à 11 | non commencées |
| 12 — Publication | faite |

## Phase 1 — Le calque physique

Structures Go, sérialisation déterministe, empreinte, JSON Schema v1 publié.
Chargement et validation. Aucun accès base : le format d'abord, écrit à la main
sur un exemple réel.

## Phase 2 — Introspection PostgreSQL

Premier pilote, via `pg_catalog`. Tables, colonnes, PK, FK, unicités, index,
`CHECK`, séquences, types énumérés, commentaires, colonnes générées.
Conteneur de test et DDL de référence couvrant les cas difficiles.

## Phase 3 — Inférence et calque logique

Entités, propriétés, associations. Heuristiques de base : table de jointure pure,
suffixe `_id`, énumérations depuis un type natif ou un `CHECK`, traits
d'horodatage. Avertissements avec code, cible, confiance et origine. Fichier de
décisions lu et prioritaire.

L'héritage par clé primaire étrangère est **proposé, pas appliqué**, depuis la
phase 5 : le schéma autorise aussi bien « un salarié est une personne » que
« un salarié a une personne », et Doctrine exige une colonne discriminante que
les bases reprises ne portent pas. Sans décision, la table est reliée à son
parent par un-vers-un ; `heritages` déclare la hiérarchie et sa colonne.

Préfixes et singularisation sont **proposés, pas appliqués** : ils changent ce
qu'un nom désigne, et une base ne dit pas sa langue. Ils alimentent le fichier de
décisions prérempli, où l'utilisateur décommente ce qui lui convient.

## Phase 4 — Interface de sélection et d'arbitrage

`ormeau interface` : front local embarqué. Saisie de connexion, arbre des tables
alimenté par `Inventorier`, sélection avec propagation des dépendances par clé
étrangère, puis écran d'arbitrage produisant un `<base>.decisions.yaml`.

Livrés : le serveur local — port dynamique, jeton d'URL à usage unique échangé
contre un cookie —, la connexion par composants ou par DSN, l'arbre des bases,
des schémas et des tables, colonnes comprises, l'extraction en tâche de fond,
suivie étape par étape dans l'en-tête et annulable, et l'écran d'arbitrage, qui
travaille hors ligne sur le calque.

Ici et pas plus tard, pour deux raisons. L'écran de connexion et l'arbre ne
dépendent que d'`Inventorier`, écrit depuis la phase 2 : les repousser était un
défaut de séquencement. Et voir les avertissements dans une interface, avec leur
confiance et leur origine, est le meilleur moyen de juger des heuristiques de la
phase 3 — bien mieux qu'un fichier de référence.

Le prix, assumé : l'écran d'arbitrage suivra les évolutions de l'inférence dans
les phases suivantes. L'écran de connexion, lui, ne bougera plus.

## Phase 5 — Génération Doctrine

Paquet PHP, commande `ormeau:generer`, mode classe de base séparée. Attributs,
énumérations natives, traits d'horodatage, héritage joint déclaré par décision.

Plancher : PHP 8.1, Symfony 5.4 à 8, Doctrine ORM 2.14 à 3. C'est le périmètre
des applications que l'outil vise — une reprise de legacy tourne rarement sur la
version courante. ORM accepte `enumType` dès la 2.11, mais les versions 2.11 à
2.13 refusent une classe de base mappée placée sous une entité, ce que produit
la génération d'une hiérarchie : le plancher suit cette contrainte constatée.

La forme du code produit dépend de la version d'ORM installée dans l'application
cible : elle est détectée, non configurée, et annoncée en tête d'exécution.
`--cible-orm` la force, pour générer à destination d'une version qui n'est pas
installée. Un passage d'ORM 2 à ORM 3 produira donc un diff large à la
régénération suivante, et ce n'est pas une surprise à découvrir dans un
`git diff`.

Hors périmètre : Symfony 5.4 sous PHP 7. Il faudrait réécrire le bundle sans
`readonly`, `enum` ni `match`, et un second générateur en annotations visant
`doctrine/annotations`, que Doctrine a abandonné. Une application 5.4 passée à
PHP 8.1 reste couverte, et c'est le chemin que prennent la plupart des
migrations.

## Phase 6 — Aller-retour

`calque -> entités -> schema:create -> diff structurel`. C'est la phase qui
valide tout ce qui précède, et probablement celle qui fera remonter le plus de
manques dans le format.

## Phase 7 — Introspection SQL Server

Deuxième dialecte, lu dans `sys.*`. C'est là qu'on découvre ce que le calque v1
ne capture pas ; incrémenter `version_ri` si nécessaire, une seule fois pour
cette phase et la suivante — ce qui suppose de regarder d'abord ce que MySQL et
MariaDB exigeront aussi.

SQL Server d'abord parce que c'est la base du public visé : le développeur PHP
qui reprend une application dont le schéma a été écrit pour ASP classic ou les
débuts de PHP, et qui porte les traces de cette époque — `binary(64)` alimenté
hors ORM, `uniqueidentifier` par défaut `newid()`, `bit` partout où un booléen
serait attendu, colonnes calculées non persistées, noms de colonnes à espaces,
contraintes `DEFAULT` nommées automatiquement.

## Phase 8 — Introspection MySQL et MariaDB

Troisième dialecte. MariaDB partage le protocole de MySQL et se traite dans le
même paquet, la variante étant détectée à la connexion. Elle en diverge assez
pour compter comme un SGBD à part entière dans le calque : elle a de vraies
séquences là où MySQL n'a qu'`AUTO_INCREMENT`, et son type `JSON` n'est qu'un
alias de `LONGTEXT`.

## Phase 9 — Échantillonnage

Statistiques, détection d'énumérations par cardinalité, détection des clés
étrangères implicites. Optionnel, plafonné, lecture seule.

## Phase 10 — Diff

`ormeau diff` entre deux calques physiques, et `ormeau:synchroniser` entre calque
et entités existantes. Sortie lisible, sortie JSON, code de retour exploitable en
CI.

## Phase 11 — Régénération par AST

Mode avancé : réécriture ciblée avec `nikic/php-parser`, préservation des
méthodes métier et du formatage. Tests de survie des modifications manuelles.

## Phase 12 — Publication

README bilingue, documentation d'installation, avertissement d'usage, subtree
split vers Packagist, image Docker multi-arch, binaires de version.

Faite en avance : publier une version dès qu'il y avait un usage réel valait
mieux que d'attendre la fin. Le split est en place depuis la phase 5 : `php/`
est publié en miroir dans `sprimault/ormeau-doctrine`, inscrit sur Packagist
et mis à jour à chaque push du miroir.

## Chaque dialecte a sa base de test

Le DDL de référence n'est pas un schéma unique au plus petit dénominateur
commun : chaque dialecte reçoit le sien, même schéma logique décliné dans ses
types propres. C'est ce qui rend les aller-retours comparables entre eux, et
c'est la seule façon d'éprouver ce qu'un dialecte est seul à savoir faire.

Il se tient à jour en même temps que les heuristiques, et couvre délibérément
les cas que les outils existants gèrent mal. Un type qui n'y figure pas est
rendu sans avoir jamais été recréé ni relu.

## Hors périmètre v1

SQLite et Oracle (le format doit d'abord se stabiliser sur trois dialectes),
générateurs autres que Doctrine, migrations. Les données des tables, les droits,
les procédures stockées et les triggers ne sont pas dans le périmètre du calque,
jamais.

L'interface n'est ni un client SQL ni un explorateur de données : pas d'éditeur
de requêtes, pas d'affichage de lignes, pas de modification.
