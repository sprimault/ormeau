> [🇬🇧 English](README.md) · [🇫🇷 Français](README.fr.md)

# ormeau-doctrine

The Symfony bundle of [Ormeau](https://github.com/sprimault/ormeau): it
generates Doctrine entities from a legacy database, and regenerates them six
months later without discarding the code written on them since.

Doctrine removed `doctrine:mapping:import`, and nothing official replaced it.
Ormeau takes over a real legacy database — `T_` prefixes, foreign keys never
declared, booleans stored as `char(1)` — and produces associations where a
literal translation would leave integer columns.

> [!IMPORTANT]
> This repository is a **read-only mirror** of the `php/` directory of
> [sprimault/ormeau](https://github.com/sprimault/ormeau), rewritten on every
> merge. Issues, pull requests and documentation live in the main repository.

## What it needs

The bundle does not read the database. It reads a logical layer produced by the
`ormeau` binary, a single file to download from the
[releases](https://github.com/sprimault/ormeau/releases/latest), with nothing
else to install. PostgreSQL for now:

```console
$ ormeau extraire --dsn "postgres://app:secret@srv:5432/gescom" --sortie gescom.calque.json
$ ormeau inferer gescom.calque.json
```

`ormeau inferer` writes `gescom.logique.json`, and a decisions file where you
settle what the tool cannot decide on its own.

## Installation

```console
$ composer require --dev sprimault/ormeau-doctrine
$ bin/console ormeau:generer gescom.logique.json
```

One database per entities directory: files name their database, and another
database does not overwrite them without `--remplacer=<base>`. Exit codes: 0
when entities are skipped or classes diverge, 1 when a file from another
database blocked the generation.

The whole workflow — extracting the layer, inferring, arbitrating, generating —
is described in the [main README](https://github.com/sprimault/ormeau#readme),
and what each version changes in the
[changelog](https://github.com/sprimault/ormeau/blob/master/CHANGELOG.md).

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

Apache 2.0 — see [`LICENSE`](LICENSE). The entities Ormeau generates are yours:
the licence covers the tool, not its output.
