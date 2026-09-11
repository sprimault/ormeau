// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/sprimault/ormeau/internal/introspection"
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
		{http.MethodPut, "/api/extractions"},
		{http.MethodGet, "/api/extractions"},
		{http.MethodPost, "/api/extractions/evenements"},
		{http.MethodPost, "/api/calque"},
		{http.MethodPut, "/api/decisions"},
		{http.MethodGet, "/api/inference"},
		{http.MethodGet, "/api/inference/entite"},
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
	session, err := s.registre.ajouter(pilote, dsnDeTest, "postgres")
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

// TestBasesRendLesBasesDuServeur vérifie ce que l'écran propose au choix : un
// serveur en porte souvent vingt, et les faire deviner est exactement ce que le
// projet refuse.
func TestBasesRendLesBasesDuServeur(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	pilote := &piloteListeur{bases: []string{"gescom", "paie"}}
	session, err := s.registre.ajouter(pilote, dsnDeTest, "postgres")
	if err != nil {
		t.Fatalf("ajouter: %v", err)
	}

	w := httptest.NewRecorder()
	routeur.ServeHTTP(w, requeteAPI(s, http.MethodGet, "/api/bases?session="+session, nil))

	if w.Code != http.StatusOK {
		t.Fatalf("code %d, attendu %d", w.Code, http.StatusOK)
	}
	var recu ReponseBases
	if err := json.Unmarshal(w.Body.Bytes(), &recu); err != nil {
		t.Fatalf("réponse illisible : %v", err)
	}
	if !slices.Equal(recu.Bases, []string{"gescom", "paie"}) {
		t.Errorf("bases %q", recu.Bases)
	}
}

// TestBasesToleUnPiloteSansEnumeration vérifie qu'un dialecte où la notion n'a
// pas de sens ne casse pas l'écran : le front se passe du sélecteur.
func TestBasesToleUnPiloteSansEnumeration(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	session, err := s.registre.ajouter(&piloteDeTest{}, dsnDeTest, "postgres")
	if err != nil {
		t.Fatalf("ajouter: %v", err)
	}

	w := httptest.NewRecorder()
	routeur.ServeHTTP(w, requeteAPI(s, http.MethodGet, "/api/bases?session="+session, nil))

	if w.Code != http.StatusOK {
		t.Fatalf("code %d, attendu %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), `"bases":[]`) {
		t.Errorf("corps %q, attendu une liste vide", w.Body.String())
	}
}

// TestBasesNeDivulguePasLeDSN est la contrepartie du DSN gardé en session :
// il permet de changer de base sans ressaisie, il ne doit jamais ressortir.
func TestBasesNeDivulguePasLeDSN(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	pilote := &piloteListeur{echec: errors.New("connexion a " + dsnDeTest + " perdue")}
	session, err := s.registre.ajouter(pilote, dsnDeTest, "postgres")
	if err != nil {
		t.Fatalf("ajouter: %v", err)
	}

	w := httptest.NewRecorder()
	routeur.ServeHTTP(w, requeteAPI(s, http.MethodGet, "/api/bases?session="+session, nil))

	for _, interdit := range []string{"secret", "bdd-interne", dsnDeTest} {
		if strings.Contains(w.Body.String(), interdit) {
			t.Errorf("%q ressort dans %q", interdit, w.Body.String())
		}
	}
}

// TestBasculerBaseRefuseUneDemandeVide vérifie qu'on ne rouvre pas une
// connexion sur rien.
func TestBasculerBaseRefuseUneDemandeVide(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	session, err := s.registre.ajouter(&piloteDeTest{}, dsnDeTest, "postgres")
	if err != nil {
		t.Fatalf("ajouter: %v", err)
	}

	corps := strings.NewReader(`{"session": "` + session + `", "base": "  "}`)
	w := httptest.NewRecorder()
	routeur.ServeHTTP(w, requeteAPI(s, http.MethodPost, "/api/base", corps))

	if w.Code != http.StatusBadRequest {
		t.Errorf("code %d, attendu %d", w.Code, http.StatusBadRequest)
	}
}

// TestBasculerBaseRefuseUneSessionInconnue couvre l'onglet resté ouvert après
// la fermeture de la connexion.
func TestBasculerBaseRefuseUneSessionInconnue(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	corps := strings.NewReader(`{"session": "inventee", "base": "gescom"}`)
	w := httptest.NewRecorder()
	routeur.ServeHTTP(w, requeteAPI(s, http.MethodPost, "/api/base", corps))

	if w.Code != http.StatusNotFound {
		t.Errorf("code %d, attendu %d", w.Code, http.StatusNotFound)
	}
}

// TestInventaireRendLesTables vérifie le cycle nominal de l'arbre.
func TestInventaireRendLesTables(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	pilote := &piloteDeTest{sommaires: []introspection.TableSommaire{
		{Schema: "public", Nom: "clients", NbColonnes: 7, LignesEstimees: 48210, ClePrimaire: true},
		{Schema: "public", Nom: "commandes", NbColonnes: 4, ReferenceVers: []string{"public.clients"}},
	}}
	session, err := s.registre.ajouter(pilote, dsnDeTest, "postgres")
	if err != nil {
		t.Fatalf("ajouter: %v", err)
	}

	w := httptest.NewRecorder()
	routeur.ServeHTTP(w, requeteAPI(s, http.MethodGet, "/api/inventaire?session="+session, nil))

	if w.Code != http.StatusOK {
		t.Fatalf("code %d, attendu %d", w.Code, http.StatusOK)
	}
	var recu ReponseInventaire
	if err := json.Unmarshal(w.Body.Bytes(), &recu); err != nil {
		t.Fatalf("réponse illisible : %v", err)
	}
	if len(recu.Tables) != 2 {
		t.Fatalf("%d table(s) rendues, attendu 2", len(recu.Tables))
	}
	if recu.Tables[1].ReferenceVers[0] != "public.clients" {
		t.Error("les références sortantes n'arrivent pas au front, la propagation des dépendances est impossible")
	}
}

// TestInventaireTransmetLesSchemas vérifie que la sélection de schémas arrive
// au pilote, et qu'une liste vide n'invente pas de schéma fantôme.
func TestInventaireTransmetLesSchemas(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom     string
		requete string
		attendu []string
	}{
		{"aucun", "", nil},
		{"un seul", "&schemas=public", []string{"public"}},
		{"plusieurs", "&schemas=public,compta", []string{"public", "compta"}},
		{"entrée vide ignorée", "&schemas=public,,", []string{"public"}},
		{"espaces retirés", "&schemas=%20public%20,%20compta", []string{"public", "compta"}},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			s, routeur := serveurDeTest(t)
			pilote := &piloteDeTest{}
			session, err := s.registre.ajouter(pilote, dsnDeTest, "postgres")
			if err != nil {
				t.Fatalf("ajouter: %v", err)
			}

			w := httptest.NewRecorder()
			routeur.ServeHTTP(w, requeteAPI(s, http.MethodGet,
				"/api/inventaire?session="+session+c.requete, nil))

			if w.Code != http.StatusOK {
				t.Fatalf("code %d, attendu %d", w.Code, http.StatusOK)
			}
			if !slices.Equal(pilote.schemasRecus, c.attendu) {
				t.Errorf("schémas transmis %q, attendus %q", pilote.schemasRecus, c.attendu)
			}
		})
	}
}

