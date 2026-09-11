// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Package config résout les emplacements où Ormeau range ses propres données.
//
// Deux emplacements séparés par ce qu'ils contiennent, et surtout par qui en
// est propriétaire. Le répertoire de configuration de l'OS porte les données de
// la machine — préférences, profils de connexion, brouillons d'arbitrage — qui
// ne se versionnent ni ne se partagent. Le répertoire de travail porte les
// données du travail — les trois fichiers d'une base —, qui se commitent avec
// le projet. Ce paquet ne connaît que le premier.
//
// Jamais le répertoire du binaire : sous Program Files ou après un gestionnaire
// de paquets, il est en lecture seule.
//
// Feuille du graphe de dépendances, comme calque : il ne lit ni drapeau ni
// variable d'environnement, la racine lui est donnée.
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// permRepertoire vaut 0700 et non 0755 : profils.yaml porte des hôtes et des
// comptes de bases de production, que les autres comptes d'une machine Linux
// n'ont pas à lire.
const permRepertoire = 0o700

// permFichier suit la même logique que permRepertoire, un cran plus bas.
const permFichier = 0o600

// lisezMoi explique que le contenu d'etat/ est jetable. Sans lui, personne
// n'ose effacer un répertoire dont il ignore ce qu'il porte.
const lisezMoi = `Ce répertoire ne contient que des données de travail : brouillons d'arbitrage,
position dans les listes, filtres. Il peut être effacé sans rien perdre.

Les préférences et les profils de connexion sont dans le répertoire parent et ne
sont pas concernés.
`

// Emplacements donne les chemins de la configuration d'une machine, dont
// l'arborescence existe et est inscriptible.
//
// Les accesseurs disent leur nature — RepertoireEtat contre FichierProfils —
// parce que rien d'autre ne la porte : un chemin est une chaîne, et se tromper
// de nature ne se voit qu'à l'appel système qui échoue.
type Emplacements struct {
	racine string
}

// RacineParDefaut rend le répertoire de configuration de l'OS, suffixé du nom
// de l'outil. Il n'est pas créé.
//
// os.UserConfigDir couvre les trois plateformes sans table à maintenir :
// %AppData% sous Windows, $XDG_CONFIG_HOME ou ~/.config sous Linux,
// ~/Library/Application Support sous macOS.
func RacineParDefaut() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("repertoire de configuration de l'utilisateur: %w", err)
	}
	return filepath.Join(base, "ormeau"), nil
}

// Ouvrir crée l'arborescence sous racine si elle manque, vérifie qu'on peut y
// écrire, et rend ses chemins.
//
// Un répertoire qui existe déjà n'est pas resserré à 0700 : quelqu'un a pu
// l'ouvrir volontairement, et un changement de permissions qu'on n'a pas
// demandé se remarque au pire moment.
//
// Chaque erreur nomme le chemin exact qui a échoué. Une arborescence à moitié
// posée n'est pas un problème — la prochaine ouverture la reprend —, mais un
// refus qui ne dit pas où envoie chercher.
func Ouvrir(racine string) (*Emplacements, error) {
	if racine == "" {
		return nil, fmt.Errorf("racine de configuration vide")
	}

	// Absolue et nettoyée, comme le répertoire de travail et pour la même
	// raison : elle est affichée au démarrage, et un chemin relatif ou aux
	// séparateurs mélangés ne se recopie pas dans un explorateur.
	absolue, err := filepath.Abs(racine)
	if err != nil {
		return nil, fmt.Errorf("resolution de %s: %w", racine, err)
	}

	e := &Emplacements{racine: absolue}
	for _, d := range []string{e.racine, e.RepertoireEtat(), e.RepertoireSessions()} {
		if err := os.MkdirAll(d, permRepertoire); err != nil {
			return nil, fmt.Errorf("creation de %s: %w", d, err)
		}
	}

	// MkdirAll réussit sans rien vérifier quand le répertoire existe déjà.
	// Sans ce contrôle, un répertoire créé par un autre compte ou fermé par une
	// politique d'entreprise ne se découvrirait qu'à la première écriture de
	// profils.yaml, avec un message hors contexte.
	if err := verifierInscriptible(e.racine); err != nil {
		return nil, err
	}

	// Une politesse, pas une dépendance : un LISEZMOI qu'on n'a pas pu écrire
	// ne justifie pas de refuser de démarrer.
	chemin := filepath.Join(e.RepertoireEtat(), "LISEZMOI.txt")
	if _, err := os.Stat(chemin); os.IsNotExist(err) {
		_ = os.WriteFile(chemin, []byte(lisezMoi), permFichier)
	}

	return e, nil
}

// verifierInscriptible s'assure qu'un fichier peut être créé dans le
// répertoire.
//
// Par une écriture réelle, et non par les bits de permission : sous Windows ils
// ne décrivent pas les droits effectifs, et sous Linux ils ignorent les ACL
// comme le montage en lecture seule.
func verifierInscriptible(repertoire string) error {
	f, err := os.CreateTemp(repertoire, ".ormeau-*")
	if err != nil {
		return fmt.Errorf("ecriture dans %s: %w", repertoire, err)
	}
	nom := f.Name()
	_ = f.Close()
	_ = os.Remove(nom)
	return nil
}

// Racine rend le répertoire de configuration lui-même.
func (e *Emplacements) Racine() string { return e.racine }

// FichierPreferences rend le chemin des préférences d'affichage.
func (e *Emplacements) FichierPreferences() string {
	return filepath.Join(e.racine, "preferences.yaml")
}

// FichierProfils rend le chemin des connexions enregistrées.
func (e *Emplacements) FichierProfils() string {
	return filepath.Join(e.racine, "profils.yaml")
}

// RepertoireEtat rend le répertoire des données jetables.
func (e *Emplacements) RepertoireEtat() string {
	return filepath.Join(e.racine, "etat")
}

// RepertoireSessions rend le répertoire des brouillons d'arbitrage.
func (e *Emplacements) RepertoireSessions() string {
	return filepath.Join(e.RepertoireEtat(), "sessions")
}
