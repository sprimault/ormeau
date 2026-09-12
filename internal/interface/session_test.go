// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
	"github.com/sprimault/ormeau/internal/config"
)

// sessionDeTest compose un brouillon cohérent avec ce que le serveur porte.
func sessionDeTest(t *testing.T, s *serveur, base string) config.Session {
	t.Helper()

	physique, ok := s.calqueDeBase(nil, base, "")
	if !ok || physique == nil {
		t.Fatalf("calque de %s introuvable", base)
	}
	return config.Session{
		Base:            base,
		EmpreinteCalque: physique.Source.Empreinte,
		Brouillon:       `{"renommages":{"public.client":"Client"}}`,
		EntiteOuverte:   "public.client",
	}
}

// TestSessionAbsenteNEstPasUnIncident couvre la base ouverte pour la première
// fois.
func TestSessionAbsenteNEstPasUnIncident(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	ecrireCalque(t, s, "gescom", physiqueDeTest())

	reponse := decoderReponse[ReponseSession](t, lire(s, routeur, "/api/session?base=gescom"))
	if reponse.Session != nil || reponse.Ecartee != "" {
		t.Errorf("session %+v, écartée %q", reponse.Session, reponse.Ecartee)
	}
}

// TestSessionEcritePuisRelue couvre l'aller-retour du brouillon.
func TestSessionEcritePuisRelue(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	ecrireCalque(t, s, "gescom", physiqueDeTest())

	attendreStatut(t, poster(s, routeur, "/api/session", sessionDeTest(t, s, "gescom")), http.StatusOK)

	relue := decoderReponse[ReponseSession](t, lire(s, routeur, "/api/session?base=gescom"))
	if relue.Session == nil {
		t.Fatalf("brouillon perdu, écartée %q", relue.Ecartee)
	}
	if relue.Session.EntiteOuverte != "public.client" {
		t.Errorf("entité ouverte %q", relue.Session.EntiteOuverte)
	}
	if !strings.Contains(string(relue.Session.Brouillon), "Client") {
		t.Errorf("brouillon relu %s", relue.Session.Brouillon)
	}
}

// TestSessionJeteeQuandLeCalqueABouge couvre la réextraction : des décisions
// prises sur l'ancien calque porteraient sur un schéma qui a changé.
func TestSessionJeteeQuandLeCalqueABouge(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	ecrireCalque(t, s, "gescom", physiqueDeTest())
	attendreStatut(t, poster(s, routeur, "/api/session", sessionDeTest(t, s, "gescom")), http.StatusOK)

	// Une extraction qui ajoute une table change l'empreinte du calque.
	autre := physiqueDeTest()
	autre.Tables = append(autre.Tables, calque.Table{
		Nom:      "facture",
		Schema:   "public",
		Colonnes: []calque.Colonne{{Nom: "id", Position: 1, TypeBrut: "integer", TypeNormalise: calque.TypeEntier}},
	})
	ecrireCalque(t, s, "gescom", autre)

	reponse := decoderReponse[ReponseSession](t, lire(s, routeur, "/api/session?base=gescom"))
	if reponse.Session != nil {
		t.Fatal("un brouillon caduc a été rendu")
	}
	if !strings.Contains(reponse.Ecartee, "réextrait") {
		t.Errorf("raison %q", reponse.Ecartee)
	}

	// Effacé dans la foulée : le garder pour le rejeter à chaque ouverture
	// n'aide personne.
	suivante := decoderReponse[ReponseSession](t, lire(s, routeur, "/api/session?base=gescom"))
	if suivante.Ecartee != "" {
		t.Errorf("le brouillon caduc n'a pas été effacé : %q", suivante.Ecartee)
	}
}

// TestSessionJeteeQuandLesDecisionsOntChange couvre le fichier retouché dans un
// éditeur : il fait foi, et un brouillon rejoué par-dessus réintroduirait ce
// qu'on venait d'en retirer.
func TestSessionJeteeQuandLesDecisionsOntChange(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	ecrireCalque(t, s, "gescom", physiqueDeTest())
	attendreStatut(t, poster(s, routeur, "/api/session", sessionDeTest(t, s, "gescom")), http.StatusOK)

	if err := os.WriteFile(s.cheminDecisions("gescom"),
		[]byte("renommages:\n  public.client: Autre\n"), 0o600); err != nil {
		t.Fatalf("écriture des décisions : %v", err)
	}

	reponse := decoderReponse[ReponseSession](t, lire(s, routeur, "/api/session?base=gescom"))
	if reponse.Session != nil {
		t.Fatal("un brouillon caduc a été rendu")
	}
	if !strings.Contains(reponse.Ecartee, "décisions") {
		t.Errorf("raison %q", reponse.Ecartee)
	}
}

// TestSessionEffacee couvre ce que fait l'écran après un enregistrement : ce
// qui est dans le fichier n'a plus à être retenu ailleurs.
func TestSessionEffacee(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	ecrireCalque(t, s, "gescom", physiqueDeTest())
	attendreStatut(t, poster(s, routeur, "/api/session", sessionDeTest(t, s, "gescom")), http.StatusOK)

	attendreStatut(t, supprimerAPI(s, routeur, "/api/session", ReferenceSession{Base: "gescom"}),
		http.StatusOK)

	reponse := decoderReponse[ReponseSession](t, lire(s, routeur, "/api/session?base=gescom"))
	if reponse.Session != nil {
		t.Error("le brouillon a survécu à sa suppression")
	}

	// Effacer ce qui n'existe pas atteint le résultat voulu.
	attendreStatut(t, supprimerAPI(s, routeur, "/api/session", ReferenceSession{Base: "gescom"}),
		http.StatusOK)
}

// TestSessionRefuseUnNomDeBaseHorsRepertoire applique au brouillon la règle des
// autres fichiers : le nom vient du navigateur, il ne doit pas pouvoir sortir
// du répertoire des sessions.
func TestSessionRefuseUnNomDeBaseHorsRepertoire(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	attendreStatut(t, lire(s, routeur, "/api/session?base=../evasion"), http.StatusBadRequest)
	attendreStatut(t, poster(s, routeur, "/api/session", config.Session{Base: "../evasion"}),
		http.StatusBadRequest)
	attendreStatut(t, supprimerAPI(s, routeur, "/api/session", ReferenceSession{Base: "../evasion"}),
		http.StatusBadRequest)
}

// TestSessionsDeDeuxProjets couvre la collision que la clé doit éviter : deux
// projets ayant chacun une base « gescom ».
func TestSessionsDeDeuxProjets(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	ecrireCalque(t, s, "gescom", physiqueDeTest())
	premier := sessionDeTest(t, s, "gescom")
	premier.EntiteOuverte = "public.premier"
	attendreStatut(t, poster(s, routeur, "/api/session", premier), http.StatusOK)

	// Même base, autre projet : le brouillon du premier ne doit pas remonter.
	ailleurs := t.TempDir()
	attendreStatut(t, poster(s, routeur, "/api/repertoire", RequeteRepertoire{Repertoire: ailleurs}),
		http.StatusOK)
	ecrireCalque(t, s, "gescom", physiqueDeTest())

	reponse := decoderReponse[ReponseSession](t, lire(s, routeur, "/api/session?base=gescom"))
	if reponse.Session != nil {
		t.Errorf("le brouillon d'un autre projet a été rendu : %+v", reponse.Session)
	}
}
