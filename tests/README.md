> [🇬🇧 English](README.md) · [🇫🇷 Français](README.fr.md)

# Test suites

`ddl/` holds the reference databases, one per DBMS. They deliberately cover the
awkward cases: table without a primary key, composite primary key,
self-referencing foreign key, pure join table, generated column, enumeration
through a check constraint, inheritance through a foreign primary key, reserved
identifiers, accents.

`reference/` holds the inference cases: a frozen physical layer as input, an
expected logical layer as output. Inference being pure, these cases replay
without a database, and they are the project's real test suite.

When a heuristic is added, the physical layer that triggers it comes first.

Expected outputs are rewritten in bulk behind a flag, never automatically: an
expected output regenerated without being reviewed no longer tests anything.
Only two packages declare it, and each runs on its own — a `./...` would fail on
all the others:

```bash
go test ./internal/inference/ -maj-attendus   # expected logical layers, also `make maj-attendus`
go test ./internal/calque/ -maj-attendus      # physical layer serialisation
```
