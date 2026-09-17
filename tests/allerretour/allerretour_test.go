// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

//go:build allerretour

// Package allerretour vérifie la chaîne complète : la base de test extraite,
// inférée, générée en entités Doctrine, recréée par Doctrine dans une base
// vierge, extraite à nouveau, puis comparée à l'originale.
//
// Le diff n'est jamais vide : ce que Doctrine ne sait pas recréer et ce que
// l'outil écarte volontairement forment une liste fermée par SGBD, dans
// ecarts_test.go et ecarts_sqlserver_test.go. Le test échoue sur un écart que
// la liste ne couvre pas, et sur une entrée de la liste qui ne couvre plus
// rien.
//
// Le préfixe du DSN de test choisit le SGBD : postgres:// par défaut,
// sqlserver:// par make aller-retour-sqlserver.
//
// Rien ici n'est importable : le paquet ne contient que des tests.
package allerretour

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/sprimault/ormeau/internal/calque"
	"github.com/sprimault/ormeau/internal/diff"
	"github.com/sprimault/ormeau/internal/inference"
	"github.com/sprimault/ormeau/internal/introspection"
	"github.com/sprimault/ormeau/internal/introspection/ddltest"
	_ "github.com/sprimault/ormeau/internal/introspection/postgres"
	_ "github.com/sprimault/ormeau/internal/introspection/sqlserver"
)

// dsnParDefaut vise le conteneur PostgreSQL de make containers, comme les
// tests d'intégration du pilote.
const dsnParDefaut = "postgres://postgres:ormeau@127.0.0.1:35432/gescom"

// Chemins relatifs au paquet, qui est le répertoire courant d'un test Go.
const (
	cheminRecreer = "../../php/tests/AllerRetour/recreer.php"
	cheminTravail = "../../.tmp/allerretour"
	baseRecreee   = "allerretour"
	nomDuLogique  = "gescom.logique.json"
	nomDesEcarts  = "ecarts.txt"
	nomParametres = "parametres.json"
	nomDesEntites = "entites"
)

// dialecte porte ce qui change d'un SGBD à l'autre dans la chaîne. Le reste —
// inférence, génération, comparaison — est le même, et c'est ce que le test
// vérifie.
type dialecte struct {
	sgbd string // nom du pilote, passé à recreer.php
	ddl  string // DDL dont la base d'origine doit venir
	// schema est celui du DDL, seul extrait. Hors du schéma par défaut, les
	// entités l'écrivent, et Doctrine y recrée les tables.
	schema string
	port   int // port par défaut, quand le DSN n'en nomme pas
	// lireEmpreinte rend l'empreinte du DDL posée dans la base, nil si elle
	// n'en porte pas.
	lireEmpreinte func(ctx context.Context, dsn string) (*string, error)
	// recreee rend le DSN de la base recréée, sur le même serveur.
	recreee func(origine url.URL) url.URL
}

// dialectes est indexé par le préfixe du DSN.
var dialectes = map[string]dialecte{
	"postgres": {
		sgbd: "postgres", ddl: "../ddl/postgres.sql", schema: "gescom", port: 5432,
		lireEmpreinte: empreintePostgres,
		recreee: func(u url.URL) url.URL {
			u.Path = "/" + baseRecreee
			return u
		},
	},
	"sqlserver": {
		sgbd: "sqlserver", ddl: "../ddl/sqlserver.sql", schema: "ventes", port: 1433,
		lireEmpreinte: empreinteSQLServer,
		// La base d'un DSN SQL Server est un paramètre : le chemin y désigne
		// l'instance nommée.
		recreee: func(u url.URL) url.URL {
			q := u.Query()
			q.Set("database", baseRecreee)
			u.RawQuery = q.Encode()
			return u
		},
	},
}

// dsnDeTest rend le DSN de la base d'origine, qu'ORMEAU_TEST_DSN remplace.
func dsnDeTest() string {
	if dsn := os.Getenv("ORMEAU_TEST_DSN"); dsn != "" {
		return dsn
	}
	return dsnParDefaut
}

