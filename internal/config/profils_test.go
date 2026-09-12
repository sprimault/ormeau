// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"os"
	"runtime"
	"strings"
	"testing"
)

// profilDeTest rend un profil complet, sans mot de passe.
func profilDeTest(nom string) Profil {
	return Profil{
		Nom:         nom,
		SGBD:        "postgres",
		Hote:        "192.168.1.10",
		Port:        5432,
		Utilisateur: "lecture",
		Base:        "gescom",
		Repertoire:  "/projets/gescom",
	}
}

// TestProfilsAbsentsNeSontPasUneErreur couvre le premier lancement.
func TestProfilsAbsentsNeSontPasUneErreur(t *testing.T) {
	t.Parallel()

	profils, avertissement, err := emplacementsDeTest(t).LireProfils()
	if err != nil || avertissement != "" || profils != nil {
		t.Errorf("profils %v, avertissement %q, erreur %v", profils, avertissement, err)
	}
}

// TestProfilEnregistrePuisRelu vérifie l'aller-retour des champs de connexion.
func TestProfilEnregistrePuisRelu(t *testing.T) {
	t.Parallel()

	e := emplacementsDeTest(t)
	if _, err := e.EnregistrerProfil(profilDeTest("gescom production"), "", true); err != nil {
		t.Fatalf("enregistrement : %v", err)
	}

	profils, _, err := e.LireProfils()
	if err != nil {
		t.Fatalf("lecture : %v", err)
	}
	if len(profils) != 1 {
		t.Fatalf("%d profil(s)", len(profils))
	}
	if profils[0] != profilDeTest("gescom production") {
		t.Errorf("profil relu %+v", profils[0])
	}
	if profils[0].MotDePasseEnregistre() {
		t.Error("un mot de passe est enregistré alors qu'aucun n'a été donné")
	}
}

// TestMotDePasseChiffreDansLeFichier vérifie que le fichier ne porte pas le
// clair, et que la relecture le rend.
func TestMotDePasseChiffreDansLeFichier(t *testing.T) {
	t.Parallel()

	e := emplacementsDeTest(t)
	const secret = "M0tDeP4sse-Production"
	if _, err := e.EnregistrerProfil(profilDeTest("nas"), secret, true); err != nil {
		t.Fatalf("enregistrement : %v", err)
	}

	contenu, err := os.ReadFile(e.FichierProfils())
	if err != nil {
		t.Fatalf("lecture du fichier : %v", err)
	}
	if strings.Contains(string(contenu), secret) {
		t.Error("le mot de passe figure en clair dans profils.yaml")
	}

	relu, err := e.MotDePasseDuProfil("nas")
	if err != nil {
		t.Fatalf("déchiffrement : %v", err)
	}
	if relu != secret {
		t.Errorf("mot de passe relu %q", relu)
	}
}

// TestMotDePasseVideEfface couvre la case décochée sur un profil qui en
// portait un : croire l'avoir retiré alors qu'il reste sur le disque serait le
// pire des deux mondes.
func TestMotDePasseVideEfface(t *testing.T) {
	t.Parallel()

	e := emplacementsDeTest(t)
	if _, err := e.EnregistrerProfil(profilDeTest("nas"), "secret", true); err != nil {
		t.Fatalf("premier enregistrement : %v", err)
	}
	if _, err := e.EnregistrerProfil(profilDeTest("nas"), "", true); err != nil {
		t.Fatalf("second enregistrement : %v", err)
	}

	profils, _, err := e.LireProfils()
	if err != nil {
		t.Fatalf("lecture : %v", err)
	}
	if profils[0].MotDePasseEnregistre() {
		t.Error("le mot de passe survit à une case décochée")
	}
}

