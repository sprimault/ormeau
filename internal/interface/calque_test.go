// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
)

// ecrireCalque pose le calque d'une base dans le répertoire de travail, comme
// une extraction l'aurait fait.
func ecrireCalque(t *testing.T, s *serveur, base string, p *calque.Physique) {
	t.Helper()

	p.Source.ExtraitLe = "2026-09-11T10:00:00Z"
	if err := p.Ecrire(filepath.Join(s.repertoire, base+".calque.json")); err != nil {
		t.Fatalf("écriture du calque : %v", err)
	}
}

// lireCalqueAPI demande le calque d'une session.
func lireCalqueAPI(s *serveur, routeur http.Handler, session string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	routeur.ServeHTTP(w, requeteAPI(s, http.MethodGet, "/api/calque?session="+session, nil))
	return w
}

// TestCalqueRendLeFichierDeLaBase couvre le cas nominal : la base de la
// session nomme le fichier, et le contenu se relit comme un calque.
func TestCalqueRendLeFichierDeLaBase(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	ecrireCalque(t, s, "gescom", physiqueDeTest())

	w := lireCalqueAPI(s, routeur, sessionSur(t, s, "gescom"))
	if w.Code != http.StatusOK {
		t.Fatalf("code %d, attendu %d : %s", w.Code, http.StatusOK, w.Body.String())
	}

	var recu ReponseCalque
	if err := json.Unmarshal(w.Body.Bytes(), &recu); err != nil {
		t.Fatalf("réponse illisible : %v", err)
	}
	if recu.Fichier != "gescom.calque.json" || recu.ExtraitLe == "" || !strings.HasPrefix(recu.Empreinte, "sha256:") {
		t.Errorf("en-tête : %+v", recu)
	}

	var relu calque.Physique
	if err := json.Unmarshal([]byte(recu.Contenu), &relu); err != nil {
		t.Fatalf("contenu illisible : %v", err)
	}
	if len(relu.Tables) != 1 || relu.Tables[0].Nom != "client" {
		t.Errorf("tables relues : %+v", relu.Tables)
	}

	for _, interdit := range []string{"secret", "bdd-interne"} {
		if strings.Contains(w.Body.String(), interdit) {
			t.Errorf("%q ressort dans la réponse", interdit)
		}
	}
}

// TestCalqueAbsent vérifie qu'une base jamais extraite rend un message qui dit
// quel fichier manque.
func TestCalqueAbsent(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	w := lireCalqueAPI(s, routeur, sessionSur(t, s, "gescom"))
	if w.Code != http.StatusNotFound {
		t.Fatalf("code %d, attendu %d", w.Code, http.StatusNotFound)
	}
	if !strings.Contains(w.Body.String(), "gescom.calque.json") {
		t.Errorf("le message ne nomme pas le fichier : %s", w.Body.String())
	}
}

// TestCalqueRetireLesStatistiques vérifie qu'aucune valeur échantillonnée
// n'atteint le navigateur, et que l'écran sait qu'on les a retirées.
func TestCalqueRetireLesStatistiques(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	p := physiqueDeTest()
	p.Statistiques = map[string]calque.StatistiquesTable{
		"public.client": {Colonnes: map[string]calque.StatistiquesColonne{
			"id": {Echantillon: []string{"valeur-reelle"}},
		}},
	}
	ecrireCalque(t, s, "gescom", p)

	w := lireCalqueAPI(s, routeur, sessionSur(t, s, "gescom"))
	if w.Code != http.StatusOK {
		t.Fatalf("code %d, attendu %d", w.Code, http.StatusOK)
	}
	if strings.Contains(w.Body.String(), "valeur-reelle") {
		t.Error("une valeur échantillonnée atteint le navigateur")
	}

	var recu ReponseCalque
	if err := json.Unmarshal(w.Body.Bytes(), &recu); err != nil {
		t.Fatalf("réponse illisible : %v", err)
	}
	if !recu.StatistiquesRetirees {
		t.Error("statistiques retirées sans le signaler")
	}
}

// TestCalqueIllisible vérifie qu'un fichier abîmé se signale au lieu de passer
// pour absent.
func TestCalqueIllisible(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	if err := os.WriteFile(filepath.Join(s.repertoire, "gescom.calque.json"), []byte("{"), 0o600); err != nil {
		t.Fatalf("écriture : %v", err)
	}

	w := lireCalqueAPI(s, routeur, sessionSur(t, s, "gescom"))
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("code %d, attendu %d", w.Code, http.StatusUnprocessableEntity)
	}
}

// TestCalqueRefuseCeQuiNeNommePasUnFichier couvre la session inconnue et la
// base dont le nom sortirait du répertoire de travail.
func TestCalqueRefuseCeQuiNeNommePasUnFichier(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	if w := lireCalqueAPI(s, routeur, "inventee"); w.Code != http.StatusNotFound {
		t.Errorf("session inconnue : code %d, attendu %d", w.Code, http.StatusNotFound)
	}

	session, err := s.registre.ajouter(&piloteDeTest{}, "postgres://u@h:5432/..", "postgres")
	if err != nil {
		t.Fatalf("ajouter: %v", err)
	}
	if w := lireCalqueAPI(s, routeur, session); w.Code != http.StatusBadRequest {
		t.Errorf("base « .. » : code %d, attendu %d", w.Code, http.StatusBadRequest)
	}
}
