// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"bytes"
	"cmp"
	"maps"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
)

// decisionsPiegeuses remplit toutes les sections avec des valeurs que des
// règles de guillemets approximatives feraient mal relire.
func decisionsPiegeuses() *Decisions {
	return &Decisions{
		EspaceDeNoms:     `Gescom\Domaine\Entity`,
		PrefixesARetirer: []string{"T_", "- tiret", "#dièse"},
		TablesIgnorees:   []string{"public.migrations", "null", "public.a: b"},
		ColonnesIgnorees: map[string][]string{
			"public.clients":  {"photo", "blob, import", "[crochet]"},
			"dbo.T_COMMANDES": {"champ_libre_12"},
			"public.vide":     {},
		},
		Renommages: map[string]string{
			"dbo.T_CLIENTS": "Client",
			"public.true":   "Vrai",
			"public.x #y":   "l'objet",
			"yes":           " espace en tête",
			"public.ligne":  "deux\nlignes",
		},
		TypesForces: map[string]string{
			"dbo.T_CLIENTS.CLI_ACTIF": "boolean",
			"public.t.code":           "007",
		},
		RelationsForcees: []RelationForcee{
			{Source: "public.commande.client_id", Cible: "public.client.id", Genre: "plusieurs_vers_un", Nom: "client"},
			{Source: "public.a.b", Cible: "public.c.d", Genre: "un_vers_un"},
		},
		Enumerations: []EnumerationForcee{
			{Colonne: "public.client.canal", Nom: "Canal"},
			{Colonne: "public.client.actif", Nom: "OuiNon", Cas: map[string]string{"O": "Oui", "N": "Non", "0": "Zero", "~": "Tilde"}},
		},
	}
}

// sensDes rend des décisions comparables : collections triées là où l'ordre ne
// porte rien, vides et absentes confondues. C'est ce que le fichier doit
// garder ; le reste est une affaire d'écriture.
func sensDes(d *Decisions) Decisions {
	s := Decisions{EspaceDeNoms: d.EspaceDeNoms}
	if len(d.PrefixesARetirer) > 0 {
		s.PrefixesARetirer = d.PrefixesARetirer
	}
	if len(d.TablesIgnorees) > 0 {
		s.TablesIgnorees = slices.Sorted(slices.Values(d.TablesIgnorees))
	}
	if len(d.ColonnesIgnorees) > 0 {
		s.ColonnesIgnorees = map[string][]string{}
		for table, colonnes := range d.ColonnesIgnorees {
			s.ColonnesIgnorees[table] = slices.Sorted(slices.Values(colonnes))
		}
	}
	if len(d.Renommages) > 0 {
		s.Renommages = maps.Clone(d.Renommages)
	}
	if len(d.TypesForces) > 0 {
		s.TypesForces = maps.Clone(d.TypesForces)
	}
	if len(d.RelationsForcees) > 0 {
		s.RelationsForcees = slices.SortedFunc(slices.Values(d.RelationsForcees), func(x, y RelationForcee) int {
			return cmp.Or(strings.Compare(x.Source, y.Source), strings.Compare(x.Cible, y.Cible))
		})
	}
	for _, e := range d.Enumerations {
		if len(e.Cas) == 0 {
			e.Cas = nil
		}
		s.Enumerations = append(s.Enumerations, e)
	}
	slices.SortFunc(s.Enumerations, func(x, y EnumerationForcee) int {
		return strings.Compare(x.Colonne, y.Colonne)
	})
	return s
}

