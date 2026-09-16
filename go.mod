module github.com/sprimault/ormeau

go 1.27

require (
	github.com/jackc/pgx/v5 v5.10.0
	github.com/microsoft/go-mssqldb v1.11.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.16.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	golang.org/x/crypto v0.56.0 // indirect
	// Relevé au-dessus de ce que pgx demande (v0.29.0) : GO-2026-5970, boucle
	// infinie atteinte depuis ConnectConfig. À retirer quand pgx exigera au
	// moins v0.39.0 de lui-même.
	golang.org/x/text v0.41.0 // indirect
)
