> [🇬🇧 English](README.md) · [🇫🇷 Français](README.fr.md)

# ormeau-doctrine

The Symfony bundle of [Ormeau](https://github.com/sprimault/ormeau): it
generates Doctrine entities from the logical layer that `ormeau inferer`
produces, and regenerates them without discarding the work done on them since.

> [!IMPORTANT]
> This repository is a **read-only mirror** of the `php/` directory of
> [sprimault/ormeau](https://github.com/sprimault/ormeau), rewritten on every
> merge. Issues, pull requests and documentation live in the main repository.

## Installation

Not on Packagist yet: the repository is declared first.

```console
$ composer config repositories.ormeau vcs https://github.com/sprimault/ormeau-doctrine
$ composer require --dev sprimault/ormeau-doctrine
$ bin/console ormeau:generer gescom.logique.json
```

The whole workflow — extracting the layer, inferring, arbitrating, generating —
is described in the [main README](https://github.com/sprimault/ormeau#readme).

## License

Apache 2.0 — see [`LICENSE`](LICENSE). The entities Ormeau generates are yours:
the licence covers the tool, not its output.
