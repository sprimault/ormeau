// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestContexteRendLeRepertoireEtLaVersion vérifie ce que l'interface affiche en
// permanence : on doit savoir où on écrit avant de cliquer.
func TestContexteRendLeRepertoireEtLaVersion(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	w := httptest.NewRecorder()
	routeur.ServeHTTP(w, requeteAPI(s, http.MethodGet, "/api/contexte", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("code %d, attendu %d", w.Code, http.StatusOK)
	}
	var recu ReponseContexte
	if err := json.Unmarshal(w.Body.Bytes(), &recu); err != nil {
		t.Fatalf("réponse illisible : %v", err)
	}
	if recu.Repertoire != s.repertoire {
		t.Errorf("répertoire %q, attendu %q", recu.Repertoire, s.repertoire)
	}
	if recu.Version != "test" {
		t.Errorf("version %q, attendue \"test\"", recu.Version)
	}
}

// TestMethodesRefusees vérifie que chaque point d'entrée n'accepte que ses
// méthodes.
func TestMethodesRefusees(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	cas := []struct {
		methode string
		chemin  string
	}{
		{http.MethodPost, "/api/contexte"},
		{http.MethodPut, "/api/connexion"},
		{http.MethodPatch, "/api/connexion"},
	}
	for _, c := range cas {
		t.Run(c.methode+" "+c.chemin, func(t *testing.T) {
			w := httptest.NewRecorder()
			routeur.ServeHTTP(w, requeteAPI(s, c.methode, c.chemin, nil))
			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("code %d, attendu %d", w.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

// TestConnexionRefuseUnCorpsIllisible vérifie qu'un corps mal formé ne fait pas
// remonter son contenu — il peut porter un mot de passe.
func TestConnexionRefuseUnCorpsIllisible(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	corps := strings.NewReader(`{"dsn": "postgres://u:tres-secret@h/b"`)
	w := httptest.NewRecorder()
	routeur.ServeHTTP(w, requeteAPI(s, http.MethodPost, "/api/connexion", corps))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("code %d, attendu %d", w.Code, http.StatusBadRequest)
	}
	if strings.Contains(w.Body.String(), "tres-secret") {
		t.Error("le mot de passe est ressorti dans la réponse")
	}
}

// TestConnexionRefuseUneRequeteIncomplete vérifie le message rendu quand rien
// ne permet de composer une connexion.
func TestConnexionRefuseUneRequeteIncomplete(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	w := httptest.NewRecorder()
	routeur.ServeHTTP(w, requeteAPI(s, http.MethodPost, "/api/connexion", strings.NewReader(`{}`)))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("code %d, attendu %d", w.Code, http.StatusBadRequest)
	}
	var recu ReponseErreur
	if err := json.Unmarshal(w.Body.Bytes(), &recu); err != nil {
		t.Fatalf("réponse illisible : %v", err)
	}
	if recu.Erreur == "" {
		t.Error("échec sans message : le front n'a rien à afficher")
	}
}

// TestFermetureInconnueReussit vérifie qu'un onglet rechargé qui reposte sa
// fermeture n'obtient pas une erreur.
func TestFermetureInconnueReussit(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	corps := strings.NewReader(`{"session": "inconnue"}`)
	w := httptest.NewRecorder()
	routeur.ServeHTTP(w, requeteAPI(s, http.MethodDelete, "/api/connexion", corps))

	if w.Code != http.StatusNoContent {
		t.Errorf("code %d, attendu %d", w.Code, http.StatusNoContent)
	}
}

// TestFermetureLibereLaConnexion vérifie le cycle complet d'une session
// enregistrée.
func TestFermetureLibereLaConnexion(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	pilote := &piloteDeTest{}
	session, err := s.registre.ajouter(pilote)
	if err != nil {
		t.Fatalf("ajouter: %v", err)
	}

	corps := strings.NewReader(`{"session": "` + session + `"}`)
	w := httptest.NewRecorder()
	routeur.ServeHTTP(w, requeteAPI(s, http.MethodDelete, "/api/connexion", corps))

	if w.Code != http.StatusNoContent {
		t.Fatalf("code %d, attendu %d", w.Code, http.StatusNoContent)
	}
	if pilote.ferme != 1 {
		t.Errorf("pilote fermé %d fois, attendu 1", pilote.ferme)
	}
	if _, ok := s.registre.trouver(session); ok {
		t.Error("session encore présente après fermeture")
	}
}

// TestComposerChoisitLaFormeDonnee couvre les deux formes acceptées par
// l'écran de connexion, dont celle où le port désigne le SGBD.
func TestComposerChoisitLaFormeDonnee(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom     string
		requete RequeteConnexion
		attendu string
		echoue  bool
	}{
		{
			nom:     "dsn complet",
			requete: RequeteConnexion{DSN: "postgres://u:p@srv:5432/gescom"},
			attendu: "postgres://u:p@srv:5432/gescom",
		},
		{
			nom:     "dsn prioritaire sur les composants",
			requete: RequeteConnexion{DSN: "postgres://u@srv/gescom", Hote: "autre", SGBD: "mysql"},
			attendu: "postgres://u@srv/gescom",
		},
		{
			nom:     "composants avec sgbd",
			requete: RequeteConnexion{SGBD: "postgres", Hote: "srv", Utilisateur: "u", Base: "gescom"},
			attendu: "postgres://u@srv:5432/gescom",
		},
		{
			nom:     "sgbd deduit du port",
			requete: RequeteConnexion{Hote: "srv", Port: 5432, Utilisateur: "u", Base: "gescom"},
			attendu: "postgres://u@srv:5432/gescom",
		},
		{
			nom:     "mot de passe echappe",
			requete: RequeteConnexion{SGBD: "postgres", Hote: "srv", Utilisateur: "u", MotDePasse: "mot@de:passe/complique", Base: "g"},
			attendu: "postgres://u:mot%40de%3Apasse%2Fcomplique@srv:5432/g",
		},
		{
			nom:     "sans hote",
			requete: RequeteConnexion{SGBD: "postgres"},
			echoue:  true,
		},
		{
			nom:     "sans rien",
			requete: RequeteConnexion{},
			echoue:  true,
		},
		{
			nom:     "port inconnu et sgbd absent",
			requete: RequeteConnexion{Hote: "srv", Port: 7777},
			echoue:  true,
		},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			dsn, err := c.requete.composer()
			if c.echoue {
				if err == nil {
					t.Errorf("composer a rendu %q, une erreur était attendue", dsn)
				}
				return
			}
			if err != nil {
				t.Fatalf("composer: %v", err)
			}
			if dsn != c.attendu {
				t.Errorf("dsn %q, attendu %q", dsn, c.attendu)
			}
		})
	}
}

// TestSansDSN vérifie que la chaîne de connexion ne survit à aucun message.
//
// Les pilotes masquent déjà le mot de passe, mais un DSN masqué reste un hôte,
// un utilisateur et un nom de base : ni la réponse ni le journal ne doivent le
// porter.
func TestSansDSN(t *testing.T) {
	t.Parallel()

	const dsn = "postgres://gescom:tres-secret@bdd-interne:5432/gescom"

	cas := []struct {
		nom     string
		message string
	}{
		{"dsn en clair", "connexion a " + dsn + ": refuse"},
		{"dsn masque", "connexion a postgres://gescom:***@bdd-interne:5432/gescom: refuse"},
		{"sans dsn", "mot de passe refuse par le serveur"},
	}
	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			nettoye := sansDSN(c.message, dsn)
			for _, interdit := range []string{"tres-secret", "bdd-interne", "***"} {
				if strings.Contains(nettoye, interdit) {
					t.Errorf("%q subsiste dans %q", interdit, nettoye)
				}
			}
			if nettoye == "" {
				t.Error("message vidé : le front n'a plus rien à afficher")
			}
		})
	}
}
