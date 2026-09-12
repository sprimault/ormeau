// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sprimault/ormeau/internal/config"
)

// profilPoste enregistre un profil par l'API.
func profilPoste(s *serveur, routeur http.Handler, p config.Profil,
	motDePasse string, avec, remplacer bool) *httptest.ResponseRecorder {
	return poster(s, routeur, "/api/profils", RequeteProfil{
		Profil:                p,
		MotDePasse:            motDePasse,
		EnregistrerMotDePasse: avec,
		Remplacer:             remplacer,
	})
}

// supprimerAPI envoie un corps JSON par DELETE.
func supprimerAPI(s *serveur, routeur http.Handler, chemin string, corps any) *httptest.ResponseRecorder {
	donnees, _ := json.Marshal(corps)
	w := httptest.NewRecorder()
	routeur.ServeHTTP(w, requeteAPI(s, http.MethodDelete, chemin, bytes.NewReader(donnees)))
	return w
}

// TestProfilsEnregistresEtListes couvre l'aller-retour par l'API.
func TestProfilsEnregistresEtListes(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	vide := decoderReponse[ReponseProfils](t, lire(s, routeur, "/api/profils"))
	if len(vide.Profils) != 0 {
		t.Errorf("profils au premier lancement : %+v", vide.Profils)
	}

	reponse := decoderReponse[ReponseProfils](t, profilPoste(s, routeur,
		config.Profil{Nom: "nas", Hote: "192.168.1.10", Base: "gescom"}, "secret", true, false))
	if len(reponse.Profils) != 1 || !reponse.Profils[0].MotDePasseEnregistre {
		t.Fatalf("profils après enregistrement : %+v", reponse.Profils)
	}
}

// TestProfilNeRendJamaisLeMotDePasse vérifie la frontière : le chiffré lui-même
// ne descend pas dans le navigateur, seulement le fait qu'il existe.
func TestProfilNeRendJamaisLeMotDePasse(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	const secret = "M0tDeP4sse-Production"

	attendreStatut(t, profilPoste(s, routeur,
		config.Profil{Nom: "nas", Hote: "h"}, secret, true, false), http.StatusOK)

	corps := lire(s, routeur, "/api/profils").Body.String()
	if strings.Contains(corps, secret) {
		t.Error("le mot de passe en clair atteint le navigateur")
	}
	if strings.Contains(corps, "mot_de_passe\":\"") {
		t.Errorf("le chiffré atteint le navigateur : %s", corps)
	}
	if !strings.Contains(corps, `"mot_de_passe_enregistre":true`) {
		t.Errorf("l'écran ne peut pas savoir qu'un mot de passe existe : %s", corps)
	}
}

// TestProfilExistantDemandeConfirmation vérifie le refus et son code, que
// l'écran distingue pour demander plutôt que d'écraser.
func TestProfilExistantDemandeConfirmation(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	depart := config.Profil{Nom: "nas", Hote: "192.168.1.10", Base: "gescom"}
	attendreStatut(t, profilPoste(s, routeur, depart, "", false, false), http.StatusOK)

	modifie := depart
	modifie.Base = "paie"
	attendreRefus(t, profilPoste(s, routeur, modifie, "", false, false), CodeProfilExistant)

	// Le refus ne touche à rien.
	liste := decoderReponse[ReponseProfils](t, lire(s, routeur, "/api/profils"))
	if liste.Profils[0].Profil.Base != "gescom" {
		t.Errorf("le refus a quand même écrit : %+v", liste.Profils[0])
	}

	attendreStatut(t, profilPoste(s, routeur, modifie, "", true, true), http.StatusOK)
	liste = decoderReponse[ReponseProfils](t, lire(s, routeur, "/api/profils"))
	if liste.Profils[0].Profil.Base != "paie" {
		t.Errorf("le remplacement confirmé n'a pas eu lieu : %+v", liste.Profils[0])
	}
}

// TestCaseDecocheeEffaceLeMotDePasse couvre le geste inverse : ce qui est
// décoché doit disparaître du disque, pas seulement de l'écran.
func TestCaseDecocheeEffaceLeMotDePasse(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	p := config.Profil{Nom: "nas", Hote: "h"}

	attendreStatut(t, profilPoste(s, routeur, p, "secret", true, false), http.StatusOK)
	reponse := decoderReponse[ReponseProfils](t, profilPoste(s, routeur, p, "secret", false, true))

	if reponse.Profils[0].MotDePasseEnregistre {
		t.Error("le mot de passe survit à une case décochée")
	}
}

