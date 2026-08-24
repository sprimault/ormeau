// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package calque

import (
	"strings"
	"testing"
)

// Le calque de référence sert de témoin : s'il produisait une anomalie, tous
// les autres cas deviendraient illisibles.
func TestValiderCalqueSain(t *testing.T) {
	t.Parallel()

	if a := physiqueDeReference().Valider(); len(a) != 0 {
		t.Errorf("calque de référence invalide : %+v", a)
	}
}

// Chaque cas casse une seule chose, et attend le code correspondant. Les codes
// servent de filtre en CI : les renommer casserait les pipelines.
func TestValiderDetecteLesAnomalies(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom   string
		muter func(*Physique)
		code  string
	}{
		{"version future", func(p *Physique) {
			p.VersionRI = VersionCourante + 1
		}, CodeVersionInconnue},

		{"version absente", func(p *Physique) {
			p.VersionRI = 0
		}, CodeVersionInconnue},

		{"sgbd hors vocabulaire", func(p *Physique) {
			p.Source.SGBD = "cobol"
		}, CodeSGBDInconnu},

		{"catalogue vide", func(p *Physique) {
			p.Source.Catalogue = ""
		}, CodeChampRequisVide},

		{"empreinte malformée", func(p *Physique) {
			p.Source.Empreinte = "sha256:pas-une-empreinte"
		}, CodeEmpreinteMalformee},

		{"table sans colonne", func(p *Physique) {
			p.TableParNom("public", "client").Colonnes = nil
		}, CodeTableSansColonne},

		{"table déclarée deux fois", func(p *Physique) {
			p.Tables = append(p.Tables, Table{
				Nom: "client", Schema: "public",
				Colonnes: []Colonne{{Nom: "id", Position: 1, TypeBrut: "integer", TypeNormalise: TypeEntier}},
			})
		}, CodeTableDupliquee},

		{"colonne déclarée deux fois", func(p *Physique) {
			tbl := p.TableParNom("public", "client")
			tbl.Colonnes = append(tbl.Colonnes, Colonne{
				Nom: "id", Position: 3, TypeBrut: "integer", TypeNormalise: TypeEntier,
			})
		}, CodeColonneDupliquee},

		{"type normalisé inventé", func(p *Physique) {
			p.TableParNom("public", "client").ColonneParNom("id").TypeNormalise = TypeNorm("entier_long")
		}, CodeTypeHorsVocabulaire},

		{"position nulle", func(p *Physique) {
			p.TableParNom("public", "client").ColonneParNom("id").Position = 0
		}, CodePositionInvalide},

		{"type brut vide", func(p *Physique) {
			p.TableParNom("public", "client").ColonneParNom("id").TypeBrut = ""
		}, CodeChampRequisVide},

		{"genre de défaut inventé", func(p *Physique) {
			p.TableParNom("public", "client").ColonneParNom("actif").Defaut.Genre = GenreDefaut("calcule")
		}, CodeGenreDefautInconnu},

		{"clé primaire sur colonne absente", func(p *Physique) {
			p.TableParNom("public", "client").ClePrimaire.Colonnes = []string{"identifiant"}
		}, CodeColonneIntrouvable},

		{"index sur colonne absente", func(p *Physique) {
			p.TableParNom("public", "commande").Index[0].Colonnes = []string{"montant"}
		}, CodeColonneIntrouvable},

		{"clé étrangère sur colonne absente", func(p *Physique) {
			p.TableParNom("public", "commande").ClesEtrangeres[0].Colonnes = []string{"cli_id"}
		}, CodeColonneIntrouvable},

		{"table cible absente du calque", func(p *Physique) {
			p.TableParNom("public", "commande").ClesEtrangeres[0].TableCible = "prospect"
		}, CodeTableCibleIntrouvable},

		{"colonne cible absente", func(p *Physique) {
			p.TableParNom("public", "commande").ClesEtrangeres[0].ColonnesCibles = []string{"numero"}
		}, CodeColonneIntrouvable},

		{"arité incohérente", func(p *Physique) {
			fk := &p.TableParNom("public", "commande").ClesEtrangeres[0]
			fk.ColonnesCibles = []string{"id", "id"}
		}, CodeAriteIncoherente},

		{"action référentielle inventée", func(p *Physique) {
			p.TableParNom("public", "commande").ClesEtrangeres[0].ALaSuppression = Action("propager")
		}, CodeActionInconnue},

		{"type énuméré non déclaré", func(p *Physique) {
			p.TableParNom("public", "client").ColonneParNom("actif").TypeEnumere = "etat_client"
		}, CodeTypeEnumereIntrouvable},

		{"type énuméré sans valeur", func(p *Physique) {
			p.TypesEnumeres[0].Valeurs = nil
		}, CodeChampRequisVide},

		{"statistiques orphelines", func(p *Physique) {
			p.Statistiques = map[string]StatistiquesTable{
				"public.prospect": {LignesEstimees: 12},
			}
		}, CodeStatistiquesOrphelines},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			p := physiqueDeReference()
			c.muter(p)

			anomalies := p.Valider()
			if len(anomalies) == 0 {
				t.Fatalf("aucune anomalie détectée, %q attendu", c.code)
			}
			for _, a := range anomalies {
				if a.Code == c.code {
					if a.Cible == "" || a.Message == "" {
						t.Errorf("anomalie sans cible ou sans message : %+v", a)
					}
					return
				}
			}
			t.Errorf("code %q absent des anomalies rendues : %+v", c.code, anomalies)
		})
	}
}

