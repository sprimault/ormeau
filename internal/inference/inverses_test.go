// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
)

// TestGenreInverse couvre le retournement de cardinalité.
func TestGenreInverse(t *testing.T) {
	t.Parallel()

	for genre, attendu := range map[calque.GenreAssociation]calque.GenreAssociation{
		calque.PlusieursVersUn: calque.UnVersPlusieurs,
		calque.UnVersUn:        calque.UnVersUn,
	} {
		if obtenu := genreInverse(genre); obtenu != attendu {
			t.Errorf("genreInverse(%s) = %s, attendu %s", genre, obtenu, attendu)
		}
	}
}

// TestNomLibre couvre la résolution des collisions.
//
// Un nom en double dans une classe PHP ne compile pas. Deux clés étrangères
// d'une même table vers la même cible — un client facturé et un client livré —
// produisent deux côtés inverses de même nom, et il faut les départager sans
// rien demander à personne.
func TestNomLibre(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom          string
		pris         map[string]bool
		souhaite     string
		discriminant string
		attendu      string
	}{
		{"nom libre", map[string]bool{}, "commande", "client", "commande"},
		{
			"collision : le discriminant departage",
			map[string]bool{"commande": true},
			"commande", "clientLivre", "commandeClientLivre",
		},
		{
			"double collision : numerotation",
			map[string]bool{"commande": true, "commandeClient": true},
			"commande", "client", "commandeClient2",
		},
		{
			"collision avec une propriete",
			map[string]bool{"client": true},
			"client", "clientId", "clientClientId",
		},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			if obtenu := nomLibre(c.pris, c.souhaite, c.discriminant); obtenu != c.attendu {
				t.Errorf("nomLibre = %q, attendu %q", obtenu, c.attendu)
			}
		})
	}
}

// TestAjouterCotesInversesApparieLesDeuxCotes vérifie que mappee_par et
// inversee_par se répondent.
//
// Doctrine n'écrit rien en base si les deux ne se désignent pas mutuellement,
// et il ne s'en plaint pas : le mapping est accepté, la relation ne persiste
// simplement jamais.
func TestAjouterCotesInversesApparieLesDeuxCotes(t *testing.T) {
	t.Parallel()

	logique := &calque.Logique{
		Entites: []calque.Entite{
			{Nom: "Client"},
			{
				Nom: "Commande",
				Associations: []calque.Association{{
					Nom:          "client",
					Genre:        calque.PlusieursVersUn,
					Cible:        "Client",
					Proprietaire: true,
				}},
			},
		},
	}

	ajouterCotesInverses(logique)

	proprietaire := logique.Entites[1].Associations[0]
	if len(logique.Entites[0].Associations) != 1 {
		t.Fatalf("Client porte %d association(s), attendue une", len(logique.Entites[0].Associations))
	}
	inverse := logique.Entites[0].Associations[0]

	if proprietaire.InverseePar != inverse.Nom {
		t.Errorf("inversee_par = %q, attendu %q", proprietaire.InverseePar, inverse.Nom)
	}
	if inverse.MappeePar != proprietaire.Nom {
		t.Errorf("mappee_par = %q, attendu %q", inverse.MappeePar, proprietaire.Nom)
	}
	if inverse.Proprietaire {
		t.Error("le cote inverse ne peut pas etre proprietaire")
	}
	if inverse.Genre != calque.UnVersPlusieurs {
		t.Errorf("genre inverse = %s, attendu un_vers_plusieurs", inverse.Genre)
	}
}

// TestAjouterCotesInversesEviteLesCollisions vérifie le cas des deux clés
// étrangères vers la même table.
func TestAjouterCotesInversesEviteLesCollisions(t *testing.T) {
	t.Parallel()

	logique := &calque.Logique{
		Entites: []calque.Entite{
			{Nom: "Client"},
			{
				Nom: "Commande",
				Associations: []calque.Association{
					{Nom: "clientFacture", Genre: calque.PlusieursVersUn, Cible: "Client", Proprietaire: true},
					{Nom: "clientLivre", Genre: calque.PlusieursVersUn, Cible: "Client", Proprietaire: true},
				},
			},
		},
	}

	ajouterCotesInverses(logique)

	noms := map[string]bool{}
	for _, a := range logique.Entites[0].Associations {
		if noms[a.Nom] {
			t.Errorf("le nom %q apparait deux fois sur Client", a.Nom)
		}
		noms[a.Nom] = true
	}
	if len(noms) != 2 {
		t.Errorf("Client porte %d association(s), attendues deux", len(noms))
	}
}

// TestAjouterCotesInversesNeCollisionnePasAvecUnePropriete vérifie qu'une
// association ne prend pas le nom d'une propriété existante.
//
// Une entité Client qui porte déjà une colonne nommée commande recevrait deux
// membres du même nom, et la classe ne compilerait pas.
func TestAjouterCotesInversesNeCollisionnePasAvecUnePropriete(t *testing.T) {
	t.Parallel()

	logique := &calque.Logique{
		Entites: []calque.Entite{
			{Nom: "Client", Proprietes: []calque.Propriete{{Nom: "commande", Colonne: "commande"}}},
			{
				Nom: "Commande",
				Associations: []calque.Association{{
					Nom: "client", Genre: calque.PlusieursVersUn, Cible: "Client", Proprietaire: true,
				}},
			},
		},
	}

	ajouterCotesInverses(logique)

	if nom := logique.Entites[0].Associations[0].Nom; nom == "commande" {
		t.Error("l'association a pris le nom d'une propriete existante")
	}
}

// TestAjouterCotesInversesIgnoreUneCibleAbsente vérifie qu'une cible hors du
// calque ne fait pas paniquer.
//
// Le cas arrive dès qu'une portée restreint l'extraction, ou qu'une décision
// écarte une table encore désignée par une clé étrangère.
func TestAjouterCotesInversesIgnoreUneCibleAbsente(t *testing.T) {
	t.Parallel()

	logique := &calque.Logique{
		Entites: []calque.Entite{{
			Nom: "Commande",
			Associations: []calque.Association{{
				Nom: "client", Genre: calque.PlusieursVersUn, Cible: "Client", Proprietaire: true,
			}},
		}},
	}

	ajouterCotesInverses(logique)

	if logique.Entites[0].Associations[0].InverseePar != "" {
		t.Error("un cote inverse a ete pose sur une entite absente du calque")
	}
}
