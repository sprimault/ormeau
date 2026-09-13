> [🇬🇧 English](README.md) · [🇫🇷 Français](README.fr.md)

# ormeau-doctrine

Le bundle Symfony d'[Ormeau](https://github.com/sprimault/ormeau) : il génère
des entités Doctrine à partir du calque logique que produit `ormeau inferer`, et
les régénère sans écraser le travail fait dessus entre-temps.

> [!IMPORTANT]
> Ce dépôt est un **miroir en lecture seule** du répertoire `php/` de
> [sprimault/ormeau](https://github.com/sprimault/ormeau), réécrit à chaque
> fusion. Issues, pull requests et documentation vivent dans le dépôt principal.

## Installation

```console
$ composer require --dev sprimault/ormeau-doctrine
$ bin/console ormeau:generer gescom.logique.json
```

La chaîne complète — extraire le calque, inférer, arbitrer, générer — est
décrite dans le [README principal](https://github.com/sprimault/ormeau/blob/master/README.fr.md).

## Licence

Apache 2.0 — voir [`LICENSE`](LICENSE). Les entités qu'Ormeau génère vous
appartiennent : la licence couvre l'outil, pas sa sortie.
