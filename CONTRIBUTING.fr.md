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
4. **L'extraction est déterministe.** Deux extractions de la même base
   produisent deux fichiers identiques octet pour octet. Clés triées, aucun
   horodatage dans le corps du document. Le mode diff en dépend entièrement.
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

Il faut Go (la version épinglée dans `go.mod`) et Docker pour les conteneurs de
test. PHP 8.3 et Composer ne servent que pour travailler sur `php/`.

```bash
make outils   # golangci-lint, govulncheck, gosec
make test     # go test -race ./...
make lint     # golangci-lint, gofmt
```

`make test` doit passer sur un clone frais sans Docker : les tests qui exigent
un SGBD portent l'étiquette `integration` et passent par
`make test-integration`.

Avant d'ouvrir une pull request, lancer au moins `make lint` et `make test`. La
CI les exécute aussi, mais après coup, quand la branche est déjà poussée.

## Commits et pull requests

Le préfixe conventionnel est exigé — `feat:`, `fix:`, `test:`, `docs:`,
`refactor:` — parce que l'outillage le lit. **Le français comme l'anglais sont
acceptés** ; écrivez dans la langue qui vous va.

Dire ce que fait le changement et pourquoi, en quelques lignes. Le mécanisme
qu'il a fallu comprendre pour écrire le correctif va en commentaire, dans le
code, là où il sera relu avec lui.

## Ne jamais joindre un calque issu d'une base de production

Un calque porte les noms de tables, de colonnes et les commentaires métier d'un
client, et avec `--echantillonner`, des valeurs réelles. Il n'a sa place ni
dans ce dépôt ni en pièce jointe d'une issue. Les seuls calques versionnés ici
sont ceux produits depuis `tests/ddl/`.

Si une reproduction en exige un, le construire depuis `tests/ddl/` ou le
réduire aux quelques objets qui déclenchent le défaut.

## Sécurité

Ne pas ouvrir d'issue publique pour une faille — voir
[`SECURITY.fr.md`](SECURITY.fr.md) pour le canal privé.
