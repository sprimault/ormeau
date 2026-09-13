// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"strings"
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
)

// Ce que les cas de référence heritage et heritage-decide ne couvrent pas :
// les refus d'une décision, la hiérarchie sur plusieurs niveaux, et le choix
// des colonnes candidates.

// longueur rend un pointeur vers une longueur de colonne.
func longueur(n int) *int {
	return &n
}

// hierarchie rend un calque physique : personne à la racine, salarie et
// prestataire qui en héritent possiblement, cadre qui hérite possiblement de
// salarie.
func hierarchie() *calque.Physique {
	enfant := func(nom, parent string) calque.Table {
		return calque.Table{
			Nom: nom, Schema: "public",
			Colonnes: []calque.Colonne{
				{Nom: "id", Position: 1, TypeBrut: "integer", TypeNormalise: calque.TypeEntier},
			},
			ClePrimaire: &calque.ClePrimaire{Colonnes: []string{"id"}},
			ClesEtrangeres: []calque.CleEtrangere{{
				Colonnes: []string{"id"}, SchemaCible: "public", TableCible: parent, ColonnesCibles: []string{"id"},
			}},
		}
	}

	return &calque.Physique{
		VersionRI: calque.VersionCourante,
		Tables: []calque.Table{
			{
				Nom: "personne", Schema: "public",
				Colonnes: []calque.Colonne{
					{Nom: "id", Position: 1, TypeBrut: "integer", TypeNormalise: calque.TypeEntier},
					{Nom: "nature", Position: 2, TypeBrut: "character(1)", TypeNormalise: calque.TypeTexte, Longueur: longueur(1)},
				},
				ClePrimaire: &calque.ClePrimaire{Colonnes: []string{"id"}},
			},
			enfant("salarie", "personne"),
			enfant("prestataire", "personne"),
			enfant("cadre", "salarie"),
		},
	}
}

// entiteNommee rend l'entité de ce nom, ou échoue.
func entiteNommee(t *testing.T, l *calque.Logique, nom string) calque.Entite {
	t.Helper()

	for _, e := range l.Entites {
		if e.Nom == nom {
			return e
		}
	}
	t.Fatalf("entité %s absente", nom)
	return calque.Entite{}
}

// TestUneDecisionRefuseeNAppliqueRien vérifie qu'une hiérarchie refusée ne
// s'applique en rien, et que le refus dit pourquoi.
//
// Tout ou rien, racine par racine : une hiérarchie à moitié appliquée serait
// incompréhensible. Chaque refus nomme ce qui manque, et aucune entité ne
// reçoit d'héritage — salarie, valide en soi, reste en un-vers-un.
func TestUneDecisionRefuseeNAppliqueRien(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom      string
		decision HeritageDecide
		raison   string
	}{
		{
			"colonne absente",
			HeritageDecide{ColonneDiscriminante: "type", Valeurs: map[string]string{"public.personne": "P", "public.salarie": "S"}},
			"public.personne n'a pas de colonne type",
		},
		{
			"racine sans valeur",
			HeritageDecide{ColonneDiscriminante: "nature", Valeurs: map[string]string{"public.salarie": "S"}},
			"valeurs ne donne pas de valeur à la racine public.personne",
		},
		{
			"aucun enfant",
			HeritageDecide{ColonneDiscriminante: "nature", Valeurs: map[string]string{"public.personne": "P"}},
			"valeurs ne cite aucune table enfant",
		},
		{
			"valeur donnée deux fois",
			HeritageDecide{ColonneDiscriminante: "nature", Valeurs: map[string]string{"public.personne": "P", "public.salarie": "P"}},
			"la valeur P est donnée à public.personne et à public.salarie",
		},
		{
			"table absente du calque",
			HeritageDecide{ColonneDiscriminante: "nature", Valeurs: map[string]string{"public.personne": "P", "public.salarie": "S", "public.fantome": "F"}},
			"public.fantome n'est pas une table générée du calque",
		},
		{
			"petit-enfant dont le parent n'est pas cité",
			HeritageDecide{ColonneDiscriminante: "nature", Valeurs: map[string]string{"public.personne": "P", "public.cadre": "C"}},
			"public.cadre ne descend pas de public.personne par des tables citées",
		},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			logique, avertissements := Inferer(hierarchie(), &Decisions{Heritages: map[string]HeritageDecide{"public.personne": c.decision}})

			var messages []string
			for _, a := range avertissements {
				if a.Code == calque.CodeDecisionOrpheline && a.Cible == "public.personne" {
					messages = append(messages, a.Message)
				}
			}
			if len(messages) != 1 || messages[0] != "héritage non appliqué : "+c.raison {
				t.Errorf("refus = %q, attendu « héritage non appliqué : %s »", messages, c.raison)
			}
			for _, e := range logique.Entites {
				if e.Heritage != nil || e.ValeurDiscriminante != "" {
					t.Errorf("%s porte un héritage malgré le refus", e.Nom)
				}
			}
			if a := entiteNommee(t, logique, "Salarie").Associations; len(a) == 0 || a[0].Genre != calque.UnVersUn {
				t.Errorf("Salarie devrait rester relié par un-vers-un, associations = %#v", a)
			}
		})
	}
}

