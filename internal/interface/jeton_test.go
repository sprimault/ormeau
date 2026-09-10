// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"testing"
	"time"
)

// TestJetonsTiresSeparement vérifie que le jeton d'URL et le cookie de session
// sont deux valeurs distinctes. Les confondre laisserait le jeton d'URL valide
// après consommation, et l'usage unique n'en serait plus un.
func TestJetonsTiresSeparement(t *testing.T) {
	t.Parallel()

	a, err := nouvelAcces()
	if err != nil {
		t.Fatalf("nouvelAcces: %v", err)
	}
	if a.jetonURL == a.jetonSession {
		t.Error("le cookie reprend le jeton d'URL")
	}
	if len(a.jetonURL) < 40 {
		t.Errorf("jeton trop court : %d caractères", len(a.jetonURL))
	}
}

// TestJetonURLNeSertQuUneFois vérifie qu'un jeton consommé ne vaut plus rien.
// Une URL fuit par l'historique, par un Referer, et sous Windows par la ligne
// de commande du navigateur.
func TestJetonURLNeSertQuUneFois(t *testing.T) {
	t.Parallel()

	a, err := nouvelAcces()
	if err != nil {
		t.Fatalf("nouvelAcces: %v", err)
	}
	if !a.consommer(a.jetonURL) {
		t.Fatal("premier échange refusé")
	}
	if a.consommer(a.jetonURL) {
		t.Error("second échange accepté")
	}
}

// TestJetonURLExpire vérifie que le jeton meurt même s'il n'a pas servi.
func TestJetonURLExpire(t *testing.T) {
	t.Parallel()

	a, err := nouvelAcces()
	if err != nil {
		t.Fatalf("nouvelAcces: %v", err)
	}
	a.expiration = time.Now().Add(-time.Second)
	if a.consommer(a.jetonURL) {
		t.Error("jeton expiré accepté")
	}
}

// TestJetonURLRefuseUneAutreValeur couvre la comparaison elle-même, y compris
// la valeur vide qu'enverrait une requête sans paramètre.
func TestJetonURLRefuseUneAutreValeur(t *testing.T) {
	t.Parallel()

	a, err := nouvelAcces()
	if err != nil {
		t.Fatalf("nouvelAcces: %v", err)
	}

	for _, fourni := range []string{"", "autre", a.jetonURL + "x", a.jetonSession} {
		if a.consommer(fourni) {
			t.Errorf("jeton %q accepté", fourni)
		}
	}
}

// TestValideSession vérifie que le cookie n'est reconnu que pour sa valeur
// exacte.
func TestValideSession(t *testing.T) {
	t.Parallel()

	a, err := nouvelAcces()
	if err != nil {
		t.Fatalf("nouvelAcces: %v", err)
	}
	if !a.valideSession(a.jetonSession) {
		t.Error("cookie légitime refusé")
	}
	for _, fourni := range []string{"", a.jetonURL, a.jetonSession + "x"} {
		if a.valideSession(fourni) {
			t.Errorf("cookie %q accepté", fourni)
		}
	}
}

// TestJetonsUniques vérifie que deux lancements ne partagent pas leurs secrets.
func TestJetonsUniques(t *testing.T) {
	t.Parallel()

	vus := map[string]bool{}
	for range 32 {
		jeton, err := genererJeton()
		if err != nil {
			t.Fatalf("genererJeton: %v", err)
		}
		if vus[jeton] {
			t.Fatalf("jeton répété : %s", jeton)
		}
		vus[jeton] = true
	}
}
