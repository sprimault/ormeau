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

**`version_ri` passe à 2.** Un calque physique déjà enregistré se lit
toujours. Un calque logique en version 1 est refusé par le paquet PHP : il se
recalcule hors ligne par `ormeau inferer <base>.calque.json --decisions
<base>.decisions.yaml --sortie <base>.logique.json`, que le message donne. Le
calque logique perd `type_php`, la stratégie d'identifiant `aucune` et
`orphelins_supprimes`, que rien ne produisait ; un générateur tiers déduit le
type PHP de `type_doctrine`. Le calque physique ne reporte plus l'index qui
soutient une contrainte d'unicité, que la contrainte dit déjà : il est retiré à
la lecture d'un calque en version 1.

**Un type Doctrine forcé que le générateur ne connaît pas se déclare `mixed`**
dans `Base/`, au lieu du type PHP de la colonne d'origine, qui levait une
`TypeError` dès que le type rendait un objet.

### Corrigé

- **Un texte Unicode forcé en `json` est signalé.** Doctrine le recrée en
  `VARCHAR(MAX)`, et la décision passait sans avertissement :
  `json_sans_unicode` dit que la conversion proposée par `migrations:diff` ne
  doit pas s'appliquer, et qu'une base recréée perd l'Unicode écrit hors
  Doctrine.
- **Un défaut `getdate()` de SQL Server est reporté.** Il n'était pas reconnu :
  la date, l'heure ou l'instant était à fournir par l'application. Sur une date
  ou une heure, Doctrine écrit l'instant courant sous DBAL 4.4 ; avant, il ne
  sait l'écrire qu'en `CONVERT(date, GETDATE())`, que `migrations:diff`
  propose de recréer à chaque passage sans rien changer — le rapport de
  génération le dit. `getdate()` sur un `datetimeoffset`, qui stocke l'heure
  locale avec un décalage nul, et `sysdatetimeoffset()`, que Doctrine n'écrit
  pas, restent signalés par `defaut_non_reporte`.
