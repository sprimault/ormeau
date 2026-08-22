// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/sprimault/ormeau/internal/introspection"
)

// Sans l'init, `ormeau extraire` échouerait sur « aucun pilote » et rien à la
// compilation ne le signalerait.
func TestPiloteEnregistreAuChargement(t *testing.T) {
	t.Parallel()

	_, err := introspection.Ouvrir(context.Background(), "postgres", "postgres://hote/base")
	if err == nil {
		t.Fatal("aucune erreur alors que le pilote n'est pas écrit")
	}
	if strings.Contains(err.Error(), "aucun pilote") {
		t.Error("le pilote postgres n'est pas enregistré dans le registre")
	}
}

// Une erreur retournée, pas un panic qui remonterait jusqu'à main.
func TestOuvrirRetourneUneErreur(t *testing.T) {
	t.Parallel()

	pilote, err := Ouvrir(context.Background(), "postgres://hote/base")
	if !errors.Is(err, errNonImplemente) {
		t.Errorf("erreur %v, attendue %v", err, errNonImplemente)
	}
	if pilote != nil {
		t.Errorf("pilote non nul rendu avec une erreur : %T", pilote)
	}
}

// Le DSN ne transite ni dans les journaux, ni dans le calque, ni dans une erreur.
func TestOuvrirNeDivulguePasLeDSN(t *testing.T) {
	t.Parallel()

	const dsn = "postgres://utilisateur:motdepasse-secret@hote:5432/base"

	_, err := Ouvrir(context.Background(), dsn)
	if err == nil {
		t.Fatal("aucune erreur")
	}
	for _, fragment := range []string{"motdepasse-secret", "utilisateur", "hote:5432"} {
		if strings.Contains(err.Error(), fragment) {
			t.Errorf("l'erreur divulgue %q : %v", fragment, err)
		}
	}
}

// Fermer fait exception : sans connexion ouverte, il n'y a rien à libérer.
func TestMethodesRetournentUneErreur(t *testing.T) {
	t.Parallel()

	p := &Pilote{}

	if _, err := p.Inventorier(context.Background(), []string{"public"}); !errors.Is(err, errNonImplemente) {
		t.Errorf("Inventorier : %v", err)
	}
	if _, err := p.Extraire(context.Background(), introspection.Portee{}); !errors.Is(err, errNonImplemente) {
		t.Errorf("Extraire : %v", err)
	}
	if err := p.Fermer(); err != nil {
		t.Errorf("Fermer sur un pilote sans connexion : %v", err)
	}
}

// Vérifié à la compilation, mais énoncé ici pour qu'un changement d'interface
// échoue sur un test nommé.
func TestPiloteSatisfaitLInterface(t *testing.T) {
	t.Parallel()

	var _ introspection.Introspecteur = (*Pilote)(nil)
}

// Constantes, jamais assemblées : l'injection est structurellement impossible.
// Et pg_catalog, pas information_schema, qui perd trop.
func TestRequetesDeCatalogue(t *testing.T) {
	t.Parallel()

	requetes := map[string]string{
		"colonnes":       requeteColonnes,
		"contraintes":    requeteContraintes,
		"index":          requeteIndex,
		"types énumérés": requeteTypesEnumeres,
		"inventaire":     requeteInventaire,
	}

	for nom, requete := range requetes {
		t.Run(nom, func(t *testing.T) {
			t.Parallel()

			if strings.Contains(strings.ToLower(requete), "information_schema") {
				t.Error("la requête passe par information_schema plutôt que pg_catalog")
			}
			if !strings.Contains(requete, "$1") {
				t.Error("les schémas ne sont pas passés en paramètre")
			}
			if !strings.Contains(strings.ToUpper(requete), "ORDER BY") {
				t.Error("aucun ORDER BY : l'ordre des lignes rendrait l'extraction non déterministe")
			}
			for _, ecriture := range []string{"INSERT ", "UPDATE ", "DELETE ", "DROP ", "ALTER ", "CREATE "} {
				if strings.Contains(strings.ToUpper(requete), ecriture) {
					t.Errorf("la requête contient %q alors que la connexion est en lecture seule", ecriture)
				}
			}
		})
	}
}
