// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
)

// calqueDEssai écrit un calque physique minimal dans un répertoire temporaire
// et rend son chemin.
func calqueDEssai(t *testing.T, nom string) string {
	t.Helper()

	physique := &calque.Physique{
		VersionRI: calque.VersionCourante,
		Source: calque.Source{
			SGBD: "postgres", Version: "16.2",
			Catalogue: "gescom", Schema: "public",
		},
		Tables: []calque.Table{{
			Nom: "t_client", Schema: "public",
			Colonnes: []calque.Colonne{{
				Nom: "id", Position: 1, TypeBrut: "integer",
				TypeNormalise: calque.TypeEntier, AutoIncrement: true,
			}},
			ClePrimaire: &calque.ClePrimaire{Nom: "pk", Colonnes: []string{"id"}},
		}},
	}

	chemin := filepath.Join(t.TempDir(), nom)
	if err := physique.Ecrire(chemin); err != nil {
		t.Fatalf("ecriture du calque d'essai : %v", err)
	}
	return chemin
}

// TestInfererDeduitLesTroisChemins vérifie qu'un seul argument suffit.
//
// Les trois fichiers d'une base partagent leur préfixe. Devoir taper les deux
// autres à chaque fois serait une occasion de les faire diverger.
func TestInfererDeduitLesTroisChemins(t *testing.T) {
	t.Parallel()

	chemin := calqueDEssai(t, "gescom.calque.json")
	repertoire := filepath.Dir(chemin)

	if err := inferer([]string{chemin}); err != nil {
		t.Fatalf("inferer : %v", err)
	}

	for _, attendu := range []string{"gescom.logique.json", "gescom.decisions.yaml"} {
		if _, err := os.Stat(filepath.Join(repertoire, attendu)); err != nil {
			t.Errorf("%s non produit : %v", attendu, err)
		}
	}
}

// TestInfererNEcrasePasLesDecisions est le test qui protège le travail de
// l'utilisateur.
//
// Le fichier porte des arbitrages relus un par un. Les écraser parce qu'une
// commande a été relancée est la faute que l'outil existe pour éviter.
func TestInfererNEcrasePasLesDecisions(t *testing.T) {
	t.Parallel()

	chemin := calqueDEssai(t, "gescom.calque.json")
	cheminDecisions := filepath.Join(filepath.Dir(chemin), "gescom.decisions.yaml")

	arbitrage := "renommages:\n  public.t_client: Acheteur\n"
	if err := os.WriteFile(cheminDecisions, []byte(arbitrage), 0o600); err != nil {
		t.Fatalf("ecriture des decisions : %v", err)
	}

	if err := inferer([]string{chemin}); err != nil {
		t.Fatalf("inferer : %v", err)
	}

	apres, err := os.ReadFile(cheminDecisions)
	if err != nil {
		t.Fatalf("relecture : %v", err)
	}
	if string(apres) != arbitrage {
		t.Errorf("le fichier de decisions a ete reecrit :\n%s", apres)
	}

	// Et l'arbitrage a bien été appliqué, sinon le fichier serait intact pour
	// la mauvaise raison.
	logique, err := calque.LireLogique(filepath.Join(filepath.Dir(chemin), "gescom.logique.json"))
	if err != nil {
		t.Fatalf("lecture du logique : %v", err)
	}
	if len(logique.Entites) != 1 || logique.Entites[0].Nom != "Acheteur" {
		t.Errorf("entite %+v, attendue Acheteur", logique.Entites)
	}
}

// TestInfererEcritLesDecisionsPreremplies vérifie le premier passage.
//
// Le fichier écrit ne doit rien décider : c'est ce qui permet de le produire
// sans risque, et de le relire ensuite à tête reposée.
func TestInfererEcritLesDecisionsPreremplies(t *testing.T) {
	t.Parallel()

	chemin := calqueDEssai(t, "gescom.calque.json")

	if err := inferer([]string{chemin}); err != nil {
		t.Fatalf("inferer : %v", err)
	}

	contenu, err := os.ReadFile(filepath.Join(filepath.Dir(chemin), "gescom.decisions.yaml"))
	if err != nil {
		t.Fatalf("lecture : %v", err)
	}

	for numero, ligne := range strings.Split(string(contenu), "\n") {
		if nette := strings.TrimSpace(ligne); nette != "" && !strings.HasPrefix(nette, "#") {
			t.Errorf("ligne %d active dans le fichier prerempli : %q", numero+1, ligne)
		}
	}
}

