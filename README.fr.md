> [🇬🇧 English](README.md) · [🇫🇷 Français](README.fr.md)

# Ormeau

![CI](https://github.com/sprimault/ormeau/actions/workflows/ci.yml/badge.svg)
![License](https://img.shields.io/badge/license-Apache%202.0-blue)

> Reprend une vraie base legacy et en produit des entités Doctrine — puis
> recommence six mois plus tard sans écraser le travail fait entre-temps.

> [!WARNING]
> Ormeau se connecte aux bases dont vous lui donnez les identifiants et écrit
> des fichiers décrivant leur schéma. Il ne lit que le catalogue, en session de
> lecture seule imposée par le serveur, mais la base visée et le compte employé
> restent votre choix. Un calque décrit le schéma d'une base cliente : il ne se
> versionne ni ne se transmet à la légère. Préversion fournie sans garantie, aux
> termes de la licence Apache 2.0 — voir
> [État d'avancement](#état-davancement) pour ce qui fonctionne.

## Ce qu'est Ormeau

L'objectif n'est **pas** de générer des entités à partir d'une base propre — un
script y suffit, et Doctrine le faisait avant de retirer
`doctrine:mapping:import`. C'est de reprendre une base legacy réelle : préfixes
`T_`, clés étrangères jamais déclarées, tables sans clé primaire, booléens en
`char(1)`, colonnes générées. Et de le faire deux fois, six mois plus tard, sans
écraser le travail fait entre-temps sur les entités.

Cet outil vient d'un constat de terrain : reprendre une base existante pour en
tirer des entités Doctrine est une tâche récurrente en mission, et plus rien ne
la couvre depuis le retrait de `doctrine:mapping:import`.

Doctrine a retiré son reverse engineering : `doctrine:mapping:import` a disparu
du bundle et `DatabaseDriver` est parti avec ORM 3. Il ne reste rien d'officiel,
et les alternatives de l'écosystème sont soit abandonnées, soit s'arrêtent à une
transposition littérale : elles rendent `$clientId` en `integer` là où il
faudrait une association vers `Client`.

## Démarrage

Télécharger l'archive de sa plateforme depuis la
[dernière version](https://github.com/sprimault/ormeau/releases/latest), la
décompresser, exécuter. Rien d'autre à installer : aucun runtime, aucun pilote
système.

```console
$ ormeau extraire --dsn "postgres://app:secret@srv:5432/gescom" --sortie gescom.calque.json
gescom.calque.json : 10 table(s), 32 colonne(s), 0 anomalie(s)
empreinte sha256:f422f6d3e5eb455a91b096bd513bd5d8e595bd4e88aa588ef25d241993e201a1
```

La connexion s'exprime aussi par composants, ce qui évite d'échapper un mot de
passe dans une URL. Le mot de passe n'a pas de drapeau : il serait visible dans
`ps` et dans l'historique du shell.

```console
$ export ORMEAU_MDP=secret
$ ormeau extraire --sgbd postgres --hote srv --utilisateur app --base gescom --sortie gescom.calque.json
```

Sans `--base`, toutes les bases du serveur sont extraites et `--sortie` désigne
un répertoire :

```console
$ ormeau extraire --sgbd postgres --hote srv --utilisateur app --sortie calques/
calques/gescom.calque.json : 10 table(s), 32 colonne(s), 0 anomalie(s)
calques/facturation.calque.json : 24 table(s), 187 colonne(s), 0 anomalie(s)
```

Vérifier une archive téléchargée — les binaires n'étant ni signés ni notariés,
SmartScreen et Gatekeeper protesteront au premier lancement :

```console
$ gh attestation verify ormeau_v0.3.0_linux_amd64.tar.gz --repo sprimault/ormeau
```

Par conteneur, en lui donnant l'identité de l'appelant : l'image tourne sous un
utilisateur non privilégié et n'écrirait pas dans un volume Linux sans cela.

```console
$ docker run --rm --user "$(id -u):$(id -g)" \
    -e ORMEAU_DSN -v "$PWD:/sortie" \
    ghcr.io/sprimault/ormeau:v0.3.0 extraire --sortie /sortie/gescom.calque.json
```

Le calque en main, l'inférence tourne hors ligne — plus besoin de la base :

```console
$ ormeau inferer gescom.calque.json
gescom.decisions.yaml écrit, entièrement en commentaire : rien n'est appliqué tant que
vous n'avez pas décommenté. Il porte les renommages que l'outil propose.

gescom.logique.json : 12 entité(s), 31 association(s), 2 énumération(s), 1 trait(s)

3 avertissement(s) :
  table_sans_cle_primaire  public.t_log_import   aucune clé primaire : Doctrine refusera cette entité en l'état
  prefixe_detecte          public                préfixe T_ commun aux 12 tables, conservé ; prefixes_a_retirer le retirerait
  singulier_ambigu         public.categories     categories rendu Category par la règle anglaise ; le français donnerait Categorie
```

On décommente ce qui convient dans `gescom.decisions.yaml`, on relance, et le
fichier n'est jamais réécrit : les arbitrages se rejouent à chaque passage.

Entre deux versions, `go install github.com/sprimault/ormeau/cmd/ormeau@master`.

### Sous Windows

Les commandes sont les mêmes, à trois détails près : le binaire s'appelle
`.\ormeau.exe`, les variables d'environnement se posent autrement, et `--user`
est inutile — un volume Windows ne porte pas de permissions POSIX.

```powershell
$env:ORMEAU_MDP = "secret"
.\ormeau.exe extraire --sgbd postgres --hote srv --utilisateur app --base gescom --sortie gescom.calque.json

docker run --rm -e ORMEAU_DSN -v "${PWD}:/sortie" `
    ghcr.io/sprimault/ormeau:v0.3.0 extraire --sortie /sortie/gescom.calque.json
```

Vérifier l'empreinte d'une archive téléchargée :

```powershell
$attendu = (Select-String -Path SHA256SUMS -Pattern windows).Line.Split(" ")[0]
$obtenu  = (Get-FileHash ormeau_v0.3.0_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
if ($attendu -eq $obtenu) { "empreinte OK" } else { "EMPREINTE DIFFERENTE" }
```

## Ce que produit l'extraction

Un **calque** : le décalque du catalogue, en JSON, qui ne juge rien et ne perd
rien. Le nom du type est celui du serveur, jamais reconstruit ; la valeur par
défaut est structurée, pour que `DEFAULT 'now()'` et `DEFAULT now()` restent
distinguables ; longueur et précision sont absentes plutôt que nulles, pour que
`decimal(10,0)` ne se confonde pas avec `int`.

```json
{
  "nom": "cli_statut",
  "position": 4,
  "type_brut": "character varying(20)",
  "type_normalise": "texte",
  "longueur": 20,
  "nullable": false,
  "defaut": { "genre": "litteral", "valeur": "ACTIF" }
}
```

Deux extractions de la même base produisent deux fichiers **identiques octet
pour octet** — l'horodatage est exclu de l'empreinte. C'est ce qui rend le mode
diff exploitable, et ce qui permet de versionner un calque dans Git.

### Trois fichiers par base

```
gescom.calque.json       le constat, extrait de la base
gescom.decisions.yaml    vos arbitrages, écrits une fois
gescom.logique.json      le modèle objet, consommé par le générateur
```

Ils vivent dans votre projet et s'y commitent. Seul le premier vient de la base ;
les deux autres se recalculent hors ligne, ce qui permet de corriger une
inférence sans rouvrir la connexion — et de le faire sur un calque rapporté de
chez un client.

`decisions.yaml` est la pièce qui rend la deuxième extraction supportable. Six
mois plus tard, le schéma a bougé, mais vos arbitrages se rejouent dessus : ce
que vous avez corrigé une fois ne se reperd pas.

Un fichier par base, jamais un fichier commun : deux bases peuvent avoir une
table `client` sans que le même renommage convienne aux deux.

La conception derrière ce format — ses deux niveaux, ses trois propriétés, le
partage entre Go et PHP — est dans
[`docs/architecture.fr.md`](docs/architecture.fr.md).

## Les cas tordus sont le sujet

Table sans clé primaire, clé primaire composite, clé étrangère non déclarée,
date `0000-00-00`, colonne booléenne stockée en `char(1)` valant `O`/`N`, deux
tables liées par des colonnes de types différents. C'est le quotidien d'une base
reprise, et ce que les outils existants gèrent le plus mal.

La règle : produire un avertissement, jamais une exception, jamais une
invention. Un calque logique partiel accompagné de vingt avertissements précis
vaut mieux qu'une erreur fatale ou qu'un modèle silencieusement faux. Chaque
élément inféré porte son `origine` — `contrainte`, `verification`,
`cardinalite`, `nommage` ou `decision` — sans quoi l'outil n'est pas auditable.

## Sûreté

L'outil ne fait que lire. Les connexions sont en lecture seule, imposées par le
serveur et non par la discipline du code, avec un délai maximal par requête.

Le DSN est le seul secret manipulé : il n'apparaît ni dans les journaux, ni dans
les messages d'erreur, ni dans le calque.

Ormeau n'ouvre que deux connexions : votre base, et `127.0.0.1` pour l'interface
locale. Aucune télémétrie, aucun appel sortant, pas même une vérification de
version — le code est public, et `netstat` le confirme le temps d'une
extraction.

**Un calque est le schéma de la base d'un client** — noms de tables, de
colonnes, commentaires métier, et avec `--echantillonner`, des valeurs réelles.
Un calque extrait d'une base de production ne rentre jamais dans un dépôt, ni en
pièce jointe d'une issue.

## État d'avancement

L'extraction PostgreSQL et l'inférence fonctionnent : `ormeau extraire` puis
`ormeau inferer` produisent les trois fichiers. La génération d'entités n'est
pas écrite, et sa commande retourne une erreur nommant la phase qui l'apportera.
L'état par phase est dans [`ROADMAP.md`](ROADMAP.md).

La CI exécute la suite de tests avec le détecteur de courses, `golangci-lint`,
`gofmt`, `govulncheck`, `gosec` et un contrôle de validité des JSON Schema à
chaque push et chaque pull request.

## Pour aller plus loin

- [`docs/architecture.fr.md`](docs/architecture.fr.md) — le calque, ses deux
  niveaux, pourquoi deux langages
- [`docs/construction.fr.md`](docs/construction.fr.md) — compilation croisée,
  images multi-arch, signature
- [`schemas/`](schemas/) — le contrat public, versionné à part
- [`CONTRIBUTING.fr.md`](CONTRIBUTING.fr.md) — les règles sur lesquelles une
  pull request est jugée

## Retours

Bogues, demandes ou questions : ouvrir une issue sur
https://github.com/sprimault/ormeau/issues (français de préférence, anglais
bienvenu).

Les failles de sécurité passent par le canal privé décrit dans
[`SECURITY.fr.md`](SECURITY.fr.md), jamais par une issue publique.

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

Apache 2.0 — voir [`LICENSE`](LICENSE).

**Les entités, calques et fichiers de décisions produits par Ormeau vous
appartiennent.** La licence couvre l'outil, pas sa sortie : rien de ce qu'il
génère n'entre dans votre projet avec une obligation attachée.
