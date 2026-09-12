// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// tailleEmpreinteCle est la longueur du suffixe qui distingue deux projets
// travaillant la même base.
//
// Seize et non huit : trente-deux bits laisseraient une collision possible, et
// son mode d'échec est mauvais — le brouillon d'un projet rendu à un autre, en
// silence, puisque les empreintes de contenu pourraient correspondre si les
// deux travaillent la même base.
const tailleEmpreinteCle = 16

// Session est un brouillon d'arbitrage, tel qu'on le retrouve en rouvrant
// l'écran.
//
// Elle ne retient que ce qui n'est pas encore dans le fichier de décisions :
// celui-ci fait foi, et une session rejouée par-dessus une version antérieure
// réintroduirait en silence des décisions retirées à la main.
//
// Jetable par définition — c'est ce que dit le LISEZMOI d'etat/. L'effacer ne
// perd qu'un travail non enregistré.
type Session struct {
	Base string `json:"base"`
	// EmpreinteCalque et EmpreinteDecisions disent de quel état ce brouillon
	// découle. L'une des deux qui ne correspond plus rend la session caduque.
	EmpreinteCalque    string `json:"empreinte_calque"`
	EmpreinteDecisions string `json:"empreinte_decisions"`
	// Brouillon est opaque ici, et une chaîne plutôt qu'un json.RawMessage : ce
	// paquet ne connaît pas le vocabulaire des décisions et n'a pas à le
	// connaître pour le ranger. Une chaîne le dit, là où un message brut
	// laisserait croire que la forme est vérifiée quelque part.
	Brouillon string `json:"brouillon,omitempty"`
	// EntiteOuverte est la table qualifiée que l'écran montrait.
	EntiteOuverte string `json:"entite_ouverte,omitempty"`
}

// LireSession rend le brouillon d'une base pour un répertoire de travail donné.
//
// Une session absente n'est pas une erreur : c'est le cas de toute base qu'on
// ouvre pour la première fois.
func (e *Emplacements) LireSession(base, repertoire string) (*Session, error) {
	contenu, err := os.ReadFile(e.fichierSession(base, repertoire)) // #nosec G304
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("lecture de la session: %w", err)
	}

	var s Session
	if err := json.Unmarshal(contenu, &s); err != nil {
		// Un brouillon illisible se jette : il n'a pas plus de valeur que le
		// travail non enregistré qu'il portait, et le redemander à l'utilisateur
		// coûterait plus que de repartir de zéro.
		return nil, nil
	}
	return &s, nil
}

// EcrireSession enregistre le brouillon d'une base.
func (e *Emplacements) EcrireSession(s Session, repertoire string) error {
	if !nomBase.MatchString(s.Base) {
		return fmt.Errorf("nom de base invalide: %q", s.Base)
	}

	contenu, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("serialisation de la session: %w", err)
	}
	if err := os.WriteFile(e.fichierSession(s.Base, repertoire), contenu, permFichier); err != nil {
		return fmt.Errorf("ecriture de la session: %w", err)
	}
	return nil
}

// SupprimerSession efface le brouillon d'une base.
//
// Un fichier déjà absent n'est pas une erreur : le résultat voulu est atteint.
func (e *Emplacements) SupprimerSession(base, repertoire string) error {
	err := os.Remove(e.fichierSession(base, repertoire))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("suppression de la session: %w", err)
	}
	return nil
}

// fichierSession compose le chemin du brouillon d'une base.
//
// Le nom porte la base en clair, pour qu'on reconnaisse le fichier avant de
// l'effacer, et l'empreinte du répertoire de travail, parce que deux projets
// peuvent chacun avoir une base « gescom » sans que leurs brouillons se
// confondent.
func (e *Emplacements) fichierSession(base, repertoire string) string {
	return filepath.Join(e.RepertoireSessions(),
		base+"-"+empreinteRepertoire(repertoire)+".json")
}

// empreinteRepertoire distingue deux répertoires de travail.
//
// Mis en minuscules sous Windows, où deux graphies du même chemin désignent le
// même dossier : sans cela, lancer depuis « C:\Projets » puis « c:\projets »
// donnerait deux brouillons pour un seul projet.
func empreinteRepertoire(repertoire string) string {
	nettoye := filepath.Clean(repertoire)
	if runtime.GOOS == "windows" {
		nettoye = strings.ToLower(nettoye)
	}
	somme := sha256.Sum256([]byte(nettoye))
	return hex.EncodeToString(somme[:])[:tailleEmpreinteCle]
}