// Valider ne doit rien réparer ni réordonner : c'est un constat, et un appelant
// qui écrirait le calque ensuite ne doit pas voir son document modifié.
func TestValiderNeModifiePasLeCalque(t *testing.T) {
	t.Parallel()

	p := physiqueDeReference()
	avant, err := Serialiser(p)
	if err != nil {
		t.Fatalf("sérialisation : %v", err)
	}

	p.Valider()

	apres, err := Serialiser(p)
	if err != nil {
		t.Fatalf("sérialisation : %v", err)
	}
	if string(avant) != string(apres) {
		t.Error("le calque a été modifié par la validation")
	}
}

// L'ordre des anomalies doit être stable : sans ça, un filtre en CI verrait du
// bruit d'une exécution à l'autre. Les statistiques sont une map, c'est là que
// le risque se trouve.
func TestValiderEstDeterministe(t *testing.T) {
	t.Parallel()

	construire := func() *Physique {
		p := physiqueDeReference()
		p.Statistiques = map[string]StatistiquesTable{
			"public.zeta": {}, "public.alpha": {}, "public.mu": {},
			"archive.beta": {}, "public.omega": {},
		}
		return p
	}

	reference := construire().Valider()
	for i := 0; i < 20; i++ {
		obtenu := construire().Valider()
		if len(obtenu) != len(reference) {
			t.Fatalf("nombre d'anomalies instable : %d puis %d", len(reference), len(obtenu))
		}
		for j := range reference {
			if obtenu[j] != reference[j] {
				t.Fatalf("ordre instable au rang %d : %+v puis %+v", j, reference[j], obtenu[j])
			}
		}
	}
}

// Une table hors périmètre d'extraction produit une cible introuvable plutôt
// qu'un silence : la génération ne saura pas résoudre l'association, et
// l'utilisateur doit l'apprendre du calque, pas de Doctrine.
func TestValiderSignaleUneCibleHorsPerimetre(t *testing.T) {
	t.Parallel()

	p := physiqueDeReference()
	p.Tables = p.Tables[:len(p.Tables)-1]
	for i := range p.Tables {
		if p.Tables[i].Nom == "client" {
			p.Tables = append(p.Tables[:i], p.Tables[i+1:]...)
			break
		}
	}

	var trouve bool
	for _, a := range p.Valider() {
		if a.Code == CodeTableCibleIntrouvable && strings.Contains(a.Message, "client") {
			trouve = true
		}
	}
	if !trouve {
		t.Error("la clé étrangère vers une table retirée n'est pas signalée")
	}
}

// TestVerifierEmpreinte couvre les trois issues : conforme, absente, altérée.
// Le cas altéré est celui qui compte — c'est lui qui détecte un calque retouché
// après écriture.
func TestVerifierEmpreinte(t *testing.T) {
	t.Parallel()

	t.Run("empreinte conforme", func(t *testing.T) {
		t.Parallel()

		p := physiqueDeReference()
		empreinte, err := p.CalculerEmpreinte()
		if err != nil {
			t.Fatalf("calcul : %v", err)
		}
		p.Source.Empreinte = empreinte

		if err := p.VerifierEmpreinte(); err != nil {
			t.Errorf("empreinte pourtant conforme : %v", err)
		}
	})

	t.Run("calque retouche apres ecriture", func(t *testing.T) {
		t.Parallel()

		p := physiqueDeReference()
		empreinte, err := p.CalculerEmpreinte()
		if err != nil {
			t.Fatalf("calcul : %v", err)
		}
		p.Source.Empreinte = empreinte

		p.TableParNom("public", "client").ColonneParNom("id").TypeBrut = "bigint"

		if err := p.VerifierEmpreinte(); err == nil {
			t.Error("la retouche n'a pas été détectée")
		}
	})

	t.Run("horodatage sans effet", func(t *testing.T) {
		t.Parallel()

		p := physiqueDeReference()
		empreinte, err := p.CalculerEmpreinte()
		if err != nil {
			t.Fatalf("calcul : %v", err)
		}
		p.Source.Empreinte = empreinte
		p.Source.ExtraitLe = "2027-01-01T00:00:00Z"

		if err := p.VerifierEmpreinte(); err != nil {
			t.Errorf("l'horodatage ne doit pas invalider l'empreinte : %v", err)
		}
	})

	t.Run("empreinte absente", func(t *testing.T) {
		t.Parallel()

		if err := physiqueDeReference().VerifierEmpreinte(); err == nil {
			t.Error("aucune erreur alors qu'il n'y a rien à vérifier")
		}
	})
}

// Le calque écrit par Ecrire doit passer sa propre validation : sans ça, la
// chaîne produirait un document qu'elle refuse elle-même.
func TestCalqueEcritEstValide(t *testing.T) {
	t.Parallel()

	p := physiqueDeReference()
	if _, err := p.CalculerEmpreinte(); err != nil {
		t.Fatalf("calcul : %v", err)
	}
	empreinte, _ := p.CalculerEmpreinte()
	p.Source.Empreinte = empreinte

	if a := p.Valider(); len(a) != 0 {
		t.Errorf("calque écrit invalide : %+v", a)
	}
	if err := p.VerifierEmpreinte(); err != nil {
		t.Errorf("empreinte du calque écrit : %v", err)
	}
}
