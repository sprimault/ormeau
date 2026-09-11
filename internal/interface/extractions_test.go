// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sprimault/ormeau/internal/calque"
	"github.com/sprimault/ormeau/internal/introspection"
)

// piloteExtracteur déroule deux passes et rend un calque figé.
//
// Il ne simule aucun catalogue : ce qui est sous test, c'est ce que l'interface
// fait d'une extraction — la suivre, l'interrompre, l'écrire —, pas ce qu'elle
// contient. Ce que rend un vrai pilote se teste contre un vrai serveur.
type piloteExtracteur struct {
	piloteDeTest
	// liberer retient la passe des colonnes jusqu'à sa fermeture ; nil, elle
	// passe aussitôt.
	liberer chan struct{}
	echec   error
}

// Extraire déroule tables puis colonnes, et rend l'échec préparé s'il y en a un.
func (p *piloteExtracteur) Extraire(ctx context.Context, _ introspection.Portee) (*calque.Physique, error) {
	err := introspection.Derouler(ctx,
		introspection.Passe{Etape: introspection.EtapeTables, Lire: func(context.Context) error { return nil }},
		introspection.Passe{Etape: introspection.EtapeColonnes, Lire: func(ctx context.Context) error {
			if p.liberer != nil {
				select {
				case <-p.liberer:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return p.echec
		}},
	)
	if err != nil {
		return nil, err
	}
	return physiqueDeTest(), nil
}

// physiqueDeTest est le plus petit calque valide.
func physiqueDeTest() *calque.Physique {
	return &calque.Physique{
		VersionRI: calque.VersionCourante,
		Source:    calque.Source{SGBD: "postgres", Version: "17.2", Catalogue: "gescom", Schema: "public"},
		Tables: []calque.Table{{
			Nom:    "client",
			Schema: "public",
			Colonnes: []calque.Colonne{
				{Nom: "id", Position: 1, TypeBrut: "integer", TypeNormalise: calque.TypeEntier},
			},
		}},
	}
}

// fabriqueDeTest remplace introspection.Ouvrir : elle rend le pilote préparé
// pour la base visée, et consigne les DSN reçus.
type fabriqueDeTest struct {
	mu      sync.Mutex
	pilotes map[string]*piloteExtracteur
	recus   []string
	echec   error
}

// ouvrir rend le pilote de la base, ou un pilote qui passe aussitôt quand rien
// n'a été préparé pour elle.
func (f *fabriqueDeTest) ouvrir(_ context.Context, _, dsn string) (introspection.Introspecteur, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.recus = append(f.recus, dsn)
	if f.echec != nil {
		return nil, f.echec
	}
	if p, ok := f.pilotes[introspection.BaseDuDSN(dsn)]; ok {
		return p, nil
	}
	return &piloteExtracteur{}, nil
}

// ouvertures rend les DSN reçus jusqu'ici, un par connexion demandée.
func (f *fabriqueDeTest) ouvertures() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.recus...)
}

// extracteurDeTest monte un serveur dont les extractions passent par la
// fabrique donnée.
func extracteurDeTest(t *testing.T, f *fabriqueDeTest) (*serveur, http.Handler) {
	t.Helper()

	s, routeur := serveurDeTest(t)
	s.extractions.ouvrir = f.ouvrir
	return s, routeur
}

// dsnSur rend le DSN de test pointant sur une base.
func dsnSur(base string) string {
	return introspection.AvecBase(dsnDeTest, base)
}

// sessionSur enregistre une session ouverte sur une base et rend son
// identifiant.
func sessionSur(t *testing.T, s *serveur, base string) string {
	t.Helper()

	id, err := s.registre.ajouter(&piloteDeTest{}, dsnSur(base), "postgres")
	if err != nil {
		t.Fatalf("ajouter: %v", err)
	}
	return id
}

