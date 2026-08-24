// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
)

// colonne construit une colonne entière pour les tests de structure.
func colonne(nom string) calque.Colonne {
	return calque.Colonne{Nom: nom, TypeBrut: "integer", TypeNormalise: calque.TypeEntier}
}

// fk construit une clé étrangère à une colonne.
func fk(nom, colonne, table string) calque.CleEtrangere {
	return calque.CleEtrangere{
		Nom:            nom,
		Colonnes:       []string{colonne},
		TableCible:     table,
		SchemaCible:    "public",
		ColonnesCibles: []string{"id"},
	}
}

// TestReconnaitreJointure couvre la table qui n'existe que pour en relier deux
// autres.
//
// La condition qui compte est l'absence de toute autre colonne. Une table de
// liaison qui porte une quantité est une entité : la rendre en association
// ferait disparaître ses données du modèle, et rien ne le signalerait avant la
// première requête qui les cherche.
func TestReconnaitreJointure(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom     string
		table   calque.Table
		attendu bool
	}{
		{
			"deux cles etrangeres et rien d'autre",
			calque.Table{
				Nom: "commande_article", Schema: "public",
				Colonnes:       []calque.Colonne{colonne("commande_id"), colonne("article_id")},
				ClePrimaire:    &calque.ClePrimaire{Colonnes: []string{"commande_id", "article_id"}},
				ClesEtrangeres: []calque.CleEtrangere{fk("f1", "commande_id", "commande"), fk("f2", "article_id", "article")},
			},
			true,
		},
		{
			"une colonne de plus : c'est une entite",
			calque.Table{
				Nom: "ligne_commande", Schema: "public",
				Colonnes:       []calque.Colonne{colonne("commande_id"), colonne("article_id"), colonne("quantite")},
				ClePrimaire:    &calque.ClePrimaire{Colonnes: []string{"commande_id", "article_id"}},
				ClesEtrangeres: []calque.CleEtrangere{fk("f1", "commande_id", "commande"), fk("f2", "article_id", "article")},
			},
			false,
		},
		{
			"sans cle primaire, les doublons passent",
			calque.Table{
				Nom: "liaison", Schema: "public",
				Colonnes:       []calque.Colonne{colonne("commande_id"), colonne("article_id")},
				ClesEtrangeres: []calque.CleEtrangere{fk("f1", "commande_id", "commande"), fk("f2", "article_id", "article")},
			},
			false,
		},
		{
			"cle primaire sur une seule des deux colonnes",
			calque.Table{
				Nom: "liaison", Schema: "public",
				Colonnes:       []calque.Colonne{colonne("commande_id"), colonne("article_id")},
				ClePrimaire:    &calque.ClePrimaire{Colonnes: []string{"commande_id"}},
				ClesEtrangeres: []calque.CleEtrangere{fk("f1", "commande_id", "commande"), fk("f2", "article_id", "article")},
			},
			false,
		},
		{
			"une seule cle etrangere",
			calque.Table{
				Nom: "commande", Schema: "public",
				Colonnes:       []calque.Colonne{colonne("id"), colonne("client_id")},
				ClePrimaire:    &calque.ClePrimaire{Colonnes: []string{"id", "client_id"}},
				ClesEtrangeres: []calque.CleEtrangere{fk("f1", "client_id", "client")},
			},
			false,
		},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			if obtenu := reconnaitreJointure(&c.table) != nil; obtenu != c.attendu {
				t.Errorf("reconnaitreJointure = %v, attendu %v", obtenu, c.attendu)
			}
		})
	}
}

