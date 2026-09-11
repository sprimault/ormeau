// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"strings"
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
)

// physiqueClients rend une table à quatre colonnes, dont une clé primaire.
func physiqueClients() *calque.Physique {
	return &calque.Physique{
		VersionRI: calque.VersionCourante,
		Tables: []calque.Table{{
			Schema: "public",
			Nom:    "clients",
			Colonnes: []calque.Colonne{
				{Nom: "id", Position: 1, TypeBrut: "integer", TypeNormalise: calque.TypeEntier},
				{Nom: "nom", Position: 2, TypeBrut: "text", TypeNormalise: calque.TypeTexte},
				{Nom: "photo", Position: 3, TypeBrut: "bytea", TypeNormalise: calque.TypeBinaire, Nullable: true},
				{Nom: "blob_import", Position: 4, TypeBrut: "bytea", TypeNormalise: calque.TypeBinaire, Nullable: true},
			},
			ClePrimaire: &calque.ClePrimaire{Colonnes: []string{"id"}},
		}},
	}
}

// proprietes rend les colonnes mappées de la première entité.
func proprietes(l *calque.Logique) []string {
	var noms []string
	for _, p := range l.Entites[0].Proprietes {
		noms = append(noms, p.Colonne)
	}
	return noms
}

// codes rend les codes d'avertissement produits.
func codes(avertissements []calque.Avertissement) map[string]int {
	compte := map[string]int{}
	for _, a := range avertissements {
		compte[a.Code]++
	}
	return compte
}

// Une colonne écartée disparaît de l'entité, et de l'entité seulement : le
// calque physique qu'on lui a donné n'est pas modifié.
func TestColonnesIgnoreesRetireLaPropriete(t *testing.T) {
	t.Parallel()

	physique := physiqueClients()
	decisions := &Decisions{ColonnesIgnorees: map[string][]string{
		"public.clients": {"photo", "blob_import"},
	}}

	logique, avertissements := Inferer(physique, decisions)

	if got := proprietes(logique); len(got) != 2 {
		t.Errorf("propriétés %v, attendues id et nom", got)
	}
	if len(physique.Tables[0].Colonnes) != 4 {
		t.Error("le calque physique a perdu des colonnes : le mode diff les signalerait comme disparues")
	}
	if codes(avertissements)[calque.CodeColonneIgnoree] != 2 {
		t.Errorf("avertissements %v : chaque colonne écartée doit se voir", codes(avertissements))
	}
}

// Doctrine refuse une entité sans identifiant : la demande est signalée, pas
// appliquée.
func TestColonnesIgnoreesGardeLaClePrimaire(t *testing.T) {
	t.Parallel()

	decisions := &Decisions{ColonnesIgnorees: map[string][]string{
		"public.clients": {"id"},
	}}

	logique, avertissements := Inferer(physiqueClients(), decisions)

	var trouve bool
	for _, colonne := range proprietes(logique) {
		if colonne == "id" {
			trouve = true
		}
	}
	if !trouve {
		t.Error("la clé primaire a été retirée, l'entité ne peut plus être chargée")
	}
	if codes(avertissements)[calque.CodeClePrimaireGardee] != 1 {
		t.Errorf("le refus n'est pas signalé : %v", codes(avertissements))
	}
}

// Une décision qui ne correspond à rien signale que la base a bougé sous le
// fichier — c'est ce qu'on veut apprendre en régénérant six mois plus tard.
func TestColonnesIgnoreesSignaleUneCibleAbsente(t *testing.T) {
	t.Parallel()

	decisions := &Decisions{ColonnesIgnorees: map[string][]string{
		"public.clients": {"colonne_supprimee_en_prod"},
	}}

	logique, avertissements := Inferer(physiqueClients(), decisions)

	if got := proprietes(logique); len(got) != 4 {
		t.Errorf("propriétés %v : rien ne devait être retiré", got)
	}
	if codes(avertissements)[calque.CodeDecisionOrpheline] != 1 {
		t.Errorf("la cible absente n'est pas signalée : %v", codes(avertissements))
	}
}

// Une table absente du fichier garde toutes ses colonnes.
func TestColonnesIgnoreesNeTouchePasLesAutresTables(t *testing.T) {
	t.Parallel()

	decisions := &Decisions{ColonnesIgnorees: map[string][]string{
		"public.commandes": {"photo"},
	}}

	logique, _ := Inferer(physiqueClients(), decisions)

	if got := proprietes(logique); len(got) != 4 {
		t.Errorf("propriétés %v : la décision visait une autre table", got)
	}
}

// Le fichier documente l'arbitrage : personne ne lit docs/ avant de corriger un
// mapping. Prérempli, la section reste commentée, rien ne s'applique sans
// relecture ; arbitrée, la décision s'y écrit hors commentaire.
func TestFichierDocumenteLesColonnesIgnorees(t *testing.T) {
	t.Parallel()

	prerempli := string(EcrireDecisions(physiqueClients(), &Decisions{}, "clients"))
	for _, attendu := range []string{"colonnes_ignorees", "clé primaire", "#colonnes_ignorees: {}"} {
		if !strings.Contains(prerempli, attendu) {
			t.Errorf("%q absent du fichier prérempli", attendu)
		}
	}

	decisions := &Decisions{ColonnesIgnorees: map[string][]string{
		"public.clients": {"photo", "blob_import"},
	}}
	arbitre := string(EcrireDecisions(physiqueClients(), decisions, "clients"))
	if !strings.Contains(arbitre, "\ncolonnes_ignorees:\n  public.clients: [blob_import, photo]\n") {
		t.Errorf("décision absente ou commentée :\n%s", arbitre)
	}
}