// TestProfilEnregistreDepuisUnDSN couvre la seconde forme de saisie : sans
// décomposition, le profil ne garderait ni hôte ni base et ne servirait à rien
// la fois suivante.
func TestProfilEnregistreDepuisUnDSN(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	reponse := decoderReponse[ReponseProfils](t, poster(s, routeur, "/api/profils", RequeteProfil{
		Profil:                config.Profil{Nom: "depuis dsn"},
		DSN:                   "postgresql://postgres:secret@192.168.0.184:30432/cadensio_main",
		EnregistrerMotDePasse: true,
	}))

	if len(reponse.Profils) != 1 {
		t.Fatalf("profils %+v", reponse.Profils)
	}
	p := reponse.Profils[0].Profil
	switch {
	// postgresql:// et postgres:// désignent le même : c'est le nom du SGBD que
	// le profil porte, pas le préfixe tel qu'il a été tapé.
	case p.SGBD != "postgres":
		t.Errorf("sgbd %q", p.SGBD)
	case p.Hote != "192.168.0.184" || p.Port != 30432:
		t.Errorf("hôte %s:%d", p.Hote, p.Port)
	case p.Utilisateur != "postgres" || p.Base != "cadensio_main":
		t.Errorf("utilisateur %q, base %q", p.Utilisateur, p.Base)
	case !reponse.Profils[0].MotDePasseEnregistre:
		t.Error("le mot de passe du DSN n'a pas été retenu")
	}

	clair, err := s.emplacements.MotDePasseDuProfil("depuis dsn")
	if err != nil || clair != "secret" {
		t.Errorf("mot de passe relu %q, erreur %v", clair, err)
	}
}

// TestDSNSansCaseNEnregistrePasLeMotDePasse vérifie que la case gouverne aussi
// le mot de passe caché dans une chaîne de connexion.
func TestDSNSansCaseNEnregistrePasLeMotDePasse(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	reponse := decoderReponse[ReponseProfils](t, poster(s, routeur, "/api/profils", RequeteProfil{
		Profil: config.Profil{Nom: "depuis dsn"},
		DSN:    "postgresql://postgres:secret@192.168.0.184:30432/cadensio_main",
	}))

	if reponse.Profils[0].MotDePasseEnregistre {
		t.Error("un mot de passe de DSN est enregistré sans que la case soit cochée")
	}
	if reponse.Profils[0].Profil.Hote != "192.168.0.184" {
		t.Errorf("le reste du DSN n'a pas été retenu : %+v", reponse.Profils[0].Profil)
	}
}

// TestProfilDepuisUnDSNIllisible vérifie le refus, et qu'il ne cite pas la
// chaîne — elle porte un mot de passe.
func TestProfilDepuisUnDSNIllisible(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	w := poster(s, routeur, "/api/profils", RequeteProfil{
		Profil: config.Profil{Nom: "casse"},
		DSN:    "postgres://utilisateur:secret@\x7f/base",
	})
	attendreStatut(t, w, http.StatusBadRequest)
	if strings.Contains(w.Body.String(), "secret") {
		t.Errorf("le refus cite le mot de passe : %s", w.Body.String())
	}
}

// TestProfilSupprime couvre le retrait et le nom inconnu.
func TestProfilSupprime(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	attendreStatut(t, profilPoste(s, routeur, config.Profil{Nom: "nas", Hote: "h"}, "", false, false),
		http.StatusOK)

	w := supprimerAPI(s, routeur, "/api/profils", ReferenceProfil{Nom: "nas"})
	attendreStatut(t, w, http.StatusOK)

	liste := decoderReponse[ReponseProfils](t, lire(s, routeur, "/api/profils"))
	if len(liste.Profils) != 0 {
		t.Errorf("profils après suppression : %+v", liste.Profils)
	}

	attendreStatut(t, supprimerAPI(s, routeur, "/api/profils", ReferenceProfil{Nom: "jamais"}),
		http.StatusNotFound)
}

// TestConnexionParProfilInconnu vérifie qu'un nom qui ne désigne rien est
// refusé avant toute tentative de connexion.
func TestConnexionParProfilInconnu(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	attendreStatut(t, poster(s, routeur, "/api/connexion", RequeteConnexion{Profil: "jamais"}),
		http.StatusNotFound)
}

