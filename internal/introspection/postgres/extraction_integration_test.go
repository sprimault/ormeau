// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

//go:build integration

package postgres

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sprimault/ormeau/internal/calque"
	"github.com/sprimault/ormeau/internal/introspection"
)

func extraireOuEchouer(t *testing.T) *calque.Physique {
	t.Helper()

	p := ouvrirOuEchouer(t)
	ctx, annuler := context.WithTimeout(context.Background(), 30*time.Second)
	defer annuler()

	physique, err := p.Extraire(ctx, introspection.Portee{Schemas: []string{"gescom"}})
	if err != nil {
		t.Fatalf("extraction : %v", err)
	}
	return physique
}

// Un calque extrait doit passer sa propre validation, sinon la chaîne produit
// un document qu'elle refuse elle-même.
func TestExtraireProduitUnCalqueValide(t *testing.T) {
	physique := extraireOuEchouer(t)

	if a := physique.Valider(); len(a) != 0 {
		t.Errorf("calque invalide : %+v", a)
	}
	if physique.Source.SGBD != "postgres" || physique.Source.Catalogue != "gescom" {
		t.Errorf("source : %+v", physique.Source)
	}
	if physique.Source.Version == "" {
		t.Error("version du serveur absente")
	}
}

// Le test qui porte le mode diff : deux extractions de la même base rendent le
// même document, octet pour octet.
func TestExtraireEstDeterministe(t *testing.T) {
	premier := extraireOuEchouer(t)
	second := extraireOuEchouer(t)

	empreinteA, err := premier.CalculerEmpreinte()
	if err != nil {
		t.Fatalf("empreinte : %v", err)
	}
	empreinteB, err := second.CalculerEmpreinte()
	if err != nil {
		t.Fatalf("empreinte : %v", err)
	}
	if empreinteA != empreinteB {
		t.Errorf("deux extractions, deux empreintes :\n  %s\n  %s", empreinteA, empreinteB)
	}

	octetsA, err := calque.Serialiser(premier)
	if err != nil {
		t.Fatalf("serialisation : %v", err)
	}
	octetsB, err := calque.Serialiser(second)
	if err != nil {
		t.Fatalf("serialisation : %v", err)
	}
	if string(octetsA) != string(octetsB) {
		t.Error("deux extractions produisent des documents differents")
	}
}

