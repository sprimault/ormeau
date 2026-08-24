// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
)

// tables construit un jeu de tables réduit à leurs noms, tout ce dont la
// détection de préfixe a besoin.
func tables(noms ...string) []calque.Table {
	jeu := make([]calque.Table, 0, len(noms))
	for _, nom := range noms {
		jeu = append(jeu, calque.Table{Nom: nom, Schema: "public"})
	}
	return jeu
}

// TestPrefixeCommun couvre la détection.
//
// Elle est délibérément conservatrice : dans le doute elle ne rend rien, et
// l'utilisateur garde des noms de classes fidèles à ses tables. L'inverse —
// amputer un nom à tort — donne une classe que personne n'a demandée et qui
// ressemble à un bug de l'outil.
func TestPrefixeCommun(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom     string
		tables  []calque.Table
		attendu string
	}{
		{
			"convention sur toutes les tables",
			tables("T_CLIENTS", "T_COMMANDES", "T_ARTICLES"),
			"T_",
		},
		{
			"prefixe plus long",
			tables("tbl_client", "tbl_commande", "tbl_article"),
			"tbl_",
		},
		{
			"debut commun plus long que le prefixe reel",
			tables("T_CLIENTS", "T_CLOTURES", "T_COMMANDES"),
			"T_",
		},
		{
			"une seule table hors convention annule tout",
			tables("T_CLIENTS", "T_COMMANDES", "PARAMETRES"),
			"",
		},
		{
			"debut commun sans separateur",
			tables("client", "commande", "commentaire"),
			"",
		},
		{
			"debut commun trop long pour etre une convention",
			tables("commande_ligne", "commande_entete", "commande_taxe"),
			"",
		},
		{
			"deux tables ne suffisent pas a conclure",
			tables("T_CLIENTS", "T_COMMANDES"),
			"",
		},
		{
			"aucune table",
			nil,
			"",
		},
		{
			"le prefixe consommerait un nom entier",
			tables("T_", "T_CLIENTS", "T_COMMANDES"),
			"",
		},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			if obtenu := prefixeCommun(c.tables); obtenu != c.attendu {
				t.Errorf("prefixeCommun = %q, attendu %q", obtenu, c.attendu)
			}
		})
	}
}

// TestRetirerPrefixe couvre le retrait proprement dit.
func TestRetirerPrefixe(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom, entree string
		prefixes    []string
		attendu     string
	}{
		{"correspondance exacte", "T_CLIENTS", []string{"T_"}, "CLIENTS"},
		{"casse differente", "t_clients", []string{"T_"}, "clients"},
		{"aucune correspondance", "PARAMETRES", []string{"T_"}, "PARAMETRES"},
		{"premier qui correspond", "tbl_client", []string{"T_", "tbl_"}, "client"},
		{"liste vide", "T_CLIENTS", nil, "T_CLIENTS"},
		{"prefixe vide ignore", "T_CLIENTS", []string{"", "T_"}, "CLIENTS"},
		{"ne laisserait rien", "T_", []string{"T_"}, "T_"},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			if obtenu := retirerPrefixe(c.entree, c.prefixes); obtenu != c.attendu {
				t.Errorf("retirerPrefixe(%q, %v) = %q, attendu %q", c.entree, c.prefixes, obtenu, c.attendu)
			}
		})
	}
}

// TestPrefixesRetenusPrivilegieLaDecision vérifie qu'une décision remplace la
// détection au lieu de s'y ajouter.
//
// Celui qui a écrit prefixes_a_retirer a regardé sa base. Ajouter une
// trouvaille de l'outil par-dessus produirait un nom que personne n'a demandé,
// et le rendrait difficile à expliquer.
func TestPrefixesRetenusPrivilegieLaDecision(t *testing.T) {
	t.Parallel()

	jeu := tables("T_CLIENTS", "T_COMMANDES", "T_ARTICLES")

	retenus, detecte := prefixesRetenus(jeu, &Decisions{PrefixesARetirer: []string{"TBL_"}})

	if len(retenus) != 1 || retenus[0] != "TBL_" {
		t.Errorf("prefixes retenus = %v, attendu [TBL_]", retenus)
	}
	if detecte != "" {
		t.Errorf("detecte = %q, attendu vide : une decision ne se signale pas", detecte)
	}
}

// TestPrefixesRetenusNAppliquePasCeQuIlDetecte vérifie que la détection ne
// touche pas au calque.
//
// C'est la distinction qui porte tout le nommage : l'outil repère le préfixe et
// le signale, mais ne l'enlève pas. Un nom de table est un constat, et
// l'amputer se décide.
func TestPrefixesRetenusNAppliquePasCeQuIlDetecte(t *testing.T) {
	t.Parallel()

	jeu := tables("T_CLIENTS", "T_COMMANDES", "T_ARTICLES")

	retenus, detecte := prefixesRetenus(jeu, &Decisions{})

	if len(retenus) != 0 {
		t.Errorf("prefixes retenus = %v, attendu aucun sans decision", retenus)
	}
	if detecte != "T_" {
		t.Errorf("detecte = %q, attendu T_", detecte)
	}
}

// TestPrefixesRetenusEstDeterministe vérifie que l'ordre des préfixes décidés
// ne dépend pas de celui du fichier.
//
// Deux inférences du même calque doivent rendre les mêmes octets, et la liste
// sert à un parcours dont le premier élément gagne.
func TestPrefixesRetenusEstDeterministe(t *testing.T) {
	t.Parallel()

	jeu := tables("T_CLIENTS", "T_COMMANDES", "T_ARTICLES")

	premier, _ := prefixesRetenus(jeu, &Decisions{PrefixesARetirer: []string{"z_", "a_", "m_"}})
	second, _ := prefixesRetenus(jeu, &Decisions{PrefixesARetirer: []string{"m_", "z_", "a_"}})

	for i := range premier {
		if premier[i] != second[i] {
			t.Fatalf("ordre instable : %v puis %v", premier, second)
		}
	}
}
