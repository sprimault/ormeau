> [🇬🇧 English](README.md) · [🇫🇷 Français](README.fr.md)

# Ormeau

![CI](https://github.com/sprimault/ormeau/actions/workflows/ci.yml/badge.svg)
![License](https://img.shields.io/badge/license-Apache%202.0-blue)

> Takes over a real legacy database and produces Doctrine entities from it —
> then does it again six months later without discarding the work done since.

> [!WARNING]
> Ormeau connects to the databases whose credentials you give it and writes
> files describing their schema. It only ever reads the catalogue, in a
> read-only session enforced by the server, but the target database and the
> account used remain your choice. A layer describes a customer's database
> schema: it is not to be versioned or passed around lightly. Preview release,
> provided without warranty under the Apache 2.0 licence — see
> [Status](#status) for what works.

## What Ormeau is

The point is **not** to generate entities from a clean database — a script does
that, and Doctrine itself did before dropping `doctrine:mapping:import`. The
point is to take over a real legacy database: `T_` prefixes, foreign keys never
declared, tables without a primary key, booleans stored as `char(1)`, generated
columns. And to do it twice, six months apart, without overwriting the work done
on the entities in between.

This tool comes from field experience: taking over an existing database to
derive Doctrine entities from it is a recurring task on client work, and nothing
has covered it since `doctrine:mapping:import` was removed.

Doctrine removed its reverse engineering: `doctrine:mapping:import` is gone from
the bundle and `DatabaseDriver` left with ORM 3. Nothing official remains, and
the ecosystem alternatives are either abandoned or stop at a literal
transposition: they return `$clientId` as an `integer` where an association to
`Client` is needed.

## Getting started

Download the archive for your platform from the
[latest release](https://github.com/sprimault/ormeau/releases/latest), unpack it,
run it. Nothing else to install: no runtime, no system driver.

```console
$ export ORMEAU_DSN="postgres://app:secret@srv:5432/gescom"
$ ormeau extraire --sortie gescom.calque.json
gescom.calque.json : 10 table(s), 32 colonne(s), 0 anomalie(s)
empreinte sha256:f422f6d3e5eb455a91b096bd513bd5d8e595bd4e88aa588ef25d241993e201a1
```

The string goes through `ORMEAU_DSN` rather than `--dsn`, which `ps` would show.

A connection can also be given as components, which avoids escaping a password
inside a URL. The password has no flag: it would be visible in `ps` and in the
shell history.

```console
$ export ORMEAU_MDP=secret
$ ormeau extraire --sgbd postgres --hote srv --utilisateur app --base gescom --sortie gescom.calque.json
```

Without `--base`, every database on the server is extracted and `--sortie`
designates a directory:

```console
$ ormeau extraire --sgbd postgres --hote srv --utilisateur app --sortie calques/
calques/gescom.calque.json : 10 table(s), 32 colonne(s), 0 anomalie(s)
calques/facturation.calque.json : 24 table(s), 187 colonne(s), 0 anomalie(s)
```

Verifying a downloaded archive — the binaries being neither signed nor
notarized, SmartScreen and Gatekeeper will complain on first launch:

```console
$ gh attestation verify ormeau_v0.5.2_linux_amd64.tar.gz --repo sprimault/ormeau
```

Through a container, giving it the caller's identity: the image runs as an
unprivileged user and would not write to a Linux volume otherwise.

```console
$ docker run --rm --user "$(id -u):$(id -g)" \
    -e ORMEAU_DSN -v "$PWD:/sortie" \
    ghcr.io/sprimault/ormeau:v0.5.2 extraire --sortie /sortie/gescom.calque.json
```

With the layer in hand, inference runs offline — the database is no longer
needed:

```console
$ ormeau inferer gescom.calque.json
gescom.decisions.yaml écrit, entièrement en commentaire : rien n'est appliqué tant que
vous n'avez pas décommenté. Il porte les renommages que l'outil propose.

gescom.logique.json : 12 entité(s), 31 association(s), 2 énumération(s), 1 trait(s)

2 avertissement(s) :
  prefixe_detecte          public                           préfixe T_ commun aux 12 tables, conservé ; prefixes_a_retirer le retirerait des noms de classes
  table_sans_cle_primaire  public.t_log_import              aucune clé primaire : Doctrine refusera cette entité en l'état

Par code :
  prefixe_detecte          1
  table_sans_cle_primaire  1
```

You uncomment what suits you in `gescom.decisions.yaml`, run it again, and your
rulings replay on every pass: the command line never rewrites an existing file.
The interface does regenerate it when you save, and asks first if it was edited
by hand.

Between releases, build from a clone: see
[Getting set up](CONTRIBUTING.md#getting-set-up). `go install` does not work, as
the embedded interface is not versioned.

### Generating entities

The Symfony bundle is its own Composer package on Packagist,
`sprimault/ormeau-doctrine`, published from a read-only mirror of `php/`:

```console
$ composer require --dev sprimault/ormeau-doctrine
```

It requires PHP 8.1, Symfony 5.4 to 8 and Doctrine ORM 2.14 to 3. Symfony Flex
registers the bundle for `dev` and `test`; without Flex, add
`Ormeau\Doctrine\OrmeauDoctrineBundle::class => ['dev' => true, 'test' => true]`
to `config/bundles.php`.

```console
$ bin/console ormeau:generer gescom.logique.json
Cible détectée : PHP 8.4, Doctrine ORM 3.7, DBAL 4.4
Base : gescom
créé     src/Entity/Enum/StatutClient.php
créé     src/Entity/Base/ClientBase.php
créé     src/Entity/Client.php
```

`Base/`, `Enum/` and `Trait/` belong to the tool and are rewritten on every run.
`Client.php` is created once and never touched again: your business methods go
there. It can be moved into a subdirectory, `src/Entity/Ventes/Client.php` with
its namespace: the next run finds it by its base class, and the other classes
refer to it where it now lives. When it no longer matches the layer — a renamed
table, a newly declared subclass —, the command names the file, the line and
the expected attribute.
`--repertoire` writes somewhere other than `src/Entity`, and `--cible-orm=2`
or `3` targets an ORM version other than the installed one. The PHP type of
some columns also depends on DBAL, which ORM 3 accepts in version 3 as in
version 4: a forced ORM version assumes one, which the announcement calls
deduced, and `--cible-dbal=3.10` corrects it.

Each file names in its header the database it comes from, read from the layer's
file name: `gescom` for `gescom.logique.json`. Another database generated into
the same directory does not rewrite those files; the command names the file and
both databases. Give each database its own directory and namespace, or pass
`--remplacer=gescom` to explicitly accept overwriting `gescom`'s files — a
renamed layer, a directory taken over. Only what prevents the requested
generation returns 1: a refused overwrite, or a PHP file under the entities
directory that cannot be read, before anything is written. A skipped entity and
a divergence return 0: they say what to rework.

This two-class shape is the price of regenerating without overwriting, and it
changes a few habits. Properties are declared in `Base/`, which is rewritten:
validation constraints and serialization groups do not go on them. They go in
`Client.php`, on a redeclared getter — `#[Assert\NotBlank] public function
getRaisonSociale(): string { return parent::getRaisonSociale(); }` —, or in the
mapping files of `config/validator/` and `config/serializer/`, which name the
inherited property. `make:entity` does not suit these classes: it would add
fields to `Client.php`, outside what the layer describes. A column is added in
the database, and the entities are regenerated.

Table and column comments become docblocks. The sentences the tool writes are
in French, like all of its output, while a comment taken from the database
keeps the database's language: a docblock may mix both, and that is expected.

### Local interface

`ormeau interface` opens your browser on a connection screen, where you type a
host and a login rather than assemble a connection string. The DBMS follows from
the port, and the server then says what it actually is.

```console
$ cd ~/projects/gescom && ormeau interface
Interface sur http://127.0.0.1:53412
Répertoire de travail : /home/steff/projects/gescom
```

The left-hand tree lists the server's databases, their schemas and their tables;
each table expands into its columns. You tick what you want to extract, and the
screen flags referenced tables missing from the selection — a foreign key left
pointing nowhere makes the association vanish silently.

Unticking a column does not remove it from the layer: it fills
`colonnes_ignorees` in the decisions file, which drops it from the entity. The
layer keeps everything, and you can change your mind without reopening the
connection.

**Extract** writes `<base>.calque.json` as a background task, without blocking
the screen; a running extraction can be cancelled.

The **Review** tab works offline, on the layer in the working directory: no
connection is needed. It shows the entities generation will produce, and saves
your rulings to `<base>.decisions.yaml`.

It listens on `127.0.0.1` only, on a port drawn at startup, and the token in the
URL is good for a single use. Files land in the directory shown on screen;
`--repertoire` points somewhere else, and the screen can change it without a
restart.

A connection can be saved as a **profile**: give it a name before connecting,
and it is kept once the connection succeeds. A profile carries the working
directory that goes with it, and remembers the password if you ask — encrypted
on the machine, with the limits [`SECURITY.md`](SECURITY.md) spells out.

Profiles, display preferences and review drafts live in the system
configuration directory, never in your project:
[`docs/emplacements.md`](docs/emplacements.md) says what goes where, and why.

### On Windows

The commands are the same, bar three details: the binary is called
`.\ormeau.exe`, environment variables are set differently, and `--user` is
pointless — a Windows volume carries no POSIX permissions.

```powershell
$env:ORMEAU_MDP = "secret"
.\ormeau.exe extraire --sgbd postgres --hote srv --utilisateur app --base gescom --sortie gescom.calque.json

docker run --rm -e ORMEAU_DSN -v "${PWD}:/sortie" `
    ghcr.io/sprimault/ormeau:v0.5.2 extraire --sortie /sortie/gescom.calque.json
```

Verifying the checksum of a downloaded archive:

```powershell
$attendu = (Select-String -Path SHA256SUMS -Pattern windows).Line.Split(" ")[0]
$obtenu  = (Get-FileHash ormeau_v0.5.2_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
if ($attendu -eq $obtenu) { "empreinte OK" } else { "EMPREINTE DIFFERENTE" }
```

## What extraction produces

A **layer** (*calque*): a tracing of the catalogue, in JSON, that judges nothing
and loses nothing. The type name is the server's own, never rebuilt; the default
value is structured, so that `DEFAULT 'now()'` and `DEFAULT now()` stay
distinguishable; length and precision are absent rather than zero, so that
`decimal(10,0)` is not confused with `int`.

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

Two extractions of the same database differ **only by `source.extrait_le`**, the
extraction time, and share the same fingerprint, which excludes it. That is what
makes the diff mode usable, and what allows a layer to be versioned in Git.

### Three files per database

```
gescom.calque.json       the observation, extracted from the database
gescom.decisions.yaml    your rulings, written once
gescom.logique.json      the object model, consumed by the generator
```

They live in your project and are committed with it. Only the first comes from
the database; the other two are recomputed offline, so an inference can be
corrected without reopening the connection — including on a layer brought back
from a customer site.

`decisions.yaml` is what makes the second extraction bearable. Six months later
the schema has moved, but your rulings replay on top of it: what you corrected
once is not lost again.

One file per database, never a shared one: two databases may both have a
`client` table without the same rename suiting either.

The design behind this format — its two levels, its three properties, the split
between Go and PHP — is in [`docs/architecture.md`](docs/architecture.md).

## The messy cases are the subject

Table with no primary key, composite primary key, undeclared foreign key,
`0000-00-00` dates, a boolean stored as `char(1)` holding `O`/`N`, two tables
linked by columns of different types. That is the daily reality of a database
being taken over, and what existing tools handle worst.

The rule: emit a warning, never an exception, never an invention. A partial
logical layer with twenty precise warnings beats a fatal error or a silently
wrong model. Every inferred element carries its `origine` — `contrainte`,
`verification`, `cardinalite`, `nommage` or `decision` — without which the tool
is not auditable.

## Safety

The tool only ever reads. Connections are read-only, enforced by the server
rather than by the code's discipline, with a per-query timeout. The session is
read back when it opens: if an intermediary dropped read-only mode on the way,
extraction is refused.

The DSN is the only secret handled: it appears neither in logs, nor in error
messages, nor in the layer.

Ormeau opens two connections and no others: your database, and `127.0.0.1` for
the local interface. No telemetry, no outbound call, not even an update check —
the code is public, and `netstat` confirms it during an extraction.

**A layer is a customer's database schema** — table names, column names,
business comments, and with `--echantillonner`, once sampling ships, real
values. A layer extracted from a production database never enters a repository,
nor an issue attachment.

## Status

PostgreSQL extraction, inference and Doctrine entity generation work:
`ormeau extraire` then `ormeau inferer` produce the three files, and
`bin/console ormeau:generer` writes the entities, their associations,
enumerations and traits, and the inheritance declared in the decisions file,
skipping what Doctrine cannot represent and saying so. Comparing the database
with existing entities (`ormeau:synchroniser`) is still to come. The state per
phase is in [`ROADMAP.md`](ROADMAP.md).

CI runs the test suite with the race detector, `golangci-lint`, `gofmt`,
`govulncheck`, `gosec` and a JSON Schema validity check on every push and pull
request. The PHP package goes through PHPUnit, PHPStan and PHP-CS-Fixer, and its
tests run under four combinations of PHP, Symfony and Doctrine ORM, from PHP 8.1
with Symfony 5.4 to PHP 8.4 with Symfony 8. Integration tests run there against
a real database server, never a simulated catalog, and the extraction of the
test database is compared byte for byte with a reference layer.

CI only confirms: every change first goes through the same full validation, on
Windows and Linux machines, before it is pushed.

## Going further

- [`docs/architecture.md`](docs/architecture.md) — the layer, its two levels,
  why two languages
- [`docs/construction.md`](docs/construction.md) — cross-compilation, multi-arch
  images, signing
- [`docs/emplacements.md`](docs/emplacements.md) — where the project files go,
  and where what concerns only your machine goes
- [`schemas/`](schemas/) — the public contract, versioned separately
- [`CONTRIBUTING.md`](CONTRIBUTING.md) — the rules a pull request is judged
  against

## Feedback

Bugs, feature requests or questions: open an issue at
https://github.com/sprimault/ormeau/issues (French preferred, English welcome).

Security flaws go through the private channel described in
[`SECURITY.md`](SECURITY.md), never through a public issue.

## Where the name comes from

*Ormeau* — pronounced roughly *or-MOH* — is French for a young elm, and for the
abalone, a shellfish whose shell is built from stacked layers of nacre. It also
happens to begin with ORM, which is convenient for a tool that produces ORM
entities.

The pivot format is called a **calque**, in the tracing-paper sense: a faithful
copy of the catalogue, with no interpretation. In linguistics a calque is a
structural borrowing from one language into another — English *skyscraper*
becoming French *gratte-ciel*. That is precisely the operation: carrying the
structure of a relational schema over into another language's type system.

## License

Apache 2.0 — see [`LICENSE`](LICENSE).

**The entities, layers and decision files Ormeau produces are yours.** The
licence covers the tool, not its output: nothing it generates enters your
project with an obligation attached.
