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

// Personne ne déclare son SGBD dans un formulaire quand le port le dit déjà.
// C'est l'outil qui aiguille, pas l'utilisateur.
func TestDSNDeduitLeSGBDDuPort(t *testing.T) {
	t.Parallel()

	cas := []struct {
		port    int
		attendu string
	}{
		{5432, "postgres://u@hote:5432/base"},
		{1433, "sqlserver://u@hote:1433?database=base"},
		{3306, "mysql://u@hote:3306/base"},
	}

	for _, c := range cas {
		obtenu, err := Connexion{Hote: "hote", Port: c.port, Utilisateur: "u", Base: "base"}.DSN()
		if err != nil {
			t.Errorf("port %d : %v", c.port, err)
			continue
		}
		if obtenu != c.attendu {
			t.Errorf("port %d donne %q, attendu %q", c.port, obtenu, c.attendu)
		}
	}
}

// Un SGBD déclaré l'emporte sur ce que le port suggère : le port par défaut de
// l'un peut être occupé par l'autre.
func TestDSNPrefereLeSGBDDeclare(t *testing.T) {
	t.Parallel()

	obtenu, err := Connexion{SGBD: "sqlserver", Hote: "hote", Port: 5432, Utilisateur: "u", Base: "base"}.DSN()
	if err != nil {
		t.Fatalf("DSN: %v", err)
	}
	if attendu := "sqlserver://u@hote:5432?database=base"; obtenu != attendu {
		t.Errorf("obtenu %q, attendu %q", obtenu, attendu)
	}
}

// Rien ne se tente à l'aveugle quand le port ne tranche pas : chaque essai
// enverrait des identifiants, et une politique qui compte les échecs
// d'authentification verrouillerait le compte.
func TestDSNRefuseUnPortInconnuSansSGBD(t *testing.T) {
	t.Parallel()

	_, err := Connexion{Hote: "hote", Port: 7777, Utilisateur: "u"}.DSN()
	if err == nil {
		t.Fatal("aucune erreur sur un port inconnu")
	}
	if !strings.Contains(err.Error(), "sgbd") {
		t.Errorf("le message ne dit pas quoi préciser : %q", err)
	}
}

// La forme clé/valeur se lit comme pgx la lit : blancs autour du « = »,
// apostrophes, échappements, et dernière valeur retenue pour une clé répétée.
// Une lecture qui s'en écarte enregistre un profil avec un autre mot de passe,
// ou extrait une autre base que celle qu'on a nommée.
func TestLectureCleValeur(t *testing.T) {
	t.Parallel()

	connexion, err := ConnexionDepuisDSN(`host = hote port=5432 user=u password='un \'secret\'  long' dbname=a dbname=gescom`)
	if err != nil {
		t.Fatalf("ConnexionDepuisDSN: %v", err)
	}
	attendue := Connexion{SGBD: "postgres", Hote: "hote", Port: 5432, Utilisateur: "u", MotDePasse: `un 'secret'  long`, Base: "gescom"}
	if connexion != attendue {
		t.Errorf("connexion %+v, attendue %+v", connexion, attendue)
	}

	cas := map[string]string{
		"host=h dbname = gescom":        "gescom",
		"host=h dbname='ma base'":       "ma base",
		"host=h DBNAME=gescom":          "",
		"host=h dbname='non fermee":     "",
		"host=h password='a  b' user=u": "",
	}
	for dsn, attendu := range cas {
		if obtenu := BaseDuDSN(dsn); obtenu != attendu {
			t.Errorf("BaseDuDSN(%q) = %q, attendu %q", dsn, obtenu, attendu)
		}
	}
}

// AvecBase ne touche qu'à la base : un mot de passe entre apostrophes garde
// ses blancs, et un nom de base qui en a besoin est cité.
func TestAvecBaseCleValeur(t *testing.T) {
	t.Parallel()

	cas := []struct {
		dsn, base, attendu string
	}{
		{"host=h password='a  b' dbname = postgres sslmode=disable", "gescom", "host=h password='a  b' dbname = gescom sslmode=disable"},
		{"host=h dbname=x dbname=y", "gescom", "host=h dbname=gescom dbname=gescom"},
		{"host=h password='a  b'", "ma base", `host=h password='a  b' dbname='ma base'`},
		{"host=h", `l'autre`, `host=h dbname='l\'autre'`},
	}

	for _, c := range cas {
		if obtenu := AvecBase(c.dsn, c.base); obtenu != c.attendu {
			t.Errorf("AvecBase(%q, %q) = %q, attendu %q", c.dsn, c.base, obtenu, c.attendu)
		}
		if base := BaseDuDSN(AvecBase(c.dsn, c.base)); base != c.base {
			t.Errorf("AvecBase(%q, %q) relue avec la base %q", c.dsn, c.base, base)
		}
	}
}