// TestReconnaitreHeritage couvre la clé primaire qui est aussi étrangère.
func TestReconnaitreHeritage(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom     string
		table   calque.Table
		attendu bool
	}{
		{
			"cle primaire pointant sur la table mere",
			calque.Table{
				Nom: "salarie", Schema: "public",
				Colonnes:       []calque.Colonne{colonne("id"), colonne("matricule")},
				ClePrimaire:    &calque.ClePrimaire{Colonnes: []string{"id"}},
				ClesEtrangeres: []calque.CleEtrangere{fk("f1", "id", "personne")},
			},
			true,
		},
		{
			"cle primaire composite dont une seule colonne est etrangere",
			calque.Table{
				Nom: "affectation", Schema: "public",
				Colonnes:       []calque.Colonne{colonne("salarie_id"), colonne("debut")},
				ClePrimaire:    &calque.ClePrimaire{Colonnes: []string{"salarie_id", "debut"}},
				ClesEtrangeres: []calque.CleEtrangere{fk("f1", "salarie_id", "salarie")},
			},
			false,
		},
		{
			"cle etrangere hors de la cle primaire",
			calque.Table{
				Nom: "commande", Schema: "public",
				Colonnes:       []calque.Colonne{colonne("id"), colonne("client_id")},
				ClePrimaire:    &calque.ClePrimaire{Colonnes: []string{"id"}},
				ClesEtrangeres: []calque.CleEtrangere{fk("f1", "client_id", "client")},
			},
			false,
		},
		{
			"auto-reference : une table n'herite pas d'elle-meme",
			calque.Table{
				Nom: "categorie", Schema: "public",
				Colonnes:       []calque.Colonne{colonne("id")},
				ClePrimaire:    &calque.ClePrimaire{Colonnes: []string{"id"}},
				ClesEtrangeres: []calque.CleEtrangere{fk("f1", "id", "categorie")},
			},
			false,
		},
		{
			"sans cle primaire",
			calque.Table{
				Nom: "journal", Schema: "public",
				Colonnes:       []calque.Colonne{colonne("id")},
				ClesEtrangeres: []calque.CleEtrangere{fk("f1", "id", "personne")},
			},
			false,
		},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			if obtenu := reconnaitreHeritage(&c.table) != nil; obtenu != c.attendu {
				t.Errorf("reconnaitreHeritage = %v, attendu %v", obtenu, c.attendu)
			}
		})
	}
}

// TestRetirerSuffixeIdentifiant couvre les marques de clé étrangère.
//
// C'est du retrait de suffixe, pas de la morphologie : rien n'est deviné sur la
// langue ni sur le nombre, contrairement au nommage des classes.
func TestRetirerSuffixeIdentifiant(t *testing.T) {
	t.Parallel()

	for entree, attendu := range map[string]string{
		"client_id":             "client",
		"CLIENT_ID":             "CLIENT",
		"client_fk":             "client",
		"client_code":           "client",
		"client_num":            "client",
		"client_no":             "client",
		"client_ident":          "client",
		"commande_remplacee_id": "commande_remplacee",

		// Rien à retirer : la colonne désigne la table elle-même, ou ne porte
		// aucune marque.
		"id":     "",
		"_id":    "",
		"client": "",
		"nom":    "",
	} {
		if obtenu := retirerSuffixeIdentifiant(entree); obtenu != attendu {
			t.Errorf("retirerSuffixeIdentifiant(%q) = %q, attendu %q", entree, obtenu, attendu)
		}
	}
}

// TestNomAssociation couvre le nom de la propriété qui porte la relation.
func TestNomAssociation(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom      string
		cle      calque.CleEtrangere
		nomCible string
		attendu  string
	}{
		{
			"convention client_id",
			fk("f1", "client_id", "client"), "Client", "client",
		},
		{
			"colonne sans suffixe : le nom de la classe cible",
			fk("f1", "proprietaire", "personne"), "Personne", "personne",
		},
		{
			"colonne nommee id : la classe cible, sinon la propriete s'appellerait id",
			fk("f1", "id", "personne"), "Personne", "personne",
		},
		{
			"cle composite : la classe cible",
			calque.CleEtrangere{
				Colonnes:       []string{"societe_id", "etablissement_id"},
				TableCible:     "etablissement",
				ColonnesCibles: []string{"societe_id", "id"},
			},
			"Etablissement", "etablissement",
		},
		{
			"nom explicite conserve",
			fk("f1", "commande_remplacee_id", "commande"), "Commande", "commandeRemplacee",
		},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			if obtenu := nomAssociation(&c.cle, c.nomCible); obtenu != c.attendu {
				t.Errorf("nomAssociation = %q, attendu %q", obtenu, c.attendu)
			}
		})
	}
}

// TestUnicitePorte vérifie ce qui distingue un-vers-un de plusieurs-vers-un.
//
// La distinction ne se lit ni dans le nommage ni dans la clé étrangère : c'est
// l'unicité sur les colonnes portantes, et rien d'autre.
func TestUnicitePorte(t *testing.T) {
	t.Parallel()

	table := calque.Table{
		Unicites: []calque.Contrainte{
			{Nom: "uq_client", Colonnes: []string{"client_id"}},
			{Nom: "uq_client_date", Colonnes: []string{"societe_id", "debut"}},
		},
	}

	cas := []struct {
		nom      string
		colonnes []string
		attendu  bool
	}{
		{"unicite exacte", []string{"client_id"}, true},
		{"aucune unicite", []string{"article_id"}, false},
		{"unicite composite couvrant les memes colonnes", []string{"debut", "societe_id"}, true},
		{"sous-ensemble d'une unicite composite", []string{"societe_id"}, false},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			if obtenu := unicitePorte(&table, c.colonnes); obtenu != c.attendu {
				t.Errorf("unicitePorte(%v) = %v, attendu %v", c.colonnes, obtenu, c.attendu)
			}
		})
	}
}