// TestProfilCompleteLaConnexion vérifie que le serveur reprend du profil ce que
// la requête ne dit pas, mot de passe compris — le front n'envoie qu'un nom.
func TestProfilCompleteLaConnexion(t *testing.T) {
	t.Parallel()

	s, _ := serveurDeTest(t)
	if err := s.emplacements.EnregistrerProfil(config.Profil{
		Nom: "nas", SGBD: "postgres", Hote: "192.168.1.10", Port: 5432,
		Utilisateur: "lecture", Base: "gescom",
	}, "secret", true); err != nil {
		t.Fatalf("enregistrement : %v", err)
	}

	requete := RequeteConnexion{Profil: "nas"}
	if _, err := s.connexionDuProfil(&requete); err != nil {
		t.Fatalf("complétion : %v", err)
	}
	if requete.Hote != "192.168.1.10" || requete.Port != 5432 || requete.Base != "gescom" {
		t.Errorf("requête complétée %+v", requete)
	}
	if requete.MotDePasse != "secret" {
		t.Errorf("mot de passe non repris du profil")
	}
}

// TestProfilSansDSNDepuisLOngletChaine couvre la connexion par profil depuis
// l'onglet « Chaîne de connexion », la chaîne laissée vide.
//
// Le front n'envoie alors qu'un nom de profil : sans ce chemin, se connecter
// depuis cet onglet partirait sans mot de passe et la base refuserait sans que
// rien ne l'explique.
func TestProfilSansDSNDepuisLOngletChaine(t *testing.T) {
	t.Parallel()

	s, _ := serveurDeTest(t)
	if err := s.emplacements.EnregistrerProfil(config.Profil{
		Nom: "nas", SGBD: "postgres", Hote: "192.168.0.184", Port: 30432,
		Utilisateur: "postgres", Base: "cadensio_main",
	}, "secret", true); err != nil {
		t.Fatalf("enregistrement : %v", err)
	}

	requete := RequeteConnexion{Profil: "nas", DSN: ""}
	if _, err := s.connexionDuProfil(&requete); err != nil {
		t.Fatalf("complétion : %v", err)
	}

	dsn, err := requete.composer()
	if err != nil {
		t.Fatalf("composition : %v", err)
	}
	if !strings.Contains(dsn, "192.168.0.184:30432") || !strings.Contains(dsn, "cadensio_main") {
		t.Errorf("dsn composé sans le profil")
	}
	if !strings.Contains(dsn, "secret") {
		t.Error("le mot de passe enregistré n'est pas dans le DSN composé")
	}
}

// TestDSNSaisiLEmporteSurLeProfil : une chaîne tapée décrit la connexion
// voulue, elle n'est pas un complément du profil.
func TestDSNSaisiLEmporteSurLeProfil(t *testing.T) {
	t.Parallel()

	s, _ := serveurDeTest(t)
	if err := s.emplacements.EnregistrerProfil(config.Profil{
		Nom: "nas", SGBD: "postgres", Hote: "192.168.0.184", Base: "cadensio_main",
	}, "secret", true); err != nil {
		t.Fatalf("enregistrement : %v", err)
	}

	requete := RequeteConnexion{Profil: "nas", DSN: "postgres://u:p@autre:5432/ailleurs"}
	if _, err := s.connexionDuProfil(&requete); err != nil {
		t.Fatalf("complétion : %v", err)
	}

	dsn, err := requete.composer()
	if err != nil {
		t.Fatalf("composition : %v", err)
	}
	if dsn != "postgres://u:p@autre:5432/ailleurs" {
		t.Errorf("dsn composé %q", dsn)
	}
}

// TestProfilAvecBaseVerrouilleLaSession vérifie le cadrage : un profil qui
// nomme une base dit sur quoi l'on travaille, et l'écran ne propose pas le
// reste.
//
// Ce n'est pas une barrière et le test ne prétend pas l'être : le compte garde
// le droit d'ouvrir les autres bases par un autre outil ou sans profil. La
// seule vraie limite se pose côté serveur.
func TestProfilAvecBaseVerrouilleLaSession(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	session := sessionImposee(t, s, "gescom")

	bases := decoderReponse[ReponseBases](t, lire(s, routeur, "/api/bases?session="+session))
	if len(bases.Bases) != 1 || bases.Bases[0] != "gescom" {
		t.Errorf("bases proposées %v", bases.Bases)
	}

	w := poster(s, routeur, "/api/base", RequeteBase{Session: session, Base: "paie"})
	attendreStatut(t, w, http.StatusConflict)
	if !strings.Contains(w.Body.String(), "gescom") {
		t.Errorf("le refus ne nomme pas la base du profil : %s", w.Body.String())
	}
}