// sslmode passe dans la chaîne composée. Sans lui, une connexion par
// composants partait toujours avec le défaut du pilote, prefer : TLS non
// vérifié, et repli en clair.
func TestDSNPorteLeModeSSL(t *testing.T) {
	t.Parallel()

	avec, err := Connexion{SGBD: "postgres", Hote: "h", Base: "gescom", SSLMode: "verify-full"}.DSN()
	if err != nil {
		t.Fatalf("composition : %v", err)
	}
	if !strings.HasSuffix(avec, "/gescom?sslmode=verify-full") {
		t.Errorf("dsn %q sans sslmode", Masquer(avec))
	}

	sans, err := Connexion{SGBD: "postgres", Hote: "h", Base: "gescom"}.DSN()
	if err != nil {
		t.Fatalf("composition : %v", err)
	}
	if strings.Contains(sans, "?") {
		t.Errorf("dsn %q : une requête sans mode demandé", Masquer(sans))
	}
}

// Un mode hors vocabulaire ne part pas au pilote, et sslmode n'a de sens que
// pour PostgreSQL, seul SGBD qui l'accepte ici.
func TestDSNRefuseUnModeSSLInconnu(t *testing.T) {
	t.Parallel()

	for _, c := range []Connexion{
		{SGBD: "postgres", Hote: "h", SSLMode: "verify"},
		{SGBD: "postgres", Hote: "h", SSLMode: "VERIFY-FULL"},
		{SGBD: "mysql", Hote: "h", SSLMode: "require"},
	} {
		if dsn, err := c.DSN(); err == nil || !strings.Contains(err.Error(), "sslmode") {
			t.Errorf("%+v : dsn %q, erreur %v", c, Masquer(dsn), err)
		}
	}
}

// Enregistrer un profil depuis une chaîne garde son sslmode, dans les deux
// formes. La forme clé/valeur est celle de libpq : elle désigne PostgreSQL
// même sans port, qui était jusqu'ici le seul indice lu.
func TestConnexionDepuisDSNGardeLeModeSSL(t *testing.T) {
	t.Parallel()

	cas := []struct {
		dsn, sgbd, mode string
	}{
		{"postgres://u@h:5432/gescom?sslmode=verify-full", "postgres", "verify-full"},
		{"postgresql://u@h/gescom?serverVersion=16&sslmode=require&charset=utf8", "postgres", "require"},
		{"postgres://u@h/gescom", "postgres", ""},
		{"host=h dbname=gescom sslmode=verify-ca", "postgres", "verify-ca"},
		{"host=h port=5432 dbname=gescom", "postgres", ""},
	}

	for _, c := range cas {
		connexion, err := ConnexionDepuisDSN(c.dsn)
		if err != nil {
			t.Fatalf("%q : %v", c.dsn, err)
		}
		if connexion.SGBD != c.sgbd || connexion.SSLMode != c.mode {
			t.Errorf("%q : sgbd %q, sslmode %q ; attendu %q, %q", c.dsn, connexion.SGBD, connexion.SSLMode, c.sgbd, c.mode)
		}
	}
}

// Un sslmode que le pilote refuserait est refusé dès la lecture, plutôt que
// d'entrer dans un profil qui ne se rouvrirait pas.
func TestConnexionDepuisDSNRefuseUnModeSSLInconnu(t *testing.T) {
	t.Parallel()

	for _, dsn := range []string{"postgres://u@h/gescom?sslmode=verif", "host=h sslmode=strict"} {
		if _, err := ConnexionDepuisDSN(dsn); err == nil || !strings.Contains(err.Error(), "sslmode") {
			t.Errorf("%q : erreur %v, attendu un refus qui nomme sslmode", dsn, err)
		}
	}
}