// dialecteDe lit le SGBD dans le préfixe du DSN de test, qui doit nommer un
// utilisateur : recreer.php se connecte avec les mêmes identifiants.
func dialecteDe(dsn string) (dialecte, *url.URL, error) {
	adresse, err := url.Parse(dsn)
	if err != nil || adresse.User == nil {
		return dialecte{}, nil, errors.New("l'aller-retour exige un DSN de test en URL qui nomme un utilisateur")
	}
	schema := adresse.Scheme
	if schema == "postgresql" {
		schema = "postgres"
	}
	d, connu := dialectes[schema]
	if !connu {
		return dialecte{}, nil, fmt.Errorf("aller-retour : SGBD %q non pris en charge", adresse.Scheme)
	}
	return d, adresse, nil
}

// TestMain refuse de lancer l'aller-retour contre une base qui ne vient pas du
// DDL du dépôt : un diff calculé sur l'ancien schéma, ou sur une autre base,
// validerait une liste d'écarts qui ne correspond à rien.
func TestMain(m *testing.M) {
	if err := controlerBaseDeTest(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

// controlerBaseDeTest compare l'empreinte posée dans la base d'origine à celle
// du DDL du dépôt.
func controlerBaseDeTest() error {
	d, _, err := dialecteDe(dsnDeTest())
	if err != nil {
		return err
	}

	ctx, annuler := context.WithTimeout(context.Background(), 10*time.Second)
	defer annuler()

	empreinte, err := d.lireEmpreinte(ctx, dsnDeTest())
	if err != nil {
		return err
	}
	return ddltest.Controler(empreinte, d.ddl)
}

// empreintePostgres lit le commentaire de base que tests/ddl/20-empreinte.sh
// a posé.
func empreintePostgres(ctx context.Context, dsn string) (*string, error) {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connexion a la base de test (make containers ?): %w", err)
	}
	defer func() { _ = conn.Close(ctx) }()

	var commentaire *string
	err = conn.QueryRow(ctx, `
SELECT shobj_description(oid, 'pg_database')
FROM pg_database
WHERE datname = current_database()`).Scan(&commentaire)
	if err != nil {
		return nil, fmt.Errorf("lecture de l'empreinte du DDL: %w", err)
	}
	return commentaire, nil
}

// empreinteSQLServer lit la propriété étendue de base que
// tests/ddl/sqlserver-init.sh a posée.
func empreinteSQLServer(ctx context.Context, dsn string) (*string, error) {
	db, err := sql.Open("sqlserver", dsn)
	if err != nil {
		return nil, fmt.Errorf("dsn de test illisible (%s)", introspection.Masquer(dsn))
	}
	defer func() { _ = db.Close() }()

	var valeur string
	err = db.QueryRowContext(ctx, `
SELECT CONVERT(nvarchar(200), value)
FROM sys.extended_properties
WHERE class = 0 AND name = N'ormeau_ddl'`).Scan(&valeur)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("lecture de l'empreinte du DDL (make containers ?): %w", err)
	}
	return &valeur, nil
}

// sortiePHP est ce que recreer.php écrit sur sa sortie standard.
type sortiePHP struct {
	ORM      string `json:"orm"`
	DBAL     string `json:"dbal"`
	Ecartees []struct {
		Entite string `json:"entite"`
		Raison string `json:"raison"`
	} `json:"ecartees"`
}

// TestAllerRetour déroule la chaîne et confronte le diff à la liste fermée.
func TestAllerRetour(t *testing.T) {
	travail, err := filepath.Abs(cheminTravail)
	if err != nil {
		t.Fatalf("répertoire de travail : %v", err)
	}
	if err := os.RemoveAll(travail); err != nil {
		t.Fatalf("nettoyage de %s : %v", travail, err)
	}
	if err := os.MkdirAll(travail, 0o750); err != nil {
		t.Fatalf("création de %s : %v", travail, err)
	}

	dsnOrigine := dsnDeTest()
	d, adresse, err := dialecteDe(dsnOrigine)
	if err != nil {
		t.Fatal(err)
	}
	recreee := d.recreee(*adresse)

	origine := extraire(t, d, dsnOrigine, d.schema)
	empreinte, err := origine.CalculerEmpreinte()
	if err != nil {
		t.Fatalf("empreinte : %v", err)
	}
	origine.Source.Empreinte = empreinte

	logique, _ := inference.Inferer(origine, nil)
	cheminLogique := filepath.Join(travail, nomDuLogique)
	if err := logique.Ecrire(cheminLogique); err != nil {
		t.Fatalf("écriture du calque logique : %v", err)
	}

	sortie := recreer(t, d, travail, cheminLogique, adresse)
	cible, err := cibleDe(sortie)
	if err != nil {
		t.Fatal(err)
	}

	recree := extraire(t, d, recreee.String(), d.schema)
	ecarts := diff.Comparer(origine, recree, diff.Options{})

	c := contexte{origine: origine, recree: recree, logique: logique, php: sortie}
	couverture := confronter(t, d.sgbd, cible, ecarts, c)
	if err := os.WriteFile(filepath.Join(travail, nomDesEcarts), []byte(couverture), 0o600); err != nil {
		t.Errorf("écriture de %s : %v", nomDesEcarts, err)
	}
}

