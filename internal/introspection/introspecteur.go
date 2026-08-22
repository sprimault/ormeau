// Package introspection lit les catalogues natifs des SGBD et produit un calque
// physique. Aucun paquet d'ici ne connaît le moindre ORM.
package introspection

import (
	"context"
	"fmt"

	"github.com/sprimault/ormeau/internal/calque"
)

// Introspecteur est volontairement minimal. Pas de couche d'abstraction SQL
// partagée entre dialectes : elle finirait par contraindre chaque pilote au
// plus petit dénominateur commun, exactement ce que la lecture des catalogues
// natifs cherche à éviter. La duplication entre pilotes est assumée.
type Introspecteur interface {
	// Inventorier est une passe légère, une requête par SGBD, qui alimente
	// l'arbre de sélection de l'interface. Sans elle, afficher la liste des
	// tables imposerait une extraction complète avant même que l'utilisateur
	// ait choisi quoi que ce soit.
	Inventorier(ctx context.Context, schemas []string) ([]TableSommaire, error)
	Extraire(ctx context.Context, portee Portee) (*calque.Physique, error)
	Fermer() error
}

// TableSommaire est ce qu'on sait d'une table sans l'avoir introspectée. Aucune
// donnée métier : LignesEstimees vient des statistiques du catalogue, pas d'un
// COUNT.
type TableSommaire struct {
	Schema         string `json:"schema"`
	Nom            string `json:"nom"`
	Commentaire    string `json:"commentaire,omitempty"`
	NbColonnes     int    `json:"nb_colonnes"`
	LignesEstimees int64  `json:"lignes_estimees"`
	ClePrimaire    bool   `json:"cle_primaire"`
	// ReferenceVers liste les tables qualifiées atteintes par clé étrangère.
	// L'interface s'en sert pour proposer les dépendances quand on coche une
	// table, et pour signaler les références qui sortiraient de la sélection.
	ReferenceVers []string `json:"reference_vers,omitempty"`
}

type Portee struct {
	Schemas        []string
	TablesIncluses []string
	TablesExclues  []string
	// Echantillonner est la seule option qui autorise la lecture de données.
	// Elle alimente la détection d'énumérations et des clés étrangères
	// implicites, cas majoritaire sur du legacy.
	Echantillonner bool
	// CardinaliteMax plafonne l'échantillonnage : au-delà, une colonne produit
	// une statistique, pas un échantillon.
	CardinaliteMax int
}

type Fabrique func(ctx context.Context, dsn string) (Introspecteur, error)

var fabriques = map[string]Fabrique{}

// Enregistrer est appelé par l'init de chaque pilote.
func Enregistrer(sgbd string, f Fabrique) {
	fabriques[sgbd] = f
}

func Ouvrir(ctx context.Context, sgbd, dsn string) (Introspecteur, error) {
	f, ok := fabriques[sgbd]
	if !ok {
		return nil, fmt.Errorf("aucun pilote pour %q", sgbd)
	}
	return f(ctx, dsn)
}
