.PHONY: dev test lint outils vulncheck sec build binaries web-build web-types web-types-check web-lint web-test php-test php-lint image image-push clean

# Toute la chaîne Go travaille dans .tmp/, à l'intérieur du dépôt, plutôt
# que dans %TEMP%. `go build` et `go test` y écrivent des exécutables
# éphémères que l'antivirus du poste de développement met en quarantaine
# au moment même où le linker les produit : la compilation échoue alors
# sur un accès refusé, sans rapport apparent avec le code, et de façon
# intermittente. Le répertoire est ignoré par git.
#
# Créé au parsing plutôt que par une dépendance de cible : GOTMPDIR doit
# exister avant la première commande go, quelle que soit la cible.
TMP := $(CURDIR)/.tmp
export GOTMPDIR := $(TMP)/gobuild
_ := $(shell mkdir -p "$(GOTMPDIR)")

# Variables de développement, surchargeables à l'appel :
#   make dev ORMEAU_DEV_PORT=7777
#
# L'interface n'a pas d'URL publique à annoncer : elle tourne sur le
# poste et n'est jamais exposée. Le port reste dynamique par défaut,
# on ne le fixe que pour garder un signet stable en dev.
ORMEAU_DEV_PORT ?= 7777

# dev lance le backend Go et le serveur HMR de Vite côte à côte, puis
# les arrête ensemble. On ouvre http://127.0.0.1:5173 : Vite sert le
# front et relaie /api/* vers le backend (proxy déclaré dans
# web/vite.config.ts).
#
# Dépend de web-build parce que internal/interface embarque dist/ via
# //go:embed et refuse de compiler s'il est absent — un clone frais ne
# l'a jamais.
#
# Le rechargement à chaud du Go passe par wgo, mais son absence ne
# bloque pas : sans lui on perd le rechargement, pas la commande. Sur
# un dépôt public, `make dev` doit fonctionner sur un clone frais sans
# installer quoi que ce soit d'abord.
dev: web-build
	@trap 'kill 0' EXIT INT TERM; \
	(cd web && npm run dev) & \
	if command -v wgo >/dev/null 2>&1; then \
		wgo run ./cmd/ormeau interface --port $(ORMEAU_DEV_PORT) --sans-navigateur; \
	else \
		echo "wgo absent : pas de rechargement automatique du backend"; \
		echo "  go install github.com/bokwoon95/wgo@latest"; \
		go run ./cmd/ormeau interface --port $(ORMEAU_DEV_PORT) --sans-navigateur; \
	fi

# Les cibles Go dépendent de web-build pour la même raison que dev.
# Tant que web/ n'existe pas — avant la phase 9 — web-build ne fait
# rien et les cibles Go restent utilisables sur un clone frais.
test: web-build
	go test -race ./...

# Les tests d'intégration exigent les conteneurs SGBD et portent
# l'étiquette `integration` : `make test` doit passer sans docker.
test-integration: containers
	go test -race -tags integration ./...

# La vérification des types générés est accrochée à lint, pas à test :
# `make test` doit rester exécutable sur un clone frais sans rien
# installer, alors que lint exige déjà golangci-lint et échoue sans lui.
# C'est aussi la cible qu'on lance avant chaque publication.
lint: web-build web-types-check
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint absent : make outils"; exit 1; }
	golangci-lint run
	gofmt -l .

# outils installe l'outillage de développement. À relancer après un changement
# de version de Go : golangci-lint refuse d'analyser du code plus récent que la
# toolchain qui l'a construit, et le message d'erreur ne dit pas quoi faire.
#
# golangci-lint est épinglé, et la CI installe la même version : deux versions
# différentes rendraient un verdict différent sur un code identique.
#
# Pas en deçà de v2.13.0 : les versions antérieures embarquent staticcheck
# v0.7.0, qui panique en analysant internal/poll de la stdlib Go 1.27. Sous
# Linux seulement — le code de ce paquet diffère sous Windows, où le défaut ne
# se manifeste pas.
GOLANGCI_VERSION ?= v2.13.0

outils:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest

# vulncheck utilise l'outil officiel de la Go Team. Il ne signale une CVE
# que si le code appelle effectivement la fonction affectée — beaucoup
# moins bruyant qu'un scan de dépendances brut.
vulncheck: web-build
	@command -v govulncheck >/dev/null 2>&1 || { echo "govulncheck absent : make outils"; exit 1; }
	govulncheck ./...

# sec exécute gosec. Sur ce projet, deux familles comptent plus que les
# autres : les credentials en dur, et la construction de requêtes SQL.
# Les requêtes de catalogue sont des constantes, aucune ne doit être
# assemblée à partir d'une entrée.
sec: web-build
	@command -v gosec >/dev/null 2>&1 || { echo "gosec absent : make outils"; exit 1; }
	gosec ./...

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

# Le binaire de développement va dans .tmp/ pour la même raison que le
# reste de la chaîne. dist/ reste réservé aux artefacts de publication,
# que binaries produit et que l'antivirus ne voit qu'une fois écrits.
build: web-build
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o .tmp/ormeau ./cmd/ormeau

# binaries produit les cinq cibles de publication depuis n'importe
# quelle machine : Go croise nativement, et c'est vrai tant qu'aucun
# pilote n'exige cgo.
#
# CGO_ENABLED=0 est explicite et non implicite : sur une machine Linux,
# cgo est actif par défaut, et un binaire lié dynamiquement à la glibc
# refuserait de démarrer sur Alpine.
#
# Dépend de web-build : sans dist/, //go:embed produit un système de
# fichiers vide et la compilation réussit quand même. On publierait
# cinq binaires avec une interface blanche sans qu'aucun avertissement
# ne le signale.
binaries: web-build
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/ormeau_windows_amd64.exe ./cmd/ormeau
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/ormeau_linux_amd64       ./cmd/ormeau
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/ormeau_linux_arm64       ./cmd/ormeau
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/ormeau_darwin_arm64      ./cmd/ormeau
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/ormeau_darwin_amd64      ./cmd/ormeau
	cd dist && sha256sum ormeau_* > SHA256SUMS

