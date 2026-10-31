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

// exclusionBracesFigee reprend le rôle que tenait l'unique entrée réelle de la
// liste jusqu'au retrait de son avis en amont. Les trois cas qui exercent les
// deux mécanismes d'exclusion — une exclusion qui couvre un avis, une exclusion
// que plus aucun avis ne rencontre — ont besoin d'une entrée qui rencontre
// quelque chose dans rapport-reel.json ; la liste de production est vide, et
// c'est son état normal.
var exclusionBracesFigee = map[string]Exclusion{
	"GHSA-vfj7-8cjw-p6xm": {
		Raison:  "rapport figé : l'avis a été retiré en amont le 2026-10-23",
		Retrait: "avec rapport-reel.json, si un jour il cesse de porter cet avis",
	},
}

// TestControler exerce les deux refus et les cas qui passent, sur des rapports
// figés. Les rapports inventés sont réduits aux champs que le contrôle lit ;
// rapport-reel.json est la sortie entière de `npm audit --json` sur un état du
// lock du dépôt, pour qu'un cas au moins ne teste pas la fidélité d'un faux.
//
// Cet état est celui qui portait encore un avis « high » et sa chaîne de
// propagation. Il est conservé tel quel : régénéré depuis un lock sans avis, le
// fichier ne porterait plus rien à exclure ni à propager, et les trois cas qui
// s'appuient dessus ne prouveraient plus rien. Que la liste de production, elle,
// soit à jour est contrôlé en CI, où une exclusion orpheline fait échouer
// l'audit.
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
			nom:      "un rapport passe quand une exclusion couvre son avis",
			fichier:  "rapport-reel.json",
			nommees:  exclusionBracesFigee,
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
			nommees:  exclusionBracesFigee,
			echoue:   true,
			fragment: "exclusion sans avis correspondant",
		},
		{
			nom:      "un avis que la liste ne nomme pas est refusé",
			fichier:  "avis-non-exclu.json",
			nommees:  exclusionBracesFigee,
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
// figé porte cinq paquets « high » pour un seul avis : un contrôle indexé par
// paquet en signalerait cinq, dont quatre qu'aucune exclusion ne pourrait
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
//
// Il ne boucle sur rien tant que la liste est vide, ce qui est son état normal :
// c'est TestMotivee qui tient la règle, et celui-ci qui l'applique aux entrées
// du jour.
func TestChaqueExclusionEstMotivee(t *testing.T) {
	t.Parallel()

	for id, e := range exclusions {
		if !e.motivee() {
			t.Errorf("%s : raison ou condition de retrait vide", id)
		}
	}
}

// TestMotivee éprouve la règle elle-même, que la liste de production ne
// déclenche plus depuis qu'elle est vide. Sans ce test, rien ne dirait qu'une
// ligne muette serait refusée le jour où quelqu'un en ajoute une.
func TestMotivee(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom     string
		e       Exclusion
		motivee bool
	}{
		{
			nom:     "les deux champs renseignés",
			e:       Exclusion{Raison: "exposition nulle", Retrait: "à la publication d'un correctif"},
			motivee: true,
		},
		{
			nom: "sans raison",
			e:   Exclusion{Retrait: "à la publication d'un correctif"},
		},
		{
			nom: "sans condition de retrait",
			e:   Exclusion{Raison: "exposition nulle"},
		},
		{
			// Des champs blancs passeraient une comparaison à la chaîne vide.
			nom: "des champs qui ne portent que des blancs",
			e:   Exclusion{Raison: "  ", Retrait: "\t\n"},
		},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			if got := c.e.motivee(); got != c.motivee {
				t.Errorf("motivee() = %v, attendu %v", got, c.motivee)
			}
		})
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
