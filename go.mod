module github.com/sprimault/ormeau

go 1.27

require (
	github.com/jackc/pgx/v5 v5.10.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.16.0 // indirect
	// Relevé au-dessus de ce que pgx demande (v0.29.0) : GO-2026-5970, boucle
	// infinie atteinte depuis ConnectConfig. À retirer quand pgx exigera au
	// moins v0.39.0 de lui-même.
	golang.org/x/text v0.39.0 // indirect
)
