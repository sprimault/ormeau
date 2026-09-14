// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	if _, err := s.emplacements.EnregistrerProfil(config.Profil{
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
	if _, err := s.emplacements.EnregistrerProfil(config.Profil{
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

// TestProfilEtChaineRefuses : une chaîne décrit à elle seule la connexion
// voulue. Tolérer un profil à côté lui imposerait la base et le répertoire du
// profil sur un serveur qui n'est peut-être pas le sien. Le front n'envoie
// jamais les deux ; le serveur le refuse avant toute tentative.
func TestProfilEtChaineRefuses(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	if _, err := s.emplacements.EnregistrerProfil(config.Profil{
		Nom: "nas", SGBD: "postgres", Hote: "192.168.1.10", Base: "gescom",
	}, "secret", true); err != nil {
		t.Fatalf("enregistrement : %v", err)
	}

	w := poster(s, routeur, "/api/connexion", RequeteConnexion{Profil: "nas", DSN: "postgres://u@autre:5432/ailleurs"})
	attendreStatut(t, w, http.StatusBadRequest)
}

// profilAvecMotDePasse enregistre le profil de référence des tests de
// destination, mot de passe compris.
func profilAvecMotDePasse(t *testing.T, s *serveur) {
	t.Helper()
	if _, err := s.emplacements.EnregistrerProfil(config.Profil{
		Nom: "production", SGBD: "postgres", Hote: "bdd.exemple", Port: 5432,
		Utilisateur: "lecture", Base: "gescom",
	}, "secret", true); err != nil {
		t.Fatalf("enregistrement : %v", err)
	}
}

// TestMotDePasseNonEnvoyeVersUneAutreDestination est le test qui protège le
// mot de passe enregistré : choisir le profil « production » puis corriger un
// champ du formulaire ne doit pas l'envoyer à un autre serveur, ni l'essayer
// pour un autre compte, où l'échec compterait contre lui. Le refus arrive avant
// tout contact avec la base, en nommant le champ.
func TestMotDePasseNonEnvoyeVersUneAutreDestination(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom     string
		requete RequeteConnexion
		champ   string
	}{
		{"autre sgbd", RequeteConnexion{SGBD: "mysql"}, "SGBD"},
		{"autre hôte", RequeteConnexion{Hote: "recette.exemple"}, "hôte"},
		{"autre port", RequeteConnexion{Port: 5433}, "port"},
		{"autre utilisateur", RequeteConnexion{Utilisateur: "admin"}, "utilisateur"},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			s, _ := serveurDeTest(t)
			profilAvecMotDePasse(t, s)

			requete := c.requete
			requete.Profil = "production"
			_, err := s.connexionDuProfil(&requete)
			if err == nil || !strings.Contains(err.Error(), c.champ) {
				t.Fatalf("erreur %v, attendu un refus qui nomme %s", err, c.champ)
			}
			if codeDe(err) != http.StatusConflict {
				t.Errorf("code %d, attendu 409", codeDe(err))
			}
			if requete.MotDePasse != "" {
				t.Error("le mot de passe enregistré a été injecté")
			}
		})
	}
}

// TestMotDePasseEnvoyeVersLaMemeDestination : la comparaison porte sur ce que
// les deux connexions visent, pas sur le texte. Un hôte en majuscules, un port
// laissé vide qui vaut celui du SGBD, une autre base du même serveur : le
// mot de passe du rôle y vaut toujours.
func TestMotDePasseEnvoyeVersLaMemeDestination(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom     string
		requete RequeteConnexion
	}{
		{"requête vide", RequeteConnexion{}},
		{"hôte en majuscules et blancs", RequeteConnexion{Hote: " BDD.exemple "}},
		{"sgbd déduit du port", RequeteConnexion{Hote: "bdd.exemple", Port: 5432}},
		{"autre base", RequeteConnexion{Base: "paie"}},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			s, _ := serveurDeTest(t)
			profilAvecMotDePasse(t, s)

			requete := c.requete
			requete.Profil = "production"
			if _, err := s.connexionDuProfil(&requete); err != nil {
				t.Fatalf("complétion : %v", err)
			}
			if requete.MotDePasse != "secret" {
				t.Error("le mot de passe enregistré n'a pas été repris")
			}
		})
	}
}

