// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"strings"
	"testing"

	"github.com/sprimault/ormeau/internal/introspection"
)

// Ces tests ne joignent aucune base : ce qui exige un serveur porte l'étiquette
// integration et vit dans postgres_integration_test.go.

// Sans l'init, `ormeau extraire` échouerait sur « aucun pilote » et rien à la
// compilation ne le signalerait.
func TestPiloteEnregistreAuChargement(t *testing.T) {
	t.Parallel()

	_, err := introspection.Ouvrir(context.Background(), "postgres", "?")
	if err != nil && strings.Contains(err.Error(), "aucun pilote") {
		t.Error("le pilote postgres n'est pas enregistré dans le registre")
	}
}

// Un DSN illisible est rejeté avant toute tentative de connexion : inutile
// d'attendre un délai réseau pour une chaîne qui ne peut pas marcher.
func TestOuvrirRefuseUnDSNIllisible(t *testing.T) {
	t.Parallel()

	if _, err := Ouvrir(context.Background(), "postgres://u:p@hote:pas-un-port/base"); err == nil {
		t.Fatal("aucune erreur sur un dsn illisible")
	}
}

// Le DSN ne transite ni dans les journaux, ni dans le calque, ni dans une erreur.
func TestOuvrirNeDivulguePasLeDSN(t *testing.T) {
	t.Parallel()

	const secret = "motdepasse-secret"

	_, err := Ouvrir(context.Background(), "postgres://u:"+secret+"@hote:pas-un-port/base")
	if err == nil {
		t.Fatal("aucune erreur")
	}
	if strings.Contains(err.Error(), secret) {
		t.Errorf("le mot de passe apparait dans l'erreur : %v", err)
	}
}

// Un intermédiaire qui ignore les paramètres de démarrage ne produit aucune
// erreur : seule la relecture le révèle, et elle doit refuser plutôt que
// laisser extraire.
func TestVerifierSession(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom, chemin, lectureSeule string
		refus                     string
	}{
		{"parametres appliques", "", "on", ""},
		{"lecture seule perdue", "", "off", "inscriptible"},
		{"chemin perdu", `"$user", public`, "on", "search_path"},
		{"tout perdu", `"$user", public`, "off", "inscriptible"},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			err := verifierSession(c.chemin, c.lectureSeule)
			switch {
			case c.refus == "" && err != nil:
				t.Errorf("refus inattendu : %v", err)
			case c.refus != "" && err == nil:
				t.Error("session acceptée")
			case c.refus != "" && !strings.Contains(err.Error(), c.refus):
				t.Errorf("refus %q, attendu une mention de %q", err, c.refus)
			}
		})
	}
}

// Vérifié à la compilation, mais énoncé ici pour qu'un changement d'interface
// échoue sur un test nommé.
func TestPiloteSatisfaitLInterface(t *testing.T) {
	t.Parallel()

	var _ introspection.Introspecteur = (*pilote)(nil)
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
