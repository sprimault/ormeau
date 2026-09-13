> [🇬🇧 English](construction.md) · [🇫🇷 Français](construction.fr.md)

# Building and distributing

## One machine produces everything

Go cross-compiles natively. A single Linux machine is enough to produce the
Windows, macOS and Linux binaries, for amd64 as well as arm64:

```bash
make binaries
```

This holds on one condition only: **no cgo**. All the chosen drivers are pure Go,
and that is what makes the build matrix trivial. The first driver requiring cgo
would impose a cross-compiler per target.

`CGO_ENABLED=0` is explicit on every build line, never implicit: on a Linux
machine cgo is on by default, and a binary dynamically linked against glibc would
refuse to start on Alpine.

## The Linux binaries depend on no distribution

A static binary has no dependency on glibc, musl, or anything else from the
system. A single `ormeau_linux_amd64` runs on Debian, Ubuntu, RHEL and Alpine
alike.

There is therefore no "Debian build" or "Alpine build". Two architectures, no
more: amd64 and arm64.

## Build order

The front end is built **before** the binaries. Vite writes to
`internal/interface/embarque`, which `go:embed` reads from its own package:
`embed` cannot reach above the directory it is declared in, and an intermediate
copy would be one more step to forget.

The `binaries` target depends on `web-build` for that reason, and a test checks
that the embedded filesystem carries a non-empty `index.html`. Without it the
safeguard would rest on the Makefile dependency alone, and a missing bundle
would surface as a blank interface.

## Local builds

Antivirus software on a Windows workstation may quarantine the temporary
executables the linker writes, and the build then fails on an access-denied
error with no apparent connection to the code. The fix is to point `GOTMPDIR`
and `GOCACHE` into `.tmp/`, inside the repository — but in `makefile.local`,
which the Makefile includes when present and git ignores, not in the versioned
Makefile: it is a constraint of that workstation, not a project decision.
[`CONTRIBUTING.md`](../CONTRIBUTING.md) gives the lines to copy.

`make build` writes to `.tmp/`; `dist/` stays reserved for release artefacts.

## Docker images

Since the binaries are already cross-compiled, the multi-arch image builds
**without QEMU emulation**: buildx fills in `TARGETOS` and `TARGETARCH`, and each
platform gets the matching binary.

```dockerfile
FROM alpine:3 AS certificats
RUN apk add --no-cache ca-certificates

FROM scratch
ARG TARGETOS TARGETARCH
COPY --from=certificats /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY dist/ormeau_${TARGETOS}_${TARGETARCH} /ormeau
ENTRYPOINT ["/ormeau"]
```

```bash
docker buildx build --platform linux/amd64,linux/arm64 \
  -t ghcr.io/sprimault/ormeau:$VERSION --push .
```

`FROM scratch` means there is no distribution in the image — and therefore **no
root certificates**. Without copying `ca-certificates.crt`, every TLS connection
to a database fails with an error that is hard to diagnose. That is the classic
`scratch` trap.

The image runs as an unprivileged user, which has a practical consequence: it
cannot write into a volume owned by somebody else. Extracting a layer therefore
requires giving it the caller's identity, otherwise the write is denied without
the message saying why:

```bash
docker run --rm --user "$(id -u):$(id -g)" \
  -e ORMEAU_DSN -v "$PWD:/sortie" \
  ghcr.io/sprimault/ormeau:VERSION extraire --sortie /sortie/gescom.calque.json
```

Lowering the image's user would fix the symptom and introduce a flaw: a tool
handling production credentials has no business running as root.

The container remains a fallback: it still has to reach the database, which the
native binary does with no network configuration.

## The PHP package mirror