// posterExtraction lance une extraction par l'API.
func posterExtraction(s *serveur, routeur http.Handler, session, portee string) *httptest.ResponseRecorder {
	corps := strings.NewReader(`{"session": "` + session + `", "portee": ` + portee + `}`)
	w := httptest.NewRecorder()
	routeur.ServeHTTP(w, requeteAPI(s, http.MethodPost, "/api/extractions", corps))
	return w
}

// lancerOuEchouer lance une extraction et rend la tâche créée.
func lancerOuEchouer(t *testing.T, s *serveur, routeur http.Handler, session string) Extraction {
	t.Helper()

	w := posterExtraction(s, routeur, session, `{}`)
	if w.Code != http.StatusAccepted {
		t.Fatalf("code %d, attendu %d : %s", w.Code, http.StatusAccepted, w.Body.String())
	}
	var lancee Extraction
	if err := json.Unmarshal(w.Body.Bytes(), &lancee); err != nil {
		t.Fatalf("réponse illisible : %v", err)
	}
	return lancee
}

// retirer annule ou retire une tâche par l'API.
func retirer(t *testing.T, s *serveur, routeur http.Handler, id string) {
	t.Helper()

	w := httptest.NewRecorder()
	routeur.ServeHTTP(w, requeteAPI(s, http.MethodDelete, "/api/extractions", strings.NewReader(`{"id": "`+id+`"}`)))
	if w.Code != http.StatusNoContent {
		t.Fatalf("code %d, attendu %d", w.Code, http.StatusNoContent)
	}
}

// tacheCourante rend l'état d'une tâche et le signal du prochain changement.
func tacheCourante(e *extractions, id string) (Extraction, bool, chan struct{}) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, t := range e.taches {
		if t.ID == id {
			return t.Extraction, true, e.signal
		}
	}
	return Extraction{}, false, e.signal
}

// attendreEtat attend qu'une tâche atteigne un état, en suivant le signal du
// registre plutôt qu'en sondant.
func attendreEtat(t *testing.T, e *extractions, id string, etat EtatExtraction) Extraction {
	t.Helper()

	delai := time.After(5 * time.Second)
	for {
		courante, _, signal := tacheCourante(e, id)
		if courante.Etat == etat {
			return courante
		}
		select {
		case <-signal:
		case <-delai:
			t.Fatalf("tâche dans l'état %q, attendu %q", courante.Etat, etat)
		}
	}
}

// TestExtractionEcritLeCalque couvre le cycle nominal : la base de la session,
// une connexion propre à la tâche, un calque écrit et relisible.
func TestExtractionEcritLeCalque(t *testing.T) {
	t.Parallel()

	pilote := &piloteExtracteur{}
	f := &fabriqueDeTest{pilotes: map[string]*piloteExtracteur{"gescom": pilote}}
	s, routeur := extracteurDeTest(t, f)
	session := sessionSur(t, s, "gescom")

	w := posterExtraction(s, routeur, session, `{"schemas": ["public"], "tables_incluses": ["public.client"]}`)
	if w.Code != http.StatusAccepted {
		t.Fatalf("code %d, attendu %d : %s", w.Code, http.StatusAccepted, w.Body.String())
	}
	var lancee Extraction
	if err := json.Unmarshal(w.Body.Bytes(), &lancee); err != nil {
		t.Fatalf("réponse illisible : %v", err)
	}
	if lancee.Base != "gescom" || lancee.Fichier != "gescom.calque.json" || lancee.NbTables != 1 {
		t.Errorf("tâche lancée : %+v", lancee)
	}

	finale := attendreEtat(t, s.extractions, lancee.ID, EtatTerminee)
	if finale.Resultat == nil || finale.Resultat.Tables != 1 || finale.Resultat.Colonnes != 1 {
		t.Fatalf("résultat : %+v", finale.Resultat)
	}
	derniere := introspection.Avancement{Etape: introspection.EtapeColonnes, Rang: 2, Total: 2}
	if finale.Avancement == nil || *finale.Avancement != derniere {
		t.Errorf("avancement final %+v, attendu %+v", finale.Avancement, derniere)
	}
	// À la milliseconde : une extraction de quelques tables tient dans la
	// seconde, et une durée arrondie à zéro n'apprendrait rien.
	for _, horodatage := range []string{finale.Debut, finale.Fin} {
		if _, err := time.Parse(formatInstant, horodatage); err != nil {
			t.Errorf("horodatage %q hors du format %s : %v", horodatage, formatInstant, err)
		}
	}

	relu, err := calque.LirePhysique(filepath.Join(s.repertoire, "gescom.calque.json"))
	if err != nil {
		t.Fatalf("calque illisible : %v", err)
	}
	if relu.Source.Empreinte != finale.Resultat.Empreinte || relu.Source.ExtraitLe == "" {
		t.Errorf("source relue : %+v", relu.Source)
	}

	if recus := f.ouvertures(); len(recus) != 1 || recus[0] != dsnSur("gescom") {
		t.Errorf("connexions ouvertes : %q", recus)
	}
	if pilote.ferme != 1 {
		t.Errorf("connexion de la tâche fermée %d fois, attendu 1", pilote.ferme)
	}
}