// Les cas tordus de tests/ddl/postgres.sql, un par un. C'est ce fichier qui
// sert de spécification, et ce test qui vérifie qu'on le lit correctement.
func TestExtraireLesCasTordus(t *testing.T) {
	p := extraireOuEchouer(t)

	t.Run("colonne generee", func(t *testing.T) {
		c := colonneOuEchouer(t, p, "t_client", "cli_ca_ht")
		if c.Generee == nil {
			t.Fatal("colonne generee non detectee")
		}
		if !c.Generee.Stockee {
			t.Error("attendue stockee")
		}
		if c.Generee.Expression == "" {
			t.Error("expression de calcul absente")
		}
		if c.Defaut != nil {
			t.Errorf("une colonne generee n'a pas de defaut : %+v", c.Defaut)
		}
	})

	t.Run("identite", func(t *testing.T) {
		if c := colonneOuEchouer(t, p, "t_client", "cli_id"); !c.AutoIncrement {
			t.Error("GENERATED ALWAYS AS IDENTITY non detectee")
		}
	})

	t.Run("decimal garde precision et echelle", func(t *testing.T) {
		c := colonneOuEchouer(t, p, "t_client", "cli_ca_ttc")
		if c.TypeNormalise != calque.TypeDecimal {
			t.Errorf("type normalise %q", c.TypeNormalise)
		}
		if c.Precision == nil || *c.Precision != 12 {
			t.Errorf("precision %v, attendue 12", c.Precision)
		}
		if c.Echelle == nil || *c.Echelle != 2 {
			t.Errorf("echelle %v, attendue 2", c.Echelle)
		}
	})

	t.Run("longueur de char", func(t *testing.T) {
		c := colonneOuEchouer(t, p, "t_client", "cli_siret")
		if c.Longueur == nil || *c.Longueur != 14 {
			t.Errorf("longueur %v, attendue 14", c.Longueur)
		}
		if c.Commentaire == "" {
			t.Error("commentaire de colonne absent")
		}
	})

	t.Run("defaut litteral distinct d'une expression", func(t *testing.T) {
		statut := colonneOuEchouer(t, p, "t_client", "cli_statut")
		if statut.Defaut == nil || statut.Defaut.Genre != calque.DefautLitteral {
			t.Fatalf("defaut de cli_statut : %+v", statut.Defaut)
		}
		if statut.Defaut.Valeur != "ACTIF" {
			t.Errorf("valeur %q, attendue ACTIF", statut.Defaut.Valeur)
		}

		cree := colonneOuEchouer(t, p, "t_client", "created_at")
		if cree.Defaut == nil || cree.Defaut.Genre != calque.DefautExpression {
			t.Errorf("now() doit etre une expression : %+v", cree.Defaut)
		}
	})

	t.Run("cle primaire composite ordonnee", func(t *testing.T) {
		tbl := tableOuEchouer(t, p, "t_client_tag")
		if tbl.ClePrimaire == nil {
			t.Fatal("aucune cle primaire")
		}
		attendu := []string{"cli_id", "tag_id"}
		if len(tbl.ClePrimaire.Colonnes) != 2 ||
			tbl.ClePrimaire.Colonnes[0] != attendu[0] ||
			tbl.ClePrimaire.Colonnes[1] != attendu[1] {
			t.Errorf("colonnes %v, attendues %v", tbl.ClePrimaire.Colonnes, attendu)
		}
	})

	t.Run("table sans cle primaire", func(t *testing.T) {
		if tbl := tableOuEchouer(t, p, "t_log_import"); tbl.ClePrimaire != nil {
			t.Errorf("cle primaire inattendue : %+v", tbl.ClePrimaire)
		}
	})

	t.Run("cle etrangere avec action", func(t *testing.T) {
		tbl := tableOuEchouer(t, p, "t_client")
		var trouvee bool
		for _, fk := range tbl.ClesEtrangeres {
			if fk.TableCible != "t_commercial" {
				continue
			}
			trouvee = true
			if fk.SchemaCible != "gescom" {
				t.Errorf("schema cible %q", fk.SchemaCible)
			}
			if fk.ALaSuppression != calque.ActionSetNull {
				t.Errorf("action %q, attendue set_null", fk.ALaSuppression)
			}
			if len(fk.Colonnes) != 1 || fk.Colonnes[0] != "cli_com_id" {
				t.Errorf("colonnes %v", fk.Colonnes)
			}
			if len(fk.ColonnesCibles) != 1 || fk.ColonnesCibles[0] != "com_id" {
				t.Errorf("colonnes cibles %v", fk.ColonnesCibles)
			}
		}
		if !trouvee {
			t.Error("cle etrangere vers t_commercial absente")
		}
	})

	t.Run("auto reference", func(t *testing.T) {
		tbl := tableOuEchouer(t, p, "t_categorie")
		if len(tbl.ClesEtrangeres) != 1 || tbl.ClesEtrangeres[0].TableCible != "t_categorie" {
			t.Errorf("auto-reference absente : %+v", tbl.ClesEtrangeres)
		}
	})

	t.Run("verification verbatim", func(t *testing.T) {
		tbl := tableOuEchouer(t, p, "t_client")
		var trouvee bool
		for _, v := range tbl.Verifications {
			if v.Nom == "ck_cli_statut" {
				trouvee = true
				if !strings.Contains(v.Expression, "ACTIF") {
					t.Errorf("expression %q", v.Expression)
				}
			}
		}
		if !trouvee {
			t.Error("contrainte de verification absente")
		}
	})

	t.Run("unicite", func(t *testing.T) {
		tbl := tableOuEchouer(t, p, "t_client")
		if len(tbl.Unicites) != 1 || tbl.Unicites[0].Nom != "uq_cli_siret" {
			t.Errorf("unicites %+v", tbl.Unicites)
		}
	})

	t.Run("index partiel garde son predicat", func(t *testing.T) {
		tbl := tableOuEchouer(t, p, "t_client")
		var partiel *calque.Index
		for i := range tbl.Index {
			if tbl.Index[i].Nom == "ix_cli_actifs" {
				partiel = &tbl.Index[i]
			}
		}
		if partiel == nil {
			t.Fatal("index partiel absent")
		}
		if partiel.Predicat == "" {
			t.Error("predicat perdu : c'est ce qu'information_schema ne rend pas")
		}
		if partiel.Methode != "btree" {
			t.Errorf("methode %q", partiel.Methode)
		}
	})

	t.Run("classe d'operateurs explicite", func(t *testing.T) {
		tbl := tableOuEchouer(t, p, "t_client")

		var explicite, implicite *calque.Index
		for i := range tbl.Index {
			switch tbl.Index[i].Nom {
			case "ix_cli_nom_prefixe":
				explicite = &tbl.Index[i]
			case "ix_cli_nom":
				implicite = &tbl.Index[i]
			}
		}

		if explicite == nil {
			t.Fatal("index a classe d'operateurs explicite absent")
		}
		if len(explicite.Operateurs) != 1 || explicite.Operateurs[0] != "text_pattern_ops" {
			t.Errorf("operateurs %v, attendu [text_pattern_ops]", explicite.Operateurs)
		}

		// Un index sans classe explicite ne doit rien porter : sinon chaque
		// btree trivial chargerait le calque de sa classe implicite.
		if implicite == nil {
			t.Fatal("index a classe implicite absent")
		}
		if len(implicite.Operateurs) != 0 {
			t.Errorf("operateurs %v sur un index sans classe explicite", implicite.Operateurs)
		}
	})

	t.Run("type enumere natif", func(t *testing.T) {
		c := colonneOuEchouer(t, p, "t_commande", "cmd_canal")
		if c.TypeNormalise != calque.TypeEnumereNorm {
			t.Errorf("type normalise %q", c.TypeNormalise)
		}
		if c.TypeEnumere != "canal" {
			t.Errorf("type enumere %q", c.TypeEnumere)
		}

		var trouve bool
		for _, te := range p.TypesEnumeres {
			if te.Nom != "canal" {
				continue
			}
			trouve = true
			attendu := []string{"web", "telephone", "agence"}
			if len(te.Valeurs) != len(attendu) {
				t.Fatalf("valeurs %v", te.Valeurs)
			}
			for i := range attendu {
				if te.Valeurs[i] != attendu[i] {
					t.Errorf("ordre de declaration perdu : %v", te.Valeurs)
					break
				}
			}
		}
		if !trouve {
			t.Error("type enumere canal absent du calque")
		}
	})

	t.Run("identifiants accentues et reserves", func(t *testing.T) {
		tbl := tableOuEchouer(t, p, "t_référence")
		noms := map[string]bool{}
		for _, c := range tbl.Colonnes {
			noms[c.Nom] = true
		}
		for _, attendu := range []string{"id", "order", "select"} {
			if !noms[attendu] {
				t.Errorf("colonne %q absente : %v", attendu, noms)
			}
		}
	})

	t.Run("vue", func(t *testing.T) {
		var trouvee bool
		for _, v := range p.Vues {
			if v.Nom == "v_client_actif" {
				trouvee = true
				if v.Definition == "" {
					t.Error("definition de vue vide")
				}
				if v.Materialisee {
					t.Error("vue marquee materialisee")
				}
			}
		}
		if !trouvee {
			t.Error("vue absente du calque")
		}
	})

	t.Run("sequences", func(t *testing.T) {
		if len(p.Sequences) == 0 {
			t.Error("aucune sequence : les colonnes IDENTITY en creent")
		}
	})

	t.Run("colonnes triees par position", func(t *testing.T) {
		for _, tbl := range p.Tables {
			for i := 1; i < len(tbl.Colonnes); i++ {
				if tbl.Colonnes[i-1].Position > tbl.Colonnes[i].Position {
					t.Errorf("%s : colonnes non triees", tbl.Nom)
					break
				}
			}
		}
	})
}

