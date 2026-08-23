// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"strings"
	"testing"
)

// La validation des drapeaux vit ici : aucun paquet interne ne lit flag.

// Le refus doit tomber avant toute tentative de connexion : ces cas ne
// joignent aucun serveur, et un DSN de préfixe inconnu s'arrête à la
// résolution du pilote.
func TestExtraireRefuseAvantDeSeConnecter(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom     string
		args    []string
		attendu string
	}{
		{"sans rien", nil, "--dsn"},
		{"sans sortie", []string{"--dsn", "postgres://hote/base"}, "--sortie"},
		{"sans dsn", []string{"--sortie", "gescom.calque.json"}, "--dsn"},
		{
			"sgbd inconnu",
			[]string{"--dsn", "db2://hote/base", "--sortie", "gescom.calque.json"},
			"prefixe",
		},
		{
			"dsn sans prefixe",
			[]string{"--dsn", "hote:5432/base", "--sortie", "gescom.calque.json"},
			"prefixe",
		},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			err := extraire(c.args)
			if err == nil {
				t.Fatal("aucune erreur")
			}
			if !strings.Contains(err.Error(), c.attendu) {
				t.Errorf("erreur %q, attendu qu'elle mentionne %q", err, c.attendu)
			}
		})
	}
}

// ORMEAU_DSN évite de laisser le mot de passe dans l'historique du shell et
// dans ps. Le drapeau reste prioritaire quand les deux sont donnés.
func TestExtraireLitLEnvironnement(t *testing.T) {
	t.Setenv("ORMEAU_DSN", "db2://hote/base")

	err := extraire([]string{"--sortie", "gescom.calque.json"})
	if err == nil {
		t.Fatal("aucune erreur")
	}
	if strings.Contains(err.Error(), "--dsn") {
		t.Errorf("ORMEAU_DSN n'a pas ete lu : %v", err)
	}
	if !strings.Contains(err.Error(), "prefixe") {
		t.Errorf("erreur inattendue : %v", err)
	}
}

// Le DSN ne ressort d'aucune erreur, y compris de celles qui le citent en
// partie pour situer la panne.
func TestExtraireNeDivulguePasLeDSN(t *testing.T) {
	t.Parallel()

	const secret = "motdepasse-secret"

	err := extraire([]string{"--dsn", "db2://utilisateur:" + secret + "@hote:5432/base", "--sortie", "s.json"})
	if err == nil {
		t.Fatal("aucune erreur")
	}
	if strings.Contains(err.Error(), secret) {
		t.Errorf("le mot de passe apparaît dans l'erreur : %v", err)
	}
}

// Les drapeaux séparés évitent de composer une URL à la main, ce qui casse dès
// que le mot de passe porte un caractère réservé.
func TestExtraireAccepteLesDrapeauxSepares(t *testing.T) {
	t.Setenv("ORMEAU_DSN", "")
	t.Setenv("ORMEAU_MDP", "mot@de:passe/complique")

	// Le sgbd est volontairement inconnu : la composition doit aboutir, et
	// l'échec tomber ensuite, à la résolution du pilote.
	err := extraire([]string{
		"--sgbd", "db2", "--hote", "serveur", "--utilisateur", "u",
		"--base", "gescom", "--sortie", "s.json",
	})
	if err == nil {
		t.Fatal("aucune erreur")
	}
	if strings.Contains(err.Error(), "--dsn") {
		t.Errorf("les drapeaux separes n'ont pas ete pris en compte : %v", err)
	}
	if strings.Contains(err.Error(), "mot@de:passe") {
		t.Errorf("le mot de passe apparait dans l'erreur : %v", err)
	}
}

// Le mot de passe ne doit exister nulle part comme drapeau : le poser en
// ligne de commande le rendrait visible dans ps.
func TestExtraireNAPasDeDrapeauMotDePasse(t *testing.T) {
	t.Parallel()

	for _, interdit := range []string{"--mdp", "--motdepasse", "--password"} {
		err := extraire([]string{interdit, "secret", "--sortie", "s.json"})
		if err == nil {
			t.Errorf("%s accepte", interdit)
		}
	}
}

// Une entrée vide n'est pas un schéma : « public, » n'en demande pas deux.
func TestDecouperLesSchemas(t *testing.T) {
	t.Parallel()

	cas := []struct {
		liste   string
		attendu []string
	}{
		{"", nil},
		{"   ", nil},
		{"public", []string{"public"}},
		{"public,gescom", []string{"public", "gescom"}},
		{" public , gescom ", []string{"public", "gescom"}},
		{"public,", []string{"public"}},
		{",public,,gescom,", []string{"public", "gescom"}},
	}

	for _, c := range cas {
		obtenu := decouper(c.liste)
		if len(obtenu) != len(c.attendu) {
			t.Errorf("decouper(%q) = %v, attendu %v", c.liste, obtenu, c.attendu)
			continue
		}
		for i := range c.attendu {
			if obtenu[i] != c.attendu[i] {
				t.Errorf("decouper(%q) = %v, attendu %v", c.liste, obtenu, c.attendu)
				break
			}
		}
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

// La version doit exister comme variable, sans quoi le -X du Makefile ne
// désigne rien : le linker ne signale pas une cible absente, et tous les
// binaires publiés porteraient « dev ».
func TestVersionAffichee(t *testing.T) {
	t.Parallel()

	if version == "" {
		t.Fatal("version vide")
	}
	if affichee := versionAffichee(); !strings.HasPrefix(affichee, "ormeau ") {
		t.Errorf("version affichee %q", affichee)
	}
	if !strings.Contains(versionAffichee(), version) {
		t.Errorf("la version %q n'apparait pas dans %q", version, versionAffichee())
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
