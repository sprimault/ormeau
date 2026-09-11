// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"net/http"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/sprimault/ormeau/internal/config"
)

// pageServie rend le HTML de la page d'entrée.
func pageServie(t *testing.T, s *serveur, routeur http.Handler) string {
	t.Helper()

	w := lire(s, routeur, "/")
	if w.Code != http.StatusOK {
		t.Fatalf("code %d, attendu %d", w.Code, http.StatusOK)
	}
	return w.Body.String()
}

// TestPageInjecteLesPreferences vérifie que thème, langue et tailles arrivent
// dans la page elle-même.
//
// C'est tout l'objet du lot : le port étant tiré à chaque lancement, le
// navigateur ne peut rien retenir, et une valeur qui arriverait par un appel
// d'API arriverait après le premier rendu, donc après la bascule qu'on veut
// supprimer.
func TestPageInjecteLesPreferences(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	ouvert := false
	if err := s.preferences.ecrire(config.Preferences{
		Theme:         "sombre",
		Langue:        "en",
		ApercuOuvert:  &ouvert,
		LargeurArbre:  320,
		HauteurApercu: 240,
	}); err != nil {
		t.Fatalf("écriture : %v", err)
	}

	page := pageServie(t, s, routeur)
	for _, attendu := range []string{
		`lang="en"`,
		`data-theme="sombre"`,
		`data-apercu-ouvert="0"`,
		"--ormeau-largeur-arbre:320px",
		"--ormeau-hauteur-apercu:240px",
	} {
		if !strings.Contains(page, attendu) {
			t.Errorf("la page ne porte pas %s", attendu)
		}
	}
	// Jamais réglée : sa variable n'est pas écrite et le composant garde son
	// défaut, plutôt que de recevoir un zéro qui écraserait la mise en page.
	if strings.Contains(page, "--ormeau-largeur-entites") {
		t.Error("une taille non réglée écrit quand même sa variable")
	}
}

// TestPageSansPreferencesPorteLesDefauts couvre le premier lancement.
func TestPageSansPreferencesPorteLesDefauts(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	page := pageServie(t, s, routeur)
	if !strings.Contains(page, `lang="fr"`) || !strings.Contains(page, `data-theme="systeme"`) {
		t.Error("la page ne porte pas les valeurs par défaut")
	}
	if strings.Contains(page, "data-apercu-ouvert") {
		t.Error("un aperçu jamais réglé écrit quand même son attribut")
	}
}

// TestPreferencesLuesEtEcrites vérifie l'aller-retour par l'API, qui sert aux
// relectures et aux réglages, jamais au premier rendu.
func TestPreferencesLuesEtEcrites(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	initiales := decoderReponse[config.Preferences](t, lire(s, routeur, "/api/preferences"))
	if initiales.Theme != "systeme" || initiales.Langue != "fr" {
		t.Errorf("défauts %+v", initiales)
	}

	w := poster(s, routeur, "/api/preferences", config.Preferences{Theme: "clair", Langue: "en", LargeurArbre: 300})
	attendreStatut(t, w, http.StatusOK)

	relues := decoderReponse[config.Preferences](t, lire(s, routeur, "/api/preferences"))
	if relues.Theme != "clair" || relues.Langue != "en" || relues.LargeurArbre != 300 {
		t.Errorf("préférences relues %+v", relues)
	}
}

// TestPreferencesRefuseUneValeurInconnue vérifie que la validation du paquet
// config s'applique aussi à ce qui vient du navigateur.
func TestPreferencesRefuseUneValeurInconnue(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	w := poster(s, routeur, "/api/preferences", config.Preferences{Theme: "fluo", Langue: "fr"})
	attendreStatut(t, w, http.StatusUnprocessableEntity)

	inchangees := decoderReponse[config.Preferences](t, lire(s, routeur, "/api/preferences"))
	if inchangees.Theme != "systeme" {
		t.Errorf("un refus a quand même changé l'état : %+v", inchangees)
	}
}

// TestIndexSansAncreRefuseDeDemarrer couvre le jour où la page produite par
// Vite ne portera plus le point d'insertion.
//
// Un démarrage qui échoue en le nommant, plutôt qu'une page servie telle
// quelle : le thème ne tiendrait plus, et rien ne rattacherait le symptôme à ce
// changement.
func TestIndexSansAncreRefuseDeDemarrer(t *testing.T) {
	t.Parallel()

	cas := map[string]string{
		"balise html modifiée": `<!doctype html><html lang="en"></head>`,
		"tête absente":         `<!doctype html><html lang="fr">`,
	}
	for nom, page := range cas {
		t.Run(nom, func(t *testing.T) {
			t.Parallel()

			_, err := indexTemplate(fstest.MapFS{"index.html": {Data: []byte(page)}})
			if err == nil {
				t.Fatal("index sans ancre accepté")
			}
			if !strings.Contains(err.Error(), "ancre") {
				t.Errorf("l'erreur ne dit pas ce qui manque : %v", err)
			}
		})
	}
}
