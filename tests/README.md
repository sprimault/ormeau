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

These physical layers are written by hand, except one: the `gescom` case, a real
extraction of `ddl/postgres.sql` by the PostgreSQL driver. An integration test
compares it byte for byte with a fresh extraction, and the same file is the
input of inference. The container image is pinned to a minor version, since the
server version is part of the file: changing it means regenerating the layer and
reviewing its diff.

`reference/extraction/sqlserver/gescom.calque.json` is the extraction of
`ddl/sqlserver.sql` by the SQL Server driver, compared the same way. It feeds no
inference case: a directory under `reference/inference/` carries an expected
logical layer and generated entities, and inference for this dialect has not
been reviewed yet. The image is pinned to a cumulative update, for the same
reason.

Expected outputs are rewritten in bulk behind a flag, never automatically: an
expected output regenerated without being reviewed no longer tests anything.
Four packages declare it, and each runs on its own — a `./...` would fail on
all the others:

```bash
go test ./internal/inference/ -maj-attendus   # expected logical layers, also `make maj-attendus`
go test ./internal/calque/ -maj-attendus      # physical layer serialisation
make maj-calque-gescom                        # PostgreSQL gescom extraction, against the container
make maj-calque-sqlserver                     # SQL Server gescom extraction, against the container
```

After `make maj-calque-gescom`, `make maj-attendus` recomputes the case's
logical layer, then `ORMEAU_MAJ_ATTENDUS=1 composer test` regenerates its
entities under `php/tests/Generation/attendus/`.

## Round trip

`allerretour/` holds the test that checks the whole chain: the test database
extracted, inferred, generated into Doctrine entities, recreated by Doctrine in
the container's blank `allerretour` database, extracted again, then compared with
the original.

```bash
make aller-retour              # PostgreSQL
make aller-retour-sqlserver    # SQL Server
```

The diff is never empty: Doctrine recreates neither views, nor CHECK
constraints, nor native enumerated types, and names its foreign keys its own
way. These differences form a closed list per DBMS, in
`allerretour/ecarts_test.go` and `allerretour/ecarts_sqlserver_test.go`, each
with its reason:

- **IMPOSSIBLE**: a limit of Doctrine or DBAL;
- **VOULU** (intended): a decision of the tool, tolerated only when the logical
  layer or the generator takes it for that object;
- **À COMBLER** (to close): a known gap, which goes away with the fix that
  closes it.

The test fails on a difference the list does not cover, and on an entry that no
longer covers anything: the list can neither hide a regression nor keep a
tolerance that became useless.

It needs PHP with the `pdo_pgsql` extension — `pdo_sqlsrv` and Microsoft's ODBC
Driver 18 for SQL Server — and the dependencies of `php/` installed.
`ORMEAU_PHP` names the interpreter when it is not `php` — for instance a
`docker run` command mounting the repository at the same path — and the
installed ORM version selects the list of differences. The report of the last
run is written to `.tmp/allerretour/ecarts.txt`.

Two targets have their list, both ends of the promised range, for each DBMS:
ORM 3 with DBAL 4, from `composer.lock`, and ORM 2.14 with DBAL 3, which CI
resolves without the lock (jobs `aller-retour-orm2` and
`aller-retour-sqlserver-orm2`). An entry that only holds for one says so; an
entry that covers nothing under a target fails, and must be restricted
explicitly.
