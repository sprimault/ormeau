// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"context"
	"errors"
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
	"github.com/sprimault/ormeau/internal/introspection"
)

// piloteDeTest ne simule aucun catalogue : il ne sert qu'à vérifier que le
// registre ouvre, retrouve et referme ce qu'on lui confie. Ce que fait un vrai
// pilote se teste contre un vrai serveur, jamais contre un faux catalogue.
type piloteDeTest struct {
	ferme int
	echec error
}

// Inventorier n'est pas exercé ici.
func (p *piloteDeTest) Inventorier(context.Context, []string) ([]introspection.TableSommaire, error) {
	return nil, nil
}

// Extraire n'est pas exercé ici.
func (p *piloteDeTest) Extraire(context.Context, introspection.Portee) (*calque.Physique, error) {
	return nil, nil
}

// Fermer compte les fermetures, ce que les tests vérifient.
func (p *piloteDeTest) Fermer() error {
	p.ferme++
	return p.echec
}

// TestRegistreOuvreEtRetrouve vérifie le cycle nominal d'une connexion.
func TestRegistreOuvreEtRetrouve(t *testing.T) {
	t.Parallel()

	r := nouveauRegistre()
	pilote := &piloteDeTest{}

	id, err := r.ajouter(pilote)
	if err != nil {
		t.Fatalf("ajouter: %v", err)
	}

	retrouve, ok := r.trouver(id)
	if !ok || retrouve != pilote {
		t.Fatal("connexion introuvable après ajout")
	}
	if _, ok := r.trouver("inconnu"); ok {
		t.Error("un identifiant inconnu a rendu une connexion")
	}
}

// TestRegistreIdentifiantsImprevisibles vérifie que deux connexions ne
// reçoivent pas des identifiants voisins : deviner celui du voisin reviendrait
// à emprunter sa base.
func TestRegistreIdentifiantsImprevisibles(t *testing.T) {
	t.Parallel()

	r := nouveauRegistre()
	premier, err := r.ajouter(&piloteDeTest{})
	if err != nil {
		t.Fatalf("ajouter: %v", err)
	}
	second, err := r.ajouter(&piloteDeTest{})
	if err != nil {
		t.Fatalf("ajouter: %v", err)
	}
	if premier == second {
		t.Error("deux connexions partagent le même identifiant")
	}
}

// TestRegistreFermeEtOublie vérifie que fermer libère le pilote, et que fermer
// deux fois reste sans effet — un onglet rechargé poste la même fermeture.
func TestRegistreFermeEtOublie(t *testing.T) {
	t.Parallel()

	r := nouveauRegistre()
	pilote := &piloteDeTest{}
	id, err := r.ajouter(pilote)
	if err != nil {
		t.Fatalf("ajouter: %v", err)
	}

	if err := r.fermer(id); err != nil {
		t.Fatalf("fermer: %v", err)
	}
	if _, ok := r.trouver(id); ok {
		t.Error("connexion encore présente après fermeture")
	}
	if err := r.fermer(id); err != nil {
		t.Errorf("seconde fermeture : %v", err)
	}
	if pilote.ferme != 1 {
		t.Errorf("pilote fermé %d fois, attendu 1", pilote.ferme)
	}
}

// TestRegistreRemonteLEchecDeFermeture vérifie que l'erreur du pilote n'est pas
// avalée : l'appelant la journalise.
func TestRegistreRemonteLEchecDeFermeture(t *testing.T) {
	t.Parallel()

	panne := errors.New("connexion deja coupee")
	r := nouveauRegistre()
	id, err := r.ajouter(&piloteDeTest{echec: panne})
	if err != nil {
		t.Fatalf("ajouter: %v", err)
	}
	if err := r.fermer(id); !errors.Is(err, panne) {
		t.Errorf("erreur rendue %v, attendue %v", err, panne)
	}
}

// TestRegistrePlafonne vérifie qu'un formulaire rejoué en boucle ne peut pas
// accumuler des connexions ouvertes sur la base d'un client.
func TestRegistrePlafonne(t *testing.T) {
	t.Parallel()

	r := nouveauRegistre()
	for i := range maxConnexions {
		if _, err := r.ajouter(&piloteDeTest{}); err != nil {
			t.Fatalf("connexion %d refusée : %v", i, err)
		}
	}

	if _, err := r.ajouter(&piloteDeTest{}); !errors.Is(err, ErrTropDeConnexions) {
		t.Errorf("erreur rendue %v, attendue ErrTropDeConnexions", err)
	}
}

// TestRegistreToutFermer vérifie qu'aucune connexion ne survit à l'arrêt du
// serveur.
func TestRegistreToutFermer(t *testing.T) {
	t.Parallel()

	r := nouveauRegistre()
	pilotes := []*piloteDeTest{{}, {}, {}}
	for _, p := range pilotes {
		if _, err := r.ajouter(p); err != nil {
			t.Fatalf("ajouter: %v", err)
		}
	}

	r.toutFermer()

	for i, p := range pilotes {
		if p.ferme != 1 {
			t.Errorf("pilote %d fermé %d fois, attendu 1", i, p.ferme)
		}
	}
	if len(r.parID) != 0 {
		t.Errorf("%d connexion(s) restante(s)", len(r.parID))
	}
}
