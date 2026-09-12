// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/sprimault/ormeau/internal/introspection"
)

// TestRepertoireChangeEtRendLeCheminResolu vérifie le cas nominal, et que la
// réponse porte le chemin résolu.
func TestRepertoireChangeEtRendLeCheminResolu(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	ailleurs := t.TempDir()

	reponse := decoderReponse[ReponseContexte](t,
		poster(s, routeur, "/api/repertoire", RequeteRepertoire{Repertoire: ailleurs}))

	if reponse.Repertoire != cheminResolu(t, ailleurs) {
		t.Errorf("répertoire rendu %s, attendu %s", reponse.Repertoire, cheminResolu(t, ailleurs))
	}
	if s.repertoireCourant() != cheminResolu(t, ailleurs) {
		t.Errorf("répertoire du serveur %s", s.repertoireCourant())
	}

	// Le contexte que l'écran relit doit dire la même chose.
	contexte := decoderReponse[ReponseContexte](t, lire(s, routeur, "/api/contexte"))
	if contexte.Repertoire != s.repertoireCourant() {
		t.Errorf("contexte %s, répertoire %s", contexte.Repertoire, s.repertoireCourant())
	}
}

// TestRepertoireRendLeCheminResoluEtNonLeSaisi couvre le « .. » qui aboutit
// ailleurs que là où on croit.
//
// L'écran affiche ce que la réponse porte : montrer le chemin saisi laisserait
// croire qu'on écrit dans un répertoire qui n'est pas celui qui recevra les
// fichiers.
func TestRepertoireRendLeCheminResoluEtNonLeSaisi(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	racine := t.TempDir()
	if err := os.Mkdir(filepath.Join(racine, "projet"), 0o755); err != nil {
		t.Fatalf("préparation : %v", err)
	}
	// Assemblé à la main : filepath.Join nettoie, et rendrait le détour
	// invisible avant même qu'il n'atteigne le serveur.
	separateur := string(filepath.Separator)
	detourne := racine + separateur + "projet" + separateur + ".." + separateur + "projet"

	reponse := decoderReponse[ReponseContexte](t,
		poster(s, routeur, "/api/repertoire", RequeteRepertoire{Repertoire: detourne}))

	if reponse.Repertoire == detourne {
		t.Errorf("le chemin saisi est rendu tel quel : %s", reponse.Repertoire)
	}
	if reponse.Repertoire != cheminResolu(t, filepath.Join(racine, "projet")) {
		t.Errorf("chemin résolu %s", reponse.Repertoire)
	}
}

// TestRepertoireSuitLesLiens vérifie qu'un lien symbolique est résolu avant
// d'être retenu : contrôler les droits d'un chemin et écrire dans un autre
// n'aurait aucun sens.
func TestRepertoireSuitLesLiens(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("la création d'un lien exige des droits particuliers sous Windows")
	}

	s, routeur := serveurDeTest(t)
	racine := t.TempDir()
	cible := filepath.Join(racine, "reel")
	lien := filepath.Join(racine, "raccourci")
	if err := os.Mkdir(cible, 0o755); err != nil {
		t.Fatalf("préparation : %v", err)
	}
	if err := os.Symlink(cible, lien); err != nil {
		t.Fatalf("lien : %v", err)
	}

	reponse := decoderReponse[ReponseContexte](t,
		poster(s, routeur, "/api/repertoire", RequeteRepertoire{Repertoire: lien}))

	if reponse.Repertoire != cheminResolu(t, cible) {
		t.Errorf("lien non résolu : %s", reponse.Repertoire)
	}
}

