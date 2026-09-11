// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// tailleMax borne les largeurs et hauteurs de panneaux. Une valeur au-delà ne
// vient pas d'un glissement de séparateur mais d'un fichier retouché à la main,
// et poserait un panneau hors de l'écran sans moyen de le rattraper.
const tailleMax = 10000

// Themes et Langues sont les vocabulaires fermés que le front connaît. Une
// valeur hors liste est refusée plutôt que transmise : le front n'a pas de
// branche pour elle et rendrait une interface sans thème.
var (
	themes  = []string{"clair", "sombre", "systeme"}
	langues = []string{"fr", "en"}
)

// Preferences porte ce que l'interface affichait au dernier lancement.
//
// Elles vivent ici et non dans le navigateur parce que le port d'écoute est
// tiré à chaque lancement : l'origine change, et avec elle tout le
// localStorage. Un thème choisi ne survivait donc pas à la fermeture.
//
// Une taille nulle vaut « jamais réglée », et le front garde alors sa valeur
// par défaut. Dupliquer ici les défauts de mise en page en ferait deux sources
// à tenir d'accord.
type Preferences struct {
	Theme          string `yaml:"theme"`
	Langue         string `yaml:"langue"`
	ApercuOuvert   *bool  `yaml:"apercu_ouvert,omitempty"`
	LargeurArbre   int    `yaml:"largeur_arbre,omitempty"`
	HauteurApercu  int    `yaml:"hauteur_apercu,omitempty"`
	LargeurEntites int    `yaml:"largeur_entites,omitempty"`
}

// PreferencesParDefaut rend ce qu'on affiche à un premier lancement : le thème
// du poste, et le français.
func PreferencesParDefaut() Preferences {
	return Preferences{Theme: "systeme", Langue: "fr"}
}

// LirePreferences rend les préférences enregistrées, un avertissement, et une
// erreur.
//
// Les trois cas ne se confondent pas. Fichier absent : les valeurs par défaut,
// sans rien signaler — c'est un premier lancement, pas un incident. Fichier
// illisible, mal formé, ou portant une valeur hors vocabulaire : les valeurs
// par défaut et un avertissement, que l'appelant journalise. Une préférence
// d'affichage ne vaut pas un refus de démarrer.
//
// L'erreur reste pour ce qui empêche de lire, sans dire que le contenu est
// mauvais : un répertoire devenu inaccessible entre l'ouverture et ici.
//
// Ce qui est rendu est toujours utilisable : la validation a lieu ici, et non
// chez l'appelant, sinon chacun la refait — ou l'oublie.
func (e *Emplacements) LirePreferences() (Preferences, string, error) {
	defauts := PreferencesParDefaut()

	contenu, err := os.ReadFile(e.FichierPreferences())
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return defauts, "", nil
	case errors.Is(err, fs.ErrPermission):
		return defauts, "", fmt.Errorf("lecture de %s: %w", e.FichierPreferences(), err)
	case err != nil:
		return defauts, fmt.Sprintf("%s illisible, valeurs par défaut appliquées : %v",
			e.FichierPreferences(), err), nil
	}

	var lues Preferences
	if err := yaml.Unmarshal(contenu, &lues); err != nil {
		return defauts, fmt.Sprintf("%s mal formé, valeurs par défaut appliquées : %v",
			e.FichierPreferences(), err), nil
	}

	return valider(lues, defauts)
}

// EcrirePreferences enregistre les préférences, après les avoir validées.
//
// Valider avant d'écrire plutôt qu'à la relecture seulement : un fichier qu'on
// vient de poser et que la lecture suivante rejettera est une incohérence qu'on
// passerait du temps à comprendre.
func (e *Emplacements) EcrirePreferences(p Preferences) error {
	retenues, avertissement, _ := valider(p, PreferencesParDefaut())
	if avertissement != "" {
		return fmt.Errorf("preferences invalides: %s", avertissement)
	}

	contenu, err := yaml.Marshal(retenues)
	if err != nil {
		return fmt.Errorf("serialisation des preferences: %w", err)
	}
	if err := os.WriteFile(e.FichierPreferences(), contenu, permFichier); err != nil {
		return fmt.Errorf("ecriture de %s: %w", e.FichierPreferences(), err)
	}
	return nil
}

// valider remplace chaque valeur hors vocabulaire ou hors bornes par son défaut
// et rend un avertissement les nommant toutes.
//
// Toutes, et pas la première : corriger un fichier à la main pour se voir
// signaler la faute suivante au lancement d'après use la patience plus vite que
// la liste ne s'allonge.
func valider(p, defauts Preferences) (Preferences, string, error) {
	var fautes []string

	if p.Theme == "" {
		p.Theme = defauts.Theme
	} else if !slices.Contains(themes, p.Theme) {
		fautes = append(fautes, fmt.Sprintf("theme %q inconnu", p.Theme))
		p.Theme = defauts.Theme
	}

	if p.Langue == "" {
		p.Langue = defauts.Langue
	} else if !slices.Contains(langues, p.Langue) {
		fautes = append(fautes, fmt.Sprintf("langue %q inconnue", p.Langue))
		p.Langue = defauts.Langue
	}

	tailles := []struct {
		nom    string
		valeur *int
	}{
		{"largeur_arbre", &p.LargeurArbre},
		{"hauteur_apercu", &p.HauteurApercu},
		{"largeur_entites", &p.LargeurEntites},
	}
	for _, t := range tailles {
		if *t.valeur < 0 || *t.valeur > tailleMax {
			fautes = append(fautes, fmt.Sprintf("%s hors bornes (%d)", t.nom, *t.valeur))
			*t.valeur = 0
		}
	}

	if len(fautes) == 0 {
		return p, "", nil
	}
	return p, "préférences ignorées : " + strings.Join(fautes, ", "), nil
}