- **Une clé SQL Server tirée d'une séquence est attribuée.** Un `DEFAULT NEXT
  VALUE FOR` n'était pas reconnu : l'identifiant restait à fournir par
  l'application. Sous ORM 2, la clé passe par la séquence nommée, avec sa
  valeur de départ ; sous ORM 3, qui ignore le générateur d'une classe de base
  mappée, par un générateur produit dans `Base/Generateur/`, que les outils de
  schéma de Doctrine ne voient pas — le rapport de génération le dit.
- **Un `datetimeoffset` écrit garde son décalage.** Rendu sans fuseau, il était
  écrit sans décalage : l'instant stocké était faux, sans erreur. Il est rendu
  en `datetimetz_immutable` ; au-delà de six décimales, précision par défaut
  de SQL Server, Doctrine ne sait pas le relire, et l'avertissement
  `fuseau_precision_non_lue` le dit.
- **Une chaîne ou un binaire de longueur fixe le reste.** Seul le
  `character(n)` de PostgreSQL était reconnu : un `nchar(n)` ou un `binary(n)`
  de SQL Server était recréé en longueur variable, sans les espaces ou octets
  nuls qui complètent ses valeurs. Un binaire de longueur déclarée est rendu en
  `binary` plutôt qu'en `blob`, et `tinyint` en `smallint` plutôt qu'en
  `integer` : `migrations:diff` proposait sinon d'élargir la colonne.
- **Un texte sans longueur déclarée est rendu en `text`.** Seuls les types
  nommés `text` l'étaient : un `varchar(max)` de SQL Server ou un `varchar`
  sans longueur de PostgreSQL devenaient une chaîne, recréée en
  `VARCHAR(255)`, que `migrations:diff` proposait d'appliquer à la base.
  `nvarchar(max)` et `ntext` restent en chaîne, avec l'avertissement
  `texte_unicode_sans_equivalent` : Doctrine les recréerait en `VARCHAR(MAX)`,
  et la conversion remplacerait sans erreur tout caractère hors de la page de
  code.
- **Une connexion SQL Server par composants atteignait `master` au lieu de la
  base demandée.** Le chemin d'une URL y désigne l'instance nommée, jamais la
  base : celle-ci passe désormais en paramètre, et l'instance a son propre
  champ. Aucune erreur ne le signalait — l'arbre montrait simplement une autre
  base.
- **Un espace autour de l'hôte ne fait plus échouer la connexion.** Seul champ
  que l'URL recevait tel quel, il la rendait illisible dès qu'un copier-coller
  y laissait un blanc, et le message parlait d'une URL que personne n'avait
  écrite. Les blancs sont retirés ; ce qu'un nettoyage ne rattrape pas est
  refusé en nommant le champ.

### Ajouté

- **Un UUID tiré par la base est reconnu.** `gen_random_uuid()`, `newid()` et
  `newsequentialid()` donnent `defaut_expression: uuid_genere` au lieu de
  `defaut_non_reporte`. Doctrine ne sait pas écrire ce défaut : l'entité le
  laisse à l'application, et l'avertissement `uuid_a_fournir` comme le rapport
  de génération disent de fournir la valeur et de ne pas appliquer la
  suppression du défaut que `migrations:diff` propose.
- **Un texte SQL Server que sa vérification déclare JSON se lit en tableau.**
  Un `varchar(max)` validé par `ISJSON` est rendu en `json`. Un
  `nvarchar(max)` reste en chaîne, Doctrine le recréant en `VARCHAR(MAX)` : le
  fichier de décisions propose de le forcer en `json`, pour qui sait que
  Doctrine est seul à écrire la colonne.
- **Le chiffrement se règle dans le vocabulaire du SGBD visé.** `sslmode` vient
  de libpq et n'existe pas sous SQL Server, qui règle `encrypt` et
  `TrustServerCertificate` ensemble ; le formulaire propose désormais l'un ou
  l'autre selon le serveur, nommés par leur effet. Un serveur sans TLS — le cas
  courant d'une base reprise — se joint en choisissant « désactivé », ce
  qu'aucun réglage ne permettait.
- **SQL Server se connecte et se parcourt depuis l'interface.** Le serveur est
  décrit à la connexion — version, édition, catalogue, schémas —, et l'arbre
  affiche bases, tables et colonnes, avec les commentaires que SQL Server range
  dans des extended properties. Un serveur à certificat auto-signé demande
  `TrustServerCertificate=true` dans le DSN.
- **SQL Server s'extrait en un calque physique.** Tables, colonnes, défauts,
  colonnes calculées stockées ou non, collations, commentaires, clés, unicités,
  `CHECK`, index filtrés, séquences et vues sont lus dans `sys.*`. Le type brut
  est celui que le serveur écrit, échelle d'un `datetime2(n)` comprise ; une
  collation égale à celle de la base sort en `default`. L'inférence et la
  génération n'ont pas encore été relues pour ce dialecte.
- **Deux champs optionnels dans le calque physique**, sans changement de
  `version_ri` : `ordres` sur un index, pour un index descendant qui serait
  sinon recréé ascendant, et `depart` sur une séquence, que le minimum ne dit
  pas. Les deux pilotes les produisent ; absents d'un calque extrait
  auparavant, ils veulent dire « inconnu ».
- **`longueur_fixe` dans le calque physique**, champ optionnel lui aussi : le
  pilote dit si une chaîne ou un binaire est de longueur fixe, que le nom du
  type ne dit pas d'un SGBD à l'autre. Un calque extrait auparavant garde la
  lecture de `character(n)`.
- **`fuseau` et `precision_fractionnaire` dans le calque physique**, optionnels
  aussi : le pilote dit si un horodatage porte son décalage, et combien de
  décimales ses secondes déclarent.
- **La séquence d'un défaut, le départ d'une séquence et le SGBD dans les
  calques**, optionnels : `defaut.sequence` au calque physique, lu dans les
  dépendances du catalogue ; `sequence_depart` et `sgbd` au calque logique, le
  rendu d'une clé par séquence dépendant de la plateforme.

### Modifié

- **L'arbitrage dit où se traite chaque avertissement.** Il comptait « à
  traiter » ce qu'aucune action de l'écran ne règle. Le compteur ne compte plus
  que ce qui se décide sur place, un badge gris compte le reste dans la liste
  des entités, et les autres avertissements disent en tête
  s'ils se font dans la sélection, le fichier de décisions, la base ou le code
  de l'application. Un texte Unicode déclaré JSON se force en `json` d'un clic,
  sous le nouveau code `texte_unicode_json_propose`.

***

**`version_ri` moves to 2.** A physical layer already saved is still read. A
version 1 logical layer is refused by the PHP package: it is recomputed offline
with `ormeau inferer <base>.calque.json --decisions <base>.decisions.yaml
--sortie <base>.logique.json`, which the message gives. The logical layer loses
`type_php`, the `aucune` identifier strategy and `orphelins_supprimes`, which
nothing produced; a third-party generator derives the PHP type from
`type_doctrine`. The physical layer no longer reports the index backing a
unique constraint, which the constraint already states: it is removed when a
version 1 layer is read.

**A forced Doctrine type the generator does not know is declared `mixed`** in
`Base/`, instead of the source column's PHP type, which raised a `TypeError` as
soon as the type returned an object.

### Fixed

- **Unicode text forced to `json` is flagged.** Doctrine recreates it as
  `VARCHAR(MAX)`, and the decision went through without a warning:
  `json_sans_unicode` says the conversion `migrations:diff` offers must not be
  applied, and that a recreated database loses Unicode written outside
  Doctrine.
- **A SQL Server `getdate()` default is carried over.** It was not recognized:
  the application had to supply the date, time or instant. On a date or time
  column, Doctrine writes the current timestamp under DBAL 4.4; before that, it
  can only write `CONVERT(date, GETDATE())`, which `migrations:diff` offers to
  recreate on every run without changing anything — the generation report says
  so. `getdate()` on a `datetimeoffset`, which stores local time with a zero
  offset, and `sysdatetimeoffset()`, which Doctrine does not write, are still
  reported by `defaut_non_reporte`.
- **A SQL Server key drawn from a sequence is assigned.** A `DEFAULT NEXT
  VALUE FOR` was not recognized: the application had to supply the
  identifier. Under ORM 2, the key goes through the named sequence, with its
  start value; under ORM 3, which ignores the generator of a mapped
  superclass, through a generator produced in `Base/Generateur/`, which
  Doctrine's schema tools do not see — the generation report says so.
- **A written `datetimeoffset` keeps its offset.** Rendered without a time
  zone, it was written without its offset: the stored instant was wrong, with
  no error. It is rendered as `datetimetz_immutable`; beyond six decimals, SQL
  Server's default precision, Doctrine cannot read it back, and the
  `fuseau_precision_non_lue` warning says so.
- **A fixed-length string or binary stays fixed.** Only PostgreSQL's
  `character(n)` was recognized: a SQL Server `nchar(n)` or `binary(n)` was
  recreated with a variable length, without the spaces or zero bytes padding
  its values. A binary with a declared length is rendered as `binary` rather
  than `blob`, and `tinyint` as `smallint` rather than `integer`:
  `migrations:diff` otherwise offered to widen the column.
- **A text column with no declared length is rendered as `text`.** Only types
  named `text` were: a SQL Server `varchar(max)` or a PostgreSQL `varchar`
  without length became a string, recreated as `VARCHAR(255)`, which
  `migrations:diff` offered to apply to the database. `nvarchar(max)` and
  `ntext` stay strings, with the `texte_unicode_sans_equivalent` warning:
  Doctrine would recreate them as `VARCHAR(MAX)`, and the conversion would
  silently replace any character outside the code page.
- **A component-based SQL Server connection reached `master` instead of the
  requested database.** There, the path of a URL names the instance, never the
  database: the latter now goes in a parameter, and the instance has its own
  field. Nothing reported it — the tree simply showed another database.
- **A space around the host no longer breaks the connection.** The only field
  the URL received verbatim, it made the string unreadable as soon as a paste
  left a blank around it, and the message spoke of a URL nobody had written.
  Blanks are trimmed; what trimming cannot fix is refused by naming the field.

### Added

- **A UUID drawn by the database is recognized.** `gen_random_uuid()`,
  `newid()` and `newsequentialid()` yield `defaut_expression: uuid_genere`
  instead of `defaut_non_reporte`. Doctrine cannot write this default: the
  entity leaves it to the application, and both the `uuid_a_fournir` warning and
  the generation report say to supply the value and not to apply the default
  removal `migrations:diff` offers.
- **SQL Server text declared JSON by its check constraint is read as an
  array.** A `varchar(max)` validated by `ISJSON` is rendered as `json`. An
  `nvarchar(max)` stays a string, since Doctrine recreates it as
  `VARCHAR(MAX)`: the decisions file offers to force it to `json`, for those
  who know Doctrine is the only writer of the column.
- **Encryption is set in the vocabulary of the target DBMS.** `sslmode` comes
  from libpq and does not exist under SQL Server, which sets `encrypt` and
  `TrustServerCertificate` together; the form now offers one or the other
  depending on the server, named after their effect. A server without TLS — the
  common case of an inherited database — is reached by choosing "disabled",
  which no setting allowed.
- **SQL Server connects and can be browsed from the interface.** The server is
  described on connection — version, edition, catalog, schemas — and the tree
  shows databases, tables and columns, including the comments SQL Server keeps
  in extended properties. A server with a self-signed certificate needs
  `TrustServerCertificate=true` in the DSN.
- **SQL Server extracts into a physical layer.** Tables, columns, defaults,
  persisted or non-persisted computed columns, collations, comments, keys,
  unique constraints, `CHECK`s, filtered indexes, sequences and views are read
  from `sys.*`. The raw type is the one the server writes, including the scale
  of a `datetime2(n)`; a collation equal to the database's comes out as
  `default`. Inference and generation have not been reviewed for this dialect
  yet.
- **Two optional fields in the physical layer**, with no `version_ri` change:
  `ordres` on an index, for a descending index that would otherwise be
  recreated ascending, and `depart` on a sequence, which the minimum does not
  tell. Both drivers produce them; missing from a layer extracted earlier,
  they mean "unknown".
- **`longueur_fixe` in the physical layer**, also optional: the driver tells
  whether a string or binary has a fixed length, which the type name does not
  tell consistently across DBMSs. A layer extracted earlier keeps the
  `character(n)` reading.
- **`fuseau` and `precision_fractionnaire` in the physical layer**, also
  optional: the driver tells whether a timestamp carries its offset, and how
  many decimals its seconds declare.
- **A default's sequence, a sequence's start and the DBMS in the layers**,
  optional: `defaut.sequence` in the physical layer, read from the catalog's
  dependencies; `sequence_depart` and `sgbd` in the logical layer, since the
  rendering of a sequence key depends on the platform.

### Changed

- **Arbitration says where each warning is dealt with.** It counted as "to
  handle" what no action on the screen settles. The counter now only counts
  what is decided there, a grey badge counts the rest in the entity list, and
  the other warnings state first whether they are
  done in the table selection, the decisions file, the database or the
  application code. Unicode text declared JSON is forced to `json` in one
  click, under the new code `texte_unicode_json_propose`.

## [0.6.1] — 2026-09-16 — Ce que rien ne signalait

### Corrigé

- **Les colonnes d'un index sont citées comme elles le sont ailleurs.**
  Doctrine recopie ces noms tels quels dans le `CREATE INDEX` : une colonne
  dont la forme exige des guillemets — majuscules, accents, espaces — y
  arrivait nue, la base la repliait en minuscules et refusait la requête.
  `schema:create` échouait donc sur la table concernée. Un mot réservé, lui,
  était déjà cité par DBAL.
- **Un index dont Doctrine refuse le nom est laissé de côté au lieu
  d'empêcher toute l'entité.** DBAL n'accepte qu'un nom en `[a-zA-Z0-9_]` :
  un index nommé `IX_Libellé hors défaut` faisait échouer `schema:create`
  pour l'entité entière, index compris. Il est désormais omis et signalé au
  rapport, avec ce qu'il en coûte quand il portait une unicité. Le renommer
  en base le rétablit.
- **`nullable` n'est plus écrit sur la colonne de jointure d'un
  identifiant.** Doctrine l'y ignore — la clé primaire rend la colonne
  obligatoire —, et l'écrire est déprécié depuis ORM 3.6, erreur annoncée en
  4.0. Une entité à identifiant dérivé ou à clé partagée change donc dans
  `Base/` à la régénération ; le schéma, lui, ne bouge pas.

### Ajouté

- **La génération prévient quand une hiérarchie jointe ne s'enregistrera
  pas.** Jusqu'à ORM 2.15, son persister garde dans l'INSERT une colonne que
  la base calcule, sans lui donner de valeur : aucune entité de la hiérarchie
  ne s'écrit, racine comprise, et rien ne le signale avant la première
  écriture. Le rapport le dit quand la cible détectée est antérieure à 2.16.
  Les entités restent générées et le code produit ne change pas : la seule
  sortie est ORM 2.16.1, aucune version corrective n'existant sur 2.14 ni
  2.15.

***

### Fixed

- **Index columns are quoted as they are elsewhere.** Doctrine copies those
  names verbatim into the `CREATE INDEX`: a column whose form requires quoting
  — uppercase, accents, spaces — arrived bare, the database folded it to
  lowercase and rejected the statement. `schema:create` therefore failed on
  that table. A reserved word was already quoted by DBAL.
- **An index whose name Doctrine rejects is left out instead of blocking the
  whole entity.** DBAL only accepts a name in `[a-zA-Z0-9_]`: an index named
  `IX_Libellé hors défaut` made `schema:create` fail for the entire entity,
  indexes included. It is now omitted and reported, stating what is lost when
  it carried a unique constraint. Renaming it in the database brings it back.
- **`nullable` is no longer written on the join column of an identifier.**
  Doctrine ignores it there — the primary key already makes the column
  required — and writing it has been deprecated since ORM 3.6, announced as an
  error in 4.0. An entity with a derived identifier or a shared key therefore
  changes in `Base/` on regeneration; the schema itself does not move.

### Added

- **Generation now warns when a joined hierarchy will not persist.** Up to ORM
  2.15, its persister keeps a database-computed column in the INSERT without
  giving it a value: no entity of the hierarchy is written, root included, and
  nothing reports it before the first write. The report says so when the
  detected target is older than 2.16. Entities are still generated and the
  produced code is unchanged: the only way out is ORM 2.16.1, as no fix
  release exists on 2.14 or 2.15.

## [0.6.0] — 2026-09-15 — Aller-retour

**Un calque PostgreSQL déjà enregistré change à sa prochaine extraction** si la
base a des colonnes d'identité, sans que le schéma ait bougé : chacune reçoit
`identite`, et l'empreinte change. `version_ri` ne bouge pas, et un calque plus
ancien reste lisible, nature d'identité inconnue. Ni le calque logique ni les
entités ne changent.

**Une entité dont la clé primaire porte une association change dans `Base/` à
la régénération** : ses membres de clé passent en tête, dans l'ordre de la clé,
et sa colonne de jointure reçoit le commentaire et le défaut de la base. Une
colonne discriminante n'en reçoit que sous ORM 3, et seulement dans une
hiérarchie nouvelle : la classe de l'utilisateur, déjà écrite, ne change pas.

**Sous Doctrine ORM 3 avec DBAL 3, les propriétés `bigint` et `binary` changent
de type dans `Base/` à la régénération** : `string` au lieu de `int`, `mixed`
au lieu de `string`. Le type PHP suit désormais la version de DBAL installée,
et non plus celle d'ORM, et l'annonce en tête d'exécution la donne.

**Une colonne à défaut calculé change dans `Base/` à la régénération**, après
`ormeau inferer` : `now()`, `CURRENT_DATE` et leurs équivalents y reçoivent
leur défaut. Un `now()` d'origine s'écrit `CURRENT_TIMESTAMP`, et
`migrations:diff` propose un `SET DEFAULT CURRENT_TIMESTAMP` de même sens, à ne
pas appliquer. Un trait d'horodatage partagé par des tables dont une partie
seulement a ce défaut se scinde en deux.

**Une entité dont la table a un index partiel change dans `Base/` à la
régénération**, après `ormeau inferer` : son `#[ORM\Index]` ou son
`#[ORM\UniqueConstraint]` reçoit le prédicat dans `options`.

