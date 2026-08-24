// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package introspection

import (
	"strings"
	"testing"
)

// Composer une URL à la main casse dès que le mot de passe porte un caractère
// réservé. C'est la raison d'être des drapeaux séparés.
func TestDSNEncodeLesCaracteresReserves(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom        string
		motDePasse string
	}{
		{"arobase", "mot@passe"},
		{"barre oblique", "mot/passe"},
		{"deux points", "mot:passe"},
		{"diese", "mot#passe"},
		{"point interrogation", "mot?passe"},
		{"tout a la fois", "a@b/c:d#e?f"},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			connexion := Connexion{
				SGBD: "postgres", Hote: "hote", Utilisateur: "u",
				MotDePasse: c.motDePasse, Base: "gescom",
			}
			dsn, err := connexion.DSN()
			if err != nil {
				t.Fatalf("composition : %v", err)
			}

			// Le DSN doit rester analysable, et rendre les mêmes composants.
			if sgbd, err := SGBDDepuisDSN(dsn); err != nil || sgbd != "postgres" {
				t.Errorf("dsn inanalysable : %q", Masquer(dsn))
			}
			if base := BaseDuDSN(dsn); base != "gescom" {
				t.Errorf("base %q, attendue gescom", base)
			}
			if strings.Contains(Masquer(dsn), c.motDePasse) {
				t.Error("le mot de passe ressort en clair du masquage")
			}
		})
	}
}

// TestDSNPortParDefaut vérifie le port implicite de chaque SGBD. Se tromper ici
// produit un refus de connexion que rien dans le message n'explique.
func TestDSNPortParDefaut(t *testing.T) {
	t.Parallel()

	cas := map[string]string{
		"postgres":  "5432",
		"mysql":     "3306",
		"mariadb":   "3306",
		"sqlserver": "1433",
		"oracle":    "1521",
	}

	for sgbd, port := range cas {
		t.Run(sgbd, func(t *testing.T) {
			t.Parallel()

			dsn, err := Connexion{SGBD: sgbd, Hote: "hote"}.DSN()
			if err != nil {
				t.Fatalf("composition : %v", err)
			}
			if !strings.Contains(dsn, ":"+port) {
				t.Errorf("dsn %q, port %s attendu", dsn, port)
			}
		})
	}
}

// TestDSNPortExplicite vérifie qu'un port donné l'emporte sur le défaut.
func TestDSNPortExplicite(t *testing.T) {
	t.Parallel()

	dsn, err := Connexion{SGBD: "postgres", Hote: "hote", Port: 35432}.DSN()
	if err != nil {
		t.Fatalf("composition : %v", err)
	}
	if !strings.Contains(dsn, ":35432") {
		t.Errorf("dsn %q", dsn)
	}
}

// TestDSNRefuseLIncomplet vérifie qu'une connexion sans SGBD, sans hôte ou sans
// base est refusée à la construction plutôt qu'au premier aller-retour réseau.
func TestDSNRefuseLIncomplet(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom       string
		connexion Connexion
	}{
		{"sans sgbd", Connexion{Hote: "hote"}},
		{"sans hote", Connexion{SGBD: "postgres"}},
		{"sgbd inconnu", Connexion{SGBD: "db2", Hote: "hote"}},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			if _, err := c.connexion.DSN(); err == nil {
				t.Error("aucune erreur")
			}
		})
	}
}

// Une base vide signifie « tout le serveur » : le DSN composé ne doit alors
// nommer aucune base, ce que l'appelant reconnaît.
func TestDSNSansBase(t *testing.T) {
	t.Parallel()

	dsn, err := Connexion{SGBD: "postgres", Hote: "hote", Utilisateur: "u"}.DSN()
	if err != nil {
		t.Fatalf("composition : %v", err)
	}
	if base := BaseDuDSN(dsn); base != "" {
		t.Errorf("base %q, attendue vide", base)
	}
}

// Un utilisateur sans mot de passe est légitime : authentification par pair,
// par certificat, ou par fichier de mots de passe.
func TestDSNSansMotDePasse(t *testing.T) {
	t.Parallel()

	dsn, err := Connexion{SGBD: "postgres", Hote: "hote", Utilisateur: "u", Base: "b"}.DSN()
	if err != nil {
		t.Fatalf("composition : %v", err)
	}
	if !strings.Contains(dsn, "//u@hote") {
		t.Errorf("dsn %q", dsn)
	}
}

// TestBaseDuDSN couvre l'extraction du nom de base, dont dépend le balayage de
// serveur pour nommer les fichiers qu'il produit.
func TestBaseDuDSN(t *testing.T) {
	t.Parallel()

	cas := map[string]string{
		"postgres://u:p@h:5432/gescom":            "gescom",
		"postgres://u:p@h:5432/gescom?sslmode=on": "gescom",
		"postgres://u:p@h:5432/":                  "",
		"postgres://h:5432":                       "",
		"host=h dbname=gescom user=u":             "gescom",
		"host=h user=u":                           "",
	}

	for dsn, attendu := range cas {
		if obtenu := BaseDuDSN(dsn); obtenu != attendu {
			t.Errorf("BaseDuDSN(%q) = %q, attendu %q", dsn, obtenu, attendu)
		}
	}
}

// AvecBase réutilise les identifiants d'une connexion pour parcourir les bases
// d'un serveur : c'est ce qui permet d'en extraire plusieurs sans les
// redemander.
func TestAvecBase(t *testing.T) {
	t.Parallel()

	cas := []struct {
		dsn     string
		base    string
		attendu string
	}{
		{"postgres://u:p@h:5432/postgres", "gescom", "gescom"},
		{"postgres://u:p@h:5432/", "gescom", "gescom"},
		{"postgres://u:p@h:5432", "gescom", "gescom"},
		{"host=h dbname=postgres user=u", "gescom", "gescom"},
		{"host=h user=u", "gescom", "gescom"},
	}

	for _, c := range cas {
		obtenu := AvecBase(c.dsn, c.base)
		if base := BaseDuDSN(obtenu); base != c.attendu {
			t.Errorf("AvecBase(%q, %q) donne la base %q, attendue %q", c.dsn, c.base, base, c.attendu)
		}
	}
}

// Changer de base ne doit pas perdre les identifiants ni les options.
func TestAvecBaseConserveLeReste(t *testing.T) {
	t.Parallel()

	obtenu := AvecBase("postgres://utilisateur:secret@hote:5432/postgres?sslmode=require", "gescom")

	for _, attendu := range []string{"utilisateur", "hote:5432", "sslmode=require", "/gescom"} {
		if !strings.Contains(obtenu, attendu) {
			t.Errorf("%q absent de %q", attendu, Masquer(obtenu))
		}
	}
}
