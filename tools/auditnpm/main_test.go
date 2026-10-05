// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// aucuneExclusion sert les cas où le contrôle doit refuser tout ce qu'il
// trouve.
var aucuneExclusion = map[string]Exclusion{}

// TestControler exerce les deux refus et les cas qui passent, sur des rapports
// figés. Les rapports inventés sont réduits aux champs que le contrôle lit ;
// rapport-reel.json est la sortie entière de `npm audit --json` sur le lock du
// dépôt, pour qu'un cas au moins ne teste pas la fidélité d'un faux.
func TestControler(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom      string
		fichier  string
		nommees  map[string]Exclusion
		echoue   bool
		fragment string
	}{
		{
			nom:      "le rapport du dépôt passe avec ses exclusions",
			fichier:  "rapport-reel.json",
			nommees:  exclusions,
			fragment: "GHSA-vfj7-8cjw-p6xm",
		},
		{
			nom:      "sans exclusion, le même rapport est refusé",
			fichier:  "rapport-reel.json",
			nommees:  aucuneExclusion,
			echoue:   true,
			fragment: "GHSA-vfj7-8cjw-p6xm",
		},
		{
			nom:      "une exclusion que plus aucun avis ne rencontre est refusée",
			fichier:  "sans-avis.json",
			nommees:  exclusions,
			echoue:   true,
			fragment: "exclusion sans avis correspondant",
		},
		{
			nom:      "un avis que la liste ne nomme pas est refusé",
			fichier:  "avis-non-exclu.json",
			nommees:  exclusions,
			echoue:   true,
			fragment: "GHSA-0000-0000-0000",
		},
		{
			nom:     "une gravité moyenne ne fait rien échouer",
			fichier: "avis-moderate.json",
			nommees: aucuneExclusion,
		},
		{
			nom:      "une version de rapport inconnue est refusée",
			fichier:  "version-inconnue.json",
			nommees:  aucuneExclusion,
			echoue:   true,
			fragment: "auditReportVersion 3",
		},
		{
			nom:      "un rapport tronqué est refusé au lieu de passer pour vide",
			fichier:  "tronque.json",
			nommees:  aucuneExclusion,
			echoue:   true,
			fragment: "rapport illisible",
		},
		{
			// Le cas insidieux : un document valide dont tous les champs
			// restent à zéro se lirait comme « aucun avis ».
			nom:      "un rapport vide est refusé plutôt que lu comme sans avis",
			fichier:  "rapport-vide.json",
			nommees:  aucuneExclusion,
			echoue:   true,
			fragment: "auditReportVersion 0",
		},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			var sortie bytes.Buffer
			err := controler(ouvrir(t, c.fichier), &sortie, c.nommees)

			if c.echoue && err == nil {
				t.Fatalf("aucune erreur, sortie %q", sortie.String())
			}
			if !c.echoue && err != nil {
				t.Fatalf("erreur inattendue : %v", err)
			}

			texte := sortie.String()
			if err != nil {
				texte = err.Error()
			}
			if !strings.Contains(texte, c.fragment) {
				t.Errorf("le texte ne cite pas %q :\n%s", c.fragment, texte)
			}
		})
	}
}

// TestControlerNeCompteQueLesAvisPropres vérifie qu'un paquet qui ne fait que
// propager l'avis d'une dépendance n'est pas signalé pour lui-même. Le rapport
// du dépôt porte cinq paquets « high » pour un seul avis : un contrôle indexé
// par paquet en signalerait cinq, dont quatre qu'aucune exclusion ne pourrait
// nommer puisqu'ils n'ont pas d'identifiant.
func TestControlerNeCompteQueLesAvisPropres(t *testing.T) {
	t.Parallel()

	var sortie bytes.Buffer
	err := controler(ouvrir(t, "rapport-reel.json"), &sortie, aucuneExclusion)
	if err == nil {
		t.Fatal("le rapport du dépôt devait être refusé sans exclusion")
	}

	lignes := strings.Count(err.Error(), "\n")
	if lignes != 1 {
		t.Errorf("%d avis signalés, 1 attendu :\n%s", lignes, err)
	}
}

// TestChaqueExclusionEstMotivee refuse une exclusion sans raison ni condition
// de retrait : une ligne muette se reconduit indéfiniment, et c'est précisément
// ce que la liste doit empêcher.
func TestChaqueExclusionEstMotivee(t *testing.T) {
	t.Parallel()

	for id, e := range exclusions {
		if strings.TrimSpace(e.Raison) == "" {
			t.Errorf("%s : raison vide", id)
		}
		if strings.TrimSpace(e.Retrait) == "" {
			t.Errorf("%s : condition de retrait vide", id)
		}
	}
}

// ouvrir rend le rapport figé du nom donné.
func ouvrir(t *testing.T, nom string) *bytes.Reader {
	t.Helper()

	contenu, err := os.ReadFile(filepath.Join("testdata", nom))
	if err != nil {
		t.Fatalf("lecture du cas %s : %v", nom, err)
	}
	return bytes.NewReader(contenu)
}
