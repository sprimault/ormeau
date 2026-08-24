> [🇬🇧 English](architecture.md) · [🇫🇷 Français](architecture.fr.md)

# Architecture

Ce document explique la conception. Ce que fait l'outil et comment s'en servir
est dans le [README](../README.fr.md).

## Un format pivot, et pourquoi

```
introspection  ->  calque physique  ->  inférence  ->  calque logique  ->  génération
   (Go)                (JSON)           (Go, pure)         (JSON)            (PHP)
```

C'est une compilation. L'introspecteur est le frontal, un par SGBD ; le
générateur est le back-end, un par ORM ; le calque est le langage intermédiaire.
Sans lui, on écrit *n × m* traducteurs ; avec lui, *n + m*.

Le nom vient du décalque — une copie fidèle de la structure — et du sens
linguistique : un calque est un emprunt structurel d'une langue vers une autre,
ce qui est exactement l'opération.

## Deux niveaux

|  | calque physique | calque logique |
|---|---|---|
| décrit | ce qui **est** en base | ce qu'on **décide** d'en faire |
| produit par | introspection | inférence |
| édité à la main | jamais | jamais |
| régénérable | en relisant la base | à partir du physique et des décisions |
| perd de l'information | non | oui, volontairement |

La séparation tient à une asymétrie : le physique est un **constat**, il ne peut
être faux que par bug ; le logique est un **jugement**, il peut être discutable.
Les mélanger interdirait toute correction hors ligne et rendrait les tests
dépendants d'un serveur.

D'où la relation qui porte tout :

```
physique + décisions -> logique
```

Fonction pure : pas de réseau, pas d'effet de bord, pas d'horloge. C'est elle
qui permet de corriger une inférence sans accès à la base, de la rejouer, et de
tester avec des fichiers de référence plutôt qu'un serveur.

## Les trois propriétés du calque physique

**Complétude.** Il doit contenir tout ce qui est nécessaire pour reconstruire un
DDL équivalent à l'original. C'est testable : `base -> calque -> DDL -> base`
doit donner un diff structurel vide. Ce qui n'y figure pas est perdu — aucune
couche en aval ne peut le retrouver.

**Neutralité.** Aucun champ ne suppose la destination. `type_normalise:
"decimal"` est neutre ; `type_doctrine` ne l'est pas, et vit dans le logique.

**Déterminisme.** Deux extractions de la même base produisent deux fichiers
identiques octet pour octet. Clés triées, aucun horodatage dans le corps du
document. Sans ça, le mode diff produit du bruit et devient inutilisable.

## Le calque logique n'est pas neutre

Il ne parle pas Doctrine, mais il parle le vocabulaire de la famille Hibernate :
entités, associations avec côté propriétaire et côté inverse, stratégies
d'héritage, identifiants composites. Doctrine et EF Core le partagent ; un ORM
bâti sur un autre modèle, comme GORM, consomme le calque physique.

Ce biais est assumé. Un calque logique vraiment universel serait si pauvre qu'il
ne porterait plus aucune décision, et chaque générateur réimplémenterait les
heuristiques : exactement ce que le découpage vise à éviter.

## Pourquoi deux langages

**Go** pour l'introspection et l'inférence. Pilotes en Go pur (`pgx`,
`go-sql-driver/mysql`, `microsoft/go-mssqldb`, `sijms/go-ora`), donc aucun cgo,
aucune dépendance système : un binaire qui tourne tel quel là où on le pose. Le
plus souvent sur le poste du développeur, qui joint la base par le réseau ; sur
le serveur du client quand elle n'est pas atteignable autrement. L'équivalent
PHP imposerait `pdo_sqlsrv` et Instant Client, et transformerait le projet en
support d'installation.

**PHP** pour la génération Doctrine. La partie difficile n'est pas d'écrire des
fichiers, c'est la régénération non destructive : relire une entité déjà
retouchée avec `nikic/php-parser`, comparer au calque logique, ne réécrire que
ce qui a bougé, conserver méthodes métier et formatage.

Le contrat entre les deux est le calque, validé par un JSON Schema versionné et
publié dans [`schemas/`](../schemas/). Conséquence : n'importe qui peut écrire
un générateur Eloquent ou EF Core sans toucher au cœur, et chaque moitié est
réécrivable indépendamment.

## Ce que le calque ne contient pas

Les données des tables — seulement des statistiques agrégées, optionnelles —,
les droits et rôles, les procédures stockées et triggers, le paramétrage
serveur. Rien de ce qui ne sert pas à produire un modèle objet. Un calque qui
dérive vers le dump complet cesse d'être stable, diffable et versionnable dans
Git.

## Versionnement du format

`version_ri` est un entier, pas du semver : la seule question qui compte est de
savoir si un générateur sait lire le fichier. Ajouter un champ optionnel ne
change pas la version ; tout le reste l'incrémente — renommage, suppression,
changement de sémantique, nouvelle valeur dans un vocabulaire fermé. Le
générateur refuse une version supérieure à celle qu'il connaît.

Le champ `empreinte` a deux rôles distincts : dans `source`, il identifie un
état de base ; dans le calque logique, `empreinte_physique` indique de quel
constat découle ce jugement — ce qui permet de détecter « la base a bougé depuis
la dernière génération » sans relire les entités.

## Conséquence pratique

On peut extraire une base chez un client, repartir avec un calque de quelques
mégaoctets, et faire toute la génération et les itérations d'heuristiques
ailleurs, sans accès réseau à cette base. Le calque se versionne dans Git, se
rejoue et se compare entre deux interventions.
