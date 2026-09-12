// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sprimault/ormeau/internal/calque"
	"github.com/sprimault/ormeau/internal/introspection"
)

// dsnDeTest est celui qu'on confie au registre. Aucun test ne s'y connecte : ce
// qui compte est qu'il ne ressorte de nulle part.
const dsnDeTest = "postgres://gescom:secret@bdd-interne:5432/gescom"

// piloteDeTest ne simule aucun catalogue : il ne sert qu'à vérifier que le
// registre ouvre, retrouve et referme ce qu'on lui confie, et que le point
// d'entrée d'inventaire transmet ce qu'on lui demande. Ce que rend un vrai
// pilote se teste contre un vrai serveur, jamais contre un faux catalogue.
type piloteDeTest struct {
	ferme int
	echec error
	// schemasRecus enregistre le dernier appel, pour vérifier ce que la couche
	// HTTP a transmis — et non ce que le catalogue aurait répondu.
	schemasRecus []string
	sommaires    []introspection.TableSommaire
	echecListe   error
	// enCours et maxSimultanes mesurent le parallélisme réellement atteint.
	enCours       atomic.Int32
	maxSimultanes atomic.Int32
}

// Inventorier enregistre les schémas demandés et rend ce qui a été préparé.
//
// Compte aussi les appels simultanés : une connexion de base de données n'en
// accepte qu'un à la fois, et c'est le seul moyen de vérifier que la couche HTTP
// les sérialise vraiment.
func (p *piloteDeTest) Inventorier(_ context.Context, schemas []string) ([]introspection.TableSommaire, error) {
	if simultanes := p.enCours.Add(1); simultanes > p.maxSimultanes.Load() {
		p.maxSimultanes.Store(simultanes)
	}
	defer p.enCours.Add(-1)

	// Laisse le temps à un appel concurrent d'entrer, s'il n'est pas retenu.
	time.Sleep(5 * time.Millisecond)

	p.schemasRecus = schemas
	return p.sommaires, p.echecListe
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

// piloteListeur ajoute au précédent la capacité d'énumérer les bases, que tous
// les dialectes n'ont pas.
type piloteListeur struct {
	piloteDeTest
	bases []string
	echec error
}

// ListerBases rend ce qui a été préparé.
func (p *piloteListeur) ListerBases(context.Context) ([]string, error) {
	return p.bases, p.echec
}

// TestRegistreOuvreEtRetrouve vérifie le cycle nominal d'une connexion.
func TestRegistreOuvreEtRetrouve(t *testing.T) {
	t.Parallel()

	r := nouveauRegistre()
	pilote := &piloteDeTest{}

	id, err := r.ajouter(pilote, dsnDeTest, "postgres", false)
	if err != nil {
		t.Fatalf("ajouter: %v", err)
	}

	retrouve, ok := r.trouver(id)
	if !ok || retrouve.pilote != pilote {
		t.Fatal("connexion introuvable après ajout")
	}
	if retrouve.dsn != dsnDeTest || retrouve.sgbd != "postgres" {
		t.Error("le registre a perdu de quoi rouvrir sur une autre base")
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
	premier, err := r.ajouter(&piloteDeTest{}, dsnDeTest, "postgres", false)
	if err != nil {
		t.Fatalf("ajouter: %v", err)
	}
	second, err := r.ajouter(&piloteDeTest{}, dsnDeTest, "postgres", false)
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
	id, err := r.ajouter(pilote, dsnDeTest, "postgres", false)
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
	id, err := r.ajouter(&piloteDeTest{echec: panne}, dsnDeTest, "postgres", false)
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
		if _, err := r.ajouter(&piloteDeTest{}, dsnDeTest, "postgres", false); err != nil {
			t.Fatalf("connexion %d refusée : %v", i, err)
		}
	}

	if _, err := r.ajouter(&piloteDeTest{}, dsnDeTest, "postgres", false); !errors.Is(err, ErrTropDeConnexions) {
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
		if _, err := r.ajouter(p, dsnDeTest, "postgres", false); err != nil {
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
