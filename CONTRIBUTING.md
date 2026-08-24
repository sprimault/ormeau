> [🇬🇧 English](CONTRIBUTING.md) · [🇫🇷 Français](CONTRIBUTING.fr.md)

# Contributing to Ormeau

Contributions are welcome. This page exists so that a well-written patch does
not get turned down for a rule nobody could have known about.

Note that the project is at an early stage: most of it is not implemented yet,
and the roadmap in [`ROADMAP.md`](ROADMAP.md) says in which order it will be.
Opening an issue before writing code is a good way to avoid working on
something already under way.

## What the project is for

> The goal is **not** to generate entities from a clean database — a naive
> script does that. It is to take over a real legacy database, and to do it
> again six months later without overwriting the work done on the entities in
> between.

Every design decision is settled in favour of that sentence, and so is every
pull request. A change that makes the clean case nicer, without helping on a
database that is actually messy, is off target — however good the code.

## Non-negotiable rules

These predate any contribution and are not up for negotiation inside a pull
request. If one of them gets in your way, the design is what we discuss, in an
issue, before any code.

1. **The physical layer loses nothing.** Anything from the catalogue needed to
   reconstruct an equivalent DDL must be captured. What extraction misses is
   lost for good: no downstream stage can recover it.
2. **The physical layer does not judge.** No renaming, no singularisation, no
   prefix stripping, no relation inference. It observes.
3. **The physical layer is neutral.** No field assumes the destination. The
   test: if an EF Core generator would find the field useless or misleading, it
   is at the wrong level.
4. **Extraction is deterministic.** Two extractions of the same database
   produce byte-for-byte identical files. Sorted keys, no timestamp in the body
   of the document. The diff mode depends entirely on this.
5. **Inference is a pure function.** `physical + decisions -> logical`. No
   network, no clock, no randomness, no disk access outside the declared
   inputs. An inference that would need to query the database is misplaced:
   what it looks for belongs in the physical layer.
6. **What is not resolved is not invented.** An uncertain inference produces an
   entry in `avertissements` with its code, target and confidence. Warnings are
   a first-class output, not a log.
7. **Every inference carries its origin.** Without it the tool is not
   auditable, and nobody will run it on their database.
8. **The generator decides nothing.** It translates the logical layer. No
   heuristic goes down into `php/`.
9. **Regeneration does not destroy human work.** Business methods, docblocks
   and formatting of an existing entity are preserved. We compare the AST, we
   do not rewrite the file.
10. **No write to the introspected database**, ever, including during
    sampling. Read-only connection, with a per-query timeout.
11. **No cgo.** Every driver is pure Go. A binary requiring a system library on
    a customer's server loses the project's main argument.

## Two rules that cost time when forgotten

**A field added to the layer means three changes**: the JSON Schema, the Go
structs, the PHP reader. Ship the three together — a divergence between them is
a defect, not a temporary gap.

**A heuristic added calls for its reference case.** Add the physical layer that
triggers it under `tests/reference/` first, then the code. Those files are the
project's real test suite.

## Getting set up

You need Go (the version pinned in `go.mod`) and Docker for the test
containers. PHP 8.3 and Composer are only needed to work on `php/`.

```bash
make outils        # golangci-lint, govulncheck, gosec
make test          # go test -race ./...
make lint          # golangci-lint, gofmt
make cover         # coverage, per-function breakdown
make maj-attendus  # rewrites the expected logical layers, review them after
```

`make test` must pass on a fresh clone without Docker: tests that need a DBMS
carry the `integration` build tag and run through `make test-integration`.

The Makefile includes `makefile.local` when present, for machine-specific
settings. If your antivirus quarantines binaries as the linker writes them —
the build then fails on a permission error unrelated to the code — that is
where you redirect `GOTMPDIR` and `GOCACHE` into `.tmp/`:

```make
export GOTMPDIR := $(TMP)/gobuild
export GOCACHE := $(TMP)/gocache
_ := $(shell mkdir -p "$(GOTMPDIR)" "$(GOCACHE)")
```

The file is not versioned, and nothing requires you to create it.

The container may run elsewhere than on your workstation: `ORMEAU_TEST_DSN`
overrides the DSN the integration tests target.

Before opening a pull request, run at least `make lint` and `make test`. CI
runs them too, but it does so after the branch is already pushed.

## Commits and pull requests

The conventional prefix is required — `feat:`, `fix:`, `test:`, `docs:`,
`refactor:` — because tooling reads it. **French and English are both
accepted**; write in whichever you are comfortable with.

Say what the change does and why, in a few lines. The mechanism you had to
understand to write the fix belongs in a comment, in the code, where it will be
read alongside it.

## Never attach a layer from a production database

A layer holds a customer's table names, column names, business comments, and
with `--echantillonner`, real values. It has no place in this repository nor in
an issue attachment. The only layers versioned here are those produced from
`tests/ddl/`.

If a reproduction needs a layer, build it from `tests/ddl/` or strip it down to
the few objects that trigger the defect.

## Security

Do not open a public issue for a security flaw — see
[`SECURITY.md`](SECURITY.md) for the private channel.
