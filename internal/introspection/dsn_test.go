// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package introspection

import (
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/sprimault/ormeau/internal/introspection/dsntest"
)

// TestSGBDDepuisDSN couvre les schémas d'URL de chaque dialecte, y compris les
// alias : postgresql vaut postgres, sqlserver vaut mssql.
func TestSGBDDepuisDSN(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom     string
		dsn     string
		attendu string
	}{
		{"postgres", "postgres://u:p@h:5432/base", "postgres"},
		{"postgresql", "postgresql://u:p@h/base", "postgres"},
		{"forme cle/valeur libpq", "host=serveur user=u password=p dbname=base", "postgres"},
		{"mysql", "mysql://u:p@h:3306/base", "mysql"},
		{"mariadb", "mariadb://u:p@h/base", "mariadb"},
		{"sqlserver", "sqlserver://u:p@h:1433?database=base", "sqlserver"},
		{"mssql", "mssql://u:p@h/base", "sqlserver"},
		{"oracle", "oracle://u:p@h:1521/SERVICE", "oracle"},
		{"sqlite", "sqlite://fichier.db", "sqlite"},
		{"prefixe en majuscules", "POSTGRES://u:p@h/base", "postgres"},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			obtenu, err := SGBDDepuisDSN(c.dsn)
			if err != nil {
				t.Fatalf("resolution : %v", err)
			}
			if obtenu != c.attendu {
				t.Errorf("sgbd %q, attendu %q", obtenu, c.attendu)
			}
		})
	}
}

// TestSGBDDepuisDSNRefuse vérifie qu'un schéma inconnu est refusé plutôt que
// deviné : choisir un pilote au hasard produirait une erreur de protocole.
func TestSGBDDepuisDSNRefuse(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom string
		dsn string
	}{
		{"vide", ""},
		{"sans prefixe", "serveur:5432/base"},
		{"prefixe inconnu", "db2://u:p@h/base"},
		{"chemin nu", "/var/lib/base.db"},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			if _, err := SGBDDepuisDSN(c.dsn); err == nil {
				t.Error("aucune erreur")
			}
		})
	}
}

// Une erreur de résolution ne doit pas recopier le DSN : un DSN mal formé
// contient quand même un mot de passe.
func TestSGBDDepuisDSNNeDivulguePasLeSecret(t *testing.T) {
	t.Parallel()

	const secret = "motdepasse-tres-secret"
	_, err := SGBDDepuisDSN("db2://utilisateur:" + secret + "@hote/base")
	if err == nil {
		t.Fatal("aucune erreur")
	}
	if strings.Contains(err.Error(), secret) {
		t.Errorf("le mot de passe apparait dans l'erreur : %v", err)
	}
}

// Le test qui compte : aucune forme de DSN ne doit laisser passer son mot de
// passe une fois masquée.
func TestMasquerNeLaissePasserAucunSecret(t *testing.T) {
	t.Parallel()

	for _, c := range dsntest.Table {
		if masque := Masquer(c.DSN); strings.Contains(masque, dsntest.Secret) {
			t.Errorf("%s : secret visible dans %q", c.Nom, masque)
		}
	}
}

// Une chaîne qu'Ormeau ne sait pas lire est refusée à l'aiguillage, avant
// qu'un pilote la lise autrement, et se masque en entier. Une chaîne lisible
// garde sa structure une fois masquée.
func TestLisibiliteDesDSN(t *testing.T) {
	t.Parallel()

	for _, c := range dsntest.Table {
		t.Run(c.Nom, func(t *testing.T) {
			t.Parallel()

			_, err := SGBDDepuisDSN(c.DSN)
			if c.Lisible && err != nil {
				t.Errorf("refusee a l'aiguillage : %v", err)
			}
			if !c.Lisible && err == nil {
				t.Error("acceptee a l'aiguillage")
			}
			if err != nil && strings.Contains(err.Error(), dsntest.Secret) {
				t.Errorf("secret visible dans l'erreur : %v", err)
			}

			if masque := Masquer(c.DSN); (masque == "***") == c.Lisible {
				t.Errorf("masque %q", masque)
			}
		})
	}
}