// evenementLu est un événement tel qu'un navigateur le reçoit.
type evenementLu struct {
	id      uint64
	nom     string
	donnees string
}

// ouvrirFlux se connecte au flux comme le ferait EventSource, avec
// Last-Event-ID quand dernier n'est pas vide.
func ouvrirFlux(t *testing.T, s *serveur, url, dernier string) *bufio.Reader {
	t.Helper()

	requete, err := http.NewRequest(http.MethodGet, url+"/api/extractions/evenements", nil)
	if err != nil {
		t.Fatalf("requête : %v", err)
	}
	requete.AddCookie(&http.Cookie{Name: nomCookie, Value: s.acces.jetonSession})
	if dernier != "" {
		requete.Header.Set("Last-Event-ID", dernier)
	}

	// Le délai borne toute lecture : un événement qui n'arrive pas fait
	// échouer le test au lieu de le bloquer.
	client := &http.Client{Timeout: 5 * time.Second}
	reponse, err := client.Do(requete)
	if err != nil {
		t.Fatalf("flux : %v", err)
	}
	t.Cleanup(func() { _ = reponse.Body.Close() })

	if reponse.StatusCode != http.StatusOK {
		t.Fatalf("code %d, attendu %d", reponse.StatusCode, http.StatusOK)
	}
	if contenu := reponse.Header.Get("Content-Type"); contenu != "text/event-stream" {
		t.Fatalf("Content-Type %q", contenu)
	}
	return bufio.NewReader(reponse.Body)
}

// lireEvenement lit un événement complet du flux.
func lireEvenement(t *testing.T, lecteur *bufio.Reader) evenementLu {
	t.Helper()

	var ev evenementLu
	for {
		ligne, err := lecteur.ReadString('\n')
		if err != nil {
			t.Fatalf("lecture du flux : %v", err)
		}
		ligne = strings.TrimSuffix(ligne, "\n")
		switch {
		case ligne == "":
			return ev
		case strings.HasPrefix(ligne, "id: "):
			ev.id, _ = strconv.ParseUint(strings.TrimPrefix(ligne, "id: "), 10, 64)
		case strings.HasPrefix(ligne, "event: "):
			ev.nom = strings.TrimPrefix(ligne, "event: ")
		case strings.HasPrefix(ligne, "data: "):
			ev.donnees = strings.TrimPrefix(ligne, "data: ")
		}
	}
}

// decoderExtraction lit la tâche portée par un événement.
func decoderExtraction(t *testing.T, ev evenementLu) Extraction {
	t.Helper()

	var e Extraction
	if err := json.Unmarshal([]byte(ev.donnees), &e); err != nil {
		t.Fatalf("événement %s illisible : %v", ev.nom, err)
	}
	return e
}