// TestDestinationDivergenteRefuseeParLAPI vérifie la réponse vue par le
// navigateur : un 409 qui dit quoi faire, sans que la base ait été jointe.
func TestDestinationDivergenteRefuseeParLAPI(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	profilAvecMotDePasse(t, s)

	w := poster(s, routeur, "/api/connexion", RequeteConnexion{Profil: "production", Hote: "recette.exemple"})
	attendreStatut(t, w, http.StatusConflict)
	if reponse := decoderReponse[ReponseErreur](t, w); !strings.Contains(reponse.Erreur, "mot de passe") {
		t.Errorf("message %q", reponse.Erreur)
	}
}

// TestProfilCompleteLeModeSSL : un profil qui porte verify-full le rend à la
// connexion, et la chaîne composée le transmet au pilote.
func TestProfilCompleteLeModeSSL(t *testing.T) {
	t.Parallel()

	s, _ := serveurDeTest(t)
	if _, err := s.emplacements.EnregistrerProfil(config.Profil{
		Nom: "production", SGBD: "postgres", Hote: "bdd.exemple", SSLMode: "verify-full",
	}, "", true); err != nil {
		t.Fatalf("enregistrement : %v", err)
	}

	requete := RequeteConnexion{Profil: "production"}
	if _, err := s.connexionDuProfil(&requete); err != nil {
		t.Fatalf("complétion : %v", err)
	}
	dsn, err := requete.composer()
	if err != nil {
		t.Fatalf("composition : %v", err)
	}
	if !strings.Contains(dsn, "sslmode=verify-full") {
		t.Errorf("dsn composé sans sslmode")
	}
}

