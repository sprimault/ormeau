.PHONY: dev test cover maj-attendus maj-calque-gescom maj-calque-sqlserver aller-retour aller-retour-sqlserver lint outils vulncheck sec build binaries web-deps web-build web-types web-types-check web-lint web-test php-changelog php-test php-lint image image-tags image-push clean

# Répertoire de travail local, ignoré par git : sorties de `make build`,
# profils de couverture, tout ce qui ne se publie pas.
TMP := $(CURDIR)/.tmp
_ := $(shell mkdir -p "$(TMP)")

# Réglages propres au poste : chemins de cache Go, port de développement,
# DSN de travail. Non versionné, et absent chez tout le monde sauf celui
# qui en a besoin.
#
# C'est là que se pose GOTMPDIR sur une machine dont l'antivirus met en
# quarantaine les exécutables au moment où le linker les produit. Rien
# n'oblige les autres à s'en soucier, et rediriger GOCACHE par défaut
# leur coûterait un cache de compilation froid à chaque clone — sans
# parler de celui de la CI, qui vise ~/.cache/go-build.
-include makefile.local

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
# Dépend de web-build parce que internal/interface embarque le bundle
# de Vite via //go:embed et refuse de compiler sans lui — un clone frais
# ne l'a jamais.
#
# Le rechargement à chaud du Go passe par wgo, mais son absence ne
# bloque pas : sans lui on perd le rechargement, pas la commande. Sur
# un dépôt public, `make dev` doit fonctionner sur un clone frais avec Go
# et Node pour seuls prérequis : les dépendances du front s'installent
# au premier passage.
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

# Les cibles Go dépendent de web-build pour la même raison que dev : sans
# bundle, la compilation de cmd/ormeau échoue.
test: web-build
	go test -race ./...

# maj-attendus réécrit les calques logiques attendus des cas de
# référence. Cible séparée, jamais appelée par `make test` : un attendu
# régénéré sans être relu n'enregistre pas le comportement voulu mais le
# comportement courant, bugs compris, et le déclare correct.
#
# Relire le diff avant de commiter fait partie de la manœuvre.
maj-attendus:
	go test ./internal/inference/ -run TestReference -maj-attendus
	@echo "Attendus reecrits. Relire 'git diff tests/reference/' avant de commiter."

# cover passe par le Makefile et pas par la main : `go tool cover` bâtit
# son propre exécutable, et ne l'appeler qu'ici garantit qu'il atterrit
# dans .tmp/ comme le reste.
cover: web-build
	go test -coverprofile="$(TMP)/cover.out" ./...
	go tool cover -func="$(TMP)/cover.out"
	@echo "Detail par ligne : go tool cover -html=$(TMP)/cover.out"

# Les tests d'intégration exigent les conteneurs SGBD et portent
# l'étiquette `integration` : `make test` doit passer sans docker.
#
# -count=1 désactive le cache : Go le réutilise tant que le code ne bouge pas,
# alors que le résultat dépend ici de l'état de la base. Un « ok (cached) »
# devant un conteneur recréé ne prouve rien.
#
# Seuls les paquets qui portent l'étiquette, et pas ./... : internal/interface
# ne compile pas sans le front construit (go:embed), et un pilote SQL n'a pas à
# exiger Node pour se tester. Les tests unitaires restent à `make test`.
# --untracked : un fichier d'intégration neuf compte avant d'être ajouté.
PAQUETS_INTEGRATION = $(sort $(foreach f,$(shell git grep --untracked -l '^//go:build integration' -- '*_test.go'),./$(dir $(f))))

test-integration: containers
	go test -race -count=1 -tags integration $(PAQUETS_INTEGRATION)

# aller-retour extrait gescom, en génère les entités, laisse Doctrine recréer
# leur schéma dans la base allerretour du conteneur, extrait celle-ci et
# compare. Il exige PHP avec pdo_pgsql et les dépendances de php/ installées :
# ORMEAU_PHP désigne l'interpréteur quand ce n'est pas `php` (une commande
# docker run qui monte le dépôt au même chemin, par exemple), et la version
# d'ORM installée fixe la cible, donc la liste d'écarts. Pas de -race : le
# harnais ne lance aucune goroutine, et le détecteur ralentit l'extraction.
aller-retour: CONTENEURS = postgres
aller-retour: containers
	go test -count=1 -tags allerretour ./tests/allerretour/

# La même chaîne sous SQL Server. Il exige PHP avec pdo_sqlsrv et le pilote
# ODBC 18 de Microsoft ; la base allerretour est recréée à chaque passage par
# le script PHP, qui s'y connecte avec le compte sa du conteneur.
DSN_ALLERRETOUR_SQLSERVER = sqlserver://sa:Ormeau!2026@127.0.0.1:31433?database=gescom&TrustServerCertificate=true

aller-retour-sqlserver: CONTENEURS = sqlserver
aller-retour-sqlserver: containers
	ORMEAU_TEST_DSN='$(DSN_ALLERRETOUR_SQLSERVER)' go test -count=1 -tags allerretour ./tests/allerretour/