// La table elle-même : un cas sans le secret ne prouverait rien, et ferait
// passer au vert chaque test qui la parcourt.
func TestChaqueCasPorteLeSecret(t *testing.T) {
	t.Parallel()

	for _, c := range dsntest.Table {
		if !strings.Contains(c.DSN, dsntest.Secret) {
			t.Errorf("%s : %q ne porte pas %q", c.Nom, c.DSN, dsntest.Secret)
		}
	}
}

// Chaque clé de clesAffichees garde sa valeur, dans les deux formes ; une clé
// qui n'y figure pas perd la sienne, qu'on la sache secrète ou non. Le test lit
// la même liste que Masquer : il vérifie qu'elle est appliquée partout, pas
// qu'elle est complète — c'est l'inversion de la liste qui en tient lieu.
func TestMasquerSuitLaListeDesClesAffichees(t *testing.T) {
	t.Parallel()

	for cle := range clesAffichees {
		t.Run(cle, func(t *testing.T) {
			t.Parallel()

			if !strings.ContainsAny(cle, espaces) {
				dsn := "host=hote " + cle + "=valeur"
				if masque := Masquer(dsn); !strings.Contains(masque, cle+"=valeur") {
					t.Errorf("forme cle/valeur : %q", masque)
				}
			}

			dsn := "postgres://u@hote/base?" + url.Values{cle: {"valeur"}}.Encode()
			if masque := Masquer(dsn); !strings.Contains(masque, "=valeur") {
				t.Errorf("forme url : %q", masque)
			}
		})
	}

	for _, dsn := range []string{
		"host=hote options=valeur",
		"host=hote OPTIONS=valeur",
		"postgres://u@hote/base?options=valeur",
	} {
		if masque := Masquer(dsn); strings.Contains(masque, "valeur") {
			t.Errorf("Masquer(%q) = %q, attendu la valeur masquee", dsn, masque)
		}
	}
}

// Le front masque un DSN saisi avec sa propre copie de clesAffichees. Une clé
// ajoutée d'un seul côté y ferait afficher, ou masquer, ce que l'autre traite
// autrement : les deux listes doivent être identiques.
func TestClesAfficheesIdentiquesAuFront(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile(filepath.Join("..", "..", "web", "src", "shared", "lib", "strings.ts"))
	if err != nil {
		t.Fatalf("lecture de strings.ts : %v", err)
	}
	_, apres, trouve := strings.Cut(string(source), "const CLES_AFFICHEES = new Set([")
	liste, _, ferme := strings.Cut(apres, "])")
	if !trouve || !ferme {
		t.Fatal("CLES_AFFICHEES introuvable dans strings.ts")
	}

	front := map[string]bool{}
	for _, m := range regexp.MustCompile(`'([^']+)'`).FindAllStringSubmatch(liste, -1) {
		front[m[1]] = true
	}
	for cle := range clesAffichees {
		if !front[cle] {
			t.Errorf("%q affichee par le Go, masquee par le front", cle)
		}
	}
	for cle := range front {
		if !clesAffichees[cle] {
			t.Errorf("%q affichee par le front, masquee par le Go", cle)
		}
	}
}

// Le masquage remplace la valeur sur place : les blancs, les apostrophes des
// autres valeurs et l'ordre des paramètres restent ceux qu'on a écrits.
func TestMasquerCleValeurGardeLaForme(t *testing.T) {
	t.Parallel()

	cas := map[string]string{
		"host=hote password = secret dbname=gescom":        "host=hote password = *** dbname=gescom",
		"host=hote  password='un secret'\tdbname='a b'":    "host=hote  password=***\tdbname='a b'",
		"host=hote user= dbname=gescom password=":          "***",
		"host=hote user='' dbname=gescom password=":        "host=hote user='' dbname=gescom password=",
		`host=hote password='l\'autre' application_name=x`: "host=hote password=*** application_name=x",
	}

	for dsn, attendu := range cas {
		if obtenu := Masquer(dsn); obtenu != attendu {
			t.Errorf("Masquer(%q) = %q, attendu %q", dsn, obtenu, attendu)
		}
	}
}