// TestFluxSuitUneExtraction vérifie ce que l'en-tête affichera : un instantané
// à l'ouverture, puis chaque étape, dans l'ordre et sans saut de numéro.
func TestFluxSuitUneExtraction(t *testing.T) {
	t.Parallel()

	s, routeur := extracteurDeTest(t, &fabriqueDeTest{})
	session := sessionSur(t, s, "gescom")
	serveurHTTP := httptest.NewServer(routeur)
	t.Cleanup(serveurHTTP.Close)

	flux := ouvrirFlux(t, s, serveurHTTP.URL, "")
	if ev := lireEvenement(t, flux); ev.nom != evenementEtat || ev.donnees != `{"extractions":[]}` {
		t.Fatalf("premier événement : %+v", ev)
	}

	lancerOuEchouer(t, s, routeur, session)

	attendus := []string{"en_attente", "en_cours", "en_cours tables 1/2", "en_cours colonnes 2/2", "terminee"}
	var precedent uint64
	for i, attendu := range attendus {
		ev := lireEvenement(t, flux)
		if ev.nom != evenementExtraction {
			t.Fatalf("événement %d : %q", i, ev.nom)
		}
		if ev.id != precedent+1 {
			t.Errorf("événement %d numéroté %d après %d", i, ev.id, precedent)
		}
		precedent = ev.id

		e := decoderExtraction(t, ev)
		obtenu := string(e.Etat)
		if e.Etat == EtatEnCours && e.Avancement != nil {
			obtenu = fmt.Sprintf("%s %s %d/%d", e.Etat, e.Avancement.Etape, e.Avancement.Rang, e.Avancement.Total)
		}
		if obtenu != attendu {
			t.Errorf("événement %d : %q, attendu %q", i, obtenu, attendu)
		}
	}
}

// TestFluxReprendSansDoublonNiTrou vérifie la reconnexion d'EventSource : les
// événements manqués et eux seuls, puis la suite.
func TestFluxReprendSansDoublonNiTrou(t *testing.T) {
	t.Parallel()

	s, routeur := extracteurDeTest(t, &fabriqueDeTest{})
	session := sessionSur(t, s, "gescom")
	serveurHTTP := httptest.NewServer(routeur)
	t.Cleanup(serveurHTTP.Close)

	lancee := lancerOuEchouer(t, s, routeur, session)
	attendreEtat(t, s.extractions, lancee.ID, EtatTerminee)

	flux := ouvrirFlux(t, s, serveurHTTP.URL, "2")
	for _, attendu := range []uint64{3, 4, 5} {
		if ev := lireEvenement(t, flux); ev.id != attendu || ev.nom != evenementExtraction {
			t.Fatalf("événement %+v, attendu le numéro %d", ev, attendu)
		}
	}

	retirer(t, s, routeur, lancee.ID)
	if ev := lireEvenement(t, flux); ev.id != 6 || ev.nom != evenementRetrait {
		t.Errorf("après la reprise : %+v, attendu le retrait numéro 6", ev)
	}
}

// TestFluxSansRepriseRendUnInstantane couvre l'identifiant qu'on ne sait pas
// reprendre : binaire relancé, valeur illisible, ou journal déjà tourné.
func TestFluxSansRepriseRendUnInstantane(t *testing.T) {
	t.Parallel()

	e := nouvellesExtractions(t.Context(), t.TempDir())
	e.mu.Lock()
	for range tailleJournal + 10 {
		e.emettre(evenementExtraction, Extraction{ID: "x"})
	}
	e.mu.Unlock()

	cas := []struct {
		nom     string
		dernier uint64
		reprise bool
	}{
		{"premier chargement", 0, false},
		{"lancement précédent", e.seq + 40, true},
		{"sorti du journal", 3, true},
	}
	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			evenements, _ := e.depuis(c.dernier, c.reprise)
			if len(evenements) != 1 || evenements[0].nom != evenementEtat || evenements[0].id != e.seq {
				t.Errorf("%d événement(s), attendu un instantané numéroté %d", len(evenements), e.seq)
			}
		})
	}

	evenements, _ := e.depuis(e.seq-1, true)
	if len(evenements) != 1 || evenements[0].id != e.seq {
		t.Errorf("reprise dans le journal : %d événement(s)", len(evenements))
	}
}

