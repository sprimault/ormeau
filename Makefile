.PHONY: construire front binaires tester verifier conteneurs propre

BINAIRE := ormeau
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

construire: front
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINAIRE) ./cmd/ormeau

front:
	cd web && npm ci && npm run build

# Le front se construit avant : sinon les binaires embarquent une interface
# vide, et rien ne le signale à la compilation.
binaires: front
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/$(BINAIRE)_windows_amd64.exe ./cmd/ormeau
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/$(BINAIRE)_linux_amd64      ./cmd/ormeau
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/$(BINAIRE)_linux_arm64      ./cmd/ormeau
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/$(BINAIRE)_darwin_arm64     ./cmd/ormeau
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/$(BINAIRE)_darwin_amd64     ./cmd/ormeau
	cd dist && sha256sum * > SHA256SUMS

tester:
	go test -race ./...

verifier:
	gofmt -l .
	go vet ./...
	govulncheck ./...

conteneurs:
	docker compose -f tests/docker-compose.yml up -d --wait

propre:
	rm -rf $(BINAIRE) dist web/dist
	docker compose -f tests/docker-compose.yml down -v
