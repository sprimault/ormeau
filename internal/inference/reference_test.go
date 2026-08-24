// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"flag"
	"os"
	"path/filepath"
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