// TestExtractionsPlafonnees vérifie qu'au-delà de deux extractions la suivante
// attend, puis démarre dès qu'une place se libère.
func TestExtractionsPlafonnees(t *testing.T) {
	t.Parallel()

	pilotes := map[string]*piloteExtracteur{
		"gescom": {liberer: make(chan struct{})},
		"paie":   {liberer: make(chan struct{})},
		"stock":  {liberer: make(chan struct{})},
	}
	s, routeur := extracteurDeTest(t, &fabriqueDeTest{pilotes: pilotes})

	premiere := lancerOuEchouer(t, s, routeur, sessionSur(t, s, "gescom"))
	seconde := lancerOuEchouer(t, s, routeur, sessionSur(t, s, "paie"))
	troisieme := lancerOuEchouer(t, s, routeur, sessionSur(t, s, "stock"))

	if premiere.Etat != EtatEnCours || seconde.Etat != EtatEnCours {
		t.Errorf("états %q et %q, les deux premières devraient tourner", premiere.Etat, seconde.Etat)
	}
	if troisieme.Etat != EtatEnAttente {
		t.Errorf("troisième dans l'état %q, attendu en attente", troisieme.Etat)
	}

	close(pilotes["gescom"].liberer)
	attendreEtat(t, s.extractions, premiere.ID, EtatTerminee)
	attendreEtat(t, s.extractions, troisieme.ID, EtatEnCours)

	close(pilotes["paie"].liberer)
	close(pilotes["stock"].liberer)
	attendreEtat(t, s.extractions, troisieme.ID, EtatTerminee)
}

// TestExtractionRefuseLaMemeBase vérifie que deux extractions n'écrivent pas le
// même fichier, et qu'une extraction finie ne bloque pas la suivante.
func TestExtractionRefuseLaMemeBase(t *testing.T) {
	t.Parallel()

	pilote := &piloteExtracteur{liberer: make(chan struct{})}
	s, routeur := extracteurDeTest(t, &fabriqueDeTest{pilotes: map[string]*piloteExtracteur{"gescom": pilote}})
	session := sessionSur(t, s, "gescom")

	lancee := lancerOuEchouer(t, s, routeur, session)

	w := posterExtraction(s, routeur, session, `{}`)
	if w.Code != http.StatusConflict {
		t.Fatalf("code %d, attendu %d", w.Code, http.StatusConflict)
	}
	if !strings.Contains(w.Body.String(), "gescom") {
		t.Errorf("le refus ne nomme pas la base : %s", w.Body.String())
	}

	close(pilote.liberer)
	attendreEtat(t, s.extractions, lancee.ID, EtatTerminee)
	lancerOuEchouer(t, s, routeur, session)
}

// TestExtractionRefuseLEchantillonnage vérifie qu'une option sans effet est
// refusée plutôt qu'ignorée, avant qu'aucune connexion ne s'ouvre.
func TestExtractionRefuseLEchantillonnage(t *testing.T) {
	t.Parallel()

	f := &fabriqueDeTest{}
	s, routeur := extracteurDeTest(t, f)

	w := posterExtraction(s, routeur, sessionSur(t, s, "gescom"), `{"echantillonner": true}`)
	if w.Code != http.StatusNotImplemented {
		t.Errorf("code %d, attendu %d", w.Code, http.StatusNotImplemented)
	}
	if recus := f.ouvertures(); len(recus) != 0 {
		t.Errorf("%d connexion(s) ouverte(s) malgré le refus", len(recus))
	}
}

