// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package introspection

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
)

// piloteFactice vérifie le registre, pas un dialecte : on ne simule jamais un
// catalogue, c'est un vrai SGBD ou un calque figé.
type piloteFactice struct{}

// Seul le routage du registre est sous test.
func (piloteFactice) Inventorier(context.Context, []string) ([]TableSommaire, error) {
	return nil, nil
}

// Idem.
func (piloteFactice) Extraire(context.Context, Portee) (*calque.Physique, error) {
	return nil, nil
}

// Rien à libérer.
func (piloteFactice) Fermer() error { return nil }

// Ces tests écrivent dans le registre global : pas de t.Parallel().
func TestOuvrirUtiliseLePiloteEnregistre(t *testing.T) {
	sgbd := "sgbd-de-test-ouvrir"
	Enregistrer(sgbd, func(context.Context, string) (Introspecteur, error) {
		return piloteFactice{}, nil
	})
	t.Cleanup(func() { delete(fabriques, sgbd) })

	obtenu, err := Ouvrir(context.Background(), sgbd, "dsn-quelconque")
	if err != nil {
		t.Fatalf("ouverture : %v", err)
	}
	if _, ok := obtenu.(piloteFactice); !ok {
		t.Errorf("pilote inattendu : %T", obtenu)
	}
}

// L'erreur doit nommer le SGBD : c'est ce qui aide à corriger l'appel.
func TestOuvrirSGBDInconnu(t *testing.T) {
	_, err := Ouvrir(context.Background(), "sgbd-jamais-enregistre", "dsn")
	if err == nil {
		t.Fatal("aucune erreur pour un SGBD sans pilote")
	}
	if !strings.Contains(err.Error(), "sgbd-jamais-enregistre") {
		t.Errorf("l'erreur ne nomme pas le SGBD demandé : %v", err)
	}
}

// Le DSN est le seul secret manipulé.
func TestOuvrirNeDivulguePasLeDSN(t *testing.T) {
	const dsn = "postgres://utilisateur:motdepasse-secret@hote:5432/base"

	_, err := Ouvrir(context.Background(), "sgbd-jamais-enregistre", dsn)
	if err == nil {
		t.Fatal("aucune erreur")
	}
	if strings.Contains(err.Error(), "motdepasse-secret") {
		t.Errorf("le mot de passe apparaît dans l'erreur : %v", err)
	}
}

// L'erreur du pilote remonte telle quelle : l'appelant doit savoir pourquoi.
func TestOuvrirRemonteLErreurDuPilote(t *testing.T) {
	sgbd := "sgbd-de-test-erreur"
	attendue := errors.New("connexion refusee")
	Enregistrer(sgbd, func(context.Context, string) (Introspecteur, error) {
		return nil, attendue
	})
	t.Cleanup(func() { delete(fabriques, sgbd) })

	_, err := Ouvrir(context.Background(), sgbd, "dsn")
	if !errors.Is(err, attendue) {
		t.Errorf("erreur remontée %v, attendue %v", err, attendue)
	}
}

// Le dernier enregistrement gagne, plutôt qu'un ordre d'itération de map.
func TestEnregistrerRemplaceLeMemeSGBD(t *testing.T) {
	sgbd := "sgbd-de-test-remplacement"
	Enregistrer(sgbd, func(context.Context, string) (Introspecteur, error) {
		return nil, errors.New("premier")
	})
	Enregistrer(sgbd, func(context.Context, string) (Introspecteur, error) {
		return nil, errors.New("second")
	})
	t.Cleanup(func() { delete(fabriques, sgbd) })

	_, err := Ouvrir(context.Background(), sgbd, "dsn")
	if err == nil || err.Error() != "second" {
		t.Errorf("le second enregistrement n'a pas remplacé le premier : %v", err)
	}
}

// --echantillonner est la seule option qui lit des données : jamais par défaut.
func TestPorteeParDefautNEchantillonnePas(t *testing.T) {
	t.Parallel()

	var p Portee
	if p.Echantillonner {
		t.Error("l'échantillonnage est actif par défaut")
	}
	if p.CardinaliteMax != 0 {
		t.Errorf("plafond de cardinalité par défaut : %d", p.CardinaliteMax)
	}
}