// Une portée sans schéma prend toute la base, et non « public » seul : les
// tables de référence vivent dans gescom, et s'en tenir à public rendrait un
// calque vide.
func TestExtrairePorteeSansSchema(t *testing.T) {
	p := ouvrirOuEchouer(t)

	ctx, annuler := context.WithTimeout(context.Background(), 30*time.Second)
	defer annuler()

	physique, err := p.Extraire(ctx, introspection.Portee{})
	if err != nil {
		t.Fatalf("extraction : %v", err)
	}
	if len(physique.Tables) == 0 {
		t.Fatal("aucune table : la portee par defaut ne couvre pas la base")
	}
	if physique.TableParNom("gescom", "t_client") == nil {
		t.Error("gescom.t_client absente d'une extraction sans schema demande")
	}
}

// La portée doit filtrer, et une table exclue ne doit pas revenir par une autre
// passe de collecte.
func TestExtrairePorteeFiltrante(t *testing.T) {
	p := ouvrirOuEchouer(t)
	ctx, annuler := context.WithTimeout(context.Background(), 30*time.Second)
	defer annuler()

	physique, err := p.Extraire(ctx, introspection.Portee{
		Schemas:        []string{"gescom"},
		TablesIncluses: []string{"gescom.t_client", "gescom.t_commercial"},
	})
	if err != nil {
		t.Fatalf("extraction : %v", err)
	}
	if len(physique.Tables) != 2 {
		var noms []string
		for _, tbl := range physique.Tables {
			noms = append(noms, tbl.Nom)
		}
		t.Errorf("%d tables retenues : %v", len(physique.Tables), noms)
	}

	exclu, err := p.Extraire(ctx, introspection.Portee{
		Schemas:       []string{"gescom"},
		TablesExclues: []string{"gescom.t_log_import"},
	})
	if err != nil {
		t.Fatalf("extraction : %v", err)
	}
	for _, tbl := range exclu.Tables {
		if tbl.Nom == "t_log_import" {
			t.Error("table exclue presente dans le calque")
		}
	}
}

func tableOuEchouer(t *testing.T, p *calque.Physique, nom string) *calque.Table {
	t.Helper()

	tbl := p.TableParNom("gescom", nom)
	if tbl == nil {
		t.Fatalf("table %s absente du calque", nom)
	}
	return tbl
}

func colonneOuEchouer(t *testing.T, p *calque.Physique, table, colonne string) *calque.Colonne {
	t.Helper()

	c := tableOuEchouer(t, p, table).ColonneParNom(colonne)
	if c == nil {
		t.Fatalf("colonne %s.%s absente", table, colonne)
	}
	return c
}