**Une colonne `char(n)`, `jsonb` ou `real` change dans `Base/` à la
régénération**, après `ormeau inferer` : `options: ['fixed' => true]`, type
`jsonb` ou option `jsonb`, type `smallfloat` selon la version de DBAL. Une
colonne discriminante `char(n)` ne le reçoit que sous ORM 3, et seulement dans
une hiérarchie nouvelle.

**Une colonne générée change dans `Base/` à la régénération**, après `ormeau
inferer` : `generated: 'ALWAYS'` et un docblock qui cite son calcul. Doctrine
relit la valeur après chaque écriture ; en contrepartie, le `flush` qui suit
une écriture appelle `#[PreUpdate]` une fois sans rien écrire, et un
`PreUpdate` qui modifie une colonne l'écrit alors une fois de plus.
`schema:create` recrée toujours une colonne ordinaire.

**Un calque PostgreSQL déjà enregistré change à sa prochaine extraction** si
une colonne a une collation hors de `pg_catalog` : elle gagne
`collation_schema`. **Une colonne à collation explicite change dans `Base/`**
après `ormeau inferer` : `options: ['collation' => …]`. Une collation d'un
autre schéma n'est pas écrite. Un `ALTER … TYPE` que `migrations:diff`
propose pour une autre raison fait perdre la collation, DBAL n'y écrivant pas
`COLLATE` : à reprendre à la main avant de l'appliquer.

### Corrigé

- **Sous ORM 3 avec DBAL 3, une entité à colonne `binary` se charge, et une
  entité à colonne `bigint` n'est plus crue modifiée à chaque flush.** Le
  générateur choisissait le type PHP d'après la version d'ORM, alors que DBAL
  3 hydrate une ressource et une chaîne : le chargement levait une
  `TypeError`, et chaque flush émettait un `UPDATE`. `--cible-dbal` s'écarte
  de la version déduite quand `--cible-orm` vise une version non installée.
- **`migrations:diff` ne propose plus de supprimer puis recréer la clé
  primaire d'une table dont la clé commence par une clé étrangère.** Le
  générateur déclarait l'association de clé après les propriétés, et Doctrine
  formait la clé dans l'ordre inverse de la base, sous ORM 2 comme sous ORM 3.
- **Le commentaire et le défaut d'une colonne de jointure qui fait partie de
  la clé ne sont plus perdus**, ni ceux de la colonne discriminante sous
  ORM 3. ORM 2.14 refuse ces options sur la colonne discriminante : elles n'y
  sont pas écrites.
- **`migrations:diff` ne propose plus de supprimer un index partiel pour le
  recréer sans son prédicat.** Appliquée, une unicité partielle devenait plus
  stricte que la base, et refusait des lignes que celle-ci accepte.
- **`migrations:diff` ne propose plus de convertir un `char(n)` en
  `varchar(n)`**, ni sous DBAL 4 un `jsonb` en `json` et un `real` en
  `double precision`. Appliquée, la conversion changeait ce que la colonne
  stocke et comment elle se compare.
- **La valeur d'une colonne générée n'est plus `null` en mémoire après un
  `flush`.** Exclue de l'INSERT et de l'UPDATE, elle n'était relue qu'au
  rechargement de l'entité.
- **Une collation explicite n'est plus perdue par `schema:create`.**

### Ajouté

- **Le calque physique distingue une identité toujours générée d'une identité
  par défaut** (`identite` : `toujours` ou `par_defaut`). `GENERATED ALWAYS`
  refuse une valeur explicite, `GENERATED BY DEFAULT` l'accepte : un DDL
  reconstruit sans cette distinction change le comportement de la table. Le
  vocabulaire se définit par ce comportement, pour valoir aussi pour
  `IDENTITY` de SQL Server et `AUTO_INCREMENT` de MySQL.
- **Le commentaire d'une table de jointure est conservé.** Le calque logique le
  porte (`table_jointure.commentaire`), et `#[ORM\JoinTable]` le reçoit : une
  association plusieurs-vers-plusieurs change dans `Base/` à la régénération
  quand sa table en a un. Il faut relancer `ormeau inferer`.
- **Un défaut calculé est reporté quand son sens est certain**
  (`propriete.defaut_expression` : `horodatage_courant`, `date_courante`,
  `heure_courante`), lu sur l'expression et le type de la colonne. Il est
  écrit en objet `DefaultExpression` sous DBAL 4.4, qui déprécie la chaîne
  `'CURRENT_TIMESTAMP'`, et en chaîne avant. Tout autre défaut calculé,
  `gen_random_uuid()` ou `clock_timestamp()`, produit l'avertissement
  `defaut_non_reporte`. La propriété n'est pas initialisée : persister une
  entité sans lui donner de valeur échoue toujours.
- **Le prédicat d'un index partiel est reporté** (`index.predicat`), tel que
  le catalogue le rend, et écrit dans `options: ['where' => …]`. Un index dont
  le prédicat cite une colonne de `colonnes_ignorees` part avec elle. Le texte
  arrive tel quel dans le DDL de Doctrine : un calque logique reçu d'ailleurs
  se relit comme une migration.
- **La longueur fixe d'une chaîne est reportée** (`propriete.longueur_fixe`),
  et **`jsonb` et `real` gardent leur type** (`type_doctrine` `jsonb` et
  `smallfloat`). Le générateur se replie sur ce que DBAL connaît : `json` avec
  l'option `jsonb` avant DBAL 4.3, `float` avant DBAL 4.1, que `schema:create`
  recrée en `double precision`.
- **Le calcul d'une colonne générée passe dans le calque logique**
  (`propriete.generee` : `expression`, `stockee`), repris tel quel du calque
  physique.
- **Le calque physique garde le schéma d'une collation** hors de `pg_catalog`
  (`colonne.collation_schema`), et le calque logique la collation explicite
  d'une propriété (`propriete.collation`). Une collation d'un autre schéma
  n'y passe pas : Doctrine l'écrirait en un seul identifiant, et son nom seul
  dépendrait du `search_path` de l'application. Elle produit l'avertissement
  `collation_non_reportee`.

***

**A PostgreSQL layer already saved changes on its next extraction** if the
database has identity columns, with no schema change: each gains `identite`,
and the fingerprint changes. `version_ri` does not move, and an older layer
stays readable, with an unknown identity kind. Neither the logical layer nor
the entities change.

**An entity whose primary key carries an association changes in `Base/` on
regeneration**: its key members move first, in key order, and its join column
gets the database comment and default. A discriminator column only gets them
under ORM 3, and only in a new hierarchy: the user class, already written, does
not change.

**Under Doctrine ORM 3 with DBAL 3, `bigint` and `binary` properties change
type in `Base/` on regeneration**: `string` instead of `int`, `mixed` instead
of `string`. The PHP type now follows the installed DBAL version rather than
the ORM one, and the announcement at the start of the run gives it.

**A column with a computed default changes in `Base/` on regeneration**, after
`ormeau inferer`: `now()`, `CURRENT_DATE` and their equivalents get their
default there. An original `now()` is written `CURRENT_TIMESTAMP`, and
`migrations:diff` proposes a `SET DEFAULT CURRENT_TIMESTAMP` with the same
meaning, not to be applied. A timestamp trait shared by tables of which only
some have that default splits in two.

**An entity whose table has a partial index changes in `Base/` on
regeneration**, after `ormeau inferer`: its `#[ORM\Index]` or
`#[ORM\UniqueConstraint]` gets the predicate in `options`.

**A `char(n)`, `jsonb` or `real` column changes in `Base/` on regeneration**,
after `ormeau inferer`: `options: ['fixed' => true]`, the `jsonb` type or
option, the `smallfloat` type, depending on the DBAL version. A `char(n)`
discriminator column only gets it under ORM 3, and only in a new hierarchy.

**A generated column changes in `Base/` on regeneration**, after `ormeau
inferer`: `generated: 'ALWAYS'` and a docblock quoting its computation.
Doctrine reads the value back after every write; in return, the `flush` that
follows a write calls `#[PreUpdate]` once without writing anything, and a
`PreUpdate` that modifies a column then writes it once more. `schema:create`
still recreates an ordinary column.

**A PostgreSQL layer already saved changes on its next extraction** if a column
has a collation outside `pg_catalog`: it gains `collation_schema`. **A column
with an explicit collation changes in `Base/`** after `ormeau inferer`:
`options: ['collation' => …]`. A collation from another schema is not written.
An `ALTER … TYPE` that `migrations:diff` proposes for another reason drops the
collation, since DBAL does not write `COLLATE` there: fix it by hand before
applying it.

### Fixed

- **Under ORM 3 with DBAL 3, an entity with a `binary` column loads, and an
  entity with a `bigint` column is no longer seen as changed on every flush.**
  The generator chose the PHP type from the ORM version, while DBAL 3 hydrates
  a resource and a string: loading raised a `TypeError`, and every flush issued
  an `UPDATE`. `--cible-dbal` departs from the deduced version when
  `--cible-orm` targets a version that is not installed.
- **`migrations:diff` no longer proposes to drop and recreate the primary key
  of a table whose key starts with a foreign key.** The generator declared the
  key association after the properties, and Doctrine built the key in the
  reverse order of the database, under ORM 2 as under ORM 3.
- **The comment and default of a join column that is part of the key are no
  longer lost**, nor those of the discriminator column under ORM 3. ORM 2.14
  refuses these options on the discriminator column: they are not written
  there.
- **`migrations:diff` no longer proposes to drop a partial index and recreate
  it without its predicate.** Once applied, a partial unique index became
  stricter than the database, and rejected rows the database accepts.
