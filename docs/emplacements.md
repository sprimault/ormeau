> [🇬🇧 English](emplacements.md) · [🇫🇷 Français](emplacements.fr.md)

# Where Ormeau keeps what

Two locations, told apart by what they hold and, above all, by who owns them.

|  | work data | machine data |
|---|---|---|
| where | your project | the system configuration directory |
| what | a database's three files | preferences, profiles, drafts |
| versioned | yes, with your code | never |
| shared with the team | yes | never |
| losing the file costs | the review work | retyping |

The rule fits in one sentence: **what describes the database goes in the
project, what describes your machine stays on your machine.**

## Work data

The working directory is the one you ran the command from, or the one
`--repertoire` points at. The interface shows it at all times and lets you
change it without restarting; a connection profile can remember it.

```
gescom.calque.json       the observation, extracted from the database
gescom.decisions.yaml    your decisions
gescom.logique.json      the object model, consumed by the generator
```

In practice, the Symfony project that `bin/console ormeau:generer` will run
from. These files are meant to be committed: that is what makes the second
extraction bearable, six months later.

A layer carries a database's table names, column names and business comments —
and, with `--echantillonner`, once sampling ships, real values. Commit it to the
customer's repository, never anywhere else.

## Machine data

| system | location |
|---|---|
| Windows | `%AppData%\ormeau\` |
| Linux | `$XDG_CONFIG_HOME/ormeau/`, else `~/.config/ormeau/` |
| macOS | `~/Library/Application Support/ormeau/` |

Created on first launch, with `0700` — the profiles file carries hosts and
accounts of production databases, which other accounts on the machine have no
business reading.

```
ormeau/
  preferences.yaml     theme, language, panel sizes
  profils.yaml         saved connections
  cle.bin              encryption key for saved passwords
  etat/                disposable — delete it and lose nothing
    LISEZMOI.txt
    sessions/          review drafts in progress
```

**`ORMEAU_CONFIG_DIR` overrides this location.** Handy for portable use — the
binary and its configuration on the same stick — and it is also what keeps the
test suite out of your real configuration.

### preferences.yaml

Theme, language, and the panel sizes you set. They live here rather than in the
browser for a precise reason: the interface listens on a port drawn at each
launch, so the page origin changes every time, and any local storage would start
empty.

An unreadable or malformed file does not prevent startup: the interface opens
with its defaults and says so once.

### profils.yaml

The connections you named: DBMS, host, port, user, database, and the working
directory that goes with them. Picking a profile therefore takes you to the
right project too.

A profile is created by giving a name before connecting — it is saved once the
connection succeeds, and there is no other button. A profile that names a
database keeps the session on it: the server's other databases are not offered.
This frames the work rather than restricting rights, which only the server can
do.

### The saved password

Only if you tick the box, which is unticked by default. It is then encrypted
with AES-256-GCM using `cle.bin`, drawn at random on first write.

**What this guards against**: accidental reading — a file opened by mistake, a
backup someone browses, a `cat` during a screen share.

**What it does not guard against**: someone with access to your account. The key
lives on the same disk as the file it encrypts. This is the same model as
DBeaver or SSMS, and [`SECURITY.md`](../SECURITY.md) has the details. Anyone
needing more leaves the box unticked: the password is then retyped on every
connection and written nowhere.

If `cle.bin` is missing or unusable, or `profils.yaml` copied from another
machine, the profiles stay usable without their passwords, and the screen says
so once.

An unusable key is replaced on the next saved password. The screen then warns
that the passwords encrypted with it are lost, and the old one stays alongside,
as `cle.bin.invalide`.

### etat/

Disposable, and its `LISEZMOI.txt` says as much. It holds review drafts: what
you decided without having saved it to `<base>.decisions.yaml` yet.

A draft carries the fingerprint of the layer and of the decisions file it came
from. If either changed — a re-extraction, a file edited in an editor — it is
discarded and the screen announces it: **the decisions file has the final
word**, and replaying a draft over an earlier version would silently bring back
what had just been removed from it.

Its name carries the database and a fingerprint of the working directory: two
projects that each have a `gescom` database never collide.

## What does not survive a restart

The interface draws a free port at each launch, which has two visible
consequences.

The **session cookie** dies with the tab, and the URL token is single-use:
reopening an address noted earlier grants no access to the API.

The **URL fragment** — `#arbitrage/gescom` — survives a page reload, but not a
restart of the binary: the address changed port. That is intended, and it is why
the review draft is kept in `etat/` instead.
