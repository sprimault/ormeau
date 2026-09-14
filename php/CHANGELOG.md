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