- **`migrations:diff` no longer proposes to convert a `char(n)` into
  `varchar(n)`**, nor under DBAL 4 a `jsonb` into `json` and a `real` into
  `double precision`. Once applied, the conversion changed what the column
  stores and how it compares.
- **The value of a generated column is no longer `null` in memory after a
  `flush`.** Left out of the INSERT and the UPDATE, it was only read back when
  the entity was reloaded.
- **An explicit collation is no longer lost by `schema:create`.**

### Added

- **The physical layer tells an always-generated identity from a by-default
  one** (`identite`: `toujours` or `par_defaut`). `GENERATED ALWAYS` refuses an
  explicit value, `GENERATED BY DEFAULT` accepts it: a DDL rebuilt without that
  distinction changes how the table behaves. The vocabulary is defined by that
  behaviour, so that it also holds for SQL Server `IDENTITY` and MySQL
  `AUTO_INCREMENT`.
- **The comment of a join table is kept.** The logical layer carries it
  (`table_jointure.commentaire`), and `#[ORM\JoinTable]` receives it: a
  many-to-many association changes in `Base/` on regeneration when its table
  has one. `ormeau inferer` must be run again.
- **A computed default is carried over when its meaning is certain**
  (`propriete.defaut_expression`: `horodatage_courant`, `date_courante`,
  `heure_courante`), read from the expression and the column type. It is
  written as a `DefaultExpression` object under DBAL 4.4, which deprecates the
  `'CURRENT_TIMESTAMP'` string, and as a string before. Any other computed
  default, `gen_random_uuid()` or `clock_timestamp()`, raises the
  `defaut_non_reporte` warning. The property is not initialised: persisting an
  entity without giving it a value still fails.
- **The predicate of a partial index is carried over** (`index.predicat`), as
  the catalogue renders it, and written into `options: ['where' => …]`. An
  index whose predicate mentions a column from `colonnes_ignorees` goes with
  it. The text reaches Doctrine's DDL unchanged: a logical layer received from
  elsewhere is to be reviewed like a migration.
- **The fixed length of a string is carried over** (`propriete.longueur_fixe`),
  and **`jsonb` and `real` keep their type** (`type_doctrine` `jsonb` and
  `smallfloat`). The generator falls back on what DBAL knows: `json` with the
  `jsonb` option before DBAL 4.3, `float` before DBAL 4.1, which
  `schema:create` recreates as `double precision`.
- **The computation of a generated column reaches the logical layer**
  (`propriete.generee`: `expression`, `stockee`), copied as is from the
  physical layer.
- **The physical layer keeps the schema of a collation** outside `pg_catalog`
  (`colonne.collation_schema`), and the logical layer the explicit collation
  of a property (`propriete.collation`). A collation from another schema does
  not get there: Doctrine would write it as a single identifier, and its bare
  name would depend on the application's `search_path`. It raises the
  `collation_non_reportee` warning.

## [0.5.2] — 2026-09-14 — Ce qui tenait sans être vérifié

**Un calque PostgreSQL déjà enregistré change à sa prochaine extraction**, sans
que la base ait bougé : les objets de `public` y sortent qualifiés, dans
`type_brut` (`public.citext`), les défauts, les vérifications, les prédicats
d'index et les vues. `version_ri` ne bouge pas, et un calque plus ancien reste
lisible. L'extraction seule ne change ni le calque logique ni les entités, sauf
une séquence hors de `public` extraite par une session dont le `search_path`
contenait son schéma : sous ORM 2, son nom se qualifie dans `Base/`.

**Sous ORM 2, une clé dont la séquence n'a pas 1 pour minimum change à la
régénération** : `#[ORM\SequenceGenerator]` reçoit `initialValue`. Relancer
`ormeau inferer`, puis `ormeau:generer` ; le calque logique porte désormais
l'incrément et le minimum de la séquence.

**Un profil de connexion enregistré depuis une chaîne portant `sslmode` l'a
perdu** : il se rouvre avec le défaut du pilote, `prefer`. Le choisir, régler le
chiffrement dans le nouveau champ, puis « Mettre à jour ce profil ».

### Corrigé

- **Une classe d'entité illisible arrête `ormeau:generer` avant toute
  écriture.** Rangée dans un sous-répertoire et en cours d'édition, elle était
  ignorée, et la génération en créait une seconde à la racine. La commande
  nomme désormais chaque fichier illisible avec sa ligne, n'écrit rien et rend
  1.
- **Une chaîne de parents qui boucle, ou deux classes d'une hiérarchie
  partageant une valeur discriminante, écartent la hiérarchie avec sa raison.**
  Un calque retouché produisait des classes en héritage circulaire, ou une
  carte amputée que Doctrine refusait de charger.
- **Une association qui porte `orphelins_supprimes`, ou un identifiant de
  stratégie `aucune`, écartent l'entité avec sa raison.** Aucune version ne
  produit ces valeurs : un calque retouché obtenait un `orphanRemoval` jamais
  essayé, que Doctrine refuse sur un plusieurs-vers-un. Le code
  d'avertissement `fk_implicite_probable`, jamais émis, est retiré.
- **Un profil enregistré depuis une chaîne `clé=valeur` sans port se rouvre.**
  Son SGBD restait vide et la connexion échouait. Cette forme est celle de
  libpq : elle désigne PostgreSQL.

- **`migrations:diff` ne propose plus de modifier le minimum d'une séquence**,
  sous ORM 2. DBAL 3 le relit comme valeur initiale et le comparait au 1 par
  défaut : l'`ALTER SEQUENCE` proposé réussissait, y compris sur une base
  partagée entre applications. Quand l'incrément diffère de 1, l'`ALTER` reste
  proposé, et `ormeau:generer` dit de ne pas l'appliquer : un incrément de 10
  réserve des blocs ou sépare plusieurs nœuds, la base ne dit pas lequel, et
  aligner `allocationSize` produirait des collisions d'identifiants dans le
  second cas.
- **`ormeau inferer` refuse un fichier qui n'est pas un calque physique.**
  Passé par erreur, un calque logique était accepté, et la commande écrivait un
  `gescom.logique.logique.json` vide avec son fichier de décisions. Un champ
  requis absent ou une `version_ri` inférieure à 1 sont refusés à la lecture,
  côté Go comme côté PHP, et l'erreur nomme le champ.
- **Un calque logique sans entité est valide.** Quand toutes les tables sont
  écartées, il s'écrivait `"entites": null`, refusé par le schéma, et
  `ormeau:generer` le disait absent. Il s'écrit désormais avec une liste vide.

- **Deux extractions de la même base ne dépendent plus de la session.** Le
  `search_path` du rôle ou du DSN décidait quels noms le catalogue qualifiait :
  deux personnes produisaient deux calques, et la comparaison signalait des
  écarts inexistants. L'extraction se fait sous un `search_path` vide. Une
  séquence de `public` garde son nom nu dans le calque logique : qualifiée,
  DBAL 3 la croirait absente et `migrations:diff` proposerait de la créer.
- **`make test` et `make lint` passent sur un clone frais.** Les dépendances
  du front n'étaient jamais installées, et `make outils` ne posait pas tygo.
  Elles s'installent au premier passage, et la mise en route demande Node.
  Le README ne propose plus `go install @master`, qui ne compilait pas sans
  l'interface embarquée.
- **L'aide ne présente plus `ormeau diff` comme disponible.**

### Sécurité

- **La lecture seule de la connexion est vérifiée.** Elle était demandée au
  démarrage sans être relue : derrière un intermédiaire qui ignore les
  paramètres de démarrage, comme pgbouncer avec `ignore_startup_parameters`,
  la session restait inscriptible sans erreur. L'extraction est désormais
  refusée si la session n'est pas en lecture seule.
- **Un opérateur ou une fonction créés dans `public` ne peuvent plus se
  substituer à ceux du catalogue** dans les requêtes d'extraction
  (CVE-2018-1058), le `search_path` étant vide.
- **Le mot de passe enregistré d'un profil ne part plus vers un autre
  serveur.** Choisir un profil puis corriger l'hôte l'envoyait au nouvel hôte,
  et changer d'utilisateur l'essayait pour un autre compte. Il n'est envoyé que
  si le SGBD, l'hôte, le port et l'utilisateur sont ceux du profil ; sinon la
  connexion est refusée avant tout contact avec le serveur, et l'écran demande
  de le ressaisir.
- **Le chiffrement demandé n'est plus perdu.** `sslmode` était retiré à
  l'enregistrement d'un profil depuis une chaîne, et une connexion par
  composants partait toujours avec le défaut du pilote, `prefer` : TLS sans
  vérification du certificat, avec repli en clair. Il se choisit dans le
  formulaire, se garde dans le profil et parvient au pilote.
- **`ormeau extraire` sans `--base` n'écrit plus hors du répertoire de
  sortie.** Le nom d'une base servait tel quel à nommer son calque, et
  PostgreSQL accepte `../x` comme nom de base. Une base dont le nom ne peut pas
  nommer un fichier est sautée et signalée, avec la règle déjà appliquée par
  l'interface.

***

**A PostgreSQL layer already saved changes at its next extraction**, although
the database has not moved: objects in `public` come out qualified, in
`type_brut` (`public.citext`), defaults, checks, index predicates and views.
`version_ri` does not change, and an older layer stays readable. Extraction
alone changes neither the logical layer nor the entities, except for a sequence
outside `public` extracted by a session whose `search_path` contained its
schema: under ORM 2, its name becomes qualified in `Base/`.

**Under ORM 2, a key whose sequence does not have 1 as its minimum changes on
regeneration**: `#[ORM\SequenceGenerator]` gets `initialValue`. Run
`ormeau inferer`, then `ormeau:generer`; the logical layer now carries the
sequence's increment and minimum.

**A connection profile saved from a string carrying `sslmode` lost it**: it
reopens with the driver default, `prefer`. Pick it, set the encryption in the
new field, then “Update this profile”.

### Fixed