# ── Front (phase 9) ─────────────────────────────────────────────────
# Feature-Sliced Design obligatoire, vérifiée par steiger dans web-lint.
#
# Les cibles ci-dessous ne font rien tant que web/ n'existe pas : le
# squelette doit rester utilisable avant la phase 9, et un `make test`
# qui échoue sur un répertoire absent apprend surtout à ignorer le
# Makefile.
web-build:
	@if [ -f web/package.json ]; then cd web && npm run build; \
	else echo "web/ absent (phase 9), rien a construire"; fi

web-lint:
	cd web && npm run lint && npx steiger ./src

web-test:
	cd web && npm run audit:high && npm run test:run

# web-types régénère les types TypeScript de l'API à partir des
# structures Go de internal/calque et internal/introspection via tygo.
#
# Sur ce projet, ces types ne sont pas un confort : le calque est un
# contrat public versionné, et un front qui lit un champ que le Go
# n'expose plus produit un écran silencieusement faux. À relancer après
# tout changement de structure exposée.
#
# Version épinglée, et la CI installe la même : deux versions de tygo
# qui ne placent pas les commentaires au même endroit feraient échouer
# le contrôle sans qu'aucun type Go n'ait bougé.
TYGO_VERSION ?= v0.2.21

web-types:
	@command -v tygo >/dev/null 2>&1 || { echo "tygo absent : go install github.com/gzuidhof/tygo@$(TYGO_VERSION)"; exit 1; }
	tygo generate --config tools/tygo/tygo.yaml

TYPES_GENERES = web/src/shared/model/calque.ts web/src/shared/model/api.ts

# web-types-check régénère et refuse toute dérive. Même contrôle que la
# CI, à la même source, pour que les deux ne puissent pas diverger.
#
# La comparaison porte sur les fichiers d'avant régénération et non sur
# git : `make lint` tourne au milieu d'un travail en cours, et un diff
# contre HEAD échouerait sur des modifications légitimes non commitées
# — c'est-à-dire exactement quand on vient de toucher une structure.
#
# --strip-trailing-cr : sur un poste Windows, git repose les fichiers
# en CRLF à chaque bascule de branche alors que tygo écrit en LF.
web-types-check:
	@if [ ! -f tools/tygo/tygo.yaml ]; then echo "tygo non configure (phase 9), controle ignore"; exit 0; fi; \
	command -v tygo >/dev/null 2>&1 || { echo "tygo absent : go install github.com/gzuidhof/tygo@$(TYGO_VERSION)"; exit 1; }; \
	tmp=$$(mktemp -d) && cp $(TYPES_GENERES) "$$tmp/" && \
	tygo generate --config tools/tygo/tygo.yaml >/dev/null && \
	ecart=0; for f in $(TYPES_GENERES); do \
		diff -u --strip-trailing-cr "$$tmp/$$(basename $$f)" "$$f" || ecart=1; \
	done; rm -rf "$$tmp"; \
	if [ $$ecart -ne 0 ]; then \
		echo "Les types generes avaient derive des structures Go."; \
		echo "Le resultat est deja regenere : relire et commiter."; \
		exit 1; \
	fi; \
	echo "Types generes conformes aux structures Go"

# ── Paquet Doctrine ─────────────────────────────────────────────────
php-test:
	cd php && composer install --no-interaction && composer test

php-lint:
	cd php && composer analyse && vendor/bin/php-cs-fixer fix --dry-run --diff

# ── SGBD de test ────────────────────────────────────────────────────
containers:
	docker compose -f tests/docker-compose.yml up -d --wait

containers-down:
	docker compose -f tests/docker-compose.yml down -v

# ── Image OCI multi-arch ────────────────────────────────────────────
# Nécessite Docker Buildx et un builder buildkit actif (typiquement
# `docker buildx create --use --name ormeau-builder` une fois pour
# toutes).
#
# Les binaires étant déjà croisés par binaries, buildx n'a pas besoin
# d'émulation QEMU : le Dockerfile copie l'artefact correspondant à
# TARGETOS/TARGETARCH.
#
# Publier une version : une seule construction, deux étiquettes. En
# deux appels, la même source produit deux index différents — les
# horodatages de couches ne sont pas reproductibles — et repointer
# ensuite l'un sur l'autre laisse un index sans étiquette au registre.
IMAGE_TAG  ?= ghcr.io/sprimault/ormeau:latest
IMAGE_TAGS ?= $(IMAGE_TAG)
TAG_FLAGS   = $(foreach t,$(IMAGE_TAGS),-t $(t))
PLATFORMS  ?= linux/amd64,linux/arm64
REVISION   ?= $(shell git rev-parse HEAD 2>/dev/null || echo unknown)

image: binaries
	docker buildx build --platform $(PLATFORMS) \
		--build-arg REVISION=$(REVISION) \
		$(TAG_FLAGS) -f deploy/Dockerfile .

image-push: binaries
	docker buildx build --platform $(PLATFORMS) --push \
		--build-arg REVISION=$(REVISION) \
		$(TAG_FLAGS) -f deploy/Dockerfile .

clean:
	rm -rf .tmp dist web/dist internal/interface/dist
