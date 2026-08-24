// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"reflect"
	"testing"
)

// TestDecouperIdentifiant couvre les conventions qu'une base reprise mélange.
//
// C'est la fonction dont tout le nommage dépend : cli_nom, cliNom et CLI_NOM
// désignent la même colonne et doivent donner la même propriété. Une coupure
// ratée ici se propage à chaque nom de classe et de propriété du calque.
func TestDecouperIdentifiant(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom     string
		entree  string
		attendu []string
	}{
		{"souligne", "raison_sociale", []string{"raison", "sociale"}},
		{"tiret", "raison-sociale", []string{"raison", "sociale"}},
		{"espace", "raison sociale", []string{"raison", "sociale"}},
		{"casse chameau", "raisonSociale", []string{"raison", "Sociale"}},
		{"capitales", "CLI_NOM", []string{"CLI", "NOM"}},
		{"acronyme colle", "CLICode", []string{"CLICode"}},
		{"mot seul", "client", []string{"client"}},
		{"soulignes multiples", "cli__nom", []string{"cli", "nom"}},
		{"souligne en tete", "_prive", []string{"prive"}},
		{"souligne en fin", "prive_", []string{"prive"}},
		{"chaine vide", "", nil},
		{"soulignes seuls", "___", nil},
		{"accents", "référence_client", []string{"référence", "client"}},
		{"chiffres", "adresse_ligne2", []string{"adresse", "ligne2"}},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			if obtenu := decouperIdentifiant(c.entree); !reflect.DeepEqual(obtenu, c.attendu) {
				t.Errorf("decouperIdentifiant(%q) = %#v, attendu %#v", c.entree, obtenu, c.attendu)
			}
		})
	}
}

// TestPascalCase vérifie les noms de classe.
//
// Les préfixes ne sont pas retirés à ce stade : T_CLIENTS donne TClients, et
// c'est volontaire. Retirer un préfixe est un jugement, il vient avec les
// heuristiques de nommage et son propre cas de référence.
func TestPascalCase(t *testing.T) {
	t.Parallel()

	cas := []struct {
		entree, attendu string
	}{
		{"client", "Client"},
		{"CLIENT", "Client"},
		{"t_clients", "TClients"},
		{"ligne_commande", "LigneCommande"},
		{"ligneCommande", "LigneCommande"},
		{"référence", "Référence"},
		{"t_référence", "TRéférence"},
		{"", ""},
	}

	for _, c := range cas {
		t.Run(c.entree, func(t *testing.T) {
			t.Parallel()

			if obtenu := pascalCase(c.entree); obtenu != c.attendu {
				t.Errorf("pascalCase(%q) = %q, attendu %q", c.entree, obtenu, c.attendu)
			}
		})
	}
}

// TestCamelCase vérifie les noms de propriété.
func TestCamelCase(t *testing.T) {
	t.Parallel()

	cas := []struct {
		entree, attendu string
	}{
		{"nom", "nom"},
		{"CLI_NOM", "cliNom"},
		{"CLI_CA_TTC", "cliCaTtc"},
		{"raison_sociale", "raisonSociale"},
		{"raisonSociale", "raisonSociale"},
		{"ID", "id"},
		{"", ""},
	}

	for _, c := range cas {
		t.Run(c.entree, func(t *testing.T) {
			t.Parallel()

			if obtenu := camelCase(c.entree); obtenu != c.attendu {
				t.Errorf("camelCase(%q) = %q, attendu %q", c.entree, obtenu, c.attendu)
			}
		})
	}
}

// TestNommageIdentifiantsInvalides vérifie que le nom produit est toujours un
// identifiant PHP.
//
// C'est la garantie qui compte le plus ici : un nom mal formé ne se voit pas
// dans le calque, il se voit au moment où le fichier généré refuse de compiler,
// chez l'utilisateur et loin de l'inférence qui l'a causé.
//
// N° Commande, Prix (HT) et 2024_ventes viennent tous de bases réelles.
func TestNommageIdentifiantsInvalides(t *testing.T) {
	t.Parallel()

	cas := []struct {
		entree, classe, propriete string
	}{
		{"N° Commande", "NCommande", "nCommande"},
		{"Prix (HT)", "PrixHt", "prixHt"},
		{"% remise", "Remise", "remise"},
		{"CA TTC", "CaTtc", "caTtc"},
		{"Date de création", "DateDeCréation", "dateDeCréation"},

		// Un identifiant PHP ne commence pas par un chiffre.
		{"2024_ventes", "_2024Ventes", "_2024Ventes"},
		{"3EME_TRIMESTRE", "_3EmeTrimestre", "_3EmeTrimestre"},
	}

	for _, c := range cas {
		t.Run(c.entree, func(t *testing.T) {
			t.Parallel()

			if obtenu := pascalCase(c.entree); obtenu != c.classe {
				t.Errorf("pascalCase(%q) = %q, attendu %q", c.entree, obtenu, c.classe)
			}
			if obtenu := camelCase(c.entree); obtenu != c.propriete {
				t.Errorf("camelCase(%q) = %q, attendu %q", c.entree, obtenu, c.propriete)
			}
		})
	}
}

// TestValideEnPHP couvre la garde du premier caractère.
func TestValideEnPHP(t *testing.T) {
	t.Parallel()

	for entree, attendu := range map[string]string{
		"Client":     "Client",
		"_prive":     "_prive",
		"2024Ventes": "_2024Ventes",
		"":           "",
		"Étage":      "Étage",
	} {
		if obtenu := valideEnPHP(entree); obtenu != attendu {
			t.Errorf("valideEnPHP(%q) = %q, attendu %q", entree, obtenu, attendu)
		}
	}
}

// TestCapitaliserPreserveLesAccents vérifie qu'un identifiant accentué n'est pas
// translittéré.
//
// Remplacer é par e donnerait une classe qui ne correspond plus à sa table, et
// le lien entre les deux est tout ce qui permet de régénérer.
func TestCapitaliserPreserveLesAccents(t *testing.T) {
	t.Parallel()

	for entree, attendu := range map[string]string{
		"référence": "Référence",
		"ÉTAT":      "État",
		"élève":     "Élève",
		"ça":        "Ça",
	} {
		if obtenu := capitaliser(entree); obtenu != attendu {
			t.Errorf("capitaliser(%q) = %q, attendu %q", entree, obtenu, attendu)
		}
	}
}