// TestExtractionEtNomDeBase vérifie que seul un nom qui ne sort pas du
// répertoire de travail nomme un fichier, accents compris.
func TestExtractionEtNomDeBase(t *testing.T) {
	t.Parallel()

	cas := []struct {
		dsn     string
		accepte bool
	}{
		{"postgres://u@h:5432/gescom", true},
		{"postgres://u@h:5432/r%C3%A9f%C3%A9rence", true},
		{"postgres://u@h:5432/paie_2024-bis", true},
		{"postgres://u@h:5432/..", false},
		{"postgres://u@h:5432/a/b", false},
		{"postgres://u@h:5432/ma%20base", false},
		{"postgres://u@h:5432/", false},
	}
	for _, c := range cas {
		t.Run(c.dsn, func(t *testing.T) {
			t.Parallel()

			f := &fabriqueDeTest{}
			s, routeur := extracteurDeTest(t, f)
			session, err := s.registre.ajouter(&piloteDeTest{}, c.dsn, "postgres")
			if err != nil {
				t.Fatalf("ajouter: %v", err)
			}

			w := posterExtraction(s, routeur, session, `{}`)
			if c.accepte && w.Code != http.StatusAccepted {
				t.Errorf("code %d, attendu %d : %s", w.Code, http.StatusAccepted, w.Body.String())
			}
			if !c.accepte && w.Code != http.StatusBadRequest {
				t.Errorf("code %d, attendu %d", w.Code, http.StatusBadRequest)
			}
			if !c.accepte && len(f.ouvertures()) != 0 {
				t.Error("une connexion a été ouverte pour un nom refusé")
			}
		})
	}
}

// TestExtractionAnnuleeEnCours vérifie qu'une extraction arrêtée n'écrit rien,
// referme sa connexion, et se retire ensuite.
func TestExtractionAnnuleeEnCours(t *testing.T) {
	t.Parallel()

	pilote := &piloteExtracteur{liberer: make(chan struct{})}
	s, routeur := extracteurDeTest(t, &fabriqueDeTest{pilotes: map[string]*piloteExtracteur{"gescom": pilote}})

	lancee := lancerOuEchouer(t, s, routeur, sessionSur(t, s, "gescom"))
	retirer(t, s, routeur, lancee.ID)
	finale := attendreEtat(t, s.extractions, lancee.ID, EtatAnnulee)

	if finale.Resultat != nil || finale.Erreur != "" {
		t.Errorf("tâche annulée : %+v", finale)
	}
	if _, err := os.Stat(filepath.Join(s.repertoire, "gescom.calque.json")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("calque écrit malgré l'annulation : %v", err)
	}
	if pilote.ferme != 1 {
		t.Errorf("connexion fermée %d fois, attendu 1", pilote.ferme)
	}

	retirer(t, s, routeur, lancee.ID)
	if _, trouvee, _ := tacheCourante(s.extractions, lancee.ID); trouvee {
		t.Error("tâche finie encore présente après retrait")
	}
}

// TestExtractionAnnuleeEnAttente vérifie qu'une tâche retirée de la file ne
// démarre jamais, même quand une place se libère.
func TestExtractionAnnuleeEnAttente(t *testing.T) {
	t.Parallel()

	pilotes := map[string]*piloteExtracteur{
		"gescom": {liberer: make(chan struct{})},
		"paie":   {liberer: make(chan struct{})},
	}
	f := &fabriqueDeTest{pilotes: pilotes}
	s, routeur := extracteurDeTest(t, f)

	premiere := lancerOuEchouer(t, s, routeur, sessionSur(t, s, "gescom"))
	lancerOuEchouer(t, s, routeur, sessionSur(t, s, "paie"))
	enAttente := lancerOuEchouer(t, s, routeur, sessionSur(t, s, "stock"))

	retirer(t, s, routeur, enAttente.ID)
	attendreEtat(t, s.extractions, enAttente.ID, EtatAnnulee)

	close(pilotes["gescom"].liberer)
	attendreEtat(t, s.extractions, premiere.ID, EtatTerminee)
	close(pilotes["paie"].liberer)

	if courante, _, _ := tacheCourante(s.extractions, enAttente.ID); courante.Etat != EtatAnnulee {
		t.Errorf("tâche retirée dans l'état %q", courante.Etat)
	}
	for _, dsn := range f.ouvertures() {
		if introspection.BaseDuDSN(dsn) == "stock" {
			t.Error("une connexion a été ouverte pour la tâche retirée")
		}
	}
}

