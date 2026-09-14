// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
)

// Répertoire des cas, un sous-répertoire par cas.
const racineReference = "../../tests/reference/inference"

// majAttendus réécrit les calques logiques attendus au lieu de les comparer.
//
// Derrière un drapeau, et jamais automatique : un attendu régénéré sans être
// relu ne teste plus rien, il enregistre le comportement courant — bugs
// compris — et le déclare correct.
var majAttendus = flag.Bool("maj-attendus", false, "réécrit les calques logiques attendus")

// TestReference est le vrai jeu de tests de l'inférence.
//
// L'inférence étant pure, elle se teste sans base de données : un calque
// physique figé en entrée, un calque logique attendu en sortie, comparaison
// octet pour octet. Chaque heuristique ajoutée apporte son cas ici avant son
// code.
//
// La comparaison porte sur les octets sérialisés et non sur les structures Go :
// c'est le fichier que lit le paquet PHP, et c'est donc lui qui est le contrat.
func TestReference(t *testing.T) {
	t.Parallel()

	for _, cas := range casDeReference(t) {
		t.Run(cas, func(t *testing.T) {
			t.Parallel()

			repertoire := filepath.Join(racineReference, cas)

			physique, err := calque.LirePhysique(filepath.Join(repertoire, "physique.json"))
			if err != nil {
				t.Fatalf("lecture du physique : %v", err)
			}

			decisions, err := LireDecisions(cheminDecisions(repertoire))
			if err != nil {
				t.Fatalf("lecture des decisions : %v", err)
			}

			logique, _ := Inferer(physique, decisions)

			obtenu, err := calque.Serialiser(logique)
			if err != nil {
				t.Fatalf("serialisation : %v", err)
			}

			attendu := filepath.Join(repertoire, "logique.json")
			if *majAttendus {
				if err := os.WriteFile(attendu, obtenu, 0o600); err != nil {
					t.Fatalf("ecriture de l'attendu : %v", err)
				}
				t.Logf("attendu reecrit, a relire avant de commiter")
				return
			}

			reference, err := os.ReadFile(attendu)
			if err != nil {
				t.Fatalf("lecture de l'attendu : %v (relancer avec -maj-attendus pour le creer)", err)
			}
			if string(obtenu) != string(reference) {
				t.Errorf("le logique produit differe de l'attendu\n--- attendu ---\n%s\n--- obtenu ---\n%s",
					reference, obtenu)
			}
		})
	}
}

// TestReferenceEstDeterministe vérifie que deux inférences du même physique
// rendent les mêmes octets.
//
// Sans lui, une itération de map non triée passerait inaperçue : les cas de
// référence ne l'attraperaient qu'une exécution sur deux, ce qui se lit comme
// un test instable plutôt que comme le défaut de déterminisme qu'il est.
func TestReferenceEstDeterministe(t *testing.T) {
	t.Parallel()

	for _, cas := range casDeReference(t) {
		t.Run(cas, func(t *testing.T) {
			t.Parallel()

			repertoire := filepath.Join(racineReference, cas)

			physique, err := calque.LirePhysique(filepath.Join(repertoire, "physique.json"))
			if err != nil {
				t.Fatalf("lecture du physique : %v", err)
			}
			decisions, err := LireDecisions(cheminDecisions(repertoire))
			if err != nil {
				t.Fatalf("lecture des decisions : %v", err)
			}

			premier, second := inferer(t, physique, decisions), inferer(t, physique, decisions)
			if premier != second {
				t.Errorf("deux inferences du meme physique different\n--- 1 ---\n%s\n--- 2 ---\n%s",
					premier, second)
			}
		})
	}
}