// TestRepertoireRefuseCeQuiNEstPasUtilisable couvre ce que l'écran peut
// envoyer de travers.
func TestRepertoireRefuseCeQuiNEstPasUtilisable(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	depart := s.repertoireCourant()

	fichier := filepath.Join(t.TempDir(), "gescom.calque.json")
	if err := os.WriteFile(fichier, []byte("{}"), 0o600); err != nil {
		t.Fatalf("préparation : %v", err)
	}

	cas := map[string]string{
		"vide":               "",
		"relatif":            "projets/gescom",
		"absent":             filepath.Join(t.TempDir(), "jamais-cree"),
		"fichier et non dir": fichier,
	}
	for nom, chemin := range cas {
		t.Run(nom, func(t *testing.T) {
			t.Parallel()

			attendreStatut(t, poster(s, routeur, "/api/repertoire",
				RequeteRepertoire{Repertoire: chemin}), http.StatusUnprocessableEntity)
		})
	}

	if s.repertoireCourant() != depart {
		t.Errorf("un refus a changé le répertoire : %s", s.repertoireCourant())
	}
	// Rien n'est créé : une faute de frappe ne doit pas semer un dossier vide
	// que personne ne saura rattacher à quoi que ce soit.
	if _, err := os.Stat(cas["absent"]); !os.IsNotExist(err) {
		t.Error("un répertoire absent a été créé")
	}
}

// TestRepertoireRefuseUnDossierSansEcriture couvre le seul contrôle des droits
// effectifs : sans lui, l'écran accepterait un répertoire en lecture seule et
// l'échec ne se verrait qu'à l'écriture du calque, en fin d'extraction.
func TestRepertoireRefuseUnDossierSansEcriture(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("un chmod 0500 ne ferme pas l'écriture sous Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root écrit dans un répertoire en lecture seule")
	}

	s, routeur := serveurDeTest(t)
	depart := s.repertoireCourant()
	ferme := t.TempDir()
	if err := os.Chmod(ferme, 0o500); err != nil {
		t.Fatalf("fermeture : %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(ferme, 0o700) })

	w := poster(s, routeur, "/api/repertoire", RequeteRepertoire{Repertoire: ferme})
	attendreStatut(t, w, http.StatusUnprocessableEntity)
	if !strings.Contains(w.Body.String(), "écriture") {
		t.Errorf("le refus ne dit pas pourquoi : %s", w.Body.String())
	}
	if s.repertoireCourant() != depart {
		t.Errorf("un répertoire en lecture seule a été retenu : %s", s.repertoireCourant())
	}
}

// TestExtractionGardeSonRepertoire vérifie qu'une tâche écrit là où elle a été
// lancée, même si l'écran change de répertoire ensuite.
func TestExtractionGardeSonRepertoire(t *testing.T) {
	t.Parallel()

	e := nouvellesExtractions(t.Context())
	fabrique := &fabriqueDeTest{}
	e.ouvrir = fabrique.ouvrir
	t.Cleanup(e.attendre)

	lancement := t.TempDir()
	lancee, err := e.lancer("gescom", dsnSur("gescom"), "postgres", lancement, introspection.Portee{})
	if err != nil {
		t.Fatalf("lancer : %v", err)
	}
	attendreEtat(t, e, lancee.ID, EtatTerminee)

	if _, err := os.Stat(filepath.Join(lancement, "gescom.calque.json")); err != nil {
		t.Errorf("le calque n'est pas dans le répertoire du lancement : %v", err)
	}
}

// cheminResolu rend ce que le serveur retiendra d'un chemin.
//
// Sur macOS, t.TempDir rend un chemin sous /var, lui-même un lien vers
// /private/var : comparer au chemin brut ferait échouer le test là-bas et
// nulle part ailleurs.
func cheminResolu(t *testing.T, chemin string) string {
	t.Helper()

	resolu, err := filepath.EvalSymlinks(filepath.Clean(chemin))
	if err != nil {
		t.Fatalf("résolution de %s : %v", chemin, err)
	}
	absolu, err := filepath.Abs(resolu)
	if err != nil {
		t.Fatalf("chemin absolu de %s : %v", resolu, err)
	}
	return absolu
}

// TestRepertoireRefuseLesAutresMethodes garde le point d'entrée sur le seul
// verbe qui le concerne.
func TestRepertoireRefuseLesAutresMethodes(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	w := lire(s, routeur, "/api/repertoire")
	attendreStatut(t, w, http.StatusMethodNotAllowed)
	if !strings.Contains(w.Body.String(), "méthode") {
		t.Errorf("refus muet : %s", w.Body.String())
	}
}
