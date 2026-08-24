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
$ ormeau extraire --dsn "postgres://app:secret@srv:5432/gescom" --sortie gescom.calque.json
gescom.calque.json : 10 table(s), 32 colonne(s), 0 anomalie(s)
empreinte sha256:f422f6d3e5eb455a91b096bd513bd5d8e595bd4e88aa588ef25d241993e201a1
```

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
$ gh attestation verify ormeau_v0.2.0_linux_amd64.tar.gz --repo sprimault/ormeau
```

Through a container, giving it the caller's identity: the image runs as an
unprivileged user and would not write to a Linux volume otherwise.

```console
$ docker run --rm --user "$(id -u):$(id -g)" \
    -e ORMEAU_DSN -v "$PWD:/sortie" \
    ghcr.io/sprimault/ormeau:v0.2.0 extraire --sortie /sortie/gescom.calque.json
```

Between releases, `go install github.com/sprimault/ormeau/cmd/ormeau@master`.

### On Windows

The commands are the same, bar three details: the binary is called
`.\ormeau.exe`, environment variables are set differently, and `--user` is
pointless — a Windows volume carries no POSIX permissions.

```powershell
$env:ORMEAU_MDP = "secret"
.\ormeau.exe extraire --sgbd postgres --hote srv --utilisateur app --base gescom --sortie gescom.calque.json

docker run --rm -e ORMEAU_DSN -v "${PWD}:/sortie" `
    ghcr.io/sprimault/ormeau:v0.2.0 extraire --sortie /sortie/gescom.calque.json
```

Verifying the checksum of a downloaded archive:

```powershell
$attendu = (Select-String -Path SHA256SUMS -Pattern windows).Line.Split(" ")[0]
$obtenu  = (Get-FileHash ormeau_v0.2.0_windows_amd64.zip -Algorithm SHA256).Hash.ToLower()
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

Two extractions of the same database produce **byte-for-byte identical** files —
the timestamp is excluded from the fingerprint. That is what makes the diff mode
usable, and what allows a layer to be versioned in Git.

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
rather than by the code's discipline, with a per-query timeout.

The DSN is the only secret handled: it appears neither in logs, nor in error
messages, nor in the layer.

**A layer is a customer's database schema** — table names, column names,
business comments, and with `--echantillonner`, real values. A layer extracted
from a production database never enters a repository, nor an issue attachment.

## Status

PostgreSQL extraction works. Inference and entity generation are not written:
their commands return an error naming the phase that will bring them. The state
per phase is in [`ROADMAP.md`](ROADMAP.md).

CI runs the test suite with the race detector, `golangci-lint`, `gofmt`,
`govulncheck`, `gosec` and a JSON Schema validity check on every push and pull
request.

## Going further

- [`docs/architecture.md`](docs/architecture.md) — the layer, its two levels,
  why two languages
- [`docs/construction.md`](docs/construction.md) — cross-compilation, multi-arch
  images, signing
- [`schemas/`](schemas/) — the public contract, versioned separately
- [`CONTRIBUTING.md`](CONTRIBUTING.md) — the rules a pull request is judged
  against

## Feedback

Bugs, feature requests or questions: open an issue at
https://github.com/sprimault/ormeau/issues (French preferred, English welcome).

Security flaws go through the private channel described in
[`SECURITY.md`](SECURITY.md), never through a public issue.

## License

Apache 2.0 — see [`LICENSE`](LICENSE).
