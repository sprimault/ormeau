> [🇬🇧 English](architecture.md) · [🇫🇷 Français](architecture.fr.md)

# Architecture

This document explains the design. What the tool does and how to use it is in
the [README](../README.md).

## A pivot format, and why

```
introspection  ->  physical layer  ->  inference  ->  logical layer  ->  generation
    (Go)              (JSON)          (Go, pure)        (JSON)            (PHP)
```

This is a compilation. The introspector is the front end, one per DBMS; the
generator is the back end, one per ORM; the layer is the intermediate
representation. Without it you write *n × m* translators; with it, *n + m*.

The name — *calque* in French, the project's own vocabulary — comes from tracing
paper, a faithful copy of the structure, and from the linguistic sense: a calque
is a structural borrowing from one language into another, which is exactly the
operation.

## Two levels

|  | physical layer | logical layer |
|---|---|---|
| describes | what **is** in the database | what we **decide** to make of it |
| produced by | introspection | inference |
| hand-edited | never | never |
| regenerable | by re-reading the database | from the physical layer and the decisions |
| loses information | no | yes, deliberately |

The separation rests on an asymmetry: the physical level is an **observation**,
it can only be wrong through a bug; the logical level is a **judgement**, it can
be argued with. Mixing them would forbid any offline correction and make the
tests depend on a running server.

Hence the relation that carries everything:

```
physical + decisions -> logical
```

A pure function: no network, no side effects, no clock. That is what allows an
inference to be corrected without access to the database, replayed, and tested
against reference files rather than a server.

## The three properties of the physical layer

**Completeness.** It must contain everything needed to reconstruct a DDL
equivalent to the original. That is testable: `database -> layer -> DDL ->
database` must yield an empty structural diff. What is not there is lost — no
downstream stage can recover it.

**Neutrality.** No field assumes the destination. `type_normalise: "decimal"` is
neutral; `type_doctrine` is not, and belongs to the logical level.

**Determinism.** Two extractions of the same database produce two byte-for-byte
identical files. Sorted keys, no timestamp in the body of the document. Without
it, the diff mode produces noise and becomes useless.

## The logical layer is not neutral

It does not speak Doctrine, but it speaks the vocabulary of the Hibernate
family: entities, associations with an owning and an inverse side, inheritance
strategies, composite identifiers. Doctrine and EF Core share it; an ORM built
on another model, such as GORM, consumes the physical layer instead.

The bias is deliberate. A truly universal logical layer would be so poor that it
would carry no decision at all, and every generator would reimplement the
heuristics: exactly what the split exists to avoid.

## Why two languages

**Go** for introspection and inference. Pure-Go drivers (`pgx`,
`go-sql-driver/mysql`, `microsoft/go-mssqldb`, `sijms/go-ora`), therefore no cgo
and no system dependency: a binary that runs as is wherever you put it. Most
often on the developer's workstation, reaching the database over the network; on
the customer's server when it is not reachable otherwise. The PHP equivalent
would require `pdo_sqlsrv` and Instant Client, turning the project into
installation support.

**PHP** for Doctrine generation. The hard part is not writing files, it is
non-destructive regeneration: reading an already-edited entity with
`nikic/php-parser`, comparing it to the logical layer, rewriting only what
moved, keeping business methods and formatting.

The contract between the two halves is the layer, validated by a versioned JSON
Schema published in [`schemas/`](../schemas/). As a consequence: anyone can
write an Eloquent or EF Core generator without touching the core, and either
half is independently rewritable.

## What the layer does not contain

Table data — only aggregate statistics, and only optionally —, grants and roles,
stored procedures and triggers, server settings. Nothing that does not serve to
produce an object model. A layer that drifts towards a full dump stops being
stable, diffable and versionable in Git.

## Format versioning

`version_ri` is an integer, not semver: the only question that matters is
whether a generator can read the file. Adding an optional field does not change
the version; everything else increments it — renaming, removal, change of
meaning, a new value in a closed vocabulary. The generator refuses a version
higher than the one it knows.

The `empreinte` field has two distinct roles: in `source` it identifies a
database state; in the logical layer, `empreinte_physique` says which
observation this judgement derives from — which is how "the database moved since
the last generation" is detected without re-reading the entities.

## Practical consequence

You can extract a database at a customer's site, leave with a layer of a few
megabytes, and do all the generation and heuristic iterations elsewhere, with no
network access to that database. The layer is versioned in Git, replayed and
compared between two engagements.
