> [🇬🇧 English](CONTRIBUTING.md) · [🇫🇷 Français](CONTRIBUTING.fr.md)

# Contribuer à Ormeau

Les contributions sont bienvenues. Cette page existe pour qu'un correctif bien
écrit ne soit pas refusé au nom d'une règle que personne ne pouvait deviner.

Le projet en est à ses débuts : l'essentiel n'est pas encore écrit, et
[`ROADMAP.md`](ROADMAP.md) dit dans quel ordre il le sera. Ouvrir une issue
avant d'écrire du code évite de travailler sur ce qui est déjà en cours.

## Ce que vise le projet

> L'objectif n'est **pas** de générer des entités à partir d'une base propre —
> un script naïf y suffit. C'est de reprendre une base legacy réelle, et de le
> faire deux fois, six mois plus tard, sans écraser le travail fait entre-temps
> sur les entités.

Toute décision de conception s'arbitre en faveur de cette phrase, et toute pull
request aussi. Un changement qui rend le cas propre plus agréable, sans aider
sur une base réellement tordue, est hors sujet — quelle que soit la qualité du
code.

## Règles non négociables

Elles précèdent toute contribution et ne se discutent pas dans une pull
request. Si l'une d'elles vous gêne, c'est la conception qu'on discute, dans
une issue, avant tout code.

1. **Le calque physique ne perd rien.** Toute information du catalogue
   nécessaire à reconstruire un DDL équivalent doit y figurer. Ce qui n'est pas
   capturé à l'extraction est perdu définitivement : aucune couche en aval ne
   peut le retrouver.
2. **Le calque physique ne juge pas.** Aucun renommage, aucune singularisation,
   aucun retrait de préfixe, aucune inférence de relation. Il constate.
3. **Le calque physique est neutre.** Aucun champ ne suppose la destination. Le
   test : si un générateur EF Core rendait le champ inutilisable ou trompeur,
   il est au mauvais niveau.
4. **L'extraction est déterministe.** Deux extractions de la même base ne
   diffèrent que par `source.extrait_le`, que l'empreinte exclut. Clés triées,
   aucun autre horodatage dans le document. Le mode diff en dépend entièrement.
5. **L'inférence est une fonction pure.** `physique + décisions -> logique`.
   Pas de réseau, pas d'horloge, pas d'aléa, pas d'accès disque hors des
   entrées déclarées. Une heuristique qui aurait besoin d'interroger la base
   est mal placée : ce qu'elle cherche doit figurer dans le physique.
6. **Ce qui n'est pas résolu n'est pas inventé.** Une inférence incertaine
   produit une entrée dans `avertissements` avec son code, sa cible et sa
   confiance. Les avertissements sont une sortie de premier ordre, pas un
   journal.
7. **Toute inférence porte son origine.** Sans ça, l'outil n'est pas auditable,
   et personne ne le lancera sur sa base.
8. **Le générateur ne décide rien.** Il traduit le calque logique. Aucune
   heuristique ne descend dans `php/`.
9. **La régénération ne détruit pas le travail humain.** Méthodes métier,
   docblocks et formatage d'une entité existante sont préservés. On compare
   l'AST, on ne réécrit pas le fichier.
10. **Aucune écriture dans la base introspectée**, jamais, y compris pendant
    l'échantillonnage. Connexion en lecture seule, avec un délai maximal par
    requête.
11. **Aucun cgo.** Tous les pilotes sont en Go pur. Un binaire qui exigerait
    une bibliothèque système sur le serveur d'un client perd l'argument
    principal du projet.

## Deux règles qui coûtent du temps quand on les oublie

**Un champ ajouté au calque appelle trois modifications** : le JSON Schema, les
structures Go, le lecteur PHP. Livrer les trois ensemble — une divergence entre
elles est un défaut, pas un décalage temporaire.

**Une heuristique ajoutée appelle son cas de référence.** On ajoute d'abord le
calque physique qui la déclenche dans `tests/reference/`, puis le code. Ces
fichiers sont le vrai jeu de tests du projet.

## Mise en route

Il faut Go (la version épinglée dans `go.mod`), Node 22 pour l'interface
embarquée, et Docker pour les conteneurs de test. Le premier `make test` installe
lui-même les dépendances du front, par `npm ci`, quand `web/node_modules` est
absent ; après une modification de `web/package-lock.json`, lancer
`cd web && npm ci`. PHP 8.1 et Composer ne servent que pour travailler sur `php/`, qui se
publie en miroir dans `sprimault/ormeau-doctrine` : c'est ici qu'on y contribue,
jamais sur le miroir, réécrit à chaque fusion. Le mécanisme et la publication
d'une version sont décrits dans [`docs/construction.fr.md`](docs/construction.fr.md).

```bash
make outils        # golangci-lint, govulncheck, gosec, tygo
make test          # construction du front, puis go test -race ./...
make lint          # construction du front, contrôle des types générés, golangci-lint, gofmt
make cover         # couverture, détail par fonction
make maj-attendus  # réécrit les calques logiques attendus, à relire ensuite
```