// TestInfererRefuseUnCheminDeDecisionsAbsent vérifie qu'un chemin explicite
// introuvable est une erreur.
//
// Une faute de frappe bien plus souvent qu'une intention : en créer un
// silencieusement laisserait croire que les arbitrages ont été pris en compte.
func TestInfererRefuseUnCheminDeDecisionsAbsent(t *testing.T) {
	t.Parallel()

	chemin := calqueDEssai(t, "gescom.calque.json")
	absent := filepath.Join(filepath.Dir(chemin), "jamais-ecrit.yaml")

	err := inferer([]string{chemin, "--decisions", absent})
	if err == nil {
		t.Fatal("aucune erreur sur un fichier de decisions introuvable")
	}
	if !strings.Contains(err.Error(), "introuvable") {
		t.Errorf("message %q, attendu explicite", err)
	}
	if _, err := os.Stat(absent); err == nil {
		t.Error("le fichier a ete cree alors qu'il etait designe explicitement")
	}
}

// TestInfererAccepteLesDrapeauxApresLArgument couvre l'ordre que montre la
// ligne d'usage.
//
// flag s'arrête au premier argument positionnel : sans réordonnancement, la
// forme naturelle verrait --decisions traité comme un second fichier.
func TestInfererAccepteLesDrapeauxApresLArgument(t *testing.T) {
	t.Parallel()

	chemin := calqueDEssai(t, "gescom.calque.json")
	sortie := filepath.Join(filepath.Dir(chemin), "ailleurs.json")

	if err := inferer([]string{chemin, "--sortie", sortie}); err != nil {
		t.Fatalf("inferer : %v", err)
	}
	if _, err := os.Stat(sortie); err != nil {
		t.Errorf("--sortie apres l'argument non pris en compte : %v", err)
	}
}

// TestInfererRefuseAutreQuUnArgument vérifie les deux cas dégénérés.
func TestInfererRefuseAutreQuUnArgument(t *testing.T) {
	t.Parallel()

	chemin := calqueDEssai(t, "gescom.calque.json")

	for _, args := range [][]string{{}, {chemin, chemin}} {
		if err := inferer(args); err == nil {
			t.Errorf("aucune erreur sur %d argument(s)", len(args))
		}
	}
}

// TestBaseDuChemin couvre la dérivation du préfixe commun.
func TestBaseDuChemin(t *testing.T) {
	t.Parallel()

	// Le séparateur est laissé tel que l'utilisateur l'a tapé : la dérivation
	// ne fait que retirer un suffixe, elle ne normalise pas le chemin.
	for entree, attendu := range map[string]string{
		"gescom.calque.json":         "gescom",
		"calques/gescom.calque.json": "calques/gescom",
		"gescom.json":                "gescom",
		"gescom":                     "gescom",
	} {
		if obtenu := baseDuChemin(entree); obtenu != attendu {
			t.Errorf("baseDuChemin(%q) = %q, attendu %q", entree, obtenu, attendu)
		}
	}
}

// TestCodesTries vérifie l'ordre du décompte final.
//
// Du plus fréquent au moins fréquent, alphabétique à égalité : le décompte se
// lit dans un journal de CI, où une ligne qui bouge sans raison se remarque.
func TestCodesTries(t *testing.T) {
	t.Parallel()

	avertissements := []calque.Avertissement{
		{Code: "b"}, {Code: "a"}, {Code: "a"}, {Code: "c"}, {Code: "a"}, {Code: "b"},
	}

	codes := codesTries(avertissements)
	attendus := []decompte{{"a", 3}, {"b", 2}, {"c", 1}}

	for i, attendu := range attendus {
		if codes[i] != attendu {
			t.Errorf("rang %d : %+v, attendu %+v", i, codes[i], attendu)
		}
	}
}

// TestAnalyserMelangeDrapeauxEtArguments couvre le réordonnancement lui-même.
func TestAnalyserMelangeDrapeauxEtArguments(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom         string
		args        []string
		positionnel string
		valeur      string
	}{
		{"drapeau avant", []string{"--x", "v", "fichier"}, "fichier", "v"},
		{"drapeau apres", []string{"fichier", "--x", "v"}, "fichier", "v"},
		{"sans drapeau", []string{"fichier"}, "fichier", ""},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			jeu := flag.NewFlagSet("essai", flag.ContinueOnError)
			x := jeu.String("x", "", "")

			positionnels, err := analyser(jeu, c.args)
			if err != nil {
				t.Fatalf("analyser : %v", err)
			}
			if len(positionnels) != 1 || positionnels[0] != c.positionnel {
				t.Errorf("positionnels %v, attendu [%s]", positionnels, c.positionnel)
			}
			if *x != c.valeur {
				t.Errorf("--x = %q, attendu %q", *x, c.valeur)
			}
		})
	}
}
