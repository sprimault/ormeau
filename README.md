> [🇬🇧 English](README.md) · [🇫🇷 Français](README.fr.md)

# Ormeau

![CI](https://github.com/sprimault/ormeau/actions/workflows/ci.yml/badge.svg)
![License](https://img.shields.io/badge/license-Apache%202.0-blue)

> Reverse-engineers a real legacy database into Doctrine entities — and does it
> again six months later without discarding the work done in between.

> [!WARNING]
> **Not usable yet.** The pivot format and its serialization are implemented and
> tested; the PostgreSQL driver, the inference engine and the Doctrine generator
> are not. Every command currently returns an error naming the roadmap phase that
> will bring it. See [Status](#status).

## What Ormeau is

Ormeau introspects a relational schema and generates Doctrine entities from it.

The point is **not** to generate entities from a clean database — a naive script
does that, and Doctrine itself did before dropping `doctrine:mapping:import`. The
point is to take over a real legacy database: `T_` prefixes, foreign keys never
declared, tables without a primary key, booleans stored as `char(1)`, generated
columns. And to do it twice, six months apart, without overwriting the work done
on the entities in between.

Every design decision is settled in favour of that sentence.

## Why it exists

Doctrine removed its reverse engineering: `doctrine:mapping:import` is gone from
the bundle and `DatabaseDriver` left with ORM 3. Nothing official remains, and
the ecosystem alternatives are either abandoned or too naive for real legacy —
they produce `$clientId` as an `integer` instead of an association.

## The layer format

```
introspection  ->  physical layer  ->  inference  ->  logical layer  ->  generation
    (Go)              (JSON)          (Go, pure)        (JSON)            (PHP)
```

The **layer** (*calque*) is the pivot format: a data model, serialized as JSON,
describing a database with no reference to the source DBMS nor to the target
framework. Without it you write *n × m* translators; with it, *n + m*.

It comes in two levels, and the split carries the whole design:

|  | physical layer | logical layer |
|---|---|---|
| describes | what **is** in the database | what we **decide** to make of it |
| produced by | introspection | inference |
| can be wrong | only through a bug | it is a judgement, it is debatable |
| loses information | no | yes, deliberately |

The physical level is an observation; the logical level is a judgement. Keeping
them apart is what allows an inference to be corrected offline and tested with no
database at all:

```
physical + decisions -> logical
```

A pure function: no network, no clock, no randomness.

Three properties are enforced on the physical layer — **completeness** (an
equivalent DDL must be reconstructible), **neutrality** (no field assumes the
destination), **determinism** (two identical extractions produce two
byte-for-byte identical files). The last one is what makes the diff mode usable
rather than noisy.

Details in [`docs/architecture.md`](docs/architecture.md). The contract itself is
in [`schemas/`](schemas/), versioned separately from the repository.

## Two languages, one deliberate boundary

**Go** for introspection and inference. All drivers are pure Go, so no cgo and no
system dependency: a single binary you run on a customer's server without
installing anything. The PHP equivalent would require `pdo_sqlsrv` and Instant
Client, turning the project into installation support.

**PHP** for Doctrine generation. The hard part is not writing files, it is
non-destructive regeneration: reading an already-edited entity with
`nikic/php-parser`, comparing it to the logical layer, rewriting only what moved,
keeping business methods and formatting.

## The messy cases are the subject

Table with no primary key, composite primary key, undeclared foreign key,
`0000-00-00` dates, a boolean stored as `char(1)` holding `O`/`N`, two tables
linked by columns of different types. That is the daily reality of a database
being taken over, and what existing tools handle worst.

The rule: emit a warning, never an exception, never an invention. A partial
logical layer with twenty precise warnings is worth infinitely more than a fatal
error or a silently wrong model.

Every inferred element carries its `origine` — `contrainte`, `verification`,
`cardinalite`, `nommage` or `decision`. Without it the tool is not auditable, and
nobody will run it on their database.

## Intended usage

```
ormeau extraire --dsn "postgres://..." --sortie gescom.calque.json
ormeau inferer  gescom.calque.json --decisions decisions.yaml --sortie gescom.logique.json
ormeau diff     gescom.calque.json
ormeau interface

bin/console ormeau:generer      gescom.logique.json
bin/console ormeau:synchroniser gescom.calque.json
```

`ormeau:synchroniser` answers "what changed in the database since my entities" —
the inverse of `doctrine:schema:update`, and what actually serves day to day on
legacy where the schema moves without migrations.

## Safety

The tool only ever reads. Connections are read-only, including during sampling,
with a per-query timeout. No SQL string travels from the browser: the local
interface exposes fixed endpoints.

The DSN is the only secret handled. It never appears in logs, in error messages,
or in the layer file.

**A layer is a customer's database schema** — table names, column names, business
comments, and with `--echantillonner`, real values. Never commit a layer extracted
from a production database, not even as an issue attachment.

## Status

Nothing is installable yet. The roadmap is in [`ROADMAP.md`](ROADMAP.md).

| Phase | State |
|---|---|
| 1 — Physical layer: structures, deterministic serialization, fingerprint, JSON Schema | Done |
| 2 — PostgreSQL introspection | Read-only connection and inventory done; extraction still to write |
| 3 — Inference and logical layer | Structures only |
| 4 — Doctrine generation | Skeleton only |
| 5 to 11 | Not started |

CI runs the test suite with the race detector, `golangci-lint`, `gofmt`,
`govulncheck`, `gosec` and a JSON Schema validity check on every push and pull
request. The workflow is public and its runs are in the Actions tab.

## Feedback

Bugs, feature requests or questions: open an issue at
https://github.com/sprimault/ormeau/issues (French preferred, English welcome).

Planning to send a patch? [`CONTRIBUTING.md`](CONTRIBUTING.md) states the rules
a pull request is judged against — they are not obvious from the code.

Security flaws go through the private channel described in
[`SECURITY.md`](SECURITY.md), never through a public issue.

## License

Apache 2.0 — see [`LICENSE`](LICENSE).
