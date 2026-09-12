// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"context"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/sprimault/ormeau/internal/config"
)

// origineDeTest est celle qu'un serveur écoutant sur ce port annoncerait.
const origineDeTest = "http://127.0.0.1:65000"

// serveurDeTest construit un serveur complet sans écouter sur le réseau.
func serveurDeTest(t *testing.T) (*serveur, http.Handler) {
	t.Helper()

	acces, err := nouvelAcces()
	if err != nil {
		t.Fatalf("nouvelAcces: %v", err)
	}
	repertoire := t.TempDir()
	// Une configuration jetable plutôt que nil : les préférences s'écrivent
	// vraiment, et un test qui les enregistre exerce le même chemin que le
	// binaire.
	emplacements, err := config.Ouvrir(filepath.Join(t.TempDir(), "configuration"))
	if err != nil {
		t.Fatalf("config.Ouvrir: %v", err)
	}
	s := &serveur{
		acces:        acces,
		registre:     nouveauRegistre(),
		extractions:  nouvellesExtractions(t.Context()),
		preferences:  nouvellesPreferences(emplacements),
		emplacements: emplacements,
		repertoire:   repertoire,
		version:      "test",
		origine:      origineDeTest,
	}
	// t.Context est annulé avant les nettoyages : les tâches encore en cours
	// s'arrêtent, et l'attente ne retient pas le test.
	t.Cleanup(s.extractions.attendre)

	routeur, err := s.routes()
	if err != nil {
		t.Fatalf("routes: %v", err)
	}
	return s, routeur
}

// requeteAPI prépare un appel d'API muni du cookie et de l'origine attendus.
func requeteAPI(s *serveur, methode, chemin string, corps io.Reader) *http.Request {
	r := httptest.NewRequest(methode, chemin, corps)
	r.Host = "127.0.0.1:65000"
	r.Header.Set("Origin", s.origine)
	r.AddCookie(&http.Cookie{Name: nomCookie, Value: s.acces.jetonSession})
	return r
}

// TestEntrerEchangeLeJetonContreLeCookie vérifie l'échange complet : cookie
// posé, redirection vers la racine pour que le jeton quitte la barre d'adresse.
func TestEntrerEchangeLeJetonContreLeCookie(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	w := httptest.NewRecorder()
	routeur.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/entrer?jeton="+s.acces.jetonURL, nil))

	if w.Code != http.StatusSeeOther {
		t.Fatalf("code %d, attendu %d", w.Code, http.StatusSeeOther)
	}
	if lieu := w.Header().Get("Location"); lieu != "/" {
		t.Errorf("redirection vers %q, attendue \"/\"", lieu)
	}

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("%d cookie(s) posé(s), attendu 1", len(cookies))
	}
	c := cookies[0]
	switch {
	case c.Name != nomCookie:
		t.Errorf("cookie nommé %q", c.Name)
	case c.Value != s.acces.jetonSession:
		t.Error("le cookie ne porte pas le jeton de session")
	case !c.HttpOnly:
		t.Error("cookie accessible par script")
	case c.SameSite != http.SameSiteStrictMode:
		t.Error("cookie sans SameSite=Strict : une page tierce pourrait viser 127.0.0.1")
	}
}

// TestEntrerRefuseUnJetonInvalide couvre l'URL bricolée et le second passage.
func TestEntrerRefuseUnJetonInvalide(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	cas := []struct {
		nom   string
		jeton string
	}{
		{"absent", ""},
		{"inventé", "jeton-invente"},
		{"jeton de session", s.acces.jetonSession},
	}
	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			w := httptest.NewRecorder()
			routeur.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/entrer?jeton="+c.jeton, nil))
			if w.Code != http.StatusForbidden {
				t.Errorf("code %d, attendu %d", w.Code, http.StatusForbidden)
			}
			if len(w.Result().Cookies()) != 0 {
				t.Error("un cookie a été posé malgré le refus")
			}
		})
	}
}

// TestEntrerNeSertQuUneFois vérifie que rejouer l'URL de démarrage ne redonne
// pas de session.
func TestEntrerNeSertQuUneFois(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	url := "/entrer?jeton=" + s.acces.jetonURL

	premier := httptest.NewRecorder()
	routeur.ServeHTTP(premier, httptest.NewRequest(http.MethodGet, url, nil))
	if premier.Code != http.StatusSeeOther {
		t.Fatalf("premier passage : code %d", premier.Code)
	}

	second := httptest.NewRecorder()
	routeur.ServeHTTP(second, httptest.NewRequest(http.MethodGet, url, nil))
	if second.Code != http.StatusForbidden {
		t.Errorf("second passage : code %d, attendu %d", second.Code, http.StatusForbidden)
	}
}

// TestAPIExigeLaSession vérifie qu'aucun processus local ne peut atteindre
// l'API sans avoir échangé le jeton.
func TestAPIExigeLaSession(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	cas := []struct {
		nom    string
		cookie *http.Cookie
	}{
		{"sans cookie", nil},
		{"cookie vide", &http.Cookie{Name: nomCookie, Value: ""}},
		{"cookie inventé", &http.Cookie{Name: nomCookie, Value: "valeur-inventee"}},
		{"jeton d'URL en cookie", &http.Cookie{Name: nomCookie, Value: s.acces.jetonURL}},
	}
	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/api/contexte", nil)
			r.Header.Set("Origin", s.origine)
			if c.cookie != nil {
				r.AddCookie(c.cookie)
			}
			w := httptest.NewRecorder()
			routeur.ServeHTTP(w, r)
			if w.Code != http.StatusUnauthorized {
				t.Errorf("code %d, attendu %d", w.Code, http.StatusUnauthorized)
			}
		})
	}
}

