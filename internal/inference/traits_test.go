// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
)

// horodatage construit une propriété d'horodatage pour les tests.
func horodatage(nom, colonne string, nullable bool) calque.Propriete {
	php := "\\DateTimeImmutable"
	if nullable {
		php = "?" + php
	}
	return calque.Propriete{
		Nom: nom, Colonne: colonne,
		TypePHP: php, TypeDoctrine: "datetimetz_immutable", Nullable: nullable,
	}
}

// entiteHorodatee construit une entité avec sa clé et ses horodatages.
func entiteHorodatee(nom string, horodatages ...calque.Propriete) calque.Entite {
	proprietes := []calque.Propriete{{Nom: "id", Colonne: "id", TypePHP: "int", TypeDoctrine: "integer"}}
	proprietes = append(proprietes, horodatages...)

	return calque.Entite{
		Nom:         nom,
		Table:       calque.ReferenceTable{Nom: nom, Schema: "public"},
		Identifiant: &calque.Identifiant{Proprietes: []string{"id"}},
		Proprietes:  proprietes,
	}
}

// TestExtraireTraitsPartageEntreDeuxEntites vérifie le cas nominal.
func TestExtraireTraitsPartageEntreDeuxEntites(t *testing.T) {
	t.Parallel()

	logique := &calque.Logique{
		Entites: []calque.Entite{
			entiteHorodatee("Client",
				horodatage("createdAt", "created_at", false),
				horodatage("updatedAt", "updated_at", true)),
			entiteHorodatee("Commande",
				horodatage("createdAt", "created_at", false),
				horodatage("updatedAt", "updated_at", true)),
		},
	}

	extraireTraits(logique)

	if len(logique.Traits) != 1 || logique.Traits[0].Nom != "Horodatage" {
		t.Fatalf("traits = %+v, attendu un seul nomme Horodatage", logique.Traits)
	}
	if len(logique.Traits[0].Proprietes) != 2 {
		t.Errorf("le trait porte %d propriete(s), attendues deux", len(logique.Traits[0].Proprietes))
	}

	for i := range logique.Entites {
		e := &logique.Entites[i]
		if len(e.Traits) != 1 || e.Traits[0] != "Horodatage" {
			t.Errorf("%s : traits %v", e.Nom, e.Traits)
		}
		if len(e.Proprietes) != 1 || e.Proprietes[0].Nom != "id" {
			t.Errorf("%s garde %d propriete(s), attendue la seule cle", e.Nom, len(e.Proprietes))
		}
	}
}

// TestExtraireTraitsIgnoreUneSeuleEntite vérifie qu'un trait à un utilisateur
// n'est pas produit.
//
// Un fichier de plus pour deux propriétés qu'on lirait aussi bien dans
// l'entité : le trait ne se justifie que par le partage.
func TestExtraireTraitsIgnoreUneSeuleEntite(t *testing.T) {
	t.Parallel()

	logique := &calque.Logique{
		Entites: []calque.Entite{
			entiteHorodatee("Client",
				horodatage("createdAt", "created_at", false),
				horodatage("updatedAt", "updated_at", true)),
		},
	}

	extraireTraits(logique)

	if len(logique.Traits) != 0 {
		t.Errorf("traits = %+v, attendu aucun", logique.Traits)
	}
	if len(logique.Entites[0].Proprietes) != 3 {
		t.Errorf("l'entite a perdu des proprietes : %d restantes", len(logique.Entites[0].Proprietes))
	}
}