// TestChiffrementDetecteUneRetouche vérifie l'apport de GCM : un octet changé
// est refusé, au lieu de rendre n'importe quoi qu'on enverrait ensuite comme
// mot de passe à une base de production.
func TestChiffrementDetecteUneRetouche(t *testing.T) {
	t.Parallel()

	e := emplacementsDeTest(t)
	chiffre, _, err := e.chiffrer("secret")
	if err != nil {
		t.Fatalf("chiffrement : %v", err)
	}

	retouche := []byte(chiffre)
	retouche[len(retouche)-2] ^= 0x01
	if _, err := e.dechiffrer(string(retouche)); err == nil {
		t.Error("un chiffré retouché est accepté")
	}
}

// TestCleAbsenteNeVidePasLeProfil couvre la clé effacée alors qu'un profil
// porte un mot de passe chiffré.
func TestCleAbsenteNeVidePasLeProfil(t *testing.T) {
	t.Parallel()

	e := emplacementsDeTest(t)
	if _, err := e.EnregistrerProfil(profilDeTest("nas"), "secret", true); err != nil {
		t.Fatalf("enregistrement : %v", err)
	}
	if err := os.Remove(e.fichierCle()); err != nil {
		t.Fatalf("suppression de la clé : %v", err)
	}

	// Le profil se charge encore : seule la ressaisie revient.
	profils, _, err := e.LireProfils()
	if err != nil || len(profils) != 1 || profils[0].Hote != "192.168.1.10" {
		t.Fatalf("profils %+v, erreur %v", profils, err)
	}

	_, err = e.MotDePasseDuProfil("nas")
	if !errors.Is(err, ErrCleAbsente) {
		t.Errorf("erreur %v, attendue ErrCleAbsente", err)
	}
}

// TestCleAbimeeDegradeCommeUneCleAbsente couvre une clé tronquée : les profils
// se chargent avec un avertissement, le mot de passe enregistré devient
// indéchiffrable au lieu d'une erreur quelconque, et un nouvel enregistrement
// tire une clé neuve plutôt que d'échouer ou d'annoncer un mot de passe
// illisible.
func TestCleAbimeeDegradeCommeUneCleAbsente(t *testing.T) {
	t.Parallel()

	e := emplacementsDeTest(t)
	if _, err := e.EnregistrerProfil(profilDeTest("nas"), "secret", true); err != nil {
		t.Fatalf("enregistrement : %v", err)
	}
	if err := os.WriteFile(e.fichierCle(), []byte("abime"), permFichier); err != nil {
		t.Fatalf("clé tronquée : %v", err)
	}

	profils, avertissement, err := e.LireProfils()
	if err != nil || len(profils) != 1 {
		t.Fatalf("profils %+v, erreur %v", profils, err)
	}
	if !strings.Contains(avertissement, "inutilisable") {
		t.Errorf("avertissement %q, attendu la clé inutilisable", avertissement)
	}
	if _, err := e.MotDePasseDuProfil("nas"); !errors.Is(err, ErrMotDePasseIndechiffrable) {
		t.Errorf("erreur %v, attendue ErrMotDePasseIndechiffrable", err)
	}

	avertissement, err = e.EnregistrerProfil(profilDeTest("paie"), "autre", true)
	if err != nil {
		t.Fatalf("enregistrement avec une clé tronquée : %v", err)
	}
	if !strings.Contains(avertissement, "perdus") || !strings.Contains(avertissement, "cle.bin.invalide") {
		t.Errorf("avertissement %q, attendu la perte et la clé mise de côté", avertissement)
	}
	if clair, err := e.MotDePasseDuProfil("paie"); err != nil || clair != "autre" {
		t.Errorf("mot de passe relu %q, erreur %v", clair, err)
	}
	if cle, err := os.ReadFile(e.fichierCle()); err != nil || len(cle) != tailleCle {
		t.Errorf("clé de %d octets, erreur %v", len(cle), err)
	}
	if ancienne, err := os.ReadFile(e.fichierCle() + ".invalide"); err != nil || string(ancienne) != "abime" {
		t.Errorf("ancienne clé %q, erreur %v", ancienne, err)
	}

	// La clé neuve ne se remplace plus : l'avertissement ne revient pas.
	if avertissement, err := e.EnregistrerProfil(profilDeTest("stock"), "encore", true); err != nil || avertissement != "" {
		t.Errorf("avertissement %q, erreur %v", avertissement, err)
	}
}

