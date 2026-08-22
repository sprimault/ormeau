// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package calque

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// Serialiser produit une sortie stable : deux extractions identiques doivent
// donner deux fichiers identiques octet pour octet, sinon le mode diff est
// noyé sous le bruit. L'ordre des clés est celui des structures, l'ordre des
// collections est imposé par Trier.
func Serialiser(v any) ([]byte, error) {
	var tampon bytes.Buffer
	encodeur := json.NewEncoder(&tampon)
	encodeur.SetIndent("", "  ")
	encodeur.SetEscapeHTML(false)
	if err := encodeur.Encode(v); err != nil {
		return nil, err
	}
	return tampon.Bytes(), nil
}

// Trier impose un ordre total sur toutes les collections du calque. À appeler
// avant toute sérialisation ou tout calcul d'empreinte.
//
// Tri stable, et pas seulement trié : les clés ne sont pas toutes uniques. Un
// index ou une vérification sans nom est courant sur du legacy, et plusieurs
// homonymes se comparent alors égaux. Un tri instable les permuterait selon
// l'implémentation, en écrasant l'ordre que le ORDER BY du catalogue avait
// déjà établi — la seule chose qui départage encore ces éléments.
func (p *Physique) Trier() {
	sort.SliceStable(p.Tables, func(i, j int) bool {
		if p.Tables[i].Schema != p.Tables[j].Schema {
			return p.Tables[i].Schema < p.Tables[j].Schema
		}
		return p.Tables[i].Nom < p.Tables[j].Nom
	})

	for i := range p.Tables {
		t := &p.Tables[i]
		sort.SliceStable(t.Colonnes, func(a, b int) bool { return t.Colonnes[a].Position < t.Colonnes[b].Position })
		sort.SliceStable(t.ClesEtrangeres, func(a, b int) bool { return t.ClesEtrangeres[a].Nom < t.ClesEtrangeres[b].Nom })
		sort.SliceStable(t.Unicites, func(a, b int) bool { return t.Unicites[a].Nom < t.Unicites[b].Nom })
		sort.SliceStable(t.Index, func(a, b int) bool { return t.Index[a].Nom < t.Index[b].Nom })
		sort.SliceStable(t.Verifications, func(a, b int) bool { return t.Verifications[a].Nom < t.Verifications[b].Nom })
	}

	sort.SliceStable(p.Sequences, func(i, j int) bool {
		if p.Sequences[i].Schema != p.Sequences[j].Schema {
			return p.Sequences[i].Schema < p.Sequences[j].Schema
		}
		return p.Sequences[i].Nom < p.Sequences[j].Nom
	})
	sort.SliceStable(p.TypesEnumeres, func(i, j int) bool { return p.TypesEnumeres[i].Nom < p.TypesEnumeres[j].Nom })
	sort.SliceStable(p.Vues, func(i, j int) bool { return p.Vues[i].Nom < p.Vues[j].Nom })
}

// CalculerEmpreinte hache le calque en excluant ExtraitLe et Empreinte :
// l'horodatage varie d'une extraction à l'autre alors que le contenu, lui, est
// identique. C'est cette exclusion qui rend le diff exploitable.
func (p *Physique) CalculerEmpreinte() (string, error) {
	p.Trier()

	copie := *p
	copie.Source.ExtraitLe = ""
	copie.Source.Empreinte = ""

	donnees, err := Serialiser(&copie)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("sha256:%x", sha256.Sum256(donnees)), nil
}

// Ecrire calcule l'empreinte, la pose, puis sérialise. L'ordre compte :
// l'empreinte est celle du document sans elle.
func (p *Physique) Ecrire(chemin string) error {
	empreinte, err := p.CalculerEmpreinte()
	if err != nil {
		return err
	}
	p.Source.Empreinte = empreinte

	donnees, err := Serialiser(p)
	if err != nil {
		return err
	}
	// 0600 : un calque porte le schéma de la base d'un client, et avec
	// --echantillonner des valeurs réelles. Pas un fichier à laisser lisible par
	// tous les comptes d'un serveur partagé.
	return os.WriteFile(chemin, donnees, 0o600)
}

// LirePhysique refuse une version plus récente que celle qu'il connaît : mieux
// vaut un échec net qu'un champ ignoré en silence. Plus ancienne : accepté.
func LirePhysique(chemin string) (*Physique, error) {
	// Le chemin vient de la ligne de commande : lire le fichier que
	// l'utilisateur désigne est la fonction même de l'outil.
	donnees, err := os.ReadFile(chemin) // #nosec G304
	if err != nil {
		return nil, err
	}

	var p Physique
	if err := json.Unmarshal(donnees, &p); err != nil {
		return nil, err
	}
	if p.VersionRI > VersionCourante {
		return nil, fmt.Errorf("calque en version %d, cet outil ne connaît que la version %d", p.VersionRI, VersionCourante)
	}
	return &p, nil
}

// VersionCourante est la version du format que cet outil produit et sait lire.
// Sans rapport avec SemVer : un champ optionnel ajouté ne l'incrémente pas.
//
// Le JSON Schema, cette constante et le lecteur PHP annoncent la même valeur.
// Une divergence est un défaut, pas un décalage temporaire.
const VersionCourante = 1