// TestExtraireTraitsSepareLesSignatures vérifie que deux horodatages
// différents ne partagent pas de trait.
//
// Une entité dont updated_at est facultative et une autre où elle est
// obligatoire ne peuvent pas utiliser le même : le mapping serait faux, et
// Doctrine ne s'en plaindrait qu'à l'exécution.
func TestExtraireTraitsSepareLesSignatures(t *testing.T) {
	t.Parallel()

	logique := &calque.Logique{
		Entites: []calque.Entite{
			entiteHorodatee("Client",
				horodatage("createdAt", "created_at", false),
				horodatage("updatedAt", "updated_at", true)),
			entiteHorodatee("Commande",
				horodatage("createdAt", "created_at", false),
				horodatage("updatedAt", "updated_at", false)),
			entiteHorodatee("Facture",
				horodatage("createdAt", "created_at", false),
				horodatage("updatedAt", "updated_at", false)),
		},
	}

	extraireTraits(logique)

	if len(logique.Traits) != 1 {
		t.Fatalf("traits = %d, attendu un seul : Client est seule de sa signature", len(logique.Traits))
	}
	if len(logique.Entites[0].Traits) != 0 {
		t.Errorf("Client a recu le trait d'une autre signature : %v", logique.Entites[0].Traits)
	}
	if len(logique.Entites[1].Traits) != 1 || len(logique.Entites[2].Traits) != 1 {
		t.Error("Commande et Facture partagent une signature et devraient partager le trait")
	}
}

// TestExtraireTraitsNeSortJamaisLaCle vérifie qu'une colonne d'horodatage
// utilisée comme clé primaire reste sur l'entité.
//
// Le cas existe sur du legacy — une table d'historique dont la clé est la date
// —, et sortir la clé dans un trait rendrait l'entité ingénérable.
func TestExtraireTraitsNeSortJamaisLaCle(t *testing.T) {
	t.Parallel()

	horodatee := func(nom string) calque.Entite {
		return calque.Entite{
			Nom:         nom,
			Table:       calque.ReferenceTable{Nom: nom, Schema: "public"},
			Identifiant: &calque.Identifiant{Proprietes: []string{"createdAt"}},
			Proprietes:  []calque.Propriete{horodatage("createdAt", "created_at", false)},
		}
	}

	logique := &calque.Logique{Entites: []calque.Entite{horodatee("Trace"), horodatee("Audit")}}

	extraireTraits(logique)

	if len(logique.Traits) != 0 {
		t.Errorf("traits = %+v, attendu aucun : la propriete est la cle", logique.Traits)
	}
	for i := range logique.Entites {
		if len(logique.Entites[i].Proprietes) != 1 {
			t.Errorf("%s a perdu sa cle", logique.Entites[i].Nom)
		}
	}
}

// TestExtraireTraitsLaisseLeMetier vérifie qu'un horodatage métier reste une
// propriété.
//
// date_facture est un horodatage au même titre que created_at, mais il porte du
// sens : le sortir dans un trait technique le rendrait introuvable.
func TestExtraireTraitsLaisseLeMetier(t *testing.T) {
	t.Parallel()

	logique := &calque.Logique{
		Entites: []calque.Entite{
			entiteHorodatee("Facture", horodatage("dateFacture", "date_facture", false)),
			entiteHorodatee("Avoir", horodatage("dateFacture", "date_facture", false)),
		},
	}

	extraireTraits(logique)

	if len(logique.Traits) != 0 {
		t.Errorf("traits = %+v, attendu aucun sur un horodatage metier", logique.Traits)
	}
}

// TestNomDeTraitDepartageLesSignatures vérifie le nommage du second trait.
//
// Deux signatures dans le même calque donnent deux fichiers, et un Horodatage2
// n'apprendrait rien sur ce qui les sépare.
func TestNomDeTraitDepartageLesSignatures(t *testing.T) {
	t.Parallel()

	premier := nomDeTrait([]calque.Propriete{{Nom: "createdAt"}}, nil)
	if premier != "Horodatage" {
		t.Errorf("premier trait nomme %q, attendu Horodatage", premier)
	}

	deja := []calque.Trait{{Nom: "Horodatage"}}
	second := nomDeTrait([]calque.Propriete{{Nom: "creeLe"}, {Nom: "supprimeLe"}}, deja)
	if second != "CreeLeSupprimeLe" {
		t.Errorf("second trait nomme %q, attendu CreeLeSupprimeLe", second)
	}
}