// Un enregistrement refusé ne remplace pas la clé : l'avertissement de perte
// ne doit jamais accompagner une opération qui n'a pas eu lieu.
func TestEnregistrementRefuseGardeLaCleAbimee(t *testing.T) {
	t.Parallel()

	e := emplacementsDeTest(t)
	if _, err := e.EnregistrerProfil(profilDeTest("nas"), "secret", true); err != nil {
		t.Fatalf("enregistrement : %v", err)
	}
	if err := os.WriteFile(e.fichierCle(), []byte("abime"), permFichier); err != nil {
		t.Fatalf("clé tronquée : %v", err)
	}

	if _, err := e.EnregistrerProfil(profilDeTest("nas"), "autre", false); !errors.Is(err, ErrProfilExistant) {
		t.Fatalf("erreur %v, attendue ErrProfilExistant", err)
	}
	if cle, err := os.ReadFile(e.fichierCle()); err != nil || string(cle) != "abime" {
		t.Errorf("clé %q, erreur %v : remplacée pour un enregistrement refusé", cle, err)
	}
}

// TestRemplacementDemandeConfirmation couvre le geste qu'on fait sans y penser
// après avoir modifié un champ : écraser un profil de production n'a rien
// d'anodin.
func TestRemplacementDemandeConfirmation(t *testing.T) {
	t.Parallel()

	e := emplacementsDeTest(t)
	if _, err := e.EnregistrerProfil(profilDeTest("nas"), "", false); err != nil {
		t.Fatalf("premier enregistrement : %v", err)
	}

	modifie := profilDeTest("nas")
	modifie.Base = "paie"
	if _, err := e.EnregistrerProfil(modifie, "", false); !errors.Is(err, ErrProfilExistant) {
		t.Fatalf("erreur %v, attendue ErrProfilExistant", err)
	}

	// Le refus ne touche à rien : le profil d'origine est intact.
	profils, _, err := e.LireProfils()
	if err != nil || len(profils) != 1 || profils[0].Base != "gescom" {
		t.Fatalf("profils %+v, erreur %v", profils, err)
	}

	if _, err := e.EnregistrerProfil(modifie, "", true); err != nil {
		t.Fatalf("remplacement confirmé : %v", err)
	}
	profils, _, _ = e.LireProfils()
	if profils[0].Base != "paie" {
		t.Errorf("le remplacement confirmé n'a pas eu lieu : %+v", profils[0])
	}
}

// TestCleManquanteAvertitUneFois vérifie que la lecture des profils signale
// des mots de passe qu'on ne pourra pas relire.
//
// Une fois, au chargement, plutôt qu'à chaque tentative de connexion : c'est le
// ménage un peu large dans le répertoire de configuration, ou un profils.yaml
// recopié d'une autre machine.
func TestCleManquanteAvertitUneFois(t *testing.T) {
	t.Parallel()

	e := emplacementsDeTest(t)
	if _, err := e.EnregistrerProfil(profilDeTest("nas"), "secret", true); err != nil {
		t.Fatalf("enregistrement : %v", err)
	}
	if err := os.Remove(e.fichierCle()); err != nil {
		t.Fatalf("suppression de la clé : %v", err)
	}

	profils, avertissement, err := e.LireProfils()
	if err != nil {
		t.Fatalf("lecture : %v", err)
	}
	if len(profils) != 1 {
		t.Fatalf("les profils ne sont plus utilisables : %+v", profils)
	}
	if !strings.Contains(avertissement, "cle.bin") {
		t.Errorf("l'avertissement ne nomme pas le fichier absent : %q", avertissement)
	}
}

