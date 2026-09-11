// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// emplacementsDeTest ouvre une configuration dans un répertoire temporaire.
func emplacementsDeTest(t *testing.T) *Emplacements {
	t.Helper()

	e, err := Ouvrir(filepath.Join(t.TempDir(), "ormeau"))
	if err != nil {
		t.Fatalf("ouverture : %v", err)
	}
	return e
}

// TestPreferencesAbsentesSontUnPremierLancement vérifie qu'un fichier absent ne
// produit ni erreur ni avertissement.
func TestPreferencesAbsentesSontUnPremierLancement(t *testing.T) {
	t.Parallel()

	p, avertissement, err := emplacementsDeTest(t).LirePreferences()
	if err != nil {
		t.Fatalf("lecture : %v", err)
	}
	if avertissement != "" {
		t.Errorf("un premier lancement avertit : %s", avertissement)
	}
	if p.Theme != "systeme" || p.Langue != "fr" {
		t.Errorf("défauts %+v", p)
	}
}

// TestPreferencesRelues vérifie l'aller-retour des six valeurs.
func TestPreferencesRelues(t *testing.T) {
	t.Parallel()

	e := emplacementsDeTest(t)
	ouvert := false
	ecrites := Preferences{
		Theme:          "sombre",
		Langue:         "en",
		ApercuOuvert:   &ouvert,
		LargeurArbre:   320,
		HauteurApercu:  240,
		LargeurEntites: 400,
	}
	if err := e.EcrirePreferences(ecrites); err != nil {
		t.Fatalf("écriture : %v", err)
	}

	relues, avertissement, err := e.LirePreferences()
	if err != nil {
		t.Fatalf("lecture : %v", err)
	}
	if avertissement != "" {
		t.Errorf("avertissement sur un fichier qu'on vient d'écrire : %s", avertissement)
	}
	if relues.Theme != "sombre" || relues.Langue != "en" {
		t.Errorf("thème et langue relus : %+v", relues)
	}
	if relues.ApercuOuvert == nil || *relues.ApercuOuvert {
		t.Errorf("aperçu relu : %v", relues.ApercuOuvert)
	}
	if relues.LargeurArbre != 320 || relues.HauteurApercu != 240 || relues.LargeurEntites != 400 {
		t.Errorf("tailles relues : %+v", relues)
	}
}

// TestApercuOuvertDistingueAbsentDeFaux vérifie que le pointeur sert à ça : un
// panneau jamais replié et un panneau replié ne sont pas la même chose.
func TestApercuOuvertDistingueAbsentDeFaux(t *testing.T) {
	t.Parallel()

	e := emplacementsDeTest(t)
	if err := e.EcrirePreferences(Preferences{Theme: "clair", Langue: "fr"}); err != nil {
		t.Fatalf("écriture : %v", err)
	}

	relues, _, err := e.LirePreferences()
	if err != nil {
		t.Fatalf("lecture : %v", err)
	}
	if relues.ApercuOuvert != nil {
		t.Errorf("aperçu non réglé rendu comme %v", *relues.ApercuOuvert)
	}
}

// TestPreferencesMalFormeesRendentLesDefauts couvre le fichier tronqué par un
// arrêt brutal : valeurs par défaut, avertissement, pas d'erreur.
func TestPreferencesMalFormeesRendentLesDefauts(t *testing.T) {
	t.Parallel()

	e := emplacementsDeTest(t)
	if err := os.WriteFile(e.FichierPreferences(), []byte("theme: [sombre"), permFichier); err != nil {
		t.Fatalf("écriture : %v", err)
	}

	p, avertissement, err := e.LirePreferences()
	if err != nil {
		t.Fatalf("un fichier mal formé ne doit pas empêcher de démarrer : %v", err)
	}
	if avertissement == "" {
		t.Error("fichier mal formé accepté en silence")
	}
	if p.Theme != "systeme" || p.Langue != "fr" {
		t.Errorf("défauts non appliqués : %+v", p)
	}
}

// TestValeursHorsVocabulaireSontIgnorees vérifie que la validation a lieu à la
// lecture, et qu'elle nomme toutes les fautes plutôt que la première.
func TestValeursHorsVocabulaireSontIgnorees(t *testing.T) {
	t.Parallel()

	e := emplacementsDeTest(t)
	contenu := "theme: fluo\nlangue: klingon\nlargeur_arbre: 99999\n"
	if err := os.WriteFile(e.FichierPreferences(), []byte(contenu), permFichier); err != nil {
		t.Fatalf("écriture : %v", err)
	}

	p, avertissement, err := e.LirePreferences()
	if err != nil {
		t.Fatalf("lecture : %v", err)
	}
	if p.Theme != "systeme" || p.Langue != "fr" || p.LargeurArbre != 0 {
		t.Errorf("valeurs hors vocabulaire retenues : %+v", p)
	}
	for _, attendu := range []string{"fluo", "klingon", "largeur_arbre"} {
		if !strings.Contains(avertissement, attendu) {
			t.Errorf("l'avertissement tait %s : %s", attendu, avertissement)
		}
	}
}

// TestTailleNegativeEstIgnoree couvre l'autre borne, qu'un fichier retouché à
// la main peut franchir.
func TestTailleNegativeEstIgnoree(t *testing.T) {
	t.Parallel()

	e := emplacementsDeTest(t)
	if err := os.WriteFile(e.FichierPreferences(), []byte("hauteur_apercu: -40\n"), permFichier); err != nil {
		t.Fatalf("écriture : %v", err)
	}

	p, avertissement, err := e.LirePreferences()
	if err != nil {
		t.Fatalf("lecture : %v", err)
	}
	if p.HauteurApercu != 0 {
		t.Errorf("hauteur négative retenue : %d", p.HauteurApercu)
	}
	if !strings.Contains(avertissement, "hauteur_apercu") {
		t.Errorf("avertissement muet : %s", avertissement)
	}
}

// TestEcrirePreferencesRefuseUneValeurInvalide vérifie qu'on ne pose pas un
// fichier que la lecture suivante rejettera.
func TestEcrirePreferencesRefuseUneValeurInvalide(t *testing.T) {
	t.Parallel()

	e := emplacementsDeTest(t)
	if err := e.EcrirePreferences(Preferences{Theme: "fluo", Langue: "fr"}); err == nil {
		t.Error("thème inconnu accepté à l'écriture")
	}
	if _, err := os.Stat(e.FichierPreferences()); !os.IsNotExist(err) {
		t.Error("un fichier a été écrit malgré le refus")
	}
}
