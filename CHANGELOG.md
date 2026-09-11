# Journal des versions

Le format suit [Keep a Changelog](https://keepachangelog.com/fr/1.1.0/).

SemVer avec la clause du zéro : **en `0.x`, rien n'est imposé**. Le mineur
marque une phase de la feuille de route, pas une rupture d'API — tout le reste
s'accumule en correctif, correctifs, fonctionnalités et ruptures confondus. Le
numéro ne prévient donc de rien : ce qu'un calque enregistré, un fichier de
décisions ou un projet généré doit reprendre se dit en tête de section.

Trois numéros à ne pas confondre :

| Numéro | Où | Ce qu'il suit |
|---|---|---|
| version du dépôt | tag git, image Docker | le binaire et le paquet PHP, ensemble |
| `version_ri` | chaque calque | le format pivot, seul |
| version Packagist | tag du subtree split | la version du dépôt |

`version_ri` est un entier : ajouter un champ optionnel ne l'incrémente pas,
tout le reste l'incrémente. Une version peut sortir sans qu'il bouge ; il ne
bouge jamais sans version.

Un titre de section s'écrit `## [version] — date — titre`, et le titre d'une
version mineure reprend celui de la phase de [`ROADMAP.md`](ROADMAP.md) qu'elle
termine. La publication en tire le nom et les notes de la version : ce qui est
relu ici est ce qui sera lu sur la page des versions, et il n'y a rien à
recopier ensuite. **Une section absente arrête la publication.**

La section en cours s'écrit `## [Non publié]`, sans date ni titre, et chaque lot
y ajoute ce qu'il change : écrite au fil de l'eau, elle est relue en pull request
avec le code qu'elle décrit. **Elle prend son numéro et sa date au moment du
tag**, faute de quoi la publication cherche une section qui n'existe pas.

Chaque section est **bilingue, français d'abord, séparé par `***`**. Ce
préambule reste en français : il n'est jamais publié.

## [Non publié]

### Corrigé

- **Le refus d'un calque absent nomme le répertoire de travail.** Lancée
  ailleurs que dans le projet, l'interface annonçait un calque manquant sans
  dire où elle l'avait cherché.

***

### Fixed

- **A missing layer is now refused with the working directory named.** Launched
  outside the project, the interface reported a missing layer without saying
  where it had looked for it.

## [0.4.0] — 2026-09-11 — Interface de sélection et d'arbitrage

**Si vous avez déjà un fichier de décisions** : `relations_forcees` est
désormais appliqué. Une relation écrite dans le fichier devient une
association, et l'emporte sur une clé étrangère déclarée sur la même colonne ;
en 0.3.0, la section était lue sans effet. Seuls `plusieurs_vers_un` et
`un_vers_un` s'appliquent : un autre genre produit un avertissement.

### Ajouté

- **`ormeau interface`** ouvre un front local embarqué dans le binaire :
  connexion par composants ou par DSN, arbre des schémas et des tables,
  extraction en tâche de fond et annulable.
- **L'onglet d'arbitrage** travaille hors ligne, sur le calque, et écrit
  `<base>.decisions.yaml`.
- **`colonnes_ignorees`** retire une propriété de l'entité sans rien retirer du
  calque.
- **Le fichier de décisions porte une empreinte en tête.** L'interface demande
  confirmation avant d'écraser un fichier retouché à la main ; la ligne de
  commande ne réécrit toujours pas un fichier existant.

### À savoir

- La génération d'entités Doctrine reste à écrire : la chaîne s'arrête au
  calque logique.
- Calques produits en `version_ri` 1, inchangé depuis la 0.2.0.
- Les binaires ne sont ni signés ni notariés — SmartScreen sous Windows et
  Gatekeeper sous macOS protesteront au premier lancement. La raison est dans
  `docs/construction.fr.md`. Vérifier une archive :
  `gh attestation verify ormeau_v0.4.0_linux_amd64.tar.gz --repo sprimault/ormeau`

***

**If you already have a decisions file**: `relations_forcees` is now applied. A
relation written in the file becomes an association, and overrides a foreign
key declared on the same column; in 0.3.0 the section was read with no effect.
Only `plusieurs_vers_un` and `un_vers_un` apply: any other kind produces a
warning.

### Added

- **`ormeau interface`** opens a local front end embedded in the binary:
  connection by fields or DSN, a tree of schemas and tables, background
  extraction that can be cancelled.
- **The review tab** works offline, on the layer, and writes
  `<base>.decisions.yaml`.
- **`colonnes_ignorees`** drops a property from the entity without removing
  anything from the layer.
- **The decisions file carries a fingerprint at its top.** The interface asks
  before overwriting a file edited by hand; the command line still never
  rewrites an existing file.

### Good to know

- Doctrine entity generation remains to be written: the chain stops at the
  logical layer.
- Layers produced in `version_ri` 1, unchanged since 0.2.0.
- The binaries are neither signed nor notarized — SmartScreen on Windows and
  Gatekeeper on macOS will complain on first launch. The reason is in
  `docs/construction.md`. Verifying an archive:
  `gh attestation verify ormeau_v0.4.0_linux_amd64.tar.gz --repo sprimault/ormeau`

## [0.3.0] — 2026-08-24 — Inférence et calque logique

L'inférence est écrite. `ormeau inferer` lit un calque physique et produit le
calque logique : entités, associations avec leurs deux côtés, héritage par clé
primaire étrangère, tables de jointure, énumérations, traits d'horodatage.

Au premier passage, un fichier de décisions prérempli est écrit, entièrement en
commentaire. Préfixes de tables et singularisation y sont proposés, jamais
appliqués : un nom de table est un constat, et une base ne dit pas sa langue.
Le fichier n'est jamais réécrit ensuite.

La génération d'entités Doctrine reste à écrire : la chaîne s'arrête au calque
logique.

Calques produits en `version_ri` 1, inchangé depuis la 0.2.0.

Les binaires ne sont ni signés ni notariés — SmartScreen sous Windows et
Gatekeeper sous macOS protesteront au premier lancement. La raison est dans
`docs/construction.fr.md`.

Vérifier une archive :

    gh attestation verify ormeau_v0.3.0_linux_amd64.tar.gz --repo sprimault/ormeau

***

Inference is implemented. `ormeau inferer` reads a physical layer and produces
the logical one: entities, associations with both sides, inheritance through a
foreign primary key, join tables, enumerations, timestamp traits.

On the first pass a pre-filled decisions file is written, entirely commented
out. Table prefixes and singularisation are suggested there, never applied: a
table name is an observation, and a database does not state its language. The
file is never rewritten afterwards.

Doctrine entity generation remains to be written: the chain stops at the
logical layer.

Layers produced in `version_ri` 1, unchanged since 0.2.0.

The binaries are neither signed nor notarized — SmartScreen on Windows and
Gatekeeper on macOS will complain on first launch. The reason is in
`docs/construction.md`.

Verifying an archive:

    gh attestation verify ormeau_v0.3.0_linux_amd64.tar.gz --repo sprimault/ormeau

## [0.2.0] — 2026-08-23 — Extraction PostgreSQL

Première version publiée. Ormeau extrait le catalogue d'une base PostgreSQL et
en produit un calque. L'inférence et la génération d'entités Doctrine ne sont
pas encore écrites : leurs commandes retournent une erreur nommant leur phase.

Les binaires ne sont ni signés ni notariés — SmartScreen sous Windows et
Gatekeeper sous macOS protesteront au premier lancement. La raison est dans
`docs/construction.fr.md`.

Calques produits en `version_ri` 1.

Vérifier une archive :

    gh attestation verify ormeau_v0.2.0_linux_amd64.tar.gz --repo sprimault/ormeau

***

First published version. Ormeau extracts the catalogue of a PostgreSQL database
and produces a layer from it. Inference and Doctrine entity generation are not
written yet: their commands return an error naming their phase.

The binaries are neither signed nor notarized — SmartScreen on Windows and
Gatekeeper on macOS will complain on first launch. The reason is in
`docs/construction.md`.

Layers are produced with `version_ri` 1.

To verify an archive:

    gh attestation verify ormeau_v0.2.0_linux_amd64.tar.gz --repo sprimault/ormeau