// TestExtractionSurvitALaSession vérifie ce qui libère l'interface : fermer la
// connexion de l'arbre n'arrête pas l'extraction qu'elle a lancée.
func TestExtractionSurvitALaSession(t *testing.T) {
	t.Parallel()

	pilote := &piloteExtracteur{liberer: make(chan struct{})}
	s, routeur := extracteurDeTest(t, &fabriqueDeTest{pilotes: map[string]*piloteExtracteur{"gescom": pilote}})
	session := sessionSur(t, s, "gescom")

	lancee := lancerOuEchouer(t, s, routeur, session)

	w := httptest.NewRecorder()
	routeur.ServeHTTP(w, requeteAPI(s, http.MethodDelete, "/api/connexion", strings.NewReader(`{"session": "`+session+`"}`)))
	if w.Code != http.StatusNoContent {
		t.Fatalf("fermeture de la session : code %d", w.Code)
	}

	if courante, _, _ := tacheCourante(s.extractions, lancee.ID); courante.Etat != EtatEnCours {
		t.Errorf("tâche dans l'état %q après la fermeture de la session", courante.Etat)
	}
	close(pilote.liberer)
	attendreEtat(t, s.extractions, lancee.ID, EtatTerminee)
}

// TestExtractionEchoueSansDivulguerLeDSN vérifie qu'un échec atteint l'écran
// sans que le DSN de la session n'y figure, ni dans l'événement ni dans le
// journal de reprise.
func TestExtractionEchoueSansDivulguerLeDSN(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom      string
		fabrique *fabriqueDeTest
	}{
		{"connexion refusée", &fabriqueDeTest{echec: errors.New("connexion a " + dsnSur("gescom") + " refusee")}},
		{"extraction en échec", &fabriqueDeTest{pilotes: map[string]*piloteExtracteur{
			"gescom": {echec: errors.New("lecture sur " + introspection.Masquer(dsnSur("gescom")) + " perdue")},
		}}},
	}
	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			s, routeur := extracteurDeTest(t, c.fabrique)
			lancee := lancerOuEchouer(t, s, routeur, sessionSur(t, s, "gescom"))
			finale := attendreEtat(t, s.extractions, lancee.ID, EtatEchouee)

			if finale.Erreur == "" {
				t.Error("échec sans message : l'écran n'a rien à afficher")
			}
			evenements, _ := s.extractions.depuis(0, true)
			for _, interdit := range []string{"secret", "bdd-interne"} {
				if strings.Contains(finale.Erreur, interdit) {
					t.Errorf("%q ressort dans l'erreur %q", interdit, finale.Erreur)
				}
				for _, ev := range evenements {
					if strings.Contains(string(ev.donnees), interdit) {
						t.Errorf("%q ressort dans l'événement %d", interdit, ev.id)
					}
				}
			}
		})
	}
}