// La destination compare ce que deux connexions visent réellement : SGBD
// déduit du port, port par défaut du SGBD, hôte sans casse ni blanc autour.
func TestDestination(t *testing.T) {
	t.Parallel()

	cas := []struct {
		connexion  Connexion
		sgbd, hote string
		port       int
	}{
		{Connexion{SGBD: "postgres", Hote: "h"}, "postgres", "h", 5432},
		{Connexion{Hote: " NAS.local ", Port: 5432}, "postgres", "nas.local", 5432},
		{Connexion{SGBD: "PostgreSQL", Hote: "h", Port: 5433}, "postgres", "h", 5433},
		{Connexion{Hote: "h", Port: 9999}, "", "h", 9999},
	}

	for _, c := range cas {
		sgbd, hote, port := c.connexion.Destination()
		if sgbd != c.sgbd || hote != c.hote || port != c.port {
			t.Errorf("%+v : %q %q %d, attendu %q %q %d", c.connexion, sgbd, hote, port, c.sgbd, c.hote, c.port)
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

// Sous SQL Server, le chemin d'une URL désigne l'instance nommée : une base
// posée là ouvrait une session sur master sans qu'aucune erreur ne le dise.
func TestDSNMetLaBaseEnParametreSousSQLServer(t *testing.T) {
	t.Parallel()

	obtenu, err := Connexion{SGBD: "sqlserver", Hote: "hote", Utilisateur: "u", Base: "gescom"}.DSN()
	if err != nil {
		t.Fatalf("DSN: %v", err)
	}
	if attendu := "sqlserver://u@hote:1433?database=gescom"; obtenu != attendu {
		t.Errorf("obtenu %q, attendu %q", obtenu, attendu)
	}
}

// L'instance nommée, elle, occupe bien le chemin : c'est la forme
// « SERVEUR\COMPTA » des installations d'entreprise.
func TestDSNPorteLInstanceNommee(t *testing.T) {
	t.Parallel()

	obtenu, err := Connexion{SGBD: "sqlserver", Hote: "hote", Utilisateur: "u", Base: "gescom", Instance: "COMPTA"}.DSN()
	if err != nil {
		t.Fatalf("DSN: %v", err)
	}
	if attendu := "sqlserver://u@hote:1433/COMPTA?database=gescom"; obtenu != attendu {
		t.Errorf("obtenu %q, attendu %q", obtenu, attendu)
	}
}

// Le chiffrement se règle par son effet, et chaque effet se traduit dans le
// vocabulaire du pilote.
func TestDSNTraduitLeChiffrement(t *testing.T) {
	t.Parallel()

	cas := []struct{ chiffrement, attendu string }{
		{"desactive", "sqlserver://u@hote:1433?database=gescom&encrypt=disable"},
		{"confiance", "sqlserver://u@hote:1433?TrustServerCertificate=true&database=gescom&encrypt=true"},
		{"verifie", "sqlserver://u@hote:1433?TrustServerCertificate=false&database=gescom&encrypt=true"},
	}

	for _, c := range cas {
		obtenu, err := Connexion{SGBD: "sqlserver", Hote: "hote", Utilisateur: "u", Base: "gescom", Chiffrement: c.chiffrement}.DSN()
		if err != nil {
			t.Errorf("%s : %v", c.chiffrement, err)
			continue
		}
		if obtenu != c.attendu {
			t.Errorf("%s donne %q, attendu %q", c.chiffrement, obtenu, c.attendu)
		}
	}
}

// Chacun des deux réglages reste dans son dialecte : les mélanger produirait un
// DSN que le pilote ignorerait en silence.
func TestDSNRefuseUnReglageHorsDeSonDialecte(t *testing.T) {
	t.Parallel()

	if _, err := (Connexion{SGBD: "postgres", Hote: "hote", Utilisateur: "u", Chiffrement: "desactive"}).DSN(); err == nil {
		t.Error("le chiffrement de SQL Server a été accepté pour postgres")
	}
	if _, err := (Connexion{SGBD: "sqlserver", Hote: "hote", Utilisateur: "u", SSLMode: "disable"}).DSN(); err == nil {
		t.Error("sslmode a été accepté pour sqlserver")
	}
	if _, err := (Connexion{SGBD: "sqlserver", Hote: "hote", Utilisateur: "u", Chiffrement: "peut-etre"}).DSN(); err == nil {
		t.Error("un chiffrement hors vocabulaire a été accepté")
	}
}

// L'hôte est le seul champ que l'URL reçoit tel quel. Un blanc autour, ce que
// laisse un copier-coller, rendait l'URL illisible — et le message parlait de
// l'URL quand le défaut était dans un champ du formulaire.
func TestDSNNettoieLHote(t *testing.T) {
	t.Parallel()

	for _, hote := range []string{"  bdd", "bdd  ", " bdd "} {
		obtenu, err := Connexion{SGBD: "postgres", Hote: hote, Utilisateur: "u", Base: "b"}.DSN()
		if err != nil {
			t.Errorf("hote %q : %v", hote, err)
			continue
		}
		if attendu := "postgres://u@bdd:5432/b"; obtenu != attendu {
			t.Errorf("hote %q donne %q, attendu %q", hote, obtenu, attendu)
		}
	}
}

// Ce qu'un nettoyage ne peut pas rattraper se refuse en nommant le champ, et
// non en parlant d'une URL que l'utilisateur n'a pas écrite.
func TestDSNRefuseUnHoteInvalide(t *testing.T) {
	t.Parallel()

	for _, hote := range []string{"evo steff", "bdd/x", "bdd@autre", `SERVEUR\COMPTA`} {
		_, err := Connexion{SGBD: "postgres", Hote: hote, Utilisateur: "u"}.DSN()
		if err == nil {
			t.Errorf("hote %q accepté", hote)
			continue
		}
		if !strings.Contains(err.Error(), "hote invalide") {
			t.Errorf("hote %q : message %q, attendu qu'il nomme le champ", hote, err)
		}
	}
}

// Changer de base sous SQL Server ne doit pas toucher au chemin, qui porte
// l'instance : la session restait sur celle d'origine, et l'arbre montrait
// master alors qu'une autre base venait d'être choisie.
func TestAvecBaseSousSQLServer(t *testing.T) {
	t.Parallel()

	cas := []struct{ nom, dsn, attendu string }{
		{
			"sans base au depart",
			"sqlserver://sa@hote:1433",
			"sqlserver://sa@hote:1433?database=gescom",
		},
		{
			"base deja posee",
			"sqlserver://sa@hote:1433?database=master",
			"sqlserver://sa@hote:1433?database=gescom",
		},
		{
			"instance gardee",
			"sqlserver://sa@hote:1433/COMPTA?database=master",
			"sqlserver://sa@hote:1433/COMPTA?database=gescom",
		},
		{
			"autres parametres gardes",
			"sqlserver://sa@hote:1433?database=master&encrypt=disable",
			"sqlserver://sa@hote:1433?database=gescom&encrypt=disable",
		},
	}

	for _, c := range cas {
		if obtenu := AvecBase(c.dsn, "gescom"); obtenu != c.attendu {
			t.Errorf("%s : %q, attendu %q", c.nom, obtenu, c.attendu)
		}
	}
}

// La base d'un DSN SQL Server vit en paramètre, le chemin portant l'instance :
// lue au mauvais endroit, elle revenait vide et le calque n'avait plus de quoi
// nommer son fichier.
func TestBaseDuDSNSousSQLServer(t *testing.T) {
	t.Parallel()

	cas := map[string]string{
		"sqlserver://sa@h:1433?database=gescom":               "gescom",
		"sqlserver://sa@h:1433/COMPTA?database=gescom":        "gescom",
		"sqlserver://sa@h:1433?encrypt=disable&database=paie": "paie",
		"sqlserver://sa@h:1433":                               "",
		"sqlserver://sa@h:1433/COMPTA":                        "",
	}

	for dsn, attendu := range cas {
		if obtenu := BaseDuDSN(dsn); obtenu != attendu {
			t.Errorf("BaseDuDSN(%q) = %q, attendu %q", dsn, obtenu, attendu)
		}
	}
}

// Les trois fonctions doivent s'accorder : ce qu'AvecBase pose, BaseDuDSN le
// relit, et ce que DSN compose se relit aussi. Un désaccord entre elles est ce
// qui a fait extraire sous un nom de fichier vide.
func TestBasePoseeEtRelueSAccordent(t *testing.T) {
	t.Parallel()

	compose, err := Connexion{SGBD: "sqlserver", Hote: "h", Utilisateur: "sa", Base: "gescom"}.DSN()
	if err != nil {
		t.Fatalf("DSN: %v", err)
	}
	if base := BaseDuDSN(compose); base != "gescom" {
		t.Errorf("composé puis relu : %q", base)
	}
	if base := BaseDuDSN(AvecBase(compose, "paie")); base != "paie" {
		t.Errorf("changé puis relu : %q", base)
	}
}
