> [🇬🇧 English](SECURITY.md) · [🇫🇷 Français](SECURITY.fr.md)

# Security

## Reporting a vulnerability

**Do not open a public issue for a security flaw.**

Use the **Report a vulnerability** button in the repository's Security tab
(GitHub Private Vulnerability Reporting). The report stays private until a fix
ships.

No contact address is published, and there is no alternative channel.

## Supported versions

Ormeau is at an early stage and nothing is published yet. Once releases exist,
only the latest tag will be fixed; there will be no backport to earlier tags.

## Scope

Ormeau reads a database and writes files. It never writes to the database it
introspects, and that is an invariant rather than an intention. Two things
deserve care, and they define most of what counts as a flaw here: **the DSN is
the only secret the tool handles**, and **a layer file is a customer's database
schema**.

The following **are** in scope:

- the DSN, or its password, appearing anywhere it should not — a log line, an
  error message, the layer file, an API response of the local interface, a
  crash dump;
- any write reaching the introspected database, including during sampling;
- SQL injection through a catalogue identifier — a table, column or schema name
  that escapes into a query instead of being quoted;
- an arbitrary SQL string reaching the database from the local interface: the
  API exposes fixed endpoints, and no query is meant to travel from the browser;
- the local interface listening on anything other than the loopback address, or
  being reachable from the network;
- business data landing in a layer beyond the configured cardinality ceiling,
  or without `--echantillonner` having been asked for;
- path traversal when writing a layer, a decisions file or generated entities;
- code execution while reading a layer, a decisions file or an existing entity;
- a vulnerable dependency actually reachable from Ormeau's own code.

The following are **not** vulnerabilities:

- the SmartScreen warning on Windows and the Gatekeeper block on macOS. The
  binaries are not signed nor notarized; this is documented in
  [`docs/construction.md`](docs/construction.md) along with the reason;
- a layer containing table names, column names and business comments. That is
  what a layer is for. Protecting the file once produced is the operator's
  responsibility — the README says so, and `.gitignore` keeps layers out of
  this repository;
- connecting to whatever DSN the user supplies, including one pointing at a
  host they do not own. Choosing the database is the operator's decision, not
  the tool's;
- the `POSTGRES_PASSWORD` value in `tests/docker-compose.yml`. It belongs to an
  ephemeral test container built from `tests/ddl/`, and grants nothing;
- resource exhaustion caused by introspecting a very large schema. The tool
  runs on the operator's own machine, against a database they already have
  credentials for;
- automated scanner output with no working reproduction.

## Handling

A report must carry a reproduction: which version, which command, which DBMS,
and what an attacker obtains that the sections above do not already grant. A
report without one is closed.

The project is maintained on a voluntary basis, with no committed turnaround.
Reports are handled on a best-effort basis, worst first. There is no bug bounty
and no service level agreement.

**Never attach a layer extracted from a production database**, to a report or
to anything else. If a reproduction needs one, build it from `tests/ddl/` or
strip it down to the few objects that trigger the defect.
