// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"reflect"
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
)

// tableAvecColonnes construit une table réduite à ses noms de colonnes.
func tableAvecColonnes(noms ...string) *calque.Table {
	t := &calque.Table{Nom: "client", Schema: "public"}
	for _, nom := range noms {
		t.Colonnes = append(t.Colonnes, colonne(nom))
	}
	return t
}

// TestValeursDUneVerification couvre la reconnaissance d'une énumération dans
// un CHECK.
//
// Le jeu de cas négatifs compte plus que les positifs : la plupart des
// contraintes de vérification ne sont pas des énumérations, et en transformer
// une en enum PHP produirait un type que la base ne garantit pas.
func TestValeursDUneVerification(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom        string
		expression string
		table      *calque.Table
		colonne    string
		valeurs    []string
	}{
		{
			"forme ANY ARRAY de PostgreSQL",
			"CHECK ((((canal)::text = ANY ((ARRAY['DIRECT'::character varying, 'WEB'::character varying])::text[]))))",
			tableAvecColonnes("id", "canal"),
			"canal", []string{"DIRECT", "WEB"},
		},
		{
			"forme OR",
			"CHECK (((actif = 'O'::bpchar) OR (actif = 'N'::bpchar)))",
			tableAvecColonnes("id", "actif"),
			"actif", []string{"O", "N"},
		},
		{
			"comparaison ordonnee : pas une enumeration",
			"CHECK ((solde >= (0)::numeric))",
			tableAvecColonnes("id", "solde"),
			"", nil,
		},
		{
			"regle metier sur deux colonnes",
			"CHECK ((((etat)::text <> 'ANNULEE'::text) OR (motif IS NOT NULL)))",
			tableAvecColonnes("id", "etat", "motif"),
			"", nil,
		},
		{
			"valeur unique imposee : un enum a un cas n'apporte rien",
			"CHECK (((pays)::text = 'FR'::text))",
			tableAvecColonnes("id", "pays"),
			"", nil,
		},
		{
			"deux colonnes citees, meme sans operateur d'ordre",
			"CHECK ((((a)::text = 'X'::text) OR ((b)::text = 'Y'::text)))",
			tableAvecColonnes("a", "b"),
			"", nil,
		},
		{
			"aucun litteral",
			"CHECK ((cli_id IS NOT NULL))",
			tableAvecColonnes("cli_id"),
			"", nil,
		},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			colonne, valeurs := valeursDUneVerification(c.expression, c.table)
			if colonne != c.colonne {
				t.Errorf("colonne %q, attendue %q", colonne, c.colonne)
			}
			if c.colonne != "" && !reflect.DeepEqual(valeurs, c.valeurs) {
				t.Errorf("valeurs %v, attendues %v", valeurs, c.valeurs)
			}
		})
	}
}

// TestMentionneNeConfondPasLesPrefixes vérifie que la recherche de colonne
// porte sur des mots entiers.
//
// Sans cette borne, la colonne actif serait trouvée dans inactif, deux colonnes
// sembleraient citées, et l'énumération serait écartée à tort.
func TestMentionneNeConfondPasLesPrefixes(t *testing.T) {
	t.Parallel()

	expression := "CHECK (((inactif = 'O') OR (inactif = 'N')))"

	if mentionne(expression, "actif") {
		t.Error("actif trouvee dans inactif")
	}
	if !mentionne(expression, "inactif") {
		t.Error("inactif non trouvee")
	}
}

// TestLitterauxChaineGardeLOrdreDuCheck vérifie que l'ordre des valeurs est
// celui de la contrainte.
//
// Il porte souvent une progression métier — prospect, actif, suspendu, radié —
// que trier alphabétiquement détruirait, et qui se retrouve dans l'ordre des
// cas de l'enum généré.
func TestLitterauxChaineGardeLOrdreDuCheck(t *testing.T) {
	t.Parallel()

	expression := "ARRAY['PROSPECT'::text, 'ACTIF'::text, 'SUSPENDU'::text, 'RADIE'::text]"
	attendu := []string{"PROSPECT", "ACTIF", "SUSPENDU", "RADIE"}

	if obtenu := litterauxChaine(expression); !reflect.DeepEqual(obtenu, attendu) {
		t.Errorf("litterauxChaine = %v, attendu %v", obtenu, attendu)
	}
}

