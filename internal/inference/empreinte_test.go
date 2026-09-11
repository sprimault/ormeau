// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"slices"
	"strings"
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
)

// decisionsArbitrees sont celles d'un arbitrage ordinaire, sans type imposé :
// un test y décommente la clé de section seule.
func decisionsArbitrees() *Decisions {
	return &Decisions{
		Renommages:     map[string]string{"dbo.T_CLIENTS": "Client"},
		TablesIgnorees: []string{"dbo.T_PAYS"},
	}
}

// fichierArbitre rend le fichier que l'outil écrit pour gescom.
func fichierArbitre(t *testing.T) string {
	t.Helper()

	return string(EcrireDecisions(physiqueDeReference(t, "prefixes"), decisionsArbitrees(), "gescom"))
}

// retoucher applique une modification à la main, et échoue si le passage visé
// n'existe pas : sans cela, un test de modification passerait sans avoir rien
// modifié.
func retoucher(t *testing.T, fichier, avant, apres string) string {
	t.Helper()

	if !strings.Contains(fichier, avant) {
		t.Fatalf("passage %q absent du fichier", avant)
	}
	return strings.Replace(fichier, avant, apres, 1)
}

// TestEmpreinteSurvitAUneReextraction est le cas pour lequel l'empreinte existe :
// la même base six mois plus tard, avec une table de plus.
//
// La régénération change — la table apporte sa proposition de renommage —, et
// une détection par comparaison aurait classé manuel un fichier que personne
// n'a touché.
func TestEmpreinteSurvitAUneReextraction(t *testing.T) {
	t.Parallel()

	avant := physiqueDeReference(t, "prefixes")
	apres := *avant
	apres.Tables = append(slices.Clone(avant.Tables), calque.Table{
		Nom:      "T_FACTURES",
		Schema:   "dbo",
		Colonnes: []calque.Colonne{{Nom: "id", Position: 1, TypeBrut: "int", TypeNormalise: calque.TypeEntier}},
	})

	ecrit := EcrireDecisions(avant, decisionsArbitrees(), "gescom")
	regenere := EcrireDecisions(&apres, decisionsArbitrees(), "gescom")

	if string(ecrit) == string(regenere) {
		t.Fatal("la table ajoutée ne change pas le fichier : le cas ne teste plus le piège")
	}
	if ContenuManuel("gescom", ecrit) {
		t.Error("le fichier écrit avant la réextraction passe pour manuel")
	}
}

// TestEmpreinteIgnoreLaDocumentation vérifie qu'une version de l'outil qui
// reformule son aide ne classe pas manuels les fichiers déjà écrits.
func TestEmpreinteIgnoreLaDocumentation(t *testing.T) {
	t.Parallel()

	reformule := retoucher(t, fichierArbitre(t),
		"Une décision gagne toujours contre une heuristique",
		"Une décision l'emporte toujours sur une heuristique")

	if ContenuManuel("gescom", []byte(reformule)) {
		t.Error("une documentation reformulée classe le fichier manuel")
	}
}

// TestModificationDUneDecisionEstManuelle couvre le cas élémentaire : une valeur
// changée à la main.
func TestModificationDUneDecisionEstManuelle(t *testing.T) {
	t.Parallel()

	modifie := retoucher(t, fichierArbitre(t),
		"\nrenommages:\n  dbo.T_CLIENTS: Client\n",
		"\nrenommages:\n  dbo.T_CLIENTS: Acheteur\n")

	if !ContenuManuel("gescom", []byte(modifie)) {
		t.Error("une décision modifiée à la main passe pour intacte")
	}
}

// TestCommentaireContreUneDecisionEstManuel couvre les trois places où l'on
// annote une décision : réécrire le fichier perdrait l'explication.
func TestCommentaireContreUneDecisionEstManuel(t *testing.T) {
	t.Parallel()

	fichier := fichierArbitre(t)
	cas := map[string]string{
		"sur sa ligne":    retoucher(t, fichier, "\nrenommages:\n  dbo.T_CLIENTS: Client\n", "\nrenommages:\n  dbo.T_CLIENTS: Client  # validé avec le métier\n"),
		"dans son bloc":   retoucher(t, fichier, "\nrenommages:\n", "\nrenommages:\n  # validé avec le métier\n"),
		"juste au-dessus": retoucher(t, fichier, "\nrenommages:\n", "\n# validé avec le métier\nrenommages:\n"),
	}
	for nom, modifie := range cas {
		if !ContenuManuel("gescom", []byte(modifie)) {
			t.Errorf("commentaire %s : non protégé", nom)
		}
	}
}

