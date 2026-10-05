// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Commande auditnpm fait échouer la validation sur un avis de sécurité des
// dépendances du front, sauf ceux qu'une exclusion nomme dans exclusions.go.
//
// Elle existe parce que npm n'offre aucun moyen d'écarter un avis : `npm audit`
// ne connaît qu'un seuil global, et le baisser laisserait passer tout le reste
// (npm/cli#8886, ouvert sans réponse de mainteneur). Il faut donc lire son
// rapport. Un avis sans version corrigée publiée — il en existe — rendrait
// sinon la validation définitivement rouge, et une CI rouge en permanence ne
// dit plus rien.
//
// Elle lit le rapport sur l'entrée standard :
//
//	npm audit --json --audit-level=critical | auditnpm
//
// Le seuil ne change pas le rapport, seulement le code de retour de npm
// (vérifié sous npm 10.9.8 : les sorties de `--json` nu, `=none` et
// `=critical` sont identiques). À `critical`, npm échoue seul sur un avis
// critique, quoi que fasse ce programme, et les avis « high » sont jugés ici.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// versionRapport est la seule version du document de `npm audit --json` que ce
// contrôle sait lire. npm 6 rendait une structure sans rapport avec celle-ci ;
// une version supérieure peut déplacer les avis, et les manquer en silence
// serait pire que de s'arrêter.
const versionRapport = 2

// gravitesJugees sont les gravités que ce contrôle refuse hors exclusion. Les
// deux autres — low et moderate — sont du bruit sur un arbre de développement
// de quatre cents paquets, et les traiter noierait le signal.
var gravitesJugees = map[string]bool{"high": true, "critical": true}

func main() {
	if err := controler(os.Stdin, os.Stdout, exclusions); err != nil {
		fmt.Fprintln(os.Stderr, "audit npm:", err)
		os.Exit(1)
	}
}

// rapport est la part du document de `npm audit --json` que ce contrôle lit.
type rapport struct {
	Version        int               `json:"auditReportVersion"`
	Vulnerabilites map[string]paquet `json:"vulnerabilities"`
}

// paquet est l'entrée d'un paquet touché.
//
// Via mêle deux formes : un objet quand le paquet porte l'avis, une chaîne
// quand il ne fait que le propager depuis une dépendance. Sur le rapport de ce
// dépôt, un seul des cinq paquets signalés porte un avis propre — d'où le
// détour par json.RawMessage. Un contrôle indexé par paquet, ou qui lirait la
// gravité de l'entrée de paquet, compterait cinq fois le même avis et ne
// pourrait en nommer aucun.
type paquet struct {
	Via []json.RawMessage `json:"via"`
}

// avis est une entrée Via de forme objet, seul endroit du rapport où
// l'identifiant GitHub de l'avis apparaisse.
type avis struct {
	Paquet  string `json:"name"`
	Titre   string `json:"title"`
	URL     string `json:"url"`
	Gravite string `json:"severity"`
	Plage   string `json:"range"`
}

// controler lit un rapport et rend une erreur dès que la validation doit
// échouer : un avis jugé que nommees ne couvre pas, ou une entrée de nommees
// qui ne correspond plus à rien. Le second cas est ce qui empêche la liste de
// pourrir — sans lui, une exclusion survivrait à l'avis qu'elle couvre, et la
// première personne à la relire ne saurait pas si elle sert encore.
//
// Les exclusions sont un paramètre et non la variable de paquet : c'est ce qui
// permet d'exercer les deux refus sur des tables choisies, sans que les cas
// dépendent de ce que le dépôt exclut aujourd'hui.
func controler(entree io.Reader, sortie io.Writer, nommees map[string]Exclusion) error {
	brut, err := io.ReadAll(entree)
	if err != nil {
		return fmt.Errorf("lecture du rapport: %w", err)
	}

	var r rapport
	if err := json.Unmarshal(brut, &r); err != nil {
		return fmt.Errorf("rapport illisible: %w", err)
	}
	// Un rapport tronqué, vide ou d'une autre version se décode sans erreur en
	// laissant les champs à zéro : il passerait pour « aucun avis ».
	if r.Version != versionRapport {
		return fmt.Errorf("auditReportVersion %d, ce contrôle lit la version %d", r.Version, versionRapport)
	}

	juges := make(map[string]avis)
	exclus := make(map[string]bool)
	for _, p := range r.Vulnerabilites {
		for _, entree := range p.Via {
			a, ok := decoderAvis(entree)
			if !ok || !gravitesJugees[a.Gravite] {
				continue
			}
			id := identifiant(a.URL)
			if _, nomme := nommees[id]; nomme {
				exclus[id] = true
				continue
			}
			juges[id] = a
		}
	}

	if len(juges) > 0 {
		return fmt.Errorf("avis non couvert par une exclusion:\n%s", lister(juges))
	}

	if inutiles := exclusionsInutiles(exclus, nommees); len(inutiles) > 0 {
		return fmt.Errorf("exclusion sans avis correspondant, à retirer de tools/auditnpm/exclusions.go: %s",
			strings.Join(inutiles, ", "))
	}

	for _, id := range triees(exclus) {
		if _, err := fmt.Fprintf(sortie, "avis exclu %s: %s\n", id, nommees[id].Raison); err != nil {
			return fmt.Errorf("ecriture du compte rendu: %w", err)
		}
	}
	return nil
}

// decoderAvis distingue les deux formes d'une entrée Via. Une entrée de forme
// chaîne ne se décode pas dans une structure, et c'est cette erreur de type qui
// sert de test : npm ne marque pas autrement la différence entre un avis et sa
// propagation.
func decoderAvis(brut json.RawMessage) (avis, bool) {
	var a avis
	if err := json.Unmarshal(brut, &a); err != nil {
		return avis{}, false
	}
	return a, true
}

// identifiant tire l'identifiant GitHub de l'URL de l'avis, qui est le seul
// champ à le porter. Une entrée sans URL rend la chaîne vide, qu'aucune
// exclusion ne nomme : l'avis est alors signalé comme les autres, puisqu'on ne
// saurait pas l'exclure.
func identifiant(url string) string {
	return url[strings.LastIndex(url, "/")+1:]
}

// exclusionsInutiles nomme les exclusions qu'aucun avis du rapport n'a
// rencontrées, triées.
func exclusionsInutiles(rencontrees map[string]bool, nommees map[string]Exclusion) []string {
	var inutiles []string
	for id := range nommees {
		if !rencontrees[id] {
			inutiles = append(inutiles, id)
		}
	}
	sort.Strings(inutiles)
	return inutiles
}

// lister rend les avis en une ligne chacune, triées par identifiant : la sortie
// ne doit pas dépendre de l'ordre de parcours d'une map.
func lister(juges map[string]avis) string {
	var lignes []string
	for _, id := range triees(juges) {
		a := juges[id]
		nom := id
		if nom == "" {
			nom = "avis sans identifiant"
		}
		lignes = append(lignes, fmt.Sprintf("  %s %s %s %s — %s", a.Gravite, a.Paquet, a.Plage, nom, a.Titre))
	}
	return strings.Join(lignes, "\n")
}

// triees rend les clés d'une map, triées.
func triees[V any](m map[string]V) []string {
	cles := make([]string, 0, len(m))
	for cle := range m {
		cles = append(cles, cle)
	}
	sort.Strings(cles)
	return cles
}
