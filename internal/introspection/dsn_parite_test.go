// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package introspection

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/sprimault/ormeau/internal/introspection/dsntest"
)

// Ce fichier est le seul endroit où le paquet commun connaît pgx : la forme
// clé/valeur est la grammaire de libpq, et lireCleValeur n'a de valeur que tant
// qu'elle lit comme le pilote qui recevra la chaîne. Le test vit ici plutôt que
// dans postgres/ parce que la lecture n'est pas exportée, et ne doit pas l'être
// pour un test.

// TestPariteAvecPgx confronte lireCleValeur à pgconn.ParseConfig sur la table
// partagée, forme clé/valeur seulement.
//
// Les refus d'abord, parce que c'est d'eux que vient le défaut d'origine : une
// chaîne que pgx refuse et qu'Ormeau lirait serait masquée d'après une lecture
// que le pilote n'a pas faite, et pgx la citerait dans son erreur avec un
// masquage à lui, incomplet. L'inverse est admis : Ormeau refuse deux formes
// que pgx accepte, où une valeur glisse dans un nom qu'il affiche.
//
// Ensuite les couples, et seulement ce que pgx laisse observer : hôte, port,
// utilisateur, mot de passe et base dans son Config, les clés qu'il ne connaît
// pas dans RuntimeParams. Ce qu'il consomme lui-même — sslmode, sslpassword —
// n'y apparaît pas ; le masquage de ces clés est vérifié par dsn_test.go.
//
// Pas de t.Parallel() : t.Setenv l'interdit, et il est indispensable. pgx
// complète la chaîne par les variables PG*, ~/.pgpass et le fichier de
// services, qui appartiennent au processus entier. Sans les neutraliser, le
// test comparerait la grammaire au poste de celui qui le lance.
func TestPariteAvecPgx(t *testing.T) {
	for _, variable := range []string{"PGHOST", "PGPORT", "PGDATABASE", "PGUSER", "PGPASSWORD", "PGSERVICE", "PGOPTIONS", "PGAPPNAME"} {
		t.Setenv(variable, "")
	}
	absent := filepath.Join(t.TempDir(), "absent")
	t.Setenv("PGPASSFILE", absent)
	t.Setenv("PGSERVICEFILE", absent)

	for _, c := range dsntest.Table {
		if _, estUneURL := schema(c.DSN); estUneURL {
			continue
		}
		t.Run(c.Nom, func(t *testing.T) {
			config, errPgx := lirePgx(c.DSN)
			couples, errOrmeau := lireCleValeur(c.DSN)

			if refusGrammatical(errPgx) {
				if errOrmeau == nil {
					t.Fatalf("pgx refuse la chaine, Ormeau la lit : %v", errPgx)
				}
				return
			}
			if errOrmeau != nil || errPgx != nil {
				// Refus propre à Ormeau, ou refus de pgx sur le sens d'une
				// valeur (un port qui n'est pas un nombre) : la grammaire n'a
				// rien à comparer.
				return
			}
			comparerCouples(t, couples, config)
		})
	}
}

// Changer de base réécrit une valeur sur place : pgx doit relire la même
// connexion, avec la nouvelle base et le même mot de passe.
func TestAvecBaseRelueParPgx(t *testing.T) {
	for _, variable := range []string{"PGDATABASE", "PGPASSWORD"} {
		t.Setenv(variable, "")
	}
	t.Setenv("PGPASSFILE", filepath.Join(t.TempDir(), "absent"))

	for _, c := range dsntest.Table {
		if _, estUneURL := schema(c.DSN); estUneURL || !c.Lisible {
			continue
		}
		t.Run(c.Nom, func(t *testing.T) {
			avant, err := lirePgx(c.DSN)
			if err != nil {
				return
			}
			for _, base := range []string{"gescom", "ma base", `l'autre`} {
				apres, err := lirePgx(AvecBase(c.DSN, base))
				if err != nil {
					t.Fatalf("base %q : chaine refusee par pgx", base)
				}
				if apres.Database != base || apres.Password != avant.Password {
					t.Errorf("base %q : pgx relit la base %q et un autre mot de passe", base, apres.Database)
				}
			}
		})
	}
}

// lirePgx appelle pgconn.ParseConfig et transforme sa panique en erreur : sur
// une barre oblique inverse finale entre apostrophes, pgx sort de sa chaîne au
// lieu de la refuser.
func lirePgx(dsn string) (config *pgconn.Config, err error) {
	defer func() {
		if r := recover(); r != nil {
			config, err = nil, fmt.Errorf("%w: %v", errPaniquePgx, r)
		}
	}()
	return pgconn.ParseConfig(dsn)
}

// errPaniquePgx marque une chaîne sur laquelle pgx a paniqué.
var errPaniquePgx = errors.New("pgx a panique")

// refusGrammatical dit si pgx a refusé la chaîne sur sa forme, et non sur le
// sens d'une valeur. ParseConfigError ne l'expose que dans son message : si
// pgx en change la formulation, ce test échouera bruyamment plutôt que de
// passer à tort.
func refusGrammatical(err error) bool {
	if errors.Is(err, errPaniquePgx) {
		return true
	}
	var configError *pgconn.ParseConfigError
	return errors.As(err, &configError) && strings.Contains(err.Error(), "failed to parse as keyword/value")
}

// comparerCouples vérifie que chaque couple lu par Ormeau a la valeur que pgx
// en a tirée, quand pgx la laisse voir. La dernière occurrence d'une clé fait
// foi, chez lui comme ici.
func comparerCouples(t *testing.T, couples []couple, config *pgconn.Config) {
	t.Helper()

	derniere := map[string]string{}
	for _, p := range couples {
		derniere[p.cle] = p.valeur
	}

	for cle, valeur := range derniere {
		var vuParPgx string
		switch cle {
		case "host":
			vuParPgx = config.Host
		case "port":
			vuParPgx = strconv.Itoa(int(config.Port))
		case "user":
			// pgx ignore un utilisateur vide et prend celui du système.
			if valeur == "" {
				continue
			}
			vuParPgx = config.User
		case "password":
			vuParPgx = config.Password
		case "dbname":
			vuParPgx = config.Database
		default:
			var visible bool
			if vuParPgx, visible = config.RuntimeParams[cle]; !visible {
				continue
			}
		}
		if vuParPgx != valeur {
			t.Errorf("%s : Ormeau lit %q, pgx %q", cle, valeur, vuParPgx)
		}
	}
}
