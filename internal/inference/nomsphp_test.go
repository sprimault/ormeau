// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"testing"
)

// Ce que le cas de référence decisions-invalides ne couvre pas : la frontière
// exacte de ce que PHP accepte à chaque place, et la garantie que les décisions
// reçues ne sont pas modifiées.

// TestLaFormeDUnIdentifiantPHP vérifie la règle du lexer : ce qui passe, et
// ce qui ferait sortir un nom de sa déclaration ou de son répertoire.
func TestLaFormeDUnIdentifiantPHP(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom    string
		valide bool
	}{
		{"Client", true},
		{"_interne", true},
		{"CatégoriesDesNoms", true},
		{"T2", true},
		{"", false},
		{"2024Ventes", false},
		{"../../public/index", false},
		{`App\Entity`, false},
		{"Client Pro", false},
		{"nom;system", false},
		{"ligne\nsuivante", false},
	}
	for _, c := range cas {
		if obtenu := estIdentifiantPHP(c.nom); obtenu != c.valide {
			t.Errorf("estIdentifiantPHP(%q) = %v, attendu %v", c.nom, obtenu, c.valide)
		}
	}
}

// TestLesMotsReservesDependentDeLaPlace vérifie les écarts relevés sous PHP 8.1
// et 8.4 : un mot refusé pour une classe peut nommer un cas ou un segment
// d'espace de noms, et enum nomme une classe.
func TestLesMotsReservesDependentDeLaPlace(t *testing.T) {
	t.Parallel()

	cas := []struct {
		place    string
		nom      string
		reserves map[string]bool
		refuse   bool
	}{
		{"classe", "List", motsReservesDeClasse, true},
		{"classe", "Die", motsReservesDeClasse, true},
		{"classe", "__CLASS__", motsReservesDeClasse, true},
		{"classe", "Enum", motsReservesDeClasse, false},
		{"classe", "Resource", motsReservesDeClasse, false},
		{"cas", "Match", motsReservesDeCas, false},
		{"cas", "Class", motsReservesDeCas, true},
		{"espace", "Function", motsReservesDEspace, false},
		{"espace", "Namespace", motsReservesDEspace, true},
	}
	for _, c := range cas {
		if refuse := raisonNomRefuse(c.nom, c.reserves) != ""; refuse != c.refuse {
			t.Errorf("%s %s : refusé = %v, attendu %v", c.place, c.nom, refuse, c.refuse)
		}
	}

	for espace, valide := range map[string]bool{
		`App\Entity`:          true,
		`App\Function\Entity`: true,
		`App\\Entity`:         false,
		`\App\Entity`:         false,
		`App\Entity\`:         false,
		`App\Namespace`:       false,
	} {
		if obtenu := raisonEspaceDeNoms(espace) == ""; obtenu != valide {
			t.Errorf("espace de noms %q : valide = %v, attendu %v", espace, obtenu, valide)
		}
	}
}

// TestLesDecisionsRecuesNeSontPasModifiees vérifie que le refus agit sur une
// copie : le fichier prérempli réécrit les décisions telles que l'utilisateur
// les a écrites, et il doit y retrouver sa ligne.
func TestLesDecisionsRecuesNeSontPasModifiees(t *testing.T) {
	t.Parallel()

	d := &Decisions{
		Renommages:       map[string]string{"public.client": "List", "public.commande": "Commande"},
		RelationsForcees: []RelationForcee{{Source: "public.client.parrain_ref", Nom: "a b"}},
		Enumerations:     []EnumerationForcee{{Colonne: "public.client.actif", Nom: "OuiNon", Cas: map[string]string{"N": "class"}}},
	}

	retenues, avertissements := sansNomsInvalides(d)

	if len(avertissements) != 3 {
		t.Errorf("%d avertissements, attendu 3 : %v", len(avertissements), avertissements)
	}
	if _, garde := retenues.Renommages["public.client"]; garde || retenues.Renommages["public.commande"] != "Commande" {
		t.Errorf("renommages retenus : %v", retenues.Renommages)
	}
	if d.Renommages["public.client"] != "List" || d.RelationsForcees[0].Nom != "a b" || d.Enumerations[0].Cas["N"] != "class" {
		t.Errorf("décisions reçues modifiées : %+v", d)
	}
}
