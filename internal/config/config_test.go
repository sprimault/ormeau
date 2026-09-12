// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestOuvrirPoseLArborescence vérifie les trois répertoires et le LISEZMOI.
func TestOuvrirPoseLArborescence(t *testing.T) {
	t.Parallel()

	racine := filepath.Join(t.TempDir(), "ormeau")
	e, err := Ouvrir(racine)
	if err != nil {
		t.Fatalf("ouverture : %v", err)
	}

	for _, d := range []string{e.Racine(), e.RepertoireEtat(), e.RepertoireSessions()} {
		infos, err := os.Stat(d)
		if err != nil {
			t.Errorf("%s absent : %v", d, err)
			continue
		}
		if !infos.IsDir() {
			t.Errorf("%s n'est pas un répertoire", d)
		}
	}

	contenu, err := os.ReadFile(filepath.Join(e.RepertoireEtat(), "LISEZMOI.txt"))
	if err != nil {
		t.Fatalf("LISEZMOI : %v", err)
	}
	if !strings.Contains(string(contenu), "effacé sans rien perdre") {
		t.Errorf("LISEZMOI ne dit pas que le répertoire est jetable :\n%s", contenu)
	}
}

// TestOuvrirEstIdempotent vérifie qu'une seconde ouverture ne casse rien et
// n'écrase pas un LISEZMOI que l'utilisateur aurait annoté.
func TestOuvrirEstIdempotent(t *testing.T) {
	t.Parallel()

	racine := filepath.Join(t.TempDir(), "ormeau")
	e, err := Ouvrir(racine)
	if err != nil {
		t.Fatalf("première ouverture : %v", err)
	}

	chemin := filepath.Join(e.RepertoireEtat(), "LISEZMOI.txt")
	if err := os.WriteFile(chemin, []byte("annoté"), 0o600); err != nil {
		t.Fatalf("annotation : %v", err)
	}
	if _, err := Ouvrir(racine); err != nil {
		t.Fatalf("seconde ouverture : %v", err)
	}

	contenu, err := os.ReadFile(chemin)
	if err != nil {
		t.Fatalf("relecture : %v", err)
	}
	if string(contenu) != "annoté" {
		t.Errorf("LISEZMOI réécrit par-dessus l'annotation : %s", contenu)
	}
}

// TestOuvrirCreeEn0700 vérifie les permissions du répertoire de configuration,
// qui portera des hôtes et des comptes de bases de production.
//
// Hors Windows : les bits POSIX y sont posés sans décrire les droits effectifs,
// et l'assertion ne vérifierait rien.
func TestOuvrirCreeEn0700(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("permissions POSIX non significatives sous Windows")
	}

	racine := filepath.Join(t.TempDir(), "ormeau")
	if _, err := Ouvrir(racine); err != nil {
		t.Fatalf("ouverture : %v", err)
	}

	infos, err := os.Stat(racine)
	if err != nil {
		t.Fatalf("lecture : %v", err)
	}
	if mode := infos.Mode().Perm(); mode != permRepertoire {
		t.Errorf("permissions %04o, attendues %04o", mode, permRepertoire)
	}
}

// TestOuvrirRefuseUneRacineNonInscriptible couvre le répertoire existant mais
// fermé — créé par un autre compte, ou par une politique d'entreprise. Le refus
// doit tomber à l'ouverture, pas à la première écriture de profils.yaml.
func TestOuvrirRefuseUneRacineNonInscriptible(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("un chmod 0500 ne ferme pas l'écriture sous Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root écrit dans un répertoire en lecture seule")
	}

	racine := filepath.Join(t.TempDir(), "ormeau")
	if _, err := Ouvrir(racine); err != nil {
		t.Fatalf("préparation : %v", err)
	}
	if err := os.Chmod(racine, 0o500); err != nil {
		t.Fatalf("fermeture : %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(racine, permRepertoire) })

	_, err := Ouvrir(racine)
	if err == nil {
		t.Fatal("une racine non inscriptible est acceptée")
	}
	if !strings.Contains(err.Error(), racine) {
		t.Errorf("l'erreur ne nomme pas le chemin fautif : %v", err)
	}
}

// TestOuvrirRefuseUneRacineVide vérifie qu'aucune arborescence n'est posée dans
// le répertoire courant quand l'appelant n'a rien résolu.
func TestOuvrirRefuseUneRacineVide(t *testing.T) {
	t.Parallel()

	if _, err := Ouvrir(""); err == nil {
		t.Error("racine vide acceptée")
	}
}

// TestCheminsSontSousLaRacine vérifie la disposition annoncée par la
// documentation : deux fichiers à la racine, les sessions sous etat/.
func TestCheminsSontSousLaRacine(t *testing.T) {
	t.Parallel()

	racine := filepath.Join(t.TempDir(), "ormeau")
	e, err := Ouvrir(racine)
	if err != nil {
		t.Fatalf("ouverture : %v", err)
	}

	cas := []struct {
		nom     string
		obtenu  string
		attendu string
	}{
		{"préférences", e.FichierPreferences(), filepath.Join(racine, "preferences.yaml")},
		{"profils", e.FichierProfils(), filepath.Join(racine, "profils.yaml")},
		{"état", e.RepertoireEtat(), filepath.Join(racine, "etat")},
		{"sessions", e.RepertoireSessions(), filepath.Join(racine, "etat", "sessions")},
	}
	for _, c := range cas {
		if c.obtenu != c.attendu {
			t.Errorf("%s : %s, attendu %s", c.nom, c.obtenu, c.attendu)
		}
	}
}

// TestOuvrirRendUneRacineAbsolue vérifie la normalisation d'un chemin relatif,
// qui serait sinon affiché tel quel au démarrage et dépendrait du répertoire
// courant.
func TestOuvrirRendUneRacineAbsolue(t *testing.T) {
	t.Parallel()

	e, err := Ouvrir(filepath.Join(t.TempDir(), "ormeau", "..", "ormeau"))
	if err != nil {
		t.Fatalf("ouverture : %v", err)
	}
	if !filepath.IsAbs(e.Racine()) {
		t.Errorf("racine %s n'est pas absolue", e.Racine())
	}
	if e.Racine() != filepath.Clean(e.Racine()) {
		t.Errorf("racine %s n'est pas nettoyée", e.Racine())
	}
}

// TestRacineParDefautPorteLeNomDeLOutil vérifie que la résolution s'appuie sur
// la convention de la plateforme et ne crée rien.
func TestRacineParDefautPorteLeNomDeLOutil(t *testing.T) {
	t.Parallel()

	racine, err := RacineParDefaut()
	if err != nil {
		t.Fatalf("résolution : %v", err)
	}
	if filepath.Base(racine) != "ormeau" {
		t.Errorf("racine %s ne se termine pas par le nom de l'outil", racine)
	}
	if !filepath.IsAbs(racine) {
		t.Errorf("racine %s n'est pas absolue", racine)
	}
}