`make test` doit passer sur un clone frais sans Docker : les tests qui exigent
un SGBD portent l'étiquette `integration` et passent par
`make test-integration`.

Le Makefile inclut `makefile.local` s'il existe, pour les réglages propres à une
machine. Si votre antivirus met en quarantaine les binaires au moment du link —
la compilation échoue alors sur un accès refusé sans rapport avec le code —,
c'est là qu'on redirige `GOTMPDIR` et `GOCACHE` vers `.tmp/` :

```make
export GOTMPDIR := $(TMP)/gobuild
export GOCACHE := $(TMP)/gocache
_ := $(shell mkdir -p "$(GOTMPDIR)" "$(GOCACHE)")
```

Le fichier n'est pas versionné, et rien n'oblige à le créer.

Le conteneur peut tourner ailleurs que sur le poste : `ORMEAU_TEST_DSN`
surcharge le DSN visé par les tests d'intégration. Il doit viser une base créée
depuis `tests/ddl/` : elle porte en commentaire l'empreinte du DDL qui l'a
créée, et les tests refusent de tourner si elle manque ou ne correspond pas au
fichier du dépôt. `make containers` recrée le conteneur à chaque appel, volume
compris, pour qu'un DDL modifié soit toujours celui qu'on teste. Son image est
épinglée sur une version mineure de PostgreSQL : la version du serveur entre
dans le calque de référence de `gescom`, et monter de version se fait par
`make maj-calque-gescom`, diff relu (voir `tests/README.fr.md`).

`make aller-retour` vérifie la chaîne entière contre ce même conteneur : base
extraite, entités générées, schéma recréé par Doctrine, comparaison à une liste
fermée d'écarts ; `make aller-retour-sqlserver` fait de même sous SQL Server.
Ils exigent en plus PHP avec `pdo_pgsql` — `pdo_sqlsrv` et le pilote ODBC 18 de
Microsoft pour SQL Server — et `composer install` dans `php/` ; `ORMEAU_PHP` désigne l'interpréteur quand PHP tourne ailleurs que
sur la machine (voir `tests/README.fr.md`). Un changement qui fait apparaître ou
disparaître un écart met à jour la liste dans la même pull request.

Avant d'ouvrir une pull request, lancer au moins `make lint` et `make test`. La
CI les exécute aussi, mais après coup, quand la branche est déjà poussée.

## Style du code

Tout fichier source — `.go`, `.php`, `.ts`, `.tsx` — commence par cet en-tête :

```go
// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0
```

En Go, une ligne vide le sépare du commentaire de paquet, sans quoi il en
deviendrait la documentation. En PHP, il suit `<?php` et précède
`declare(strict_types=1)`.

**Les explications sont en français** : commentaires, godoc, docblocks, TSDoc,
messages d'erreur. Les identifiants suivent le vocabulaire du paquet qu'on
touche — le calque parle français (`Colonne`, `CleEtrangere`), et un paquet ne
mélange pas les deux langues.

**Toute déclaration a sa documentation**, tests et données de test compris :
une ligne quand elle est évidente, un paragraphe quand il y a une décision à
retrouver plus tard. Un commentaire dit pourquoi, jamais ce que la ligne
suivante fait déjà.

## Commits et pull requests

Le préfixe conventionnel est exigé — `feat:`, `fix:`, `test:`, `docs:`,
`refactor:` — parce que l'outillage le lit. **Le français comme l'anglais sont
acceptés** ; écrivez dans la langue qui vous va.

Dire ce que fait le changement et pourquoi, en quelques lignes. Le mécanisme
qu'il a fallu comprendre pour écrire le correctif va en commentaire, dans le
code, là où il sera relu avec lui.

Ce qu'un changement apporte s'écrit aussi dans
[`CHANGELOG.md`](CHANGELOG.md), sous `## [Non publié]`, dans la même pull
request : la section y est relue au moment où elle compte, et la publication en
tire les notes de la version. `make php-changelog` en reporte ensuite la partie
qui concerne le paquet PHP, à partir de la 0.5.0 : la CI refuse une copie qui
n'est pas à jour.

## Ne jamais joindre un calque issu d'une base de production

Un calque porte les noms de tables, de colonnes et les commentaires métier d'un
client, et avec `--echantillonner`, une fois l'échantillonnage livré, des
valeurs réelles. Il n'a sa place ni dans ce dépôt ni en pièce jointe d'une
issue. Les calques versionnés ici ne viennent d'aucune base réelle : ils sont
écrits à la main pour les tests, dans `tests/reference/`.

Si une reproduction en exige un, l'extraire de la base de test construite depuis
`tests/ddl/`, ou écrire à la main les quelques objets qui déclenchent le défaut.

## Sécurité

Ne pas ouvrir d'issue publique pour une faille — voir
[`SECURITY.fr.md`](SECURITY.fr.md) pour le canal privé.