- **An unreadable entity class stops `ormeau:generer` before anything is
  written.** Stored in a subdirectory and being edited, it was ignored, and the
  generation created a second one at the root. The command now names each
  unreadable file with its line, writes nothing and returns 1.
- **A chain of parents that loops, or two classes of a hierarchy sharing a
  discriminator value, set the hierarchy aside with its reason.** An edited
  layer produced classes with circular inheritance, or a truncated map that
  Doctrine refused to load.
- **An association carrying `orphelins_supprimes`, or an identifier with the
  `aucune` strategy, set the entity aside with its reason.** No version
  produces these values: an edited layer got an untried `orphanRemoval`, which
  Doctrine refuses on a many-to-one. The `fk_implicite_probable` warning code,
  never emitted, is removed.
- **A profile saved from a `key=value` string without a port reopens.** Its
  DBMS stayed empty and the connection failed. This is libpq's form: it means
  PostgreSQL.

- **`migrations:diff` no longer proposes to change a sequence's minimum**,
  under ORM 2. DBAL 3 reads it back as the initial value and compared it with
  the default 1: the proposed `ALTER SEQUENCE` succeeded, including on a
  database shared between applications. When the increment is not 1, the
  `ALTER` is still proposed, and `ormeau:generer` says not to apply it: an
  increment of 10 either reserves blocks or separates several nodes, the
  database does not say which, and aligning `allocationSize` would cause
  identifier collisions in the second case.
- **`ormeau inferer` refuses a file that is not a physical layer.** A logical
  layer passed by mistake was accepted, and the command wrote an empty
  `gescom.logique.logique.json` along with its decisions file. A missing
  required field or a `version_ri` below 1 is refused on reading, in Go as in
  PHP, and the error names the field.
- **A logical layer without entities is valid.** When every table is set
  aside, it was written as `"entites": null`, refused by the schema, and
  `ormeau:generer` reported it missing. It is now written as an empty list.

- **Two extractions of the same database no longer depend on the session.**
  The `search_path` of the role or the DSN decided which names the catalogue
  qualified: two people produced two layers, and the comparison reported
  differences that did not exist. Extraction now runs under an empty
  `search_path`. A sequence in `public` keeps its bare name in the logical
  layer: qualified, DBAL 3 would consider it missing and `migrations:diff`
  would propose to create it.
- **`make test` and `make lint` pass on a fresh clone.** The front-end
  dependencies were never installed, and `make outils` did not install tygo.
  They are installed on the first run, and the setup guide asks for Node. The
  README no longer suggests `go install @master`, which did not build without
  the embedded interface.
- **The help no longer presents `ormeau diff` as available.**

### Security

- **The read-only mode of the connection is verified.** It was requested at
  startup without being read back: behind an intermediary that ignores startup
  parameters, such as pgbouncer with `ignore_startup_parameters`, the session
  stayed writable without any error. Extraction is now refused when the
  session is not read-only.
- **An operator or function created in `public` can no longer take the place
  of the catalogue's own** in extraction queries (CVE-2018-1058), the
  `search_path` being empty.
- **A profile's saved password no longer goes to another server.** Picking a
  profile then correcting the host sent it to the new host, and changing the
  user tried it for another account. It is only sent when the DBMS, host, port
  and user are the profile's; otherwise the connection is refused before any
  contact with the server, and the screen asks for it to be typed again.
- **The requested encryption is no longer lost.** `sslmode` was dropped when
  saving a profile from a string, and a connection by fields always used the
  driver default, `prefer`: TLS without certificate verification, falling back
  to plain text. It is chosen in the form, kept in the profile and reaches the
  driver.
- **`ormeau extraire` without `--base` no longer writes outside the output
  directory.** A database name was used as is to name its layer, and PostgreSQL
  accepts `../x` as a database name. A database whose name cannot name a file
  is skipped and reported, with the rule the interface already applies.

## [0.5.1] — 2026-09-14 — Ce que les bases réelles cassaient

**Cinq changements touchent un projet déjà généré.** `version_ri` ne bouge pas,
et aucun calque n'est à extraire de nouveau : relancer `ormeau inferer`, puis
`ormeau:generer`.

- **Une entité à clé `serial` générée en 0.5.0 est inutilisable** : chaque
  `persist()` échoue sur une erreur SQL. La régénération la corrige.
- **Une table d'un schéma à majuscules, accents ou mot réservé était
  inutilisable**, INSERT en échec. Sa classe est signalée en divergence avec
  l'attribut `#[ORM\Table]` attendu : le recopier. Les tables qui
  fonctionnaient ne changent pas.
- **Chaque fichier produit nomme sa base** : la régénération réécrit la ligne
  d'en-tête de `Base/`, `Enum/` et `Trait/`, et rien d'autre. `ormeau:generer`
  rend 1 quand un fichier d'une autre base bloque la génération ; côté
  bibliothèque, `GenerateurEntite::generer()` demande le nom de la base.
- **Une colonne ignorée emporte ses index et son association** : l'entité
  régénérée change si le fichier de décisions ignore une colonne indexée ou une
  clé étrangère. `migrations:diff` propose de supprimer une colonne ignorée :
  la retirer du diff avant de l'appliquer.
- **Une colonne tableau change de type** : `integer[]`, mappé en `int`, devient
  une chaîne.

### Corrigé

- **Deux bases générées dans le même répertoire ne s'écrasent plus.** La
  seconde réécrivait en silence les classes de base, énumérations et traits de
  la première. Chaque fichier nomme sa base, lue dans le nom du calque logique ;
  un fichier d'une autre base n'est pas réécrit, l'entité qui l'aurait écrit ne
  produit rien, et la commande échoue en nommant le fichier et les deux bases.
  `--remplacer=<base>` accepte explicitement d'écraser une base nommée. Un
  fichier de la 0.5.0, qui ne nomme aucune base, est repris.

- **Un schéma, une table ou une colonne qui exige des guillemets est cité
  comme Doctrine l'attend.** Le schéma recevait des backticks que Doctrine
  citait une seconde fois, et un mot réservé (`order`, `user`) restait nu :
  PostgreSQL refusait la requête. Le nom porte désormais la citation, d'après
  la forme et une liste de mots réservés relevée sous PostgreSQL 17. Une
  association plusieurs-vers-plusieurs dont la table de jointure est dans un
  tel schéma est omise avec sa raison : Doctrine n'y cite jamais le schéma.

- **`colonnes_ignorees` retire ce qui cite la colonne.** Un index ou une
  unicité sur une colonne ignorée restait dans l'entité, et `schema:create`
  échouait sur la colonne introuvable. Une clé étrangère ignorée gardait son
  association, qui écrivait la colonne et la déclarait non nullable. Index,
  unicité et association partent avec la colonne, l'avertissement
  `colonne_ignoree` les nomme, et une relation forcée sur une colonne ignorée
  est refusée.
- **Une clé étrangère qui ne vise pas la clé primaire ne produit plus
  d'association.** Déclarée vers une colonne unique, elle donnait une
  association, une table de jointure ou un un-vers-un que Doctrine refuse : il
  n'associe que vers l'identifiant. La colonne reste une propriété, avec
  l'avertissement `reference_hors_identifiant`, et une relation forcée vers une
  autre colonne que la clé primaire est refusée.
- **Une colonne tableau n'est plus typée comme son élément.** `integer[]`
  devenait `int`, et Doctrine hydratait `{1,2}` en une valeur fausse sans
  erreur. Elle est rendue en chaîne, qui garde le littéral, avec
  l'avertissement `type_non_reconnu`.
- **Le nom de séquence d'une clé `serial` est lu dans le défaut.** Le logique
  recopiait l'expression entière, `nextval('facture_id_seq'::regclass)`, dans
  `#[ORM\SequenceGenerator]`. Un défaut qui ne nomme pas sa séquence laisse
  désormais l'identifiant à l'application, avec l'avertissement
  `sequence_non_reconnue`.
- **Une clé `serial` se génère selon la version d'ORM.** ORM 3 ignore
  `#[ORM\SequenceGenerator]` dans une classe de base et prend un nom de
  séquence par défaut, faux dès que la séquence a été renommée : la clé y est
  rendue en `IDENTITY`, qui fonctionne quel que soit ce nom. ORM 2 garde
  `SEQUENCE` et le nom de la séquence ; une séquence dont le nom porte une
  apostrophe y écarte l'entité, Doctrine ne l'échappant pas dans `NEXTVAL`.

### Modifié

- **Le paquet PHP est sur Packagist** : `composer require --dev
  sprimault/ormeau-doctrine` suffit, et la ligne `composer config repositories`
  n'est plus nécessaire. Un projet qui l'a déjà ajoutée peut la retirer ; la
  garder ne change rien.
- **Le journal des versions accompagne le paquet PHP** : Dependabot et Renovate
  le trouvent à la racine du miroir et joignent ses notes à leurs pull requests
  de mise à jour.

### Sécurité

- **Un calque logique modifié à la main ne peut plus injecter de code par un
  trait.** Le nom d'énumération d'une propriété de trait était recopié sans
  contrôle dans `Trait/`, où il s'exécutait au chargement de l'entité. Un trait
  dont une énumération est absente du calque ou refusée n'est plus écrit, et
  les entités qui l'utilisent sont écartées.
- **Un lien symbolique pendant ne fait plus écrire hors du répertoire des
  entités.** Posé à la place d'un fichier à générer, il faisait créer sa cible
  ailleurs.
- **`ormeau:generer` n'interrompt plus son compte rendu** sur un nom de table
  ou d'entité qui contient une balise de la console, comme `<error></info>`.

***

**Five changes affect a project already generated.** `version_ri` does not
change, and no layer needs to be extracted again: run `ormeau inferer`, then
`ormeau:generer`.

- **An entity with a `serial` key generated by 0.5.0 is unusable**: every
  `persist()` fails with an SQL error. Regeneration fixes it.
- **A table in a schema with capitals, accents or a reserved word was
  unusable**, its INSERT failing. Its class is reported as diverging, with the
  expected `#[ORM\Table]` attribute: copy it. Tables that worked do not change.