// TestSessionNAccepteQuUneRequeteALaFois vérifie ce qu'une connexion de base de
// données impose : un seul échange en cours.
//
// Le navigateur en lance plusieurs de front dès le chargement d'un écran —
// l'arbre et la liste des bases —, et pgx refuse la seconde par « conn busy ».
// Vu en vrai avant que ce verrou n'existe.
func TestSessionNAccepteQuUneRequeteALaFois(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	pilote := &piloteDeTest{}
	session, err := s.registre.ajouter(pilote, dsnDeTest, "postgres")
	if err != nil {
		t.Fatalf("ajouter: %v", err)
	}

	var attente sync.WaitGroup
	for range 4 {
		attente.Add(1)
		go func() {
			defer attente.Done()
			w := httptest.NewRecorder()
			routeur.ServeHTTP(w, requeteAPI(s, http.MethodGet, "/api/inventaire?session="+session, nil))
			if w.Code != http.StatusOK {
				t.Errorf("code %d, attendu %d", w.Code, http.StatusOK)
			}
		}()
	}
	attente.Wait()

	if max := pilote.maxSimultanes.Load(); max > 1 {
		t.Errorf("%d requêtes simultanées sur la même connexion, la base en refuserait", max)
	}
}

// TestInventaireRefuseUneSessionInconnue couvre l'onglet resté ouvert après la
// fermeture de la connexion.
func TestInventaireRefuseUneSessionInconnue(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	for _, session := range []string{"", "inventee"} {
		w := httptest.NewRecorder()
		routeur.ServeHTTP(w, requeteAPI(s, http.MethodGet, "/api/inventaire?session="+session, nil))
		if w.Code != http.StatusNotFound {
			t.Errorf("session %q : code %d, attendu %d", session, w.Code, http.StatusNotFound)
		}
	}
}

// TestInventaireSansTableRendUneListeVide vérifie qu'une base sans table n'est
// pas une panne, et que le front reçoit un tableau plutôt qu'un null.
func TestInventaireSansTableRendUneListeVide(t *testing.T) {
	t.Parallel()

	cas := map[string][]introspection.TableSommaire{
		"liste vide": {},
		"nil":        nil,
	}

	for nom, sommaires := range cas {
		t.Run(nom, func(t *testing.T) {
			t.Parallel()

			s, routeur := serveurDeTest(t)
			session, err := s.registre.ajouter(&piloteDeTest{sommaires: sommaires}, dsnDeTest, "postgres")
			if err != nil {
				t.Fatalf("ajouter: %v", err)
			}

			w := httptest.NewRecorder()
			routeur.ServeHTTP(w, requeteAPI(s, http.MethodGet, "/api/inventaire?session="+session, nil))

			if w.Code != http.StatusOK {
				t.Fatalf("code %d, attendu %d", w.Code, http.StatusOK)
			}
			if !strings.Contains(w.Body.String(), `"tables":[]`) {
				t.Errorf("corps %q : le front parcourt la liste, un null la ferait échouer", w.Body.String())
			}
		})
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