// TestExtractionsFiniesPlafonnees vérifie que l'historique ne grossit pas sans
// fin : au-delà du plafond, la plus ancienne tâche finie disparaît.
func TestExtractionsFiniesPlafonnees(t *testing.T) {
	t.Parallel()

	e := nouvellesExtractions(t.Context(), t.TempDir())
	e.ouvrir = (&fabriqueDeTest{}).ouvrir
	t.Cleanup(e.attendre)

	var premiere Extraction
	for i := range maxExtractionsFinies + 1 {
		base := fmt.Sprintf("base%02d", i)
		lancee, err := e.lancer(base, dsnSur(base), "postgres", introspection.Portee{})
		if err != nil {
			t.Fatalf("lancer %s : %v", base, err)
		}
		if i == 0 {
			premiere = lancee
		}
	}

	delai := time.After(5 * time.Second)
	for {
		e.mu.Lock()
		finies, restantes, signal := 0, len(e.taches), e.signal
		for _, courante := range e.taches {
			if courante.Etat.terminal() {
				finies++
			}
		}
		e.mu.Unlock()

		if finies == restantes && restantes == maxExtractionsFinies {
			break
		}
		select {
		case <-signal:
		case <-delai:
			t.Fatalf("%d tâche(s) finie(s) sur %d gardée(s)", finies, restantes)
		}
	}

	if _, trouvee, _ := tacheCourante(e, premiere.ID); trouvee {
		t.Error("la plus ancienne tâche finie est encore là")
	}
}

// TestArretAnnuleLesExtractions vérifie qu'arrêter le serveur n'attend pas la
// fin normale d'une extraction.
func TestArretAnnuleLesExtractions(t *testing.T) {
	t.Parallel()

	ctx, arreter := context.WithCancel(t.Context())
	e := nouvellesExtractions(ctx, t.TempDir())
	pilote := &piloteExtracteur{liberer: make(chan struct{})}
	e.ouvrir = (&fabriqueDeTest{pilotes: map[string]*piloteExtracteur{"gescom": pilote}}).ouvrir

	lancee, err := e.lancer("gescom", dsnSur("gescom"), "postgres", introspection.Portee{})
	if err != nil {
		t.Fatalf("lancer : %v", err)
	}
	arreter()

	fini := make(chan struct{})
	go func() {
		e.attendre()
		close(fini)
	}()
	select {
	case <-fini:
	case <-time.After(2 * time.Second):
		t.Fatal("l'arrêt attend encore l'extraction")
	}

	if courante, _, _ := tacheCourante(e, lancee.ID); courante.Etat != EtatAnnulee {
		t.Errorf("tâche dans l'état %q après l'arrêt", courante.Etat)
	}
}

// piloteDescripteur sait se décrire, ce qu'exige l'enregistrement d'une
// session.
type piloteDescripteur struct {
	piloteDeTest
	catalogue string
}

// Decrire rend le catalogue préparé.
func (p *piloteDescripteur) Decrire(context.Context) (introspection.Serveur, error) {
	return introspection.Serveur{SGBD: "postgres", Version: "17.2", Catalogue: p.catalogue, Schemas: []string{"public"}}, nil
}

// TestEnregistrerNommeLaBase vérifie qu'une session ouverte sans nommer de base
// retient celle que le serveur a choisie : c'est elle qui nomme le calque.
func TestEnregistrerNommeLaBase(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom     string
		dsn     string
		attendu string
	}{
		{"url sans base", "postgres://gescom:secret@bdd-interne:5432", "gescom"},
		{"url avec base", "postgres://gescom:secret@bdd-interne:5432/paie", "paie"},
		{"clé valeur sans base", "host=bdd-interne user=gescom password=secret", "gescom"},
	}
	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			s, _ := serveurDeTest(t)
			pilote := &piloteDescripteur{catalogue: c.attendu}
			reponse, err := s.enregistrer(t.Context(), pilote, c.dsn, "postgres")
			if err != nil {
				t.Fatalf("enregistrer : %v", err)
			}

			connexion, ok := s.registre.trouver(reponse.Session)
			if !ok {
				t.Fatal("session introuvable")
			}
			if base := introspection.BaseDuDSN(connexion.dsn); base != c.attendu {
				t.Errorf("base retenue %q, attendue %q", base, c.attendu)
			}
		})
	}
}