// TestProfilDepuisUnDSNGardeLeModeSSL : enregistré depuis une chaîne, le profil
// garde son sslmode. C'est le défaut d'origine : il se rouvrait en prefer.
func TestProfilDepuisUnDSNGardeLeModeSSL(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	reponse := decoderReponse[ReponseProfils](t, poster(s, routeur, "/api/profils", RequeteProfil{
		Profil: config.Profil{Nom: "verifie"},
		DSN:    "postgres://lecture@bdd.exemple:5432/gescom?sslmode=verify-full",
	}))
	if len(reponse.Profils) != 1 || reponse.Profils[0].Profil.SSLMode != "verify-full" {
		t.Errorf("profils %+v", reponse.Profils)
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
			if _, err := s.emplacements.EnregistrerProfil(config.Profil{
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
	if _, err := s.emplacements.EnregistrerProfil(config.Profil{Nom: "nas", Hote: "h"},
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

// TestCleAbimeeNeBloquePasLaConnexion couvre une clé tronquée alors qu'un profil
// porte un mot de passe : la connexion se prépare sans lui, comme pour une clé
// absente, au lieu d'un refus qui citerait le chemin de la clé.
func TestCleAbimeeNeBloquePasLaConnexion(t *testing.T) {
	t.Parallel()

	s, _ := serveurDeTest(t)
	if _, err := s.emplacements.EnregistrerProfil(config.Profil{Nom: "nas", Hote: "h"},
		"secret", true); err != nil {
		t.Fatalf("enregistrement : %v", err)
	}
	cle := filepath.Join(s.emplacements.Racine(), "cle.bin")
	if err := os.WriteFile(cle, []byte("abime"), 0o600); err != nil {
		t.Fatalf("clé tronquée : %v", err)
	}

	requete := RequeteConnexion{Profil: "nas"}
	if _, err := s.connexionDuProfil(&requete); err != nil {
		t.Fatalf("connexion refusée : %v", err)
	}
	if requete.Hote != "h" || requete.MotDePasse != "" {
		t.Errorf("requête %+v, attendue l'hôte du profil sans mot de passe", requete)
	}
}

// TestRemplacementDeCleAnnonceALEnregistrement vérifie que l'écran reçoit
// l'avertissement de perte au moment où il enregistre : c'est la seule fois où
// il peut être dit, la clé neuve ne signalant plus rien ensuite.
func TestRemplacementDeCleAnnonceALEnregistrement(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	if _, err := s.emplacements.EnregistrerProfil(config.Profil{Nom: "nas", Hote: "h"},
		"secret", true); err != nil {
		t.Fatalf("enregistrement : %v", err)
	}
	cle := filepath.Join(s.emplacements.Racine(), "cle.bin")
	if err := os.WriteFile(cle, []byte("abime"), 0o600); err != nil {
		t.Fatalf("clé tronquée : %v", err)
	}

	w := profilPoste(s, routeur, config.Profil{Nom: "paie", Hote: "h"}, "autre", true, false)
	if w.Code != http.StatusOK {
		t.Fatalf("code %d : %s", w.Code, w.Body.String())
	}
	var recu ReponseProfils
	if err := json.Unmarshal(w.Body.Bytes(), &recu); err != nil {
		t.Fatalf("réponse illisible : %v", err)
	}
	if !strings.Contains(recu.Avertissement, "perdus") {
		t.Errorf("avertissement %q, attendu la perte des mots de passe", recu.Avertissement)
	}
}

// TestCleAbsenteNeBloquePasLaConnexion couvre la clé effacée alors qu'un profil
// porte un mot de passe : la connexion se prépare sans lui, et c'est la saisie
// qui le fournira.
func TestCleAbsenteNeBloquePasLaConnexion(t *testing.T) {
	t.Parallel()

	s, _ := serveurDeTest(t)
	if _, err := s.emplacements.EnregistrerProfil(config.Profil{Nom: "nas", Hote: "h"},
		"secret", true); err != nil {
		t.Fatalf("enregistrement : %v", err)
	}
	if err := os.Remove(filepath.Join(s.emplacements.Racine(), "cle.bin")); err != nil {
		t.Fatalf("suppression de la clé : %v", err)
	}

	requete := RequeteConnexion{Profil: "nas"}
	if _, err := s.connexionDuProfil(&requete); err != nil {
		t.Fatalf("connexion refusée : %v", err)
	}
	if requete.Hote != "h" || requete.MotDePasse != "" {
		t.Errorf("requête %+v, attendue l'hôte du profil sans mot de passe", requete)
	}
}

// TestPlafondDeProfilsRefuseEn409 vérifie la traduction du plafond : un conflit
// que l'écran affiche, pas une erreur de saisie.
func TestPlafondDeProfilsRefuseEn409(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	for i := range 50 {
		if _, err := s.emplacements.EnregistrerProfil(config.Profil{
			Nom: fmt.Sprintf("base-%02d", i), Hote: "h",
		}, "", true); err != nil {
			t.Fatalf("profil %d : %v", i, err)
		}
	}

	attendreStatut(t, profilPoste(s, routeur, config.Profil{Nom: "de trop", Hote: "h"}, "", false, false),
		http.StatusConflict)
}

// TestProfilAuRepertoireDisparuGardeLeCourant couvre le projet déplacé ou le
// disque amovible absent : le profil sert encore à se connecter, et l'on reste
// dans le répertoire courant plutôt que d'échouer.
func TestProfilAuRepertoireDisparuGardeLeCourant(t *testing.T) {
	t.Parallel()

	s, _ := serveurDeTest(t)
	depart := s.repertoireCourant()
	disparu := filepath.Join(t.TempDir(), "projet-deplace")

	if _, err := s.emplacements.EnregistrerProfil(config.Profil{
		Nom: "ancien", Hote: "h", Repertoire: disparu,
	}, "", true); err != nil {
		t.Fatalf("enregistrement : %v", err)
	}

	requete := RequeteConnexion{Profil: "ancien"}
	if _, err := s.connexionDuProfil(&requete); err != nil {
		t.Fatalf("connexion refusée pour un répertoire disparu : %v", err)
	}
	if s.repertoireCourant() != depart {
		t.Errorf("répertoire courant %s, attendu %s", s.repertoireCourant(), depart)
	}
}

// TestProfilBasculeLeRepertoire vérifie qu'un profil emmène dans son projet, et
// qu'un profil sans répertoire laisse le courant tel quel.
func TestProfilBasculeLeRepertoire(t *testing.T) {
	t.Parallel()

	s, _ := serveurDeTest(t)
	ailleurs := t.TempDir()
	depart := s.repertoireCourant()

	if _, err := s.emplacements.EnregistrerProfil(config.Profil{
		Nom: "avec", Hote: "h", Repertoire: ailleurs,
	}, "", true); err != nil {
		t.Fatalf("enregistrement : %v", err)
	}
	if _, err := s.emplacements.EnregistrerProfil(config.Profil{Nom: "sans", Hote: "h"},
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