// TestPasDAvertissementSansMotDePasse vérifie qu'une clé absente ne dit rien
// quand aucun profil n'en dépend — le cas de tout le monde avant le premier
// enregistrement de mot de passe.
func TestPasDAvertissementSansMotDePasse(t *testing.T) {
	t.Parallel()

	e := emplacementsDeTest(t)
	if _, err := e.EnregistrerProfil(profilDeTest("nas"), "", true); err != nil {
		t.Fatalf("enregistrement : %v", err)
	}

	if _, avertissement, _ := e.LireProfils(); avertissement != "" {
		t.Errorf("avertissement sans raison : %q", avertissement)
	}
}

// TestProfilsRecopiesSansCle couvre le fichier venu d'une autre machine : les
// mots de passe sont indéchiffrables, ce qui est le comportement voulu, et le
// message ne doit pas l'appeler corruption.
func TestProfilsRecopiesSansCle(t *testing.T) {
	t.Parallel()

	source := emplacementsDeTest(t)
	if _, err := source.EnregistrerProfil(profilDeTest("nas"), "secret", true); err != nil {
		t.Fatalf("enregistrement : %v", err)
	}
	contenu, err := os.ReadFile(source.FichierProfils())
	if err != nil {
		t.Fatalf("lecture : %v", err)
	}

	// Une autre machine : le fichier voyage, la clé non.
	autre := emplacementsDeTest(t)
	if err := os.WriteFile(autre.FichierProfils(), contenu, permFichier); err != nil {
		t.Fatalf("copie : %v", err)
	}
	if _, _, err := autre.chiffrer("pour tirer une cle differente"); err != nil {
		t.Fatalf("clé de l'autre poste : %v", err)
	}

	_, err = autre.MotDePasseDuProfil("nas")
	if !errors.Is(err, ErrMotDePasseIndechiffrable) {
		t.Errorf("erreur %v, attendue ErrMotDePasseIndechiffrable", err)
	}
}

// TestCleEcriteEn0600 vérifie les permissions du fichier qui vaut les mots de
// passe qu'il protège.
func TestCleEcriteEn0600(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("permissions POSIX non significatives sous Windows")
	}

	e := emplacementsDeTest(t)
	if _, _, err := e.chiffrer("secret"); err != nil {
		t.Fatalf("chiffrement : %v", err)
	}

	infos, err := os.Stat(e.fichierCle())
	if err != nil {
		t.Fatalf("lecture : %v", err)
	}
	if mode := infos.Mode().Perm(); mode != permFichier {
		t.Errorf("permissions %04o, attendues %04o", mode, permFichier)
	}
}

// TestProfilsTriesParNom vérifie que la liste ne suit pas l'ordre des
// enregistrements, qui la rendrait imprévisible à chaque ajout.
func TestProfilsTriesParNom(t *testing.T) {
	t.Parallel()

	e := emplacementsDeTest(t)
	for _, nom := range []string{"zebra", "Alpha", "milieu"} {
		if _, err := e.EnregistrerProfil(profilDeTest(nom), "", true); err != nil {
			t.Fatalf("enregistrement de %s : %v", nom, err)
		}
	}

	profils, _, err := e.LireProfils()
	if err != nil {
		t.Fatalf("lecture : %v", err)
	}
	noms := []string{profils[0].Nom, profils[1].Nom, profils[2].Nom}
	if noms[0] != "Alpha" || noms[1] != "milieu" || noms[2] != "zebra" {
		t.Errorf("ordre %v", noms)
	}
}

