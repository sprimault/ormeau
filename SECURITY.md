> [🇬🇧 English](SECURITY.md) · [🇫🇷 Français](SECURITY.fr.md)

# Security

## Reporting a vulnerability

**Do not open a public issue for a security flaw.**

Use the **Report a vulnerability** button in the repository's Security tab
(GitHub Private Vulnerability Reporting). The report stays private until a fix
ships.

No contact address is published, and there is no alternative channel.

## Supported versions

Only the latest published release is fixed, with no backport to earlier ones.

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

### Saved password

The connection screen offers to save a profile's password, on a checkbox that
is unticked by default. It is then encrypted with AES-256-GCM in
`profils.yaml`, using a key specific to the installation, drawn at random and
stored next to it with `0600`.

**What this guards against**: accidental reading. A file opened by mistake, a
backup someone browses, a `cat` during a screen share, a glance over your
shoulder.

**What it does not guard against**: someone with access to your account. The
key lives on the same disk as the file it encrypts, and whoever can read one
can read the other. This is the same model as DBeaver or SSMS, and it is stated
here rather than left to be discovered. Anyone needing more leaves the box
unticked: the password is then retyped on every connection and written nowhere.

**Where it goes**: to the profile's destination, and nowhere else. If the DBMS,
host, port or user typed differ from the profile's, Ormeau refuses to send it,
before any contact with the server, and asks for it to be typed again.
Switching to another database on the same server does not count.

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
- being able to decrypt a saved password from the session of the user who saved
  it. That is what the section above announces, and what the checkbox says on
  screen;
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