// allerRetour écrit, relit et réécrit un fichier de décisions, et vérifie que
// rien n'a bougé : ni les octets, ni le sens. Rend les décisions relues.
//
// C'est le test qui garde le déterminisme du fichier, au même titre que celui
// du calque : l'interface le réécrit à chaque enregistrement, et une ligne qui
// changerait sans décision nouvelle serait un diff incompréhensible dans le
// dépôt de l'utilisateur.
func allerRetour(t *testing.T, p *calque.Physique, d *Decisions) *Decisions {
	t.Helper()

	premier := EcrireDecisions(p, d, "gescom")
	relues, err := DecisionsDepuis(premier)
	if err != nil {
		t.Fatalf("relecture : %v\n%s", err, premier)
	}
	if second := EcrireDecisions(p, relues, "gescom"); !bytes.Equal(premier, second) {
		t.Fatalf("la réécriture change le fichier\n--- écrit ---\n%s\n--- réécrit ---\n%s", premier, second)
	}
	if !reflect.DeepEqual(sensDes(relues), sensDes(d)) {
		t.Errorf("le fichier relu ne dit plus la même chose\nécrit : %#v\nrelu  : %#v", sensDes(d), sensDes(relues))
	}
	if ContenuManuel("gescom", premier) {
		t.Error("un fichier tout juste écrit passe pour manuel")
	}
	return relues
}

// TestEcrireDecisionsAllerRetourPrerempli couvre le premier passage : rien de
// décidé, et un fichier qui se relit à l'identique.
func TestEcrireDecisionsAllerRetourPrerempli(t *testing.T) {
	t.Parallel()

	allerRetour(t, physiqueDeReference(t, "prefixes"), &Decisions{})
}

// TestEcrireDecisionsAllerRetourValeursPiegeuses remplit toutes les sections
// avec des valeurs qui exigent des guillemets : mots réservés, indicateurs en
// tête, dièse, virgule dans une liste en ligne, retour à la ligne.
func TestEcrireDecisionsAllerRetourValeursPiegeuses(t *testing.T) {
	t.Parallel()

	allerRetour(t, physiqueDeReference(t, "prefixes"), decisionsPiegeuses())
}

// TestEcrireDecisionsGardeLeSensDesCasDeReference vérifie, cas par cas, qu'un
// fichier réécrit produit la même inférence que celui d'origine.
func TestEcrireDecisionsGardeLeSensDesCasDeReference(t *testing.T) {
	t.Parallel()

	for _, cas := range casDeReference(t) {
		t.Run(cas, func(t *testing.T) {
			t.Parallel()

			repertoire := filepath.Join(racineReference, cas)
			p, err := calque.LirePhysique(filepath.Join(repertoire, "physique.json"))
			if err != nil {
				t.Fatalf("lecture du physique : %v", err)
			}
			d, err := LireDecisions(cheminDecisions(repertoire))
			if err != nil {
				t.Fatalf("lecture des décisions : %v", err)
			}

			relues := allerRetour(t, p, d)
			if inferer(t, p, d) != inferer(t, p, relues) {
				t.Error("le fichier réécrit ne produit plus la même inférence")
			}
		})
	}
}

// TestEcrireDecisionsSepareCommentairesEtDecisions garde les deux règles de
// forme dont dépend la détection du contenu manuel : aucun commentaire généré
// dans un paragraphe de décisions, aucune clé sans valeur.
func TestEcrireDecisionsSepareCommentairesEtDecisions(t *testing.T) {
	t.Parallel()

	fichier := normaliser(EcrireDecisions(physiqueDeReference(t, "prefixes"), decisionsPiegeuses(), "gescom"))

	for _, paragraphe := range strings.Split(paragraphesActifs(fichier), "\n\n") {
		lignes := strings.Split(paragraphe, "\n")
		for i, ligne := range lignes {
			if strings.HasPrefix(strings.TrimSpace(ligne), "#") {
				t.Errorf("commentaire généré dans un paragraphe de décisions : %q", ligne)
			}
			if !strings.HasSuffix(ligne, ":") {
				continue
			}
			if i+1 == len(lignes) || indentation(lignes[i+1]) <= indentation(ligne) {
				t.Errorf("clé sans valeur : %q", ligne)
			}
		}
	}
}

// indentation compte les espaces en tête d'une ligne.
func indentation(ligne string) int {
	return len(ligne) - len(strings.TrimLeft(ligne, " "))
}
