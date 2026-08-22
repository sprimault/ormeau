// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"strings"
	"testing"
)

// La validation des drapeaux vit ici : aucun paquet interne ne lit flag.

// Le refus doit tomber avant toute tentative de connexion.
func TestExtraireExigeDSNEtSortie(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom      string
		args     []string
		manquant bool
	}{
		{"sans rien", nil, true},
		{"sans sortie", []string{"--dsn", "postgres://hote/base"}, true},
		{"sans dsn", []string{"--sortie", "gescom.calque.json"}, true},
		{
			"complet",
			[]string{"--dsn", "postgres://hote/base", "--sortie", "gescom.calque.json"},
			false,
		},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			err := extraire(c.args)
			if err == nil {
				t.Fatal("aucune erreur : la commande n'est pourtant pas écrite")
			}

			exigeUnDrapeau := strings.Contains(err.Error(), "--dsn") ||
				strings.Contains(err.Error(), "--sortie")
			if exigeUnDrapeau != c.manquant {
				t.Errorf("erreur inattendue pour ce jeu de drapeaux : %v", err)
			}
		})
	}
}

// Y compris dans l'erreur d'une commande non écrite.
func TestExtraireNeDivulguePasLeDSN(t *testing.T) {
	t.Parallel()

	const dsn = "postgres://utilisateur:motdepasse-secret@hote:5432/base"

	err := extraire([]string{"--dsn", dsn, "--sortie", "sortie.json"})
	if err == nil {
		t.Fatal("aucune erreur")
	}
	if strings.Contains(err.Error(), "motdepasse-secret") {
		t.Errorf("le mot de passe apparaît dans l'erreur : %v", err)
	}
}

// Exactement un calque. Zéro ou deux est une erreur d'invocation.
func TestInfererExigeUnCalquePhysique(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom            string
		args           []string
		attendArgument bool
	}{
		{"sans argument", nil, true},
		{"deux calques", []string{"un.calque.json", "deux.calque.json"}, true},
		{"un calque", []string{"gescom.calque.json"}, false},
		{
			"un calque et des décisions",
			[]string{"--decisions", "decisions.yaml", "gescom.calque.json"},
			false,
		},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			err := inferer(c.args)
			if err == nil {
				t.Fatal("aucune erreur : la commande n'est pourtant pas écrite")
			}

			exigeArgument := strings.Contains(err.Error(), "calque physique est attendu")
			if exigeArgument != c.attendArgument {
				t.Errorf("erreur inattendue pour ce jeu d'arguments : %v", err)
			}
		})
	}
}

// Nommer la phase distingue « pas encore fait » de « cassé ».
func TestCommandesNonEcritesNommentLeurPhase(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom     string
		appeler func() error
	}{
		{"extraire", func() error {
			return extraire([]string{"--dsn", "postgres://hote/base", "--sortie", "s.json"})
		}},
		{"inferer", func() error { return inferer([]string{"gescom.calque.json"}) }},
		{"diff", func() error { return diffuser(nil) }},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			err := c.appeler()
			if err == nil {
				t.Fatal("aucune erreur")
			}
			if !strings.Contains(err.Error(), "feuille de route") {
				t.Errorf("l'erreur ne renvoie pas à la feuille de route : %v", err)
			}
		})
	}
}

// Une commande absente de l'usage est une commande que personne ne trouvera.
func TestUsageDecritLesCommandes(t *testing.T) {
	t.Parallel()

	for _, commande := range []string{"extraire", "inferer", "diff"} {
		if !strings.Contains(usage, commande) {
			t.Errorf("la commande %q est absente de l'usage", commande)
		}
	}
	if !strings.Contains(usage, "--echantillonner") && strings.Contains(usage, "--dsn") {
		t.Log("l'usage ne mentionne pas --echantillonner ; à revoir quand la phase 7 arrivera")
	}
}
