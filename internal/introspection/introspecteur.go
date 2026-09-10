// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

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

// TableSommaire et Portee vivent dans sommaire.go : ce sont les types que
// l'interface reçoit et renvoie, et c'est ce fichier que tygo traduit.

// ListeurDeBases est implémenté par les pilotes qui savent énumérer les bases
// d'un serveur.
//
// Volontairement hors d'Introspecteur, que trois méthodes suffisent à décrire :
// énumérer les bases n'est pas introspecter, et un dialecte où la notion n'a
// pas de sens ne doit pas être forcé de la porter. L'appelant fait une
// assertion de type et se passe de la fonctionnalité si elle manque.
type ListeurDeBases interface {
	// ListerBases rend les bases exploitables du serveur, celles du système
	// exclues : elles ne produiraient que des calques sans intérêt.
	ListerBases(ctx context.Context) ([]string, error)
}

// DescripteurServeur est implémenté par les pilotes qui savent se décrire sans
// rien introspecter.
//
// Hors d'Introspecteur pour la même raison que ListeurDeBases : se décrire n'est
// pas introspecter, et l'interface doit rester à trois méthodes. C'est ce que
// l'écran de connexion affiche en retour — le serveur atteint, sa version, et
// les schémas parmi lesquels choisir.
type DescripteurServeur interface {
	Decrire(ctx context.Context) (Serveur, error)
}

// Serveur est ce qu'on apprend d'un serveur en s'y connectant.
//
// SGBD porte la variante constatée et non le préfixe du DSN : un « mysql:// »
// vers un serveur MariaDB donne « mariadb ». Le préfixe dit quel pilote
// charger, le serveur dit ce qu'il est, et c'est lui qui a raison.
type Serveur struct {
	SGBD      string   `json:"sgbd"`
	Version   string   `json:"version"`
	Catalogue string   `json:"catalogue"`
	Schemas   []string `json:"schemas"`
}

// ListeurDeColonnes est implémenté par les pilotes qui savent décrire une table
// sans l'introspecter entièrement.
//
// Hors d'Introspecteur, comme les deux précédentes. L'écran de sélection s'en
// sert au dépliement d'une table : porter les colonnes dans l'inventaire
// multiplierait sa taille par dix pour des lignes que personne n'ouvrira.
type ListeurDeColonnes interface {
	Colonnes(ctx context.Context, schema, table string) ([]ColonneSommaire, error)
}

// Fabrique ouvre une connexion et rend l'introspecteur d'un dialecte.
type Fabrique func(ctx context.Context, dsn string) (Introspecteur, error)

// Registre plutôt que switch : un SGBD s'ajoute sans toucher au code commun,
// et seuls les pilotes importés sont liés.
var fabriques = map[string]Fabrique{}

// Enregistrer est appelé par l'init de chaque pilote.
func Enregistrer(sgbd string, f Fabrique) {
	fabriques[sgbd] = f
}

// Ouvrir rend l'introspecteur du SGBD demandé. L'erreur ne nomme que le SGBD,
// jamais le DSN.
func Ouvrir(ctx context.Context, sgbd, dsn string) (Introspecteur, error) {
	f, ok := fabriques[sgbd]
	if !ok {
		return nil, fmt.Errorf("aucun pilote pour %q", sgbd)
	}
	return f(ctx, dsn)
}