# Le calque de gescom est l'extraction réelle de tests/ddl/, versionnée comme
# entrée du cas d'inférence du même nom. Il se régénère contre un conteneur
# recréé, puis make maj-attendus en tire le calque logique : les deux diffs se
# relisent avant de commiter.
maj-calque-gescom: CONTENEURS = postgres
maj-calque-gescom: containers
	go test -count=1 -tags integration ./internal/introspection/postgres/ -run TestExtraireCommeLaReference -maj-attendus
	@echo "Calque reecrit. Relire 'git diff tests/reference/inference/gescom/', puis make maj-attendus."

# Le calque SQL Server de gescom n'entre dans aucun cas d'inférence : il fige
# l'extraction seule, en attendant que l'inférence de ce dialecte soit relue.
maj-calque-sqlserver: CONTENEURS = sqlserver
maj-calque-sqlserver: containers
	go test -count=1 -tags integration ./internal/introspection/sqlserver/ -run TestExtraireCommeLaReference -maj-attendus
	@echo "Calque reecrit. Relire 'git diff tests/reference/extraction/sqlserver/'."

# La vérification des types générés est accrochée à lint, pas à test :
# `make test` doit rester exécutable sur un clone frais sans outil Go à
# installer, alors que lint exige golangci-lint et tygo, que pose
# `make outils`. lint construit aussi le front, dont go vet a besoin pour
# compiler cmd/ormeau. C'est la cible qu'on lance avant chaque publication.
#
# gofmt -l ne fait qu'afficher : l'échec se décide ici. La liste vient de git,
# fichiers neufs non ajoutés compris, pour contrôler avant le commit sans
# descendre dans ce que .gitignore écarte (.tmp/ peut contenir des sources).
# golangci-lint ne ferait pas l'affaire : il ignore les fichiers d'une
# étiquette de build inactive, comme les tests d'intégration.
lint: web-build web-types-check
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint absent : make outils"; exit 1; }
	golangci-lint run
	@ecarts=$$(git ls-files -z -co --exclude-standard '*.go' | xargs -0 gofmt -l); \
	if [ -n "$$ecarts" ]; then echo "gofmt : fichiers a formater"; echo "$$ecarts"; exit 1; fi

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

# govulncheck et gosec sont épinglés aussi, et la CI installe les mêmes : en
# @latest, un outil adopté le jour de sa sortie tourne dans des jobs dont le
# cache Go est partagé. La plus ancienne version qui analyse du Go 1.27 sans
# erreur, sous Linux comme sous Windows (essai du 2026-09-14). Épingler
# govulncheck ne fige pas sa base d'avis, qu'il interroge en direct.
GOVULNCHECK_VERSION ?= v1.7.0
GOSEC_VERSION ?= v2.28.0

outils:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
	go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)
	go install github.com/gzuidhof/tygo@$(TYGO_VERSION)

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
# Dépend de web-build : sans bundle, //go:embed refuse de compiler, et
# un répertoire présent mais vide passerait la compilation pour publier
# cinq binaires à interface blanche. C'est TestEmbarqueNonVide qui
# arrête ce second cas.
binaries: web-build
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/ormeau_windows_amd64.exe ./cmd/ormeau
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/ormeau_linux_amd64       ./cmd/ormeau
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/ormeau_linux_arm64       ./cmd/ormeau
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/ormeau_darwin_arm64      ./cmd/ormeau
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/ormeau_darwin_amd64      ./cmd/ormeau
	cd dist && sha256sum ormeau_* > SHA256SUMS

# ── Front ───────────────────────────────────────────────────────────
# Feature-Sliced Design obligatoire, vérifiée par steiger dans web-lint.
#
# Les cibles ci-dessous ne font rien tant que web/ n'existe pas : le
# dépôt doit rester utilisable avant l'interface, et un `make test` qui
# échoue sur un répertoire absent apprend surtout à ignorer le Makefile.
#
# web-deps installe les dépendances du front quand web/node_modules est
# absent, et seulement dans ce cas : sur un clone frais, il n'y a rien à
# perdre, et c'est à l'outil de le faire plutôt qu'à un message de le
# demander. Jamais relancé quand le répertoire existe : npm ci commence
# par le vider, et un poste où npm ne peut rien télécharger — proxy,
# certificat intercepté par un antivirus — perdrait des dépendances
# installées autrement. Les dates de fichier ne diraient pas si
# l'installation suit le lock, tar et git ne les posant pas de la même
# façon : un lock modifié se rattrape par `cd web && npm ci`, et la CI
# installe toujours depuis le lock.
#
# Un npm ci qui échoue laisse un node_modules partiel, que le passage
# suivant prendrait pour une installation. Il est retiré : il n'existait
# pas avant, il n'y a rien à perdre.
web-deps:
	@if [ -f web/package.json ] && [ ! -d web/node_modules ]; then \
		echo "web/node_modules absent : npm ci"; \
		cd web && npm ci || { rm -rf node_modules; echo "npm ci a echoue : verifier le proxy ou le certificat, ou installer web/node_modules sur une autre machine"; exit 1; }; \
	fi