// TestLitterauxChaineDedoublonne vérifie qu'une valeur répétée ne produit pas
// deux cas de même nom, ce qui ne compilerait pas.
func TestLitterauxChaineDedoublonne(t *testing.T) {
	t.Parallel()

	expression := "((a = 'X') OR (b = 'X') OR (a = 'Y'))"
	attendu := []string{"X", "Y"}

	if obtenu := litterauxChaine(expression); !reflect.DeepEqual(obtenu, attendu) {
		t.Errorf("litterauxChaine = %v, attendu %v", obtenu, attendu)
	}
}

// TestNomDeCas couvre le nommage des cas PHP.
//
// Le second retour distingue ce qui se lit de ce qui reste opaque : c'est lui
// qui déclenche l'avertissement, et donc la proposition d'apparier la valeur à
// un nom dans le fichier de décisions.
func TestNomDeCas(t *testing.T) {
	t.Parallel()

	cas := []struct {
		valeur, attendu string
		lisible         bool
	}{
		{"ACTIF", "Actif", true},
		{"EN_ATTENTE", "EnAttente", true},
		{"suspendu", "Suspendu", true},

		// Deux caractères ou moins n'apprennent rien : O est Oui, Ouvert ou
		// Optionnel, et seul un humain le sait.
		{"O", "O", false},
		{"N", "N", false},
		{"AB", "Ab", false},

		// Un identifiant PHP ne commence pas par un chiffre.
		{"01", "Cas01", false},
		{"1_ACTIF", "Cas1Actif", true},
	}

	for _, c := range cas {
		t.Run(c.valeur, func(t *testing.T) {
			t.Parallel()

			nom, lisible := nomDeCas(c.valeur, nil)
			if nom != c.attendu {
				t.Errorf("nomDeCas(%q) = %q, attendu %q", c.valeur, nom, c.attendu)
			}
			if lisible != c.lisible {
				t.Errorf("nomDeCas(%q) lisible = %v, attendu %v", c.valeur, lisible, c.lisible)
			}
		})
	}
}

// TestNomDeCasDecide vérifie qu'un appariement décidé gagne, et éteint
// l'avertissement.
//
// C'est la raison d'être de la section enumerations du fichier de décisions :
// un O/N en base n'a pas à produire des cas nommés O et N.
func TestNomDeCasDecide(t *testing.T) {
	t.Parallel()

	decides := map[string]string{"O": "Oui", "N": "Non"}

	nom, lisible := nomDeCas("O", decides)
	if nom != "Oui" {
		t.Errorf("nomDeCas = %q, attendu Oui", nom)
	}
	if !lisible {
		t.Error("un cas decide ne devrait plus etre signale comme opaque")
	}
}

// TestUnCheckNEcrasePasUnTypeNatif vérifie la priorité entre les deux sources.
//
// Le type énuméré natif est déclaré par le catalogue, ses valeurs sont exactes.
// Une vérification qui porte sur la même colonne les redit au mieux, les
// contredit au pire.
func TestUnCheckNEcrasePasUnTypeNatif(t *testing.T) {
	t.Parallel()

	p := &calque.Physique{
		VersionRI: calque.VersionCourante,
		Tables: []calque.Table{{
			Nom: "client", Schema: "public",
			Colonnes: []calque.Colonne{{
				Nom: "statut", TypeBrut: "statut_client",
				TypeNormalise: calque.TypeEnumereNorm, TypeEnumere: "statut_client",
			}},
			Verifications: []calque.Verification{{
				Nom:        "ck_statut",
				Expression: "CHECK (((statut)::text = ANY (ARRAY['A'::text, 'B'::text])))",
			}},
		}},
		TypesEnumeres: []calque.TypeEnumere{{
			Nom: "statut_client", Schema: "public",
			Valeurs: []string{"PROSPECT", "ACTIF"},
		}},
	}

	reconnues := enumerationsDuSchema(p, &Decisions{})
	e, connue := reconnues["public.client.statut"]
	if !connue {
		t.Fatal("aucune enumeration reconnue")
	}
	if e.origine != calque.OrigineContrainte {
		t.Errorf("origine %q, attendue contrainte : le type natif prime", e.origine)
	}
	if len(e.valeurs) != 2 || e.valeurs[0] != "PROSPECT" {
		t.Errorf("valeurs %v, attendues celles du type natif", e.valeurs)
	}
}
