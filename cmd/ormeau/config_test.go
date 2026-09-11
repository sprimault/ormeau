// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestOuvrirConfigurationSuitLaVariable vérifie le court-circuit de la
// résolution par la plateforme.
//
// C'est lui qui permet à toute la suite de tests d'exister : sans la variable,
// une exécution écrirait dans la configuration réelle du poste.
//
// Pas de t.Parallel : t.Setenv l'interdit, l'environnement étant partagé par le
// processus de test.
func TestOuvrirConfigurationSuitLaVariable(t *testing.T) {
	racine := filepath.Join(t.TempDir(), "portable")
	t.Setenv(ormeauConfigDir, racine)

	emplacements, err := ouvrirConfiguration()
	if err != nil {
		t.Fatalf("ouverture : %v", err)
	}
	if emplacements.Racine() != racine {
		t.Errorf("racine %s, attendue %s", emplacements.Racine(), racine)
	}
	if _, err := os.Stat(emplacements.RepertoireSessions()); err != nil {
		t.Errorf("arborescence non posée : %v", err)
	}
}