// Le masque doit rester lisible tel quel : url.String() encode les astérisques
// de la partie identifiants, et un %2A%2A%2A dans un message d'erreur
// n'apprend rien à personne.
func TestMasquerRendUnMasqueLisible(t *testing.T) {
	t.Parallel()

	masque := Masquer("postgres://utilisateur:secret@hote:5432/base")
	if strings.Contains(masque, "%2A") {
		t.Errorf("masque encode : %q", masque)
	}
	if !strings.Contains(masque, "***") {
		t.Errorf("masque absent : %q", masque)
	}
}

// Masquer doit rester lisible : un DSN illisible n'aide personne à diagnostiquer
// une connexion qui échoue.
func TestMasquerConserveLeReste(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom      string
		dsn      string
		attendus []string
	}{
		{
			"url complete",
			"postgres://utilisateur:secret@hote.exemple:5432/gescom?sslmode=require",
			[]string{"postgres://", "utilisateur", "hote.exemple:5432", "gescom", "sslmode=require"},
		},
		{
			"forme cle/valeur",
			"host=hote.exemple user=utilisateur password=secret dbname=gescom",
			[]string{"host=hote.exemple", "user=utilisateur", "dbname=gescom"},
		},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			masque := Masquer(c.dsn)
			for _, attendu := range c.attendus {
				if !strings.Contains(masque, attendu) {
					t.Errorf("%q absent de %q", attendu, masque)
				}
			}
		})
	}
}

// Un DSN sans mot de passe n'a rien à masquer : le rendre en étoiles rendrait
// les journaux inutiles là où il n'y a pourtant aucun risque.
func TestMasquerSansMotDePasse(t *testing.T) {
	t.Parallel()

	cas := []string{
		"postgres://hote:5432/base",
		"postgres://utilisateur@hote/base",
		"host=hote dbname=base",
	}

	for _, dsn := range cas {
		if obtenu := Masquer(dsn); obtenu != dsn {
			t.Errorf("Masquer(%q) = %q, attendu inchange", dsn, obtenu)
		}
	}
}

// Un DSN que l'analyse ne comprend pas ne doit rien laisser voir : on ne sait
// pas où est le secret, donc on ne montre rien.
func TestMasquerParDefautNeMontreRien(t *testing.T) {
	t.Parallel()

	illisible := "postgres://utilisateur:secret@hote:5432/base\x7f\x00?x=%zz"
	if masque := Masquer(illisible); strings.Contains(masque, "secret") {
		t.Errorf("secret visible sur un dsn illisible : %q", masque)
	}
}

// TestMasquerChaineVide vérifie que le masquage laisse la chaîne vide intacte :
// elle apparaît quand aucun DSN n'a été fourni, et la décorer brouillerait le
// message d'erreur.
func TestMasquerChaineVide(t *testing.T) {
	t.Parallel()

	if obtenu := Masquer(""); obtenu != "" {
		t.Errorf("Masquer(\"\") = %q", obtenu)
	}
}

// Le vocabulaire des SGBD est fermé et partagé avec calque.Source : un préfixe
// qui rendrait autre chose produirait un calque invalide.
func TestSGBDRenduAppartientAuVocabulaire(t *testing.T) {
	t.Parallel()

	vocabulaire := map[string]bool{
		"postgres": true, "mysql": true, "mariadb": true,
		"sqlserver": true, "sqlite": true, "oracle": true,
	}

	for prefixe := range prefixes {
		sgbd, err := SGBDDepuisDSN(prefixe + "://hote/base")
		if err != nil {
			t.Errorf("%s : %v", prefixe, err)
			continue
		}
		if !vocabulaire[sgbd] {
			t.Errorf("%s rend %q, hors du vocabulaire de calque.Source", prefixe, sgbd)
		}
	}
}

// Un DATABASE_URL de Symfony porte des paramètres que Doctrine comprend et
// qu'aucun SGBD ne connaît. Les laisser passer fait rejeter la connexion.
func TestNettoyerDSNRetireLesParametresDoctrine(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom     string
		dsn     string
		absents []string
	}{
		{
			"serverVersion",
			"postgresql://u:p@h:5432/base?serverVersion=16&sslmode=require",
			[]string{"serverVersion"},
		},
		{
			"charset",
			"postgresql://u:p@h/base?charset=utf8",
			[]string{"charset"},
		},
		{
			"les deux, casse melangee",
			"postgresql://u:p@h/base?serverVersion=16.2&charset=UTF8",
			[]string{"serverVersion", "charset", "UTF8"},
		},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			nettoye := NettoyerDSN(c.dsn)
			for _, absent := range c.absents {
				if strings.Contains(nettoye, absent) {
					t.Errorf("%q subsiste dans %q", absent, nettoye)
				}
			}
		})
	}
}