// TestProfilSansBaseLaisseParcourir couvre l'autre moitié de la règle : un
// profil qui ne nomme aucune base sert à parcourir le serveur.
func TestProfilSansBaseLaisseParcourir(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	session, err := s.registre.ajouter(&piloteDeTest{}, dsnSur("gescom"), "postgres", false)
	if err != nil {
		t.Fatalf("session : %v", err)
	}

	// Le pilote de test ne liste pas de bases ; ce qui compte est que le
	// point d'entrée ne court-circuite pas sur la seule base de la session.
	w := lire(s, routeur, "/api/bases?session="+session)
	attendreStatut(t, w, http.StatusOK)
	bases := decoderReponse[ReponseBases](t, w)
	if len(bases.Bases) == 1 && bases.Bases[0] == "gescom" {
		t.Error("une session libre a été traitée comme verrouillée")
	}
}

// TestBaseImposeeVientDuProfilQuiLaNomme vérifie d'où sort le verrou.
func TestBaseImposeeVientDuProfilQuiLaNomme(t *testing.T) {
	t.Parallel()

	s, _ := serveurDeTest(t)
	for _, c := range []struct {
		nom     string
		base    string
		attendu bool
	}{
		{"avec base", "gescom", true},
		{"sans base", "", false},
	} {
		t.Run(c.nom, func(t *testing.T) {
			if err := s.emplacements.EnregistrerProfil(config.Profil{
				Nom: c.nom, SGBD: "postgres", Hote: "h", Port: 5432, Base: c.base,
			}, "", true); err != nil {
				t.Fatalf("enregistrement : %v", err)
			}

			requete := RequeteConnexion{Profil: c.nom}
			impose, err := s.connexionDuProfil(&requete)
			if err != nil {
				t.Fatalf("complétion : %v", err)
			}
			if impose != c.attendu {
				t.Errorf("baseImposee %v, attendu %v", impose, c.attendu)
			}
		})
	}
}

// sessionImposee rend une session ouverte comme par un profil qui nomme sa
// base.
func sessionImposee(t *testing.T, s *serveur, base string) string {
	t.Helper()

	session, err := s.registre.ajouter(&piloteDeTest{}, dsnSur(base), "postgres", true)
	if err != nil {
		t.Fatalf("session : %v", err)
	}
	return session
}

// TestMotDePasseSaisiLEmporte : on vient de le taper, c'est qu'il a changé.
func TestMotDePasseSaisiLEmporte(t *testing.T) {
	t.Parallel()

	s, _ := serveurDeTest(t)
	if err := s.emplacements.EnregistrerProfil(config.Profil{Nom: "nas", Hote: "h"},
		"ancien", true); err != nil {
		t.Fatalf("enregistrement : %v", err)
	}

	requete := RequeteConnexion{Profil: "nas", MotDePasse: "nouveau"}
	if _, err := s.connexionDuProfil(&requete); err != nil {
		t.Fatalf("complétion : %v", err)
	}
	if requete.MotDePasse != "nouveau" {
		t.Errorf("mot de passe retenu %q", requete.MotDePasse)
	}
}

// TestProfilBasculeLeRepertoire vérifie qu'un profil emmène dans son projet, et
// qu'un profil sans répertoire laisse le courant tel quel.
func TestProfilBasculeLeRepertoire(t *testing.T) {
	t.Parallel()

	s, _ := serveurDeTest(t)
	ailleurs := t.TempDir()
	depart := s.repertoireCourant()

	if err := s.emplacements.EnregistrerProfil(config.Profil{
		Nom: "avec", Hote: "h", Repertoire: ailleurs,
	}, "", true); err != nil {
		t.Fatalf("enregistrement : %v", err)
	}
	if err := s.emplacements.EnregistrerProfil(config.Profil{Nom: "sans", Hote: "h"},
		"", true); err != nil {
		t.Fatalf("enregistrement : %v", err)
	}

	requete := RequeteConnexion{Profil: "sans"}
	if _, err := s.connexionDuProfil(&requete); err != nil {
		t.Fatalf("complétion : %v", err)
	}
	if s.repertoireCourant() != depart {
		t.Errorf("un profil sans répertoire a déplacé le courant : %s", s.repertoireCourant())
	}

	requete = RequeteConnexion{Profil: "avec"}
	if _, err := s.connexionDuProfil(&requete); err != nil {
		t.Fatalf("complétion : %v", err)
	}
	if s.repertoireCourant() == depart {
		t.Error("le répertoire du profil n'a pas été suivi")
	}
}
