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
func (p *Physique) Trier() {
	sort.Slice(p.Tables, func(i, j int) bool {
		if p.Tables[i].Schema != p.Tables[j].Schema {
			return p.Tables[i].Schema < p.Tables[j].Schema
		}
		return p.Tables[i].Nom < p.Tables[j].Nom
	})

	for i := range p.Tables {
		t := &p.Tables[i]
		sort.Slice(t.Colonnes, func(a, b int) bool { return t.Colonnes[a].Position < t.Colonnes[b].Position })
		sort.Slice(t.ClesEtrangeres, func(a, b int) bool { return t.ClesEtrangeres[a].Nom < t.ClesEtrangeres[b].Nom })
		sort.Slice(t.Unicites, func(a, b int) bool { return t.Unicites[a].Nom < t.Unicites[b].Nom })
		sort.Slice(t.Index, func(a, b int) bool { return t.Index[a].Nom < t.Index[b].Nom })
		sort.Slice(t.Verifications, func(a, b int) bool { return t.Verifications[a].Nom < t.Verifications[b].Nom })
	}

	sort.Slice(p.Sequences, func(i, j int) bool {
		if p.Sequences[i].Schema != p.Sequences[j].Schema {
			return p.Sequences[i].Schema < p.Sequences[j].Schema
		}
		return p.Sequences[i].Nom < p.Sequences[j].Nom
	})
	sort.Slice(p.TypesEnumeres, func(i, j int) bool { return p.TypesEnumeres[i].Nom < p.TypesEnumeres[j].Nom })
	sort.Slice(p.Vues, func(i, j int) bool { return p.Vues[i].Nom < p.Vues[j].Nom })
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
	return os.WriteFile(chemin, donnees, 0o644)
}

func LirePhysique(chemin string) (*Physique, error) {
	donnees, err := os.ReadFile(chemin)
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

const VersionCourante = 1
