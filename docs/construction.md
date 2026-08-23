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

The front end is built **before** the binaries. If `web/dist` does not exist at
compile time, `embed` produces an empty filesystem and the build still succeeds:
you then publish binaries with a blank interface, with no warning whatsoever.

The `binaries` target depends on `web-build` for that reason, and a test checks
that the embedded filesystem is not empty.

## Local builds

`GOTMPDIR` points at `.tmp/gobuild`, inside the repository, and the Makefile
exports it. Otherwise antivirus software on a Windows workstation quarantines the
temporary executables the linker writes to `%TEMP%`, and the build fails on an
access-denied error with no apparent connection to the code. `make build` writes
to `.tmp/` for the same reason; `dist/` stays reserved for release artefacts.

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

## Signing: the two frictions

They do not prevent publishing, but they are better documented than discovered.

**Windows.** The binary is not signed, so SmartScreen shows a warning on first
launch. Signing is technically feasible from Linux with `osslsigncode`, but it
requires a paid code-signing certificate.

**macOS.** The binary compiles from Linux but is neither signed nor notarized:
Gatekeeper blocks it on first launch. Notarization requires a Mac and a paid Apple
developer account.

In both cases the README explains the workaround. An unexplained security warning
on a tool that asks for a database password stops a careful user dead — and they
are right to stop.