// Les options légitimes doivent atteindre le serveur : un filtrage par liste
// blanche casserait celles que ce code ne connaît pas encore.
func TestNettoyerDSNConserveLesOptionsLegitimes(t *testing.T) {
	t.Parallel()

	nettoye := NettoyerDSN("postgresql://u:p@h:5432/base?serverVersion=16&sslmode=verify-full&application_name=x&search_path=gescom")

	for _, attendu := range []string{
		"postgresql://", "h:5432", "/base",
		"sslmode=verify-full", "application_name=x", "search_path=gescom",
	} {
		if !strings.Contains(nettoye, attendu) {
			t.Errorf("%q absent de %q", attendu, nettoye)
		}
	}
}

// Sans paramètre Doctrine, le DSN doit ressortir tel quel : le réencoder
// changerait une chaîne qui fonctionnait.
func TestNettoyerDSNNeTouchePasAuReste(t *testing.T) {
	t.Parallel()

	cas := []string{
		"postgres://u:p@h:5432/base",
		"postgres://u:p@h:5432/base?sslmode=disable",
		"host=hote user=u password=p dbname=base",
		"",
	}

	for _, dsn := range cas {
		if obtenu := NettoyerDSN(dsn); obtenu != dsn {
			t.Errorf("NettoyerDSN(%q) = %q, attendu inchange", dsn, obtenu)
		}
	}
}

// Le nettoyage ne doit pas devenir une voie de fuite du mot de passe.
func TestNettoyerDSNNeDivulguePasLeSecret(t *testing.T) {
	t.Parallel()

	const secret = "S3cr3t-Tr0p-Long"
	nettoye := NettoyerDSN("postgresql://u:" + secret + "@h/base?serverVersion=16")

	if strings.Contains(Masquer(nettoye), secret) {
		t.Error("le masquage ne couvre plus le dsn nettoye")
	}
}

// Le port est le second niveau d'aiguillage, après le préfixe du DSN.
func TestSGBDDepuisPort(t *testing.T) {
	t.Parallel()

	cas := map[int]string{
		5432: "postgres",
		3306: "mysql",
		1433: "sqlserver",
		1521: "oracle",
		0:    "",
		5433: "",
		8080: "",
	}

	for port, attendu := range cas {
		if obtenu := SGBDDepuisPort(port); obtenu != attendu {
			t.Errorf("port %d donne %q, attendu %q", port, obtenu, attendu)
		}
	}
}

// 3306 doit rendre le même pilote à chaque appel. La table est écrite à la
// main pour cette raison : dérivée de portsParDefaut, où le port apparaît deux
// fois, elle rendrait « mysql » ou « mariadb » selon l'ordre d'itération.
func TestSGBDDepuisPortEstStable(t *testing.T) {
	t.Parallel()

	premier := SGBDDepuisPort(3306)
	if premier != "mysql" {
		t.Fatalf("3306 donne %q, attendu \"mysql\" — la variante se lit à la connexion", premier)
	}
	for range 64 {
		if obtenu := SGBDDepuisPort(3306); obtenu != premier {
			t.Fatalf("3306 donne %q puis %q", premier, obtenu)
		}
	}
}

// Le vocabulaire de SGBDDepuisPort est celui du calque : un port ne peut pas
// désigner un SGBD que le format ne connaît pas.
func TestSGBDDepuisPortAppartientAuVocabulaire(t *testing.T) {
	t.Parallel()

	connus := map[string]bool{
		"postgres": true, "mysql": true, "mariadb": true,
		"sqlserver": true, "sqlite": true, "oracle": true,
	}
	for port, sgbd := range sgbdParPort {
		if !connus[sgbd] {
			t.Errorf("le port %d designe %q, hors du vocabulaire du calque", port, sgbd)
		}
	}
}
