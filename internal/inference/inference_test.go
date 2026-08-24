// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
)

// Ce que les fichiers de référence ne couvrent pas : les chemins qui n'ont pas
// de calque physique pour les déclencher.

// TestInfererSansDecisions vérifie qu'un appel sans fichier de décisions passe.
//
// C'est le premier passage, celui de quelqu'un qui découvre l'outil : il n'a
// pas encore de decisions.yaml, et exiger un fichier vide serait absurde.
func TestInfererSansDecisions(t *testing.T) {
	t.Parallel()

	physique := &calque.Physique{
		VersionRI: calque.VersionCourante,
		Tables: []calque.Table{{
			Nom:    "client",
			Schema: "public",
			Colonnes: []calque.Colonne{
				{Nom: "id", Position: 1, TypeBrut: "integer", TypeNormalise: calque.TypeEntier},
			},
			ClePrimaire: &calque.ClePrimaire{Nom: "client_pkey", Colonnes: []string{"id"}},
		}},
	}

	logique, _ := Inferer(physique, nil)

	if logique.EspaceDeNoms != `App\Entity` {
		t.Errorf("espace de noms = %q, attendu App\\Entity", logique.EspaceDeNoms)
	}
	if len(logique.Entites) != 1 || logique.Entites[0].Nom != "Client" {
		t.Fatalf("entites = %#v, attendue une entite Client", logique.Entites)
	}
}

// TestInfererNeRendJamaisDErreur vérifie qu'un physique vide produit un logique
// vide plutôt qu'un échec.
//
// Une inférence qui n'aboutit pas produit un avertissement, jamais une erreur :
// c'est l'invariant qui permet de rendre un calque partiel accompagné de vingt
// avertissements précis, au lieu de s'arrêter à la première table difficile.
func TestInfererNeRendJamaisDErreur(t *testing.T) {
	t.Parallel()

	logique, avertissements := Inferer(&calque.Physique{VersionRI: calque.VersionCourante}, nil)

	if logique == nil {
		t.Fatal("logique nil sur un physique vide")
	}
	if len(logique.Entites) != 0 {
		t.Errorf("entites = %#v, attendu aucune", logique.Entites)
	}
	if len(avertissements) != 0 {
		t.Errorf("avertissements = %#v, attendu aucun", avertissements)
	}
}

// TestTrierAvertissementsDepartageParCode vérifie l'ordre à cible égale.
//
// Les codes servent de filtre en CI. Deux avertissements sur la même colonne
// dans un ordre qui change d'une exécution à l'autre produiraient un diff sans
// qu'aucun schéma n'ait bougé — exactement ce que le déterminisme du calque
// existe pour éviter.
func TestTrierAvertissementsDepartageParCode(t *testing.T) {
	t.Parallel()

	avertissements := []calque.Avertissement{
		{Code: calque.CodeTypeNonReconnu, Cible: "public.client.actif"},
		{Code: calque.CodeDefautIncompatible, Cible: "public.client.actif"},
		{Code: calque.CodeCollision, Cible: "public.audit"},
	}

	trierAvertissements(avertissements)

	attendus := []string{
		calque.CodeCollision,
		calque.CodeDefautIncompatible,
		calque.CodeTypeNonReconnu,
	}
	for i, code := range attendus {
		if avertissements[i].Code != code {
			t.Errorf("rang %d : code %s, attendu %s", i, avertissements[i].Code, code)
		}
	}
}

// TestCapitaliserChaineVide vérifie que le mot vide ne fait pas paniquer.
//
// Il arrive dès qu'un identifiant se réduit à des séparateurs, ce qu'une base
// reprise produit plus souvent qu'on ne le croit.
func TestCapitaliserChaineVide(t *testing.T) {
	t.Parallel()

	if obtenu := capitaliser(""); obtenu != "" {
		t.Errorf("capitaliser(\"\") = %q, attendu la chaine vide", obtenu)
	}
}