- **Each generated file names its database**: regeneration rewrites the header
  line of `Base/`, `Enum/` and `Trait/`, and nothing else. `ormeau:generer`
  returns 1 when a file from another database blocks the generation; on the
  library side, `GenerateurEntite::generer()` requires the database name.
- **An ignored column takes its indexes and its association with it**: the
  regenerated entity changes if the decisions file ignores an indexed column or
  a foreign key. `migrations:diff` proposes to drop an ignored column: remove it
  from the diff before applying it.
- **An array column changes type**: `integer[]`, mapped to `int`, becomes a
  string.

### Fixed

- **Two databases generated into the same directory no longer overwrite each
  other.** The second silently rewrote the first one's base classes,
  enumerations and traits. Each file names its database, read from the logical
  layer's file name; a file from another database is not rewritten, the entity
  that would have written it produces nothing, and the command fails, naming
  the file and both databases. `--remplacer=<base>` explicitly accepts
  overwriting a named database. A 0.5.0 file, which names no database, is taken
  over.

- **A schema, table or column that requires quoting is quoted the way
  Doctrine expects.** The schema received backticks that Doctrine quoted a
  second time, and a reserved word (`order`, `user`) stayed bare: PostgreSQL
  rejected the query. The name now carries the quoting, based on its shape and
  on a list of reserved words taken from PostgreSQL 17. A many-to-many
  association whose join table lives in such a schema is omitted with its
  reason: Doctrine never quotes the schema there.

- **`colonnes_ignorees` removes what refers to the column.** An index or a
  unique constraint on an ignored column stayed in the entity, and
  `schema:create` failed on the missing column. An ignored foreign key kept its
  association, which wrote the column and declared it non-nullable. Index,
  unique constraint and association now go with the column, the
  `colonne_ignoree` warning names them, and a forced relation on an ignored
  column is refused.
- **A foreign key that does not target the primary key no longer produces an
  association.** Declared towards a unique column, it gave an association, a
  join table or a one-to-one that Doctrine refuses: it only associates towards
  the identifier. The column stays a property, with the
  `reference_hors_identifiant` warning, and a forced relation towards any
  column other than the primary key is refused.
- **An array column is no longer typed as its element.** `integer[]` became
  `int`, and Doctrine hydrated `{1,2}` into a wrong value without any error. It
  is rendered as a string, which keeps the literal, with the `type_non_reconnu`
  warning.
- **The sequence name of a `serial` key is read from the default.** The
  logical layer copied the whole expression, `nextval('facture_id_seq'::regclass)`,
  into `#[ORM\SequenceGenerator]`. A default that does not name its sequence now
  leaves the identifier to the application, with the `sequence_non_reconnue`
  warning.
- **A `serial` key is generated according to the ORM version.** ORM 3 ignores
  `#[ORM\SequenceGenerator]` in a base class and uses a default sequence name,
  wrong as soon as the sequence was renamed: the key is rendered as `IDENTITY`
  there, which works whatever that name. ORM 2 keeps `SEQUENCE` and the
  sequence name; a sequence whose name contains an apostrophe skips the entity
  there, as Doctrine does not escape it in `NEXTVAL`.

### Changed

- **The PHP package is on Packagist**: `composer require --dev
  sprimault/ormeau-doctrine` is enough, and the `composer config repositories`
  line is no longer needed. A project that already added it can remove it;
  keeping it changes nothing.
- **The changelog ships with the PHP package**: Dependabot and Renovate find it
  at the root of the mirror and attach its notes to their update pull requests.

### Security

- **A hand-edited logical layer can no longer inject code through a trait.**
  The enumeration name of a trait property was copied unchecked into `Trait/`,
  where it ran when the entity was loaded. A trait whose enumeration is missing
  from the layer or refused is no longer written, and the entities using it are
  skipped.
- **A dangling symbolic link no longer causes writes outside the entities
  directory.** Placed where a generated file goes, it had its target created
  elsewhere.
- **`ormeau:generer` no longer breaks off its report** on a table or entity
  name containing a console tag, such as `<error></info>`.

## [0.5.0] — 2026-09-13 — Génération Doctrine

**Trois changements touchent les calques et les fichiers de décisions déjà
enregistrés.** `version_ri` ne bouge pas.

- **Le code `heritage_deduit` a changé de sens sans changer de nom, et un
  filtre de CI sur lui ne verra rien.** Le code, sa résolution (`par_defaut`)
  et sa confiance sont les mêmes qu'avant ; il signalait un héritage appliqué,
  il signale désormais un héritage possible et non appliqué.
- **Un calque logique recalculé perd ses héritages.** Une table dont la clé
  primaire est aussi une clé étrangère est reliée à la table visée par un
  un-vers-un, qui fonctionne sur la base telle qu'elle est. Pour retrouver un
  héritage Doctrine, le déclarer dans le fichier de décisions (`heritages`) :
  la table racine, sa colonne discriminante, et une valeur par classe, racine
  comprise. Doctrine exige cette colonne ; si la base ne la porte pas, elle est
  à créer. Le message de l'avertissement dit la marche à suivre et cite les
  colonnes candidates.
- **Une décision dont un nom serait refusé par PHP est ignorée**, avec
  l'avertissement `decision_invalide` : renommage, nom d'énumération ou de cas,
  nom de relation forcée, espace de noms. L'inférence garde le nom qu'elle
  aurait produit ; corriger la décision pour retrouver le sien. La raison est
  dans « Sécurité ».

Le générateur s'installe depuis son miroir, `sprimault/ormeau-doctrine` : voir
« Générer les entités » dans le README.

### Ajouté

- **`bin/console ormeau:generer` écrit des entités Doctrine**, en mode classe de
  base : `Base/ClientBase.php` porte les propriétés, l'identifiant, les index et
  les accesseurs, et se réécrit à chaque passage ; `Client.php` porte l'entité
  et sa table, est créé une fois et n'est plus jamais modifié. Quand il ne
  correspond plus au calque — une table renommée —, la commande nomme le
  fichier, la ligne et la valeur attendue.
- **La commande annonce la version de Doctrine ORM visée** en première ligne, et
  `--cible-orm=2` ou `3` la force. Le type PHP d'une colonne en dépend : un
  `bigint` est une chaîne sous ORM 2, un entier sous ORM 3.
- **Les énumérations et les traits du calque sont générés** : `Enum/` porte des
  énumérations adossées, reliées aux colonnes par `enumType`, et un défaut
  initialise la propriété avec son cas ; `Trait/` porte les colonnes
  d'horodatage partagées, que les classes de base utilisent. Les deux sont
  réécrits à chaque passage, comme les classes de base.
- **Les associations sont générées**, dans les quatre genres et des deux côtés,
  avec leurs colonnes ou leur table de jointure et leur action à la
  suppression. Une colonne de clé étrangère n'a qu'un écrivain, l'association :
  hors clé primaire, sa propriété reste en lecture seule ; dans la clé
  primaire, l'identifiant passe par l'association. Le côté inverse n'a que son
  accesseur, puisque Doctrine n'écrit que le côté propriétaire.
- **Le fichier de décisions déclare les héritages** (`heritages`, indexé par
  la table racine), et le prérempli propose chaque hiérarchie que le schéma
  autorise, avec sa colonne candidate ou l'indication qu'aucune ne s'y prête.
  Une décision s'applique entière ou pas du tout, racine par racine.
- **L'héritage déclaré est généré**, en héritage joint : la classe de la racine
  porte la stratégie, la colonne discriminante et la carte des classes, et la
  classe de base d'une classe fille hérite de son parent. La colonne
  discriminante n'est pas une propriété, Doctrine refusant de la mapper deux
  fois. Une racine déjà générée n'est pas modifiée : chaque attribut qui lui
  manque est signalé en entier, prêt à recopier.
- **Les classes de l'utilisateur se rangent où l'on veut** sous le répertoire
  des entités : la génération suivante retrouve chacune par sa classe de base
  (`Client` qui hérite de `ClientBase`), et la cite sous son espace de noms réel
  dans les autres classes. Rien n'est recréé à la racine. Deux classes pour la
  même entité l'écartent, en nommant les deux fichiers.
- **Chaque fichier produit dit à qui il appartient**, en une ligne d'en-tête
  sans version ni date : une classe de base, une énumération ou un trait est
  généré par Ormeau et réécrit à chaque génération ; la classe de l'utilisateur
  est créée par Ormeau et jamais réécrite.
- **Le code produit se documente.** Le commentaire d'une table ou d'une colonne
  devient le docblock de la classe ou de la propriété, et la table le reporte
  dans `#[ORM\Table(options: ['comment' => …])]` pour que `schema:create` le
  recrée. Un docblock explique aussi ce qui manque : pas de mutateur sur une
  colonne qu'écrit une association, pas de propriété pour la colonne
  discriminante, une association de la clé à renseigner avant `persist()`.
  Les phrases de l'outil sont en français, et un commentaire venu de la base
  garde la langue de la base. Un tableau porte `array<mixed>`, qu'exige
  PHPStan dès le niveau 6, et le premier paragraphe d'un commentaire finit par
  une ponctuation, que PHP-CS-Fixer ajouterait sinon à chaque régénération.
- **Un commentaire de base ne peut pas injecter de code dans un fichier
  produit.** Recopié dans un docblock, il pourrait fermer le commentaire et
  faire exécuter la suite : la séquence qui le ferme est échappée, et la
  valeur exacte reste dans les options de l'attribut, en chaîne échappée.
- **Le calque logique porte le commentaire de la table** (`entite.commentaire`,
  facultatif). `version_ri` ne bouge pas ; un calque recalculé le gagne.
