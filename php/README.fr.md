> [🇬🇧 English](README.md) · [🇫🇷 Français](README.fr.md)

# ormeau-doctrine

Le bundle Symfony d'[Ormeau](https://github.com/sprimault/ormeau) : il génère
des entités Doctrine à partir d'une base legacy, et les régénère six mois plus
tard sans écraser le code écrit dessus entre-temps.

Doctrine a retiré `doctrine:mapping:import`, et rien d'officiel ne l'a
remplacé. Ormeau reprend une vraie base legacy — préfixes `T_`, clés étrangères
jamais déclarées, booléens en `char(1)` — et produit des associations là où une
traduction littérale laisserait des colonnes entières.

> [!IMPORTANT]
> Ce dépôt est un **miroir en lecture seule** du répertoire `php/` de
> [sprimault/ormeau](https://github.com/sprimault/ormeau), réécrit à chaque
> fusion. Issues, pull requests et documentation vivent dans le dépôt principal.

## Ce qu'il lui faut

Le bundle ne lit pas la base. Il lit un calque logique produit par le binaire
`ormeau`, un seul fichier à télécharger depuis les
[versions](https://github.com/sprimault/ormeau/releases/latest), sans rien
d'autre à installer. PostgreSQL pour l'instant :

```console
$ ormeau extraire --dsn "postgres://app:secret@srv:5432/gescom" --sortie gescom.calque.json
$ ormeau inferer gescom.calque.json
```

`ormeau inferer` écrit `gescom.logique.json`, et un fichier de décisions où
l'on tranche ce que l'outil ne peut pas décider seul.

## Installation

```console
$ composer require --dev sprimault/ormeau-doctrine
$ bin/console ormeau:generer gescom.logique.json
```

La chaîne complète — extraire le calque, inférer, arbitrer, générer — est
décrite dans le [README principal](https://github.com/sprimault/ormeau/blob/master/README.fr.md).

## D'où vient le nom

Un ormeau, c'est un jeune orme — et un coquillage à la coquille nacrée, faite de
couches superposées. Il commence aussi par ORM, ce qui tombe bien pour un outil
qui produit des entités ORM.

Le format pivot s'appelle un **calque**, au sens du décalque : une copie fidèle
du catalogue, sans interprétation. En linguistique, un calque est aussi un
emprunt structurel d'une langue vers une autre — « gratte-ciel » calqué sur
*skyscraper*. C'est exactement l'opération : emprunter la structure d'un schéma
relationnel dans le système de types d'un autre langage.

## Licence

Apache 2.0 — voir [`LICENSE`](LICENSE). Les entités qu'Ormeau génère vous
appartiennent : la licence couvre l'outil, pas sa sortie.
