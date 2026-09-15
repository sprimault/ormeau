// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

//go:build allerretour

// Package allerretour vérifie la chaîne complète : la base de test extraite,
// inférée, générée en entités Doctrine, recréée par Doctrine dans une base
// vierge, extraite à nouveau, puis comparée à l'originale.
//
// Le diff n'est jamais vide sur PostgreSQL : ce que Doctrine ne sait pas
// recréer et ce que l'outil écarte volontairement forment une liste fermée,
// dans ecarts_test.go. Le test échoue sur un écart que la liste ne couvre pas,
// et sur une entrée de la liste qui ne couvre plus rien.
//
// Rien ici n'est importable : le paquet ne contient que des tests.
package allerretour

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	_ "github.com/sprimault/ormeau/internal/introspection/postgres"
)

// dsnParDefaut vise le conteneur de make containers, comme les tests
// d'intégration du pilote.
const dsnParDefaut = "postgres://postgres:ormeau@127.0.0.1:35432/gescom"

// Chemins relatifs au paquet, qui est le répertoire courant d'un test Go.
const (
	cheminDDL     = "../ddl/postgres.sql"
	cheminRecreer = "../../php/tests/AllerRetour/recreer.php"
	cheminTravail = "../../.tmp/allerretour"
	baseRecreee   = "allerretour"
	schemaDeTest  = "gescom"
	nomDuLogique  = "gescom.logique.json"
	nomDesEcarts  = "ecarts.txt"
	nomParametres = "parametres.json"
	nomDesEntites = "entites"
)

// dsnDeTest rend le DSN de la base d'origine, qu'ORMEAU_TEST_DSN remplace.
func dsnDeTest() string {
	if dsn := os.Getenv("ORMEAU_TEST_DSN"); dsn != "" {
		return dsn
	}
	return dsnParDefaut
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

// controlerBaseDeTest compare l'empreinte du DDL du dépôt à celle que
// tests/ddl/20-empreinte.sh a posée en commentaire de la base.
//
// Deuxième occurrence du contrôle de internal/introspection/postgres
// (postgres_integration_test.go), où il lit la base par le pilote : un fichier
// de test ne s'importe pas. À la troisième, l'extraire.
func controlerBaseDeTest() error {
	contenu, err := os.ReadFile(cheminDDL)
	if err != nil {
		return fmt.Errorf("lecture du DDL de test: %w", err)
	}
	somme := sha256.Sum256(contenu)
	attendu := "ddl-sha256:" + hex.EncodeToString(somme[:])

	ctx, annuler := context.WithTimeout(context.Background(), 10*time.Second)
	defer annuler()

	conn, err := pgx.Connect(ctx, dsnDeTest())
	if err != nil {
		return fmt.Errorf("connexion a la base de test (make containers ?): %w", err)
	}
	defer func() { _ = conn.Close(ctx) }()

	var commentaire *string
	err = conn.QueryRow(ctx, `
SELECT shobj_description(oid, 'pg_database')
FROM pg_database
WHERE datname = current_database()`).Scan(&commentaire)
	if err != nil {
		return fmt.Errorf("lecture de l'empreinte du DDL: %w", err)
	}

	switch {
	case commentaire == nil || !strings.HasPrefix(*commentaire, "ddl-sha256:"):
		return errors.New("la base visee ne vient pas de tests/ddl/, aucune empreinte de DDL " +
			"en commentaire: verifier ORMEAU_TEST_DSN, ou recreer le conteneur par make containers")
	case *commentaire != attendu:
		return fmt.Errorf("la base de test vient d'un autre DDL que tests/ddl/postgres.sql "+
			"(base %s, fichier %s): make containers la recree", *commentaire, attendu)
	}
	return nil
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
	adresse, err := url.Parse(dsnOrigine)
	if err != nil || adresse.User == nil || (adresse.Scheme != "postgres" && adresse.Scheme != "postgresql") {
		t.Fatal("l'aller-retour exige un DSN de test en URL postgres:// qui nomme un utilisateur")
	}
	recreee := *adresse
	recreee.Path = "/" + baseRecreee

	origine := extraire(t, dsnOrigine)
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

	sortie := recreer(t, travail, cheminLogique, adresse)
	cible, err := cibleDe(sortie)
	if err != nil {
		t.Fatal(err)
	}

	recree := extraire(t, recreee.String())
	ecarts := diff.Comparer(origine, recree, diff.Options{})

	c := contexte{origine: origine, recree: recree, logique: logique, php: sortie}
	couverture := confronter(t, cible, ecarts, c)
	if err := os.WriteFile(filepath.Join(travail, nomDesEcarts), []byte(couverture), 0o600); err != nil {
		t.Errorf("écriture de %s : %v", nomDesEcarts, err)
	}
}

// extraire rend le calque du schéma de test d'une base.
func extraire(t *testing.T, dsn string) *calque.Physique {
	t.Helper()

	ctx, annuler := context.WithTimeout(context.Background(), 60*time.Second)
	defer annuler()

	pilote, err := introspection.Ouvrir(ctx, "postgres", dsn)
	if err != nil {
		t.Fatalf("connexion à %s : %v", introspection.Masquer(dsn), err)
	}
	defer func() { _ = pilote.Fermer() }()

	physique, err := pilote.Extraire(ctx, introspection.Portee{Schemas: []string{schemaDeTest}})
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
func recreer(t *testing.T, travail, cheminLogique string, adresse *url.URL) sortiePHP {
	t.Helper()

	script, err := filepath.Abs(cheminRecreer)
	if err != nil {
		t.Fatalf("chemin de recreer.php : %v", err)
	}
	port, err := strconv.Atoi(adresse.Port())
	if err != nil {
		port = 5432
	}
	motDePasse, _ := adresse.User.Password()

	// Des paramètres pour un script, pas un calque : JSON direct.
	parametres, err := json.Marshal(map[string]any{
		"logique":      cheminLogique,
		"entites":      filepath.Join(travail, nomDesEntites),
		"base":         baseRecreee,
		"schema":       schemaDeTest,
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