- **Le paquet PHP a son dépôt**,
  [sprimault/ormeau-doctrine](https://github.com/sprimault/ormeau-doctrine) :
  un miroir en lecture seule de `php/`, où Composer trouve `composer.json` à la
  racine. Il porte les versions à partir de celle-ci, et aucune antérieure.
  Issues et contributions restent sur le dépôt principal.
- **Ce qui ne se génère pas entier est écarté, en le disant** : une table sans
  clé primaire, un nom réservé de PHP, deux entités de même nom, une entité
  dont une association propriétaire vise une entité écartée, une identité
  dérivée en chaîne — que Doctrine refuse —, et une hiérarchie dont une table
  fille ne nomme pas sa clé comme la racine, Doctrine joignant par ces noms.
  Un côté inverse vers une entité écartée est omis, et l'omission figure dans
  le compte rendu.

### Modifié

- **Le paquet PHP s'installe sur PHP 8.1, Symfony 5.4 à 8 et Doctrine ORM 2.14
  à 3**, au lieu de PHP 8.3, Symfony 7 et ORM 3 : une application qu'on reprend
  tourne rarement sur la version courante. Il ne se déclare plus compatible
  avec ORM 4, qui n'est pas publié.
- **Le plancher d'ORM est 2.14, pas 2.11**, bien que `enumType` existe depuis
  la 2.11. C'est une contrainte d'ORM constatée, pas un choix de confort : les
  versions 2.11 à 2.13 refusent une classe de base mappée placée sous une
  entité, ce que produit la génération d'une hiérarchie.
- **Le paquet PHP contrôle chaque champ du calque logique à la lecture** :
  présence, type et valeurs des vocabulaires fermés, jusqu'aux objets
  imbriqués. Un calque invalide est refusé avec le chemin du champ fautif —
  `entites[12].proprietes[3].type_doctrine` — au lieu d'échouer plus loin.
- **Une classe globale n'est plus importée dans le code produit** :
  `\DateTimeImmutable` au lieu de `use DateTimeImmutable;`. PHP-CS-Fixer, en
  règles `@Symfony`, la requalifiait dans le projet, et chaque régénération la
  réimportait. La CI vérifie désormais que `@Symfony` ne trouve rien à corriger
  dans ce que le générateur réécrit.

### Corrigé

- **Un défaut vide survit à l'inférence.** Une colonne `DEFAULT ''` perdait son
  défaut dans le calque logique, où il se confondait avec l'absence de défaut ;
  la régénération du schéma l'aurait retiré. `version_ri` ne bouge pas : les
  calques logiques se recalculent avec `ormeau inferer`.
- **L'écran d'arbitrage n'affiche plus de type PHP, qui était faux.** Il
  montrait `int` pour un `bigint`, que Doctrine rend en chaîne sous ORM 2, et
  doublait le `?` d'une propriété facultative (`??string`). Le type PHP dépend
  de la version de Doctrine du projet : c'est la génération qui le fixe.
  L'écran montre le type Doctrine, et l'énumération quand il y en a une.

### Obsolète

- **`type_php` du calque logique** reste dans le format jusqu'à sa prochaine
  version, mais ne vaut que pour l'une des deux versions d'ORM. Le générateur
  Doctrine le déduit de `type_doctrine`, `nullable` et `enumeration`, et ne le
  lit que pour un type Doctrine qu'il ne connaît pas. Un outil tiers qui le
  consomme devrait faire de même.

### Sécurité

- **Un fichier de décisions ne peut plus faire écrire du code ni des fichiers
  hors de leur place.** Un renommage, un nom d'énumération ou de cas, un nom
  de relation forcée et l'espace de noms étaient recopiés tels quels dans le
  code produit et dans le chemin des fichiers écrits : `../../public/index`
  écrivait hors du répertoire des entités, et un nom contenant du code
  l'injectait. Le fichier de décisions se reçoit par une pull request et
  l'interface l'écrit : c'est un vecteur, pas seulement une faute de frappe.
  L'inférence refuse désormais tout nom que PHP refuserait à sa place,
  forme et mots réservés compris (avertissement `decision_invalide`), et
  garde le nom qu'elle aurait produit ; une décision jusqu'ici appliquée avec
  un tel nom est donc ignorée. Le générateur revérifie chaque nom du calque
  logique, écarte ce qui le porte, arrête tout sur un espace de noms refusé,
  et vérifie que chaque chemin écrit reste, une fois résolu, sous le
  répertoire des entités.

***

**Three changes affect layers and decisions files already saved.**
`version_ri` stays put.

- **The `heritage_deduit` code changed meaning without changing name, and a CI
  filter on it will not notice.** The code, its resolution (`par_defaut`) and
  its confidence are the same as before; it used to flag an applied
  inheritance, it now flags a possible, unapplied one.
- **A recomputed logical layer loses its inheritances.** A table whose primary
  key is also a foreign key is tied to the target table by a one-to-one, which
  works on the database as it is. To get a Doctrine inheritance back, declare it
  in the decisions file (`heritages`): the root table, its discriminator column,
  and one value per class, root included. Doctrine requires that column; if the
  database does not have it, it must be created. The warning message says what
  to do and names the candidate columns.
- **A decision with a name PHP would refuse is ignored**, with the
  `decision_invalide` warning: rename, enumeration or case name, forced relation
  name, namespace. Inference keeps the name it would have produced; fix the
  decision to get yours back. The reason is under "Security".

The generator installs from its mirror, `sprimault/ormeau-doctrine`: see
"Generating entities" in the README.

### Added

- **`bin/console ormeau:generer` writes Doctrine entities**, in base-class mode:
  `Base/ClientBase.php` holds the properties, identifier, indexes and
  accessors, and is rewritten on every run; `Client.php` holds the entity and
  its table, is created once and never modified again. When it no longer
  matches the layer — a renamed table —, the command names the file, the line
  and the expected value.
- **The command announces the targeted Doctrine ORM version** on its first
  line, and `--cible-orm=2` or `3` forces it. A column's PHP type depends on it:
  a `bigint` is a string under ORM 2, an integer under ORM 3.
- **The layer's enumerations and traits are generated**: `Enum/` holds backed
  enumerations, tied to their columns through `enumType`, and a default
  initialises the property with its case; `Trait/` holds the shared timestamp
  columns, which the base classes use. Both are rewritten on every run, like
  the base classes.
- **Associations are generated**, in all four kinds and on both sides, with
  their join columns or join table and their delete action. A foreign key
  column has a single writer, the association: outside the primary key, its
  property stays read-only; inside the primary key, the identifier goes
  through the association. The inverse side only gets its getter, since
  Doctrine only writes the owning side.
- **The decisions file declares inheritances** (`heritages`, keyed by the root
  table), and the pre-filled file proposes every hierarchy the schema allows,
  with its candidate column or a note that none fits. A decision applies whole
  or not at all, root by root.
- **Declared inheritance is generated**, as joined inheritance: the root class
  holds the strategy, the discriminator column and the class map, and a child
  class's base class extends its parent. The discriminator column is not a
  property, since Doctrine refuses to map it twice. A root generated earlier is
  not modified: each attribute it lacks is reported in full, ready to copy.
- **User classes can be arranged anywhere** under the entities directory: the
  next generation finds each one by its base class (`Client` extending
  `ClientBase`), and refers to it by its actual namespace in the other classes.
  Nothing is recreated at the root. Two classes for the same entity skip it,
  naming both files.
- **Every generated file says who owns it**, in a single header line with no
  version or date: a base class, an enumeration or a trait is generated by
  Ormeau and rewritten on every run; the user class is created by Ormeau and
  never rewritten.
- **Generated code documents itself.** A table or column comment becomes the
  docblock of the class or property, and the table carries it in
  `#[ORM\Table(options: ['comment' => …])]` so that `schema:create` recreates
  it. A docblock also explains what is missing: no setter on a column written
  by an association, no property for the discriminator column, a key
  association to set before `persist()`. The tool's sentences are in French,
  and a comment taken from the database keeps the database's language. An
  array carries `array<mixed>`, which PHPStan requires from level 6, and a
  comment's first paragraph ends with punctuation, which PHP-CS-Fixer would
  otherwise add on every regeneration.
- **A database comment cannot inject code into a generated file.** Copied into
  a docblock, it could close the comment and have what follows executed: the
  closing sequence is escaped, and the exact value stays in the attribute
  options, as an escaped string.
- **The logical layer carries the table comment** (`entite.commentaire`,
  optional). `version_ri` stays put; a recomputed layer gains it.
- **The PHP package has its own repository**,
  [sprimault/ormeau-doctrine](https://github.com/sprimault/ormeau-doctrine): a
  read-only mirror of `php/`, where Composer finds `composer.json` at the root.
  It carries versions from this one on, and none before. Issues and
  contributions stay on the main repository.
- **What cannot be generated whole is skipped, saying so**: a table without a
  primary key, a PHP reserved name, two entities with the same name, an entity
  whose owning association points at a skipped one, a chained derived identity
  — which Doctrine rejects —, and a hierarchy in which a child table names its
  key differently from the root, since Doctrine joins on those names. An
  inverse side pointing at a skipped entity is left out, and the report says
  so.

### Changed

- **The PHP package installs on PHP 8.1, Symfony 5.4 to 8 and Doctrine ORM 2.14
  to 3**, instead of PHP 8.3, Symfony 7 and ORM 3: an application being taken
  over rarely runs the current release. It no longer claims compatibility with
  ORM 4, which is not released.
- **The ORM floor is 2.14, not 2.11**, although `enumType` exists since 2.11.
  This is an observed ORM constraint, not a convenience: versions 2.11 to 2.13
  reject a mapped superclass placed below an entity, which generating a
  hierarchy produces.
- **The PHP package checks every field of the logical layer when reading it**:
  presence, type and closed-vocabulary values, down to nested objects. An
  invalid layer is refused with the path of the faulty field —
  `entites[12].proprietes[3].type_doctrine` — instead of failing further on.
- **A global class is no longer imported in generated code**:
  `\DateTimeImmutable` instead of `use DateTimeImmutable;`. PHP-CS-Fixer, with
  `@Symfony` rules, requalified it in the project, and every regeneration
  imported it again. CI now checks that `@Symfony` finds nothing to fix in what
  the generator rewrites.

### Fixed

- **An empty default survives inference.** A `DEFAULT ''` column lost its
  default in the logical layer, where it was mistaken for no default at all;
  regenerating the schema would have dropped it. `version_ri` stays put:
  logical layers are recomputed with `ormeau inferer`.
- **The arbitration screen no longer shows a PHP type, which was wrong.** It
  showed `int` for a `bigint`, which Doctrine returns as a string under ORM 2,
  and doubled the `?` of an optional property (`??string`). The PHP type
  depends on the project's Doctrine version: generation settles it. The screen
  shows the Doctrine type, and the enumeration when there is one.

### Deprecated

- **`type_php` in the logical layer** stays in the format until its next
  version, but only holds for one of the two ORM versions. The Doctrine
  generator derives it from `type_doctrine`, `nullable` and `enumeration`, and
  reads it only for a Doctrine type it does not know. A third-party tool
  consuming it should do the same.

### Security

- **A decisions file can no longer have code or files written out of place.**
  A rename, an enumeration or case name, a forced relation name and the
  namespace were copied verbatim into the generated code and into the paths of
  written files: `../../public/index` wrote outside the entities directory,
  and a name containing code injected it. The decisions file arrives through a
  pull request and the interface writes it: it is an attack vector, not just a
  possible typo. Inference now refuses any name PHP would refuse in its place,
  shape and reserved words included (`decision_invalide` warning), and keeps
  the name it would have produced; a decision applied so far with such a name
  is therefore ignored. The generator checks every name of the logical layer
  again, skips whatever carries an invalid one, stops altogether on a refused
  namespace, and checks that every written path, once resolved, stays under
  the entities directory.

## [0.4.2] — 2026-09-13 — Ce qui ne doit pas sortir d'Ormeau

**Deux entrées acceptées jusqu'ici sont refusées.** `version_ri` ne bouge pas :
calques et fichiers de décisions restent lisibles tels quels. Un script qui
passe l'une de ces entrées échouera désormais :

- `--echantillonner` ou `--cardinalite-max` sur `ormeau extraire` : les retirer,
  ils n'avaient aucun effet ;
- une chaîne de connexion ambiguë — un mot de passe contenant `/`, `?` ou `#`
  non encodé dans une URL, ou une valeur précédée d'un blanc et contenant `=`
  (`user= password=…`, `password = abc==`) : encoder le mot de passe, le mettre
  entre apostrophes, ou passer par `--hote`, `--utilisateur` et `ORMEAU_MDP`.

### Corrigé

- **Un mot de passe pouvait s'afficher en clair dans une erreur de
  connexion.** Il suffisait d'une chaîne clé/valeur écrite `password = secret`,
  d'une valeur entre apostrophes, ou d'un port mal saisi : `ormeau extraire`
  et l'écran de connexion recopiaient alors le secret. Toute valeur est
  désormais masquée, sauf celles qui ne peuvent rien porter de sensible —
  hôte, port, base, utilisateur, mode TLS.
- **Une chaîne de connexion ambiguë est refusée au lieu d'être devinée** : un
  mot de passe contenant `/`, `?` ou `#` non encodé dans une URL, ou
  `user= password=…`, que le pilote lisait comme un nom d'utilisateur. Le
  message dit quoi corriger.
- **Une clé de chiffrement abîmée ne bloque plus la connexion par profil.** Un
  `cle.bin` tronqué faisait échouer la connexion au lieu de la préparer sans le
  mot de passe enregistré, comme pour une clé effacée. Le prochain mot de passe
  enregistré en tire une neuve, en prévenant que les précédents sont perdus, et
  l'ancienne clé reste à côté en `cle.bin.invalide`.
- **`ormeau extraire --echantillonner` est refusé**, comme dans l'interface.
  Accepté sans effet, il produisait un calque identique à celui obtenu sans lui.
  `--cardinalite-max` suit.

***

**Two inputs accepted until now are refused.** `version_ri` stays put: layers
and decisions files remain readable as they are. A script passing one of these
inputs will now fail:

- `--echantillonner` or `--cardinalite-max` on `ormeau extraire`: remove them,
  they had no effect;
- an ambiguous connection string — a password holding an unencoded `/`, `?` or
  `#` in a URL, or a value preceded by a blank and containing `=`
  (`user= password=…`, `password = abc==`): encode the password, quote it, or
  use `--hote`, `--utilisateur` and `ORMEAU_MDP`.

### Fixed

- **A password could show in clear in a connection error.** A key/value string
  written `password = secret`, a quoted value, or a mistyped port was enough:
  `ormeau extraire` and the connection screen then echoed the secret. Every
  value is now masked, except those that cannot carry anything sensitive —
  host, port, database, user, TLS mode.
- **An ambiguous connection string is refused instead of guessed**: a password
  holding an unencoded `/`, `?` or `#` in a URL, or `user= password=…`, which
  the driver read as a user name. The message says what to fix.
- **A damaged encryption key no longer blocks connecting through a profile.** A
  truncated `cle.bin` made the connection fail instead of preparing it without
  the saved password, as for a deleted key. The next saved password draws a new
  one, warning that earlier ones are lost, and the old key stays alongside as
  `cle.bin.invalide`.
- **`ormeau extraire --echantillonner` is refused**, as in the interface.
  Accepted without effect, it produced a layer identical to one made without
  it. `--cardinalite-max` follows.

## [0.4.1] — 2026-09-12 — Où Ormeau range quoi

**Rien à reprendre dans vos projets.** `version_ri` ne bouge pas, les calques
déjà enregistrés restent lisibles, et aucun fichier de décisions n'est à
retoucher. Cette version ne change que ce qu'Ormeau retient de votre poste.

### Ajouté

- **Les connexions s'enregistrent en profils.** Donnez un nom avant de vous
  connecter : le profil est créé une fois la connexion réussie, et il n'y a pas
  d'autre bouton. Il retient le SGBD, l'hôte, l'utilisateur, la base, et le
  répertoire de travail qui va avec — le choisir vous emmène donc aussi dans le
  bon projet.
- **Le mot de passe est retenu si vous cochez la case prévue**, décochée par
  défaut. Il est chiffré sur le poste, avec une clé qui s'y trouve aussi : cela
  protège d'une lecture accidentelle, pas de quelqu'un qui a accès à votre
  session. La case le dit, et [`SECURITY.fr.md`](SECURITY.fr.md) détaille.
- **Un profil qui nomme une base y garde la session** : les autres bases du
  serveur ne sont pas proposées. C'est un cadrage de travail et non une
  protection — le compte garde ses droits, que seul le serveur restreint.
- **Le travail d'arbitrage non enregistré survit à un rechargement.** Les
  décisions en cours et l'entité ouverte sont retenues, et retrouvées à la
  réouverture de l'écran. Le fichier de décisions fait toujours foi : si la base
  a été réextraite ou le fichier retouché entre-temps, le brouillon est écarté
  et l'écran le dit.
- **Le répertoire de travail se change depuis l'interface**, sans relancer le
  binaire. Le serveur valide le chemin saisi et renvoie le chemin résolu, liens
  suivis — c'est lui qui s'affiche, et non ce qui a été tapé. Une extraction en
  cours écrit toujours là où elle a été lancée.
- **Ormeau range ses propres données dans le répertoire de configuration du
  système** : `%AppData%\ormeau`, `~/.config/ormeau` ou
  `~/Library/Application Support/ormeau`, créé au premier lancement.
  `ORMEAU_CONFIG_DIR` en désigne un autre, ce qui permet aussi un usage
  portable. [`docs/emplacements.fr.md`](docs/emplacements.fr.md) dit ce qui va
  où, et pourquoi.
- **Des aides au survol** expliquent ce qu'aucun libellé ne portait : ce que
  cocher une table entraîne, et le fait que décocher une colonne ne retire rien
  du calque.

### Corrigé

- **Le thème, la langue et la taille des panneaux sont retenus d'un lancement à
  l'autre.** Ils revenaient à leurs valeurs par défaut à chaque démarrage.
  Ils sont désormais enregistrés dans `preferences.yaml`, et appliqués avant le
  premier affichage.
- **Le refus d'un calque absent nomme le répertoire de travail.** Lancée
  ailleurs que dans le projet, l'interface annonçait un calque manquant sans
  dire où elle l'avait cherché.

***

**Nothing to change in your projects.** `version_ri` stays put, layers already
saved remain readable, and no decisions file needs editing. This release only
changes what Ormeau remembers about your machine.

### Added

- **Connections can be saved as profiles.** Give a name before connecting: the
  profile is created once the connection succeeds, and there is no other button.
  It keeps the DBMS, host, user, database, and the working directory that goes
  with them — picking it therefore takes you to the right project too.
- **The password is kept if you tick the box**, unticked by default. It is
  encrypted on this machine, with a key that lives there too: this guards
  against accidental reading, not against someone with access to your account.
  The box says so, and [`SECURITY.md`](SECURITY.md) has the details.
- **A profile that names a database keeps the session on it**: the server's
  other databases are not offered. This frames the work rather than protecting
  anything — the account keeps its rights, which only the server can restrict.
- **Unsaved review work survives a page reload.** Current decisions and the open
  entity are kept, and found again when the screen reopens. The decisions file
  still has the final word: if the database was re-extracted or the file edited
  meanwhile, the draft is discarded and the screen says so.
- **The working directory can be changed from the interface**, without
  restarting the binary. The server validates the path and returns the resolved
  one, symlinks followed — that is what is displayed, not what was typed. An
  extraction already running still writes where it was started.
- **Ormeau stores its own data in the system configuration directory**:
  `%AppData%\ormeau`, `~/.config/ormeau` or
  `~/Library/Application Support/ormeau`, created on first launch.
  `ORMEAU_CONFIG_DIR` points somewhere else, which also makes portable use
  possible. [`docs/emplacements.md`](docs/emplacements.md) says what goes where,
  and why.
- **Hover help** explains what no label carried: what ticking a table entails,
  and the fact that unticking a column removes nothing from the layer.

### Fixed

- **Theme, language and panel sizes are kept between launches.** They reverted
  to their defaults on every start. They are now stored in `preferences.yaml`,
  and applied before the first paint.
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