web-build: web-deps
	@if [ -f web/package.json ]; then cd web && npm run build; \
	else echo "web/ absent, rien a construire"; fi

# tsc avant eslint : `vite build` transpile sans contrôler les types, un
# projet peut donc se construire en étant faux. Le typage se vérifie
# séparément ou pas du tout.
web-lint: web-deps
	@if [ ! -f web/package.json ]; then echo "web/ absent, rien a controler"; exit 0; fi; \
	cd web && npx tsc -b && npm run lint && npx steiger ./src

web-test: web-deps
	@if [ ! -f web/package.json ]; then echo "web/ absent, rien a tester"; exit 0; fi; \
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

TYPES_GENERES = web/src/shared/model/calque.ts web/src/shared/model/api.ts web/src/shared/model/introspection.ts web/src/shared/model/inference.ts web/src/shared/model/config.ts

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
	@if [ ! -f tools/tygo/tygo.yaml ]; then echo "tygo non configure, controle ignore"; exit 0; fi; \
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
# Le journal du paquet s'arrête à la 0.5.0 : le miroir commence là, et les
# versions antérieures n'y ont jamais existé. La CI rejoue la cible et refuse
# un php/CHANGELOG.md qui en diffère. Le retour chariot est retiré parce qu'une
# copie Windows en core.autocrlf en porte, et qu'une ligne vide n'y serait plus
# reconnue.
php-changelog:
	awk '{ sub(/\r$$/, "") } /^## \[0\.4\.2\]/ { exit } /^$$/ { vides++; next } { while (vides) { print ""; vides-- } print }' \
		CHANGELOG.md > php/CHANGELOG.md

php-test:
	cd php && composer install --no-interaction && composer test

# composer audit interroge la base d'avis en direct : un lint vert le matin
# peut être rouge l'après-midi sur le même lock. C'est voulu, une faille
# n'attend pas la prochaine contribution.
#
# Le code produit a son propre contrôle : exclu du style du paquet, il doit
# suivre PER-CS comme du code écrit à la main. Ce que l'outil réécrit à chaque
# passage ne doit en outre rien donner à corriger à @Symfony, sans quoi chaque
# régénération produit un diff dans les projets qui suivent ces règles.
php-lint:
	cd php && composer analyse && vendor/bin/php-cs-fixer fix --dry-run --diff \
		&& vendor/bin/php-cs-fixer fix --dry-run --diff --rules=@PER-CS2.0 --using-cache=no tests/Generation/attendus \
		&& composer style-produit-symfony \
		&& composer audit

# ── SGBD de test ────────────────────────────────────────────────────
# Recréé à chaque appel, volume compris : l'image n'exécute tests/ddl/ que sur
# un volume vide, et un conteneur resté debout garderait l'ancien schéma. Les
# tests d'intégration le vérifient de toute façon (empreinte du DDL en
# commentaire de la base) ; ceci évite d'y tomber par le chemin normal.
#
# CONTENEURS restreint aux services nommés, vide pour tous. Chaque aller-retour
# ne démarre que son SGBD : l'image SQL Server pèse un gigaoctet et demi, qu'un
# job PostgreSQL téléchargerait pour rien.
CONTENEURS ?=

containers:
	docker compose -f tests/docker-compose.yml up -d --wait --force-recreate --renew-anon-volumes $(CONTENEURS)

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
# Une version se publie par release.yml, qui construit dans la chaîne
# attestée. image-push reste pour pousser à la main une image explicitement
# étiquetée : une seule construction pour plusieurs étiquettes, parce qu'en
# deux appels la même source produit deux index différents — les horodatages
# de couches ne sont pas reproductibles — et repointer ensuite l'un sur
# l'autre laisse un index sans étiquette au registre.
#
# Aucune étiquette par défaut : un :latest implicite ferait rapporter à un
# docker pull sans étiquette une image construite sur un poste, ce que
# release.yml exclut en 0.x.
IMAGE_TAGS ?=
TAG_FLAGS   = $(foreach t,$(IMAGE_TAGS),-t $(t))
PLATFORMS  ?= linux/amd64,linux/arm64
REVISION   ?= $(shell git rev-parse HEAD 2>/dev/null || echo unknown)

image: binaries
	docker buildx build --platform $(PLATFORMS) \
		--build-arg VERSION=$(VERSION) --build-arg REVISION=$(REVISION) \
		$(TAG_FLAGS) -f deploy/Dockerfile .

# Le refus passe avant binaires : sans étiquette, rien n'est construit.
image-tags:
	@test -n "$(IMAGE_TAGS)" || { echo "IMAGE_TAGS requis, par exemple IMAGE_TAGS=ghcr.io/sprimault/ormeau:vX.Y.Z ; une version se publie par release.yml"; exit 1; }

image-push: image-tags binaries
	docker buildx build --platform $(PLATFORMS) --push \
		--build-arg VERSION=$(VERSION) --build-arg REVISION=$(REVISION) \
		$(TAG_FLAGS) -f deploy/Dockerfile .

clean:
	rm -rf .tmp dist web/dist internal/interface/dist
