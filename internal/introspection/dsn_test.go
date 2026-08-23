// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package introspection

import (
	"strings"
	"testing"
)

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

	const secret = "S3cr3t-Tr0p-Long"

	dsns := []string{
		"postgres://utilisateur:" + secret + "@hote:5432/base",
		"postgres://utilisateur:" + secret + "@hote:5432/base?sslmode=require",
		"postgresql://u:" + secret + "@h/b?sslpassword=" + secret,
		"host=hote user=u password=" + secret + " dbname=base",
		"host=hote password=" + secret,
		"mysql://root:" + secret + "@tcp/base",
		"sqlserver://sa:" + secret + "@hote:1433?database=base",
		"sqlserver://sa@hote:1433?password=" + secret,
		"oracle://systeme:" + secret + "@hote:1521/ORCL",
		"pas du tout un dsn " + secret,
	}

	for _, dsn := range dsns {
		masque := Masquer(dsn)
		if strings.Contains(masque, secret) {
			t.Errorf("secret visible :\n  entree : %s\n  sortie : %s", dsn, masque)
		}
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