// TestChaqueCodeDAvertissementAUnCasDeReference exige que chaque code déclaré
// dans internal/calque apparaisse dans au moins un calque logique attendu.
//
// Les codes sont annoncés stables et servent de filtre en CI : un code déclaré
// que rien ne produit promet un signal qui n'arrive jamais (invariant 16). Le
// cas fut `fk_implicite_probable`, déclaré et traduit deux phases avant la
// détection qui l'émettrait. Un test unitaire ne suffit pas : c'est le cas de
// référence qui montre le code dans le contrat.
//
// Les codes se lisent dans le source plutôt que dans une liste recopiée, qui
// oublierait le prochain.
func TestChaqueCodeDAvertissementAUnCasDeReference(t *testing.T) {
	t.Parallel()

	declares := codesDeclares(t, "../calque/logique.go")
	if len(declares) == 0 {
		t.Fatal("aucun code Code* lu dans logique.go : le test ne mesurerait rien")
	}

	produits := map[string]bool{}
	for _, cas := range casDeReference(t) {
		logique, err := calque.LireLogique(filepath.Join(racineReference, cas, "logique.json"))
		if err != nil {
			t.Fatalf("lecture du logique de %s : %v", cas, err)
		}
		for _, a := range logique.Avertissements {
			produits[a.Code] = true
		}
	}

	for _, nom := range clesTriees(declares) {
		if !produits[declares[nom]] {
			t.Errorf("%s (%q) n'apparaît dans aucun cas de %s : ajouter le cas qui le produit, ou retirer le code tant que rien ne l'émet",
				nom, declares[nom], racineReference)
		}
	}
}

// codesDeclares rend, par nom de constante, la valeur des constantes Code* du
// fichier.
func codesDeclares(t *testing.T, chemin string) map[string]string {
	t.Helper()

	fichier, err := parser.ParseFile(token.NewFileSet(), chemin, nil, 0)
	if err != nil {
		t.Fatalf("lecture de %s : %v", chemin, err)
	}

	codes := map[string]string{}
	for _, decl := range fichier.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			valeurs := spec.(*ast.ValueSpec)
			for i, nom := range valeurs.Names {
				if !strings.HasPrefix(nom.Name, "Code") || i >= len(valeurs.Values) {
					continue
				}
				litteral, ok := valeurs.Values[i].(*ast.BasicLit)
				if !ok || litteral.Kind != token.STRING {
					t.Fatalf("%s n'est pas une chaîne littérale : le test ne sait pas le lire", nom.Name)
				}
				valeur, err := strconv.Unquote(litteral.Value)
				if err != nil {
					t.Fatalf("%s : %v", nom.Name, err)
				}
				codes[nom.Name] = valeur
			}
		}
	}
	return codes
}

// inferer rend le calque logique sérialisé, pour comparer deux exécutions.
func inferer(t *testing.T, p *calque.Physique, d *Decisions) string {
	t.Helper()

	logique, _ := Inferer(p, d)
	octets, err := calque.Serialiser(logique)
	if err != nil {
		t.Fatalf("serialisation : %v", err)
	}
	return string(octets)
}

// casDeReference rend les noms des sous-répertoires de cas, triés.
//
// Un répertoire sans physique.json est ignoré plutôt que signalé : c'est le
// répertoire en cours de création, pas un cas cassé.
func casDeReference(t *testing.T) []string {
	t.Helper()

	entrees, err := os.ReadDir(racineReference)
	if err != nil {
		t.Fatalf("lecture des cas de reference : %v", err)
	}

	var cas []string
	for _, e := range entrees {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(racineReference, e.Name(), "physique.json")); err != nil {
			continue
		}
		cas = append(cas, e.Name())
	}

	if len(cas) == 0 {
		t.Fatal("aucun cas de reference : l'inference ne serait testee par rien")
	}
	return cas
}

// cheminDecisions rend le fichier de décisions du cas, ou la chaîne vide quand
// il n'y en a pas — LireDecisions traite ce cas comme le premier passage.
func cheminDecisions(repertoire string) string {
	chemin := filepath.Join(repertoire, "decisions.yaml")
	if _, err := os.Stat(chemin); err != nil {
		return ""
	}
	return chemin
}