// TestUneHierarchieSurPlusieursNiveaux vérifie qu'un petit-enfant hérite de
// son parent direct, et que la colonne et les valeurs suivent sur chaque
// niveau.
func TestUneHierarchieSurPlusieursNiveaux(t *testing.T) {
	t.Parallel()

	decisions := &Decisions{Heritages: map[string]HeritageDecide{"public.personne": {
		ColonneDiscriminante: "nature",
		Valeurs:              map[string]string{"public.personne": "P", "public.salarie": "S", "public.cadre": "C"},
	}}}

	logique, _ := Inferer(hierarchie(), decisions)

	cadre := entiteNommee(t, logique, "Cadre")
	if cadre.Heritage == nil || cadre.Heritage.Parent != "Salarie" || cadre.Heritage.ColonneDiscriminante != "nature" {
		t.Fatalf("héritage de Cadre = %#v, attendu Salarie sur nature", cadre.Heritage)
	}
	if cadre.ValeurDiscriminante != "C" || entiteNommee(t, logique, "Personne").ValeurDiscriminante != "P" {
		t.Errorf("valeurs discriminantes mal reportées")
	}
	if len(cadre.Associations) != 0 {
		t.Errorf("la clé primaire étrangère d'un enfant décidé ne doit pas donner d'association : %#v", cadre.Associations)
	}
}

// TestLesCandidatesDiscriminantes vérifie l'ordre et les exclusions : une
// énumération reconnue d'abord, puis un texte court au nom parlant ; ni la
// clé primaire, ni une clé étrangère, ni un texte long.
func TestLesCandidatesDiscriminantes(t *testing.T) {
	t.Parallel()

	racine := &calque.Table{
		Nom: "personne", Schema: "public",
		Colonnes: []calque.Colonne{
			{Nom: "type_id", Position: 1, TypeNormalise: calque.TypeTexte, Longueur: longueur(8)},
			{Nom: "categorie", Position: 2, TypeNormalise: calque.TypeTexte, Longueur: longueur(10)},
			{Nom: "genre_long", Position: 3, TypeNormalise: calque.TypeTexte, Longueur: longueur(200)},
			{Nom: "statut", Position: 4, TypeNormalise: calque.TypeTexte, Longueur: longueur(1)},
			{Nom: "nom", Position: 5, TypeNormalise: calque.TypeTexte, Longueur: longueur(8)},
		},
		ClePrimaire: &calque.ClePrimaire{Colonnes: []string{"type_id"}},
	}
	s := &schemaLogique{enumerations: map[string]enumeree{"public.personne.statut": {nom: "Statut"}}}

	obtenu := strings.Join(candidatesDiscriminantes(racine, s), ",")
	if obtenu != "statut,categorie" {
		t.Errorf("candidates = %s, attendu statut,categorie", obtenu)
	}
}

// TestLeFichierProposeLesHierarchiesPossibles vérifie que le prérempli
// propose la racine, ses descendantes et la colonne candidate, en
// commentaire, et que la hiérarchie décidée ne se propose plus.
func TestLeFichierProposeLesHierarchiesPossibles(t *testing.T) {
	t.Parallel()

	fichier := string(EcrireDecisions(hierarchie(), &Decisions{}, "rh"))
	attendu := "#   public.personne:\n#     colonne_discriminante: nature\n#     valeurs:\n" +
		"#       public.personne: # à remplir\n#       public.cadre: # à remplir\n" +
		"#       public.prestataire: # à remplir\n#       public.salarie: # à remplir\n"
	if !strings.Contains(fichier, attendu) {
		t.Errorf("proposition absente ou mal formée :\n%s", fichier)
	}

	decide := string(EcrireDecisions(hierarchie(), &Decisions{Heritages: map[string]HeritageDecide{"public.personne": {
		ColonneDiscriminante: "nature", Valeurs: map[string]string{"public.personne": "P", "public.salarie": "S"},
	}}}, "rh"))
	if strings.Contains(decide, "# à remplir") {
		t.Errorf("une hiérarchie décidée ne doit plus être proposée :\n%s", decide)
	}
	if !strings.Contains(decide, "\nheritages:\n  public.personne:\n    colonne_discriminante: nature\n    valeurs:\n      public.personne: P\n      public.salarie: S\n") {
		t.Errorf("décision mal écrite :\n%s", decide)
	}
}

// TestUneCleCompositeSurDeuxAssociationsResteDuPlusieursVersUn vérifie que
// compter la clé primaire comme unicité ne change rien à une clé primaire
// composite portée par deux clés étrangères : chacune n'en couvre qu'une
// partie, et reste un plusieurs-vers-un.
func TestUneCleCompositeSurDeuxAssociationsResteDuPlusieursVersUn(t *testing.T) {
	t.Parallel()

	ligne := &calque.Table{
		Nom: "ligne_commande", Schema: "public",
		ClePrimaire: &calque.ClePrimaire{Colonnes: []string{"commande_id", "article_id"}},
	}

	if unicitePorte(ligne, []string{"commande_id"}) {
		t.Error("une colonne d'une clé composite n'est pas unique à elle seule")
	}
	if !unicitePorte(ligne, []string{"article_id", "commande_id"}) {
		t.Error("une clé étrangère couvrant toute la clé primaire est unique, dans n'importe quel ordre")
	}
}