// TestCommentaireIsoleNEstPasManuel vérifie l'autre moitié de l'arbitrage : un
// commentaire qu'une ligne vide sépare des décisions est réécrit sans demander.
func TestCommentaireIsoleNEstPasManuel(t *testing.T) {
	t.Parallel()

	modifie := retoucher(t, fichierArbitre(t),
		"# ── Noms de classes",
		"# note personnelle, sans rapport avec une décision\n\n# ── Noms de classes")

	if ContenuManuel("gescom", []byte(modifie)) {
		t.Error("un commentaire isolé classe le fichier manuel")
	}
}

// TestFinsDeLigneEtBOMSansEffet vérifie qu'un fichier ressorti par git en CRLF,
// ou enregistré par un éditeur avec un BOM, reste reconnu comme intact.
func TestFinsDeLigneEtBOMSansEffet(t *testing.T) {
	t.Parallel()

	fichier := fichierArbitre(t)
	cas := map[string]string{
		"CRLF":     strings.ReplaceAll(fichier, "\n", "\r\n"),
		"BOM":      bom + fichier,
		"les deux": bom + strings.ReplaceAll(fichier, "\n", "\r\n"),
	}
	for nom, variante := range cas {
		if ContenuManuel("gescom", []byte(variante)) {
			t.Errorf("%s : fichier intact classé manuel", nom)
		}
	}
}

// TestFichierRecopieDUneAutreBaseEstManuel couvre le copier-coller entre deux
// bases : l'empreinte serait valide, les décisions étrangères.
func TestFichierRecopieDUneAutreBaseEstManuel(t *testing.T) {
	t.Parallel()

	if !ContenuManuel("paie", []byte(fichierArbitre(t))) {
		t.Error("un fichier écrit pour gescom passe pour intact sous le nom de paie")
	}
}

// TestFichierSansEmpreinte couvre les fichiers d'avant l'empreinte : le
// prérempli n'a rien à perdre, le fichier écrit à la main si.
func TestFichierSansEmpreinte(t *testing.T) {
	t.Parallel()

	prerempli := string(EcrireDecisions(physiqueDeReference(t, "prefixes"), &Decisions{}, "gescom"))
	var lignes []string
	for _, ligne := range strings.Split(prerempli, "\n") {
		if strings.HasPrefix(ligne, prefixeEmpreinte) || strings.HasPrefix(ligne, "# Écrite par l'outil") {
			continue
		}
		lignes = append(lignes, ligne)
	}

	if ContenuManuel("gescom", []byte(strings.Join(lignes, "\n"))) {
		t.Error("un prérempli ancien, sans empreinte, passe pour manuel")
	}
	if !ContenuManuel("gescom", []byte(decisionsCompletes)) {
		t.Error("un fichier écrit à la main, sans empreinte, passe pour intact")
	}
}

// TestCleDeSectionSeuleEstManuellePuisRetiree couvre la clé décommentée puis
// laissée vide : une retouche humaine, vue une fois, que la réécriture retire
// sans laisser le fichier manuel.
func TestCleDeSectionSeuleEstManuellePuisRetiree(t *testing.T) {
	t.Parallel()

	modifie := fichierArbitre(t) + "\ntypes_forces:\n"
	if !ContenuManuel("gescom", []byte(modifie)) {
		t.Fatal("une clé décommentée à la main passe pour intacte")
	}

	relues, err := DecisionsDepuis([]byte(modifie))
	if err != nil {
		t.Fatalf("relecture : %v", err)
	}
	reecrit := string(EcrireDecisions(physiqueDeReference(t, "prefixes"), relues, "gescom"))

	if ContenuManuel("gescom", []byte(reecrit)) {
		t.Error("le fichier réécrit passe encore pour manuel")
	}
	if slices.Contains(strings.Split(reecrit, "\n"), "types_forces:") {
		t.Error("la clé vide a survécu à la réécriture")
	}
}