// extraire rend le calque de quelques schémas d'une base.
func extraire(t *testing.T, d dialecte, dsn string, schemas ...string) *calque.Physique {
	t.Helper()

	ctx, annuler := context.WithTimeout(context.Background(), 60*time.Second)
	defer annuler()

	pilote, err := introspection.Ouvrir(ctx, d.sgbd, dsn)
	if err != nil {
		t.Fatalf("connexion à %s : %v", introspection.Masquer(dsn), err)
	}
	defer func() { _ = pilote.Fermer() }()

	physique, err := pilote.Extraire(ctx, introspection.Portee{Schemas: schemas})
	if err != nil {
		t.Fatalf("extraction de %s : %v", introspection.Masquer(dsn), err)
	}
	return physique
}

// recreer lance recreer.php : génération des entités, puis création de leur
// schéma par Doctrine dans la base recréée.
//
// L'interpréteur est ORMEAU_PHP, découpé sur les blancs, ou php : sur une
// machine où PHP tourne en conteneur, la commande docker run qui monte le
// dépôt au même chemin. Les paramètres passent par un fichier plutôt que par
// la ligne de commande, où le mot de passe serait visible.
func recreer(t *testing.T, d dialecte, travail, cheminLogique string, adresse *url.URL) sortiePHP {
	t.Helper()

	script, err := filepath.Abs(cheminRecreer)
	if err != nil {
		t.Fatalf("chemin de recreer.php : %v", err)
	}
	port, err := strconv.Atoi(adresse.Port())
	if err != nil {
		port = d.port
	}
	motDePasse, _ := adresse.User.Password()

	// Des paramètres pour un script, pas un calque : JSON direct.
	parametres, err := json.Marshal(map[string]any{
		"sgbd":         d.sgbd,
		"logique":      cheminLogique,
		"entites":      filepath.Join(travail, nomDesEntites),
		"base":         baseRecreee,
		"schema":       d.schema,
		"hote":         adresse.Hostname(),
		"port":         port,
		"utilisateur":  adresse.User.Username(),
		"mot_de_passe": motDePasse,
	})
	if err != nil {
		t.Fatalf("paramètres : %v", err)
	}
	cheminParametres := filepath.Join(travail, nomParametres)
	if err := os.WriteFile(cheminParametres, parametres, 0o600); err != nil {
		t.Fatalf("écriture des paramètres : %v", err)
	}

	commande := strings.Fields(os.Getenv("ORMEAU_PHP"))
	if len(commande) == 0 {
		commande = []string{"php"}
	}
	commande = append(commande, script, cheminParametres)

	ctx, annuler := context.WithTimeout(context.Background(), 5*time.Minute)
	defer annuler()

	// #nosec G204 -- commande fournie par qui lance le test, sur son poste.
	cmd := exec.CommandContext(ctx, commande[0], commande[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("recreer.php a échoué (%v) :\n%s%s", err, stderr.String(), stdout.String())
	}

	var sortie sortiePHP
	if err := json.Unmarshal(stdout.Bytes(), &sortie); err != nil {
		t.Fatalf("sortie de recreer.php illisible (%v) :\n%s", err, stdout.String())
	}
	return sortie
}

// cibleDe nomme la cible d'après les majeures d'ORM et de DBAL installées : la
// liste d'écarts en dépend, le DDL étant l'affaire de DBAL.
func cibleDe(s sortiePHP) (string, error) {
	orm, _, _ := strings.Cut(strings.TrimPrefix(s.ORM, "v"), ".")
	dbal, _, _ := strings.Cut(strings.TrimPrefix(s.DBAL, "v"), ".")
	if orm == "" || dbal == "" {
		return "", fmt.Errorf("versions illisibles dans la sortie de recreer.php : orm %q, dbal %q", s.ORM, s.DBAL)
	}
	return "orm" + orm + "-dbal" + dbal, nil
}
