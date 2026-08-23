> [🇬🇧 English](README.md) · [🇫🇷 Français](README.fr.md)

# Ormeau

![CI](https://github.com/sprimault/ormeau/actions/workflows/ci.yml/badge.svg)
![License](https://img.shields.io/badge/license-Apache%202.0-blue)

> Reprend une vraie base legacy et en produit des entités Doctrine — puis
> recommence six mois plus tard sans écraser le travail fait entre-temps.

> [!WARNING]
> **Pas encore utilisable.** Le format pivot et sa sérialisation sont écrits et
> testés ; le pilote PostgreSQL, l'inférence et le générateur Doctrine, non.
> Chaque commande retourne aujourd'hui une erreur nommant la phase de la feuille
> de route qui l'apportera. Voir [État d'avancement](#état-davancement).

## Ce qu'est Ormeau

Ormeau introspecte un schéma relationnel et en génère des entités Doctrine.

L'objectif n'est **pas** de générer des entités à partir d'une base propre — un
script naïf y suffit, et Doctrine le faisait avant de retirer
`doctrine:mapping:import`. C'est de reprendre une base legacy réelle : préfixes
`T_`, clés étrangères jamais déclarées, tables sans clé primaire, booléens en
`char(1)`, colonnes générées. Et de le faire deux fois, six mois plus tard, sans
écraser le travail fait entre-temps sur les entités.

Toute décision de conception s'arbitre en faveur de cette phrase.

## Pourquoi il existe

Doctrine a retiré son reverse engineering : `doctrine:mapping:import` a disparu du
bundle et `DatabaseDriver` est parti avec ORM 3. Il ne reste rien d'officiel, et
les alternatives de l'écosystème sont soit abandonnées, soit trop naïves pour du
legacy réel — elles produisent `$clientId` en `integer` au lieu d'une association.

## Le calque

```
introspection  ->  calque physique  ->  inférence  ->  calque logique  ->  génération
   (Go)                (JSON)           (Go, pure)         (JSON)            (PHP)
```

Le **calque** est le format pivot : le modèle de données, sérialisé en JSON, qui
décrit une base sans référence ni au SGBD d'origine ni au framework de
destination. Sans lui, on écrit *n × m* traducteurs ; avec lui, *n + m*.

Le nom vient du décalque — une copie fidèle de la structure — et du sens
linguistique : un calque est un emprunt structurel d'une langue vers une autre,
ce qui est exactement l'opération.

Il a deux niveaux, et cette séparation porte toute la conception :

|  | calque physique | calque logique |
|---|---|---|
| décrit | ce qui **est** en base | ce qu'on **décide** d'en faire |
| produit par | introspection | inférence |
| peut être faux | seulement par bug | c'est un jugement, il se discute |
| perd de l'information | non | oui, volontairement |

Le physique est un constat, le logique un jugement. Les séparer est ce qui permet
de corriger une inférence hors ligne et de tester sans aucune base :

```
physique + décisions -> logique
```

Fonction pure : pas de réseau, pas d'horloge, pas d'aléa.

Trois propriétés sont tenues sur le calque physique — **complétude** (un DDL
équivalent doit pouvoir être reconstruit), **neutralité** (aucun champ ne suppose
la destination), **déterminisme** (deux extractions identiques donnent deux
fichiers identiques octet pour octet). La dernière est ce qui rend le mode diff
exploitable plutôt que bruyant.

Détails dans [`docs/architecture.fr.md`](docs/architecture.fr.md). Le contrat
lui-même est dans [`schemas/`](schemas/), versionné à part du dépôt.

## Deux langages, une frontière qui n'est pas arbitraire

**Go** pour l'introspection et l'inférence. Les pilotes sont en Go pur, donc
aucun cgo et aucune dépendance système : un binaire unique qu'on lance sur le
serveur d'un client sans rien installer. L'équivalent PHP imposerait `pdo_sqlsrv`
et Instant Client, et transformerait le projet en support d'installation.

**PHP** pour la génération Doctrine. La partie difficile n'est pas d'écrire des
fichiers, c'est la régénération non destructive : relire une entité déjà
retouchée avec `nikic/php-parser`, comparer au calque logique, ne réécrire que ce
qui a bougé, conserver méthodes métier et formatage.

## Les cas dégueulasses sont le sujet

Table sans clé primaire, clé primaire composite, clé étrangère non déclarée, date
`0000-00-00`, colonne booléenne stockée en `char(1)` valant `O`/`N`, deux tables
liées par des colonnes de types différents. C'est le quotidien d'une base
reprise, et c'est ce que les outils existants gèrent le plus mal.

La règle : produire un avertissement, jamais une exception, jamais une invention.
Un calque logique partiel accompagné de vingt avertissements précis vaut
infiniment mieux qu'une erreur fatale ou qu'un modèle silencieusement faux.

Chaque élément inféré porte son `origine` — `contrainte`, `verification`,
`cardinalite`, `nommage` ou `decision`. Sans ça, l'outil n'est pas auditable, et
personne ne le lancera sur sa base.

## Usage visé

```
ormeau extraire --dsn "postgres://..." --sortie gescom.calque.json
ormeau inferer  gescom.calque.json --decisions decisions.yaml --sortie gescom.logique.json
ormeau diff     gescom.calque.json
ormeau interface

bin/console ormeau:generer      gescom.logique.json
bin/console ormeau:synchroniser gescom.calque.json
```

`ormeau:synchroniser` répond à « qu'est-ce qui a changé en base depuis mes
entités » — l'inverse de `doctrine:schema:update`, et ce qui sert vraiment au
quotidien sur du legacy où le schéma bouge sans passer par les migrations.

## Sûreté

L'outil ne fait que lire. Les connexions sont en lecture seule, y compris pendant
l'échantillonnage, avec un délai maximal par requête. Aucune chaîne SQL ne
transite depuis le navigateur : l'interface locale expose des points d'entrée
fixes.

Le DSN est le seul secret manipulé. Il n'apparaît ni dans les journaux, ni dans
les messages d'erreur, ni dans le calque.

**Un calque est le schéma de la base d'un client** — noms de tables, de colonnes,
commentaires métier, et avec `--echantillonner`, des valeurs réelles. Un calque
extrait d'une base de production ne rentre jamais dans un dépôt, ni en pièce
jointe d'une issue.

## État d'avancement

Rien n'est installable pour l'instant. La feuille de route est dans
[`ROADMAP.md`](ROADMAP.md).

| Phase | État |
|---|---|
| 1 — Calque physique : structures, sérialisation déterministe, empreinte, JSON Schema | Terminée |
| 2 — Introspection PostgreSQL | Requêtes de catalogue écrites, pilote non implémenté |
| 3 — Inférence et calque logique | Structures seulement |
| 4 — Génération Doctrine | Squelette seulement |
| 5 à 11 | Non commencées |

La CI exécute la suite de tests avec le détecteur de courses, `golangci-lint`,
`gofmt`, `govulncheck`, `gosec` et un contrôle de validité des JSON Schema à
chaque push et chaque pull request. Le workflow est public et ses exécutions sont
dans l'onglet Actions.

## Retours

Bogues, demandes ou questions : ouvrir une issue sur
https://github.com/sprimault/ormeau/issues (français de préférence, anglais
bienvenu).

Vous comptez envoyer un correctif ? [`CONTRIBUTING.fr.md`](CONTRIBUTING.fr.md)
énonce les règles sur lesquelles une pull request est jugée — elles ne se
devinent pas à la lecture du code.

Les failles de sécurité passent par le canal privé décrit dans
[`SECURITY.fr.md`](SECURITY.fr.md), jamais par une issue publique.

## Licence

Apache 2.0 — voir [`LICENSE`](LICENSE).
