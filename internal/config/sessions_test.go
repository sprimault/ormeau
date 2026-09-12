// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestSessionIllisibleEstJetee couvre le brouillon tronqué par un arrêt brutal :
// il se lit comme une absence. Remonter l'erreur ferait refuser l'ouverture de
// l'arbitrage pour un fichier jetable.
func TestSessionIllisibleEstJetee(t *testing.T) {
	t.Parallel()

	e := emplacementsDeTest(t)
	repertoire := t.TempDir()
	if err := os.WriteFile(e.fichierSession("gescom", repertoire), []byte(`{"base": "ges`), permFichier); err != nil {
		t.Fatalf("brouillon tronqué : %v", err)
	}

	session, err := e.LireSession("gescom", repertoire)
	if session != nil || err != nil {
		t.Errorf("session %+v, erreur %v, attendu ni l'une ni l'autre", session, err)
	}
}

// TestSessionRefuseUnNomHorsDuRepertoire vérifie le dernier rempart avant le
// disque : un nom de base composé en chemin écrirait hors de etat/sessions.
func TestSessionRefuseUnNomHorsDuRepertoire(t *testing.T) {
	t.Parallel()

	e := emplacementsDeTest(t)
	if err := e.EcrireSession(Session{Base: "../evasion"}, t.TempDir()); err == nil {
		t.Error("nom de base composé en chemin accepté")
	}

	entrees, err := os.ReadDir(e.RepertoireEtat())
	if err != nil {
		t.Fatalf("lecture de etat : %v", err)
	}
	for _, entree := range entrees {
		if filepath.Ext(entree.Name()) == ".json" {
			t.Errorf("%s écrit hors de %s", entree.Name(), e.RepertoireSessions())
		}
	}
}

// TestSessionIgnoreLaCasseDuRepertoireSousWindows couvre deux graphies du même
// dossier : sans quoi le brouillon d'un projet paraîtrait perdu selon la façon
// dont on a tapé son chemin.
func TestSessionIgnoreLaCasseDuRepertoireSousWindows(t *testing.T) {
	t.Parallel()

	if runtime.GOOS != "windows" {
		t.Skip("la casse des chemins n'est indifférente que sous Windows")
	}

	e := emplacementsDeTest(t)
	if err := e.EcrireSession(Session{Base: "gescom", EntiteOuverte: "public.client"}, `C:\Projets\Gescom`); err != nil {
		t.Fatalf("écriture : %v", err)
	}

	session, err := e.LireSession("gescom", `c:\projets\gescom`)
	if err != nil || session == nil || session.EntiteOuverte != "public.client" {
		t.Errorf("session %+v, erreur %v", session, err)
	}
}