// TestAPIRefuseUneOrigineEtrangere couvre le DNS rebinding : un nom de domaine
// tiers qui résout vers 127.0.0.1 arrive avec son propre Origin.
func TestAPIRefuseUneOrigineEtrangere(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	cas := []string{
		"http://exemple.test",
		"http://127.0.0.1:9999",
		"https://127.0.0.1:65000",
		"null",
	}
	for _, origine := range cas {
		t.Run(origine, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/api/contexte", nil)
			r.Header.Set("Origin", origine)
			r.AddCookie(&http.Cookie{Name: nomCookie, Value: s.acces.jetonSession})
			w := httptest.NewRecorder()
			routeur.ServeHTTP(w, r)
			if w.Code != http.StatusForbidden {
				t.Errorf("code %d, attendu %d", w.Code, http.StatusForbidden)
			}
		})
	}
}

// TestAPIExigeUneOrigineSurLesEcritures vérifie qu'un POST sans Origin est
// refusé. Les navigateurs l'omettent sur certains GET de même origine, jamais
// sur un POST.
func TestAPIExigeUneOrigineSurLesEcritures(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	r := httptest.NewRequest(http.MethodPost, "/api/connexion", nil)
	r.AddCookie(&http.Cookie{Name: nomCookie, Value: s.acces.jetonSession})
	w := httptest.NewRecorder()
	routeur.ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("code %d, attendu %d", w.Code, http.StatusForbidden)
	}
}

// TestAPIAccepteUneLectureSansOrigine vérifie l'autre moitié de la règle : le
// GET de même origine passe, protégé par le cookie SameSite=Strict.
func TestAPIAccepteUneLectureSansOrigine(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	r := httptest.NewRequest(http.MethodGet, "/api/contexte", nil)
	r.AddCookie(&http.Cookie{Name: nomCookie, Value: s.acces.jetonSession})
	w := httptest.NewRecorder()
	routeur.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("code %d, attendu %d", w.Code, http.StatusOK)
	}
}

// TestCanoniserRedirigeLocalhost vérifie qu'arriver par « localhost » renvoie
// sur 127.0.0.1 : sans cela, la page se chargerait et tous ses appels d'API
// seraient refusés sans explication.
func TestCanoniserRedirigeLocalhost(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Host = "localhost:65000"
	w := httptest.NewRecorder()
	routeur.ServeHTTP(w, r)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("code %d, attendu %d", w.Code, http.StatusSeeOther)
	}
	if lieu := w.Header().Get("Location"); lieu != s.origine+"/" {
		t.Errorf("redirection vers %q, attendue %q", lieu, s.origine+"/")
	}
}

// TestFrontalSertLIndex vérifie que la racine rend le document embarqué.
func TestFrontalSertLIndex(t *testing.T) {
	t.Parallel()

	_, routeur := serveurDeTest(t)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Host = "127.0.0.1:65000"
	w := httptest.NewRecorder()
	routeur.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("code %d, attendu %d", w.Code, http.StatusOK)
	}
	if w.Body.Len() == 0 {
		t.Error("index.html servi vide")
	}
	if cache := w.Header().Get("Cache-Control"); cache != "no-store" {
		t.Errorf("Cache-Control %q : un index en cache référencerait des bundles disparus", cache)
	}
}

// TestFrontalReplieSurLIndex vérifie qu'une route d'écran rechargée revient sur
// l'application, alors qu'une ressource absente reste un 404.
func TestFrontalReplieSurLIndex(t *testing.T) {
	t.Parallel()

	_, routeur := serveurDeTest(t)

	cas := []struct {
		chemin string
		code   int
	}{
		{"/arbitrage", http.StatusOK},
		{"/tables/public/clients", http.StatusOK},
		{"/assets/absent.js", http.StatusNotFound},
	}
	for _, c := range cas {
		t.Run(c.chemin, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, c.chemin, nil)
			r.Host = "127.0.0.1:65000"
			w := httptest.NewRecorder()
			routeur.ServeHTTP(w, r)
			if w.Code != c.code {
				t.Errorf("code %d, attendu %d", w.Code, c.code)
			}
		})
	}
}

// TestEmbarqueNonVide est le garde-fou de la construction : sans web-build, le
// système de fichiers embarqué serait vide et le binaire publié afficherait une
// interface blanche, sans qu'aucun avertissement ne le signale.
func TestEmbarqueNonVide(t *testing.T) {
	t.Parallel()

	racine, err := fs.Sub(embarque, "embarque")
	if err != nil {
		t.Fatalf("fs.Sub: %v", err)
	}
	index, err := fs.ReadFile(racine, "index.html")
	if err != nil {
		t.Fatalf("index.html absent : le front n'a pas été construit (%v)", err)
	}
	if len(index) == 0 {
		t.Error("index.html embarqué vide")
	}
}

// TestServirRendLaMainSurAnnulation vérifie l'arrêt propre : le processus doit
// mourir sur SIGINT, et le jeton comme le cookie avec lui.
func TestServirRendLaMainSurAnnulation(t *testing.T) {
	t.Parallel()

	ctx, annuler := context.WithCancel(context.Background())
	fini := make(chan error, 1)
	go func() {
		fini <- Servir(ctx, Options{Repertoire: t.TempDir(), SansNavigateur: true})
	}()

	// Laisse le temps d'écouter avant de demander l'arrêt.
	time.Sleep(50 * time.Millisecond)
	annuler()

	select {
	case err := <-fini:
		if err != nil {
			t.Errorf("Servir: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Servir n'a pas rendu la main après annulation")
	}
}