Composer only reads a VCS repository whose `composer.json` sits at its root. The
package lives in `php/`, so it is published to
[sprimault/ormeau-doctrine](https://github.com/sprimault/ormeau-doctrine), where
`php/` becomes the root. That repository is a read-only mirror: no issues, no
pull requests, and nothing is written to it by hand.

`.github/workflows/miroir.yml` pushes `git subtree split --prefix=php` to it on
every push to `master`, and for every `v*` tag when called from `release.yml`.
The split is deterministic, and the split of an older commit is an ancestor of
the split of a newer one: the mirror always moves forward by fast-forward, and
**nothing is ever force-pushed to it**. A rejection means the mirror has
diverged; that gets investigated, not overwritten.

The mirror only carries versions in which the package works: it starts at
0.5.0, and no earlier tag will be pushed to it.

### What guards the write access

- The key is a **deploy key** of the mirror: it cannot write to any other
  repository.
- Its private half is the `MIROIR_CLE` secret of the main repository's
  **`miroir` environment**, whose deployment rule only admits `master` and `v*`
  tags. A workflow triggered by a pull request runs on another ref: GitHub does
  not open the environment to it, whatever its file says.
- On the mirror, two rules: `master` cannot be deleted or force-pushed, and only
  the deploy key updates it, the owner included; a `v*` tag cannot be moved,
  and only the key creates one. The owner can delete a tag: the recovery below
  requires it.

One-time setup, by the owner of both repositories, in a temporary directory:
never in the repository, where a `git add` would pick it up, nor in `~/.ssh`,
where it would remain a key able to write to the mirror. Once handed over, it
only exists in the secret; lost or due for rotation, it is replaced by a new
one, on both sides.

```bash
cd "$(mktemp -d)"
ssh-keygen -t ed25519 -N "" -C "miroir ormeau-doctrine" -f miroir
gh repo deploy-key add miroir.pub --repo sprimault/ormeau-doctrine --allow-write --title "miroir.yml"
gh secret set MIROIR_CLE --env miroir --repo sprimault/ormeau < miroir
rm miroir miroir.pub
```

Under PowerShell, `-N ""` does not reach `ssh-keygen`: leave it out, and
confirm an empty passphrase twice. The `<` redirection does not exist either;
pass the key through `((Get-Content miroir) -join "`n") | gh secret set …`,
which keeps LF line endings, without which ssh rejects the key on the runner.

### Publishing a version

The tag is set **before** the closing pull request is merged, so that no window
separates announcing the version from its availability:

1. The pull request that dates the `CHANGELOG` section is green.
2. The tag is set on its head and pushed:
   `git tag vX.Y.Z <PR head> && git push origin vX.Y.Z`.
3. `release.yml` first pushes the tag to the mirror, and checks that Composer
   resolves `sprimault/ormeau-doctrine:X.Y.Z` from the real mirror. Binaries,
   draft release and image wait for that proof.
4. The pull request is merged: the README announcing the version reaches
   `master` after the mirror carries it.
5. The draft is reviewed and published.

### If publishing fails

As long as the draft is not published, the number can be reused. After that,
never: a project may have locked the tag's commit, and a fix takes the next
number.

- **The mirror job fails before pushing the tag** (missing secret, rejected
  push): the tag only exists on the main repository, and nothing else has gone
  out. Delete the tag, fix the pull request, set the tag again on its new head:

  ```bash
  git push origin :refs/tags/vX.Y.Z && git tag -d vX.Y.Z
  ```

- **A job fails after the mirror tag, for a transient reason** (network,
  runner): re-run the failed jobs. The split is the same, and pushing an
  identical tag again changes nothing.

  ```bash
  gh run rerun <run id> --failed
  ```

- **The fix needs a commit**: the split changes, and the mirror tag cannot move.
  Delete the tag on both sides and the draft if there is one, fix, set the tag
  again:

  ```bash
  gh api -X DELETE repos/sprimault/ormeau-doctrine/git/refs/tags/vX.Y.Z
  gh release delete vX.Y.Z --yes
  git push origin :refs/tags/vX.Y.Z && git tag -d vX.Y.Z
  ```

  The same deletion applies to a pull request abandoned after the tag was set.

- **The mirror's `master` rejects the fast-forward**: something was written to
  it that does not come from the split. The mirror rule only allows the key, so
  the key has been used elsewhere. Revoke it (delete the deploy key, add a new
  one), compare `git ls-remote` with the local split of `master`, and only then
  put the mirror's `master` back on the split — the one case where a force push
  is allowed, done by hand by the owner, with the rule disabled for the
  duration.

## Signing: the two frictions

They do not prevent publishing, but they are better documented than discovered.

**Windows.** The binary is not signed, so SmartScreen shows a warning on first
launch. Signing is technically feasible from Linux with `osslsigncode`, but it
requires a paid code-signing certificate.

**macOS.** The binary compiles from Linux but is neither signed nor notarized:
Gatekeeper blocks it on first launch. Notarization requires a Mac and a paid Apple
developer account.

In both cases the README warns about it and shows how to verify the archive
before running it. An unexplained security warning on a tool that asks for a
database password stops a careful user dead — and they are right to stop.