// TestProfilRemplaceSansDoubler vérifie qu'un second enregistrement du même
// nom met à jour au lieu d'ajouter.
func TestProfilRemplaceSansDoubler(t *testing.T) {
	t.Parallel()

	e := emplacementsDeTest(t)
	if _, err := e.EnregistrerProfil(profilDeTest("nas"), "", true); err != nil {
		t.Fatalf("premier : %v", err)
	}
	modifie := profilDeTest("nas")
	modifie.Base = "paie"
	if _, err := e.EnregistrerProfil(modifie, "", true); err != nil {
		t.Fatalf("second : %v", err)
	}

	profils, _, err := e.LireProfils()
	if err != nil {
		t.Fatalf("lecture : %v", err)
	}
	if len(profils) != 1 || profils[0].Base != "paie" {
		t.Errorf("profils %+v", profils)
	}
}

// TestSupprimerProfil couvre le retrait et le nom inconnu.
func TestSupprimerProfil(t *testing.T) {
	t.Parallel()

	e := emplacementsDeTest(t)
	if _, err := e.EnregistrerProfil(profilDeTest("nas"), "", true); err != nil {
		t.Fatalf("enregistrement : %v", err)
	}

	if err := e.SupprimerProfil("nas"); err != nil {
		t.Fatalf("suppression : %v", err)
	}
	profils, _, err := e.LireProfils()
	if err != nil || len(profils) != 0 {
		t.Errorf("profils %+v, erreur %v", profils, err)
	}

	if err := e.SupprimerProfil("jamais"); !errors.Is(err, ErrProfilInconnu) {
		t.Errorf("erreur %v, attendue ErrProfilInconnu", err)
	}
}

// TestNomDeProfilRefuse couvre ce que l'écran ne doit pas pouvoir poser.
func TestNomDeProfilRefuse(t *testing.T) {
	t.Parallel()

	e := emplacementsDeTest(t)
	cas := map[string]string{
		"vide":       "",
		"trop long":  strings.Repeat("a", 61),
		"séparateur": "projets/gescom",
		"retour":     "gescom\nautre",
	}
	for nom, valeur := range cas {
		t.Run(nom, func(t *testing.T) {
			t.Parallel()

			p := profilDeTest(valeur)
			if _, err := e.EnregistrerProfil(p, "", true); err == nil {
				t.Errorf("nom %q accepté", valeur)
			}
		})
	}
}

// TestProfilsMalFormesNEmpechentPasDOuvrir couvre le fichier retouché à la
// main : l'écran de connexion s'ouvre sans profil, on saisit comme avant.
func TestProfilsMalFormesNEmpechentPasDOuvrir(t *testing.T) {
	t.Parallel()

	e := emplacementsDeTest(t)
	if err := os.WriteFile(e.FichierProfils(), []byte("profils: [oui"), permFichier); err != nil {
		t.Fatalf("écriture : %v", err)
	}

	profils, avertissement, err := e.LireProfils()
	if err != nil {
		t.Fatalf("un fichier mal formé ne doit pas empêcher d'ouvrir : %v", err)
	}
	if avertissement == "" {
		t.Error("fichier mal formé accepté en silence")
	}
	if len(profils) != 0 {
		t.Errorf("profils %+v", profils)
	}
}

// TestProfilAuNomRetoucheEstEcarte vérifie qu'une entrée qu'on ne saurait ni
// relire ni supprimer ne rentre pas dans la liste.
func TestProfilAuNomRetoucheEstEcarte(t *testing.T) {
	t.Parallel()

	e := emplacementsDeTest(t)
	contenu := "profils:\n  - nom: \"../evasion\"\n    hote: h\n  - nom: correct\n    hote: h\n"
	if err := os.WriteFile(e.FichierProfils(), []byte(contenu), permFichier); err != nil {
		t.Fatalf("écriture : %v", err)
	}

	profils, avertissement, err := e.LireProfils()
	if err != nil {
		t.Fatalf("lecture : %v", err)
	}
	if len(profils) != 1 || profils[0].Nom != "correct" {
		t.Errorf("profils %+v", profils)
	}
	if !strings.Contains(avertissement, "evasion") {
		t.Errorf("l'avertissement ne nomme pas l'entrée écartée : %s", avertissement)
	}
}
