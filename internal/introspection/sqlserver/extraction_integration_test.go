// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

//go:build integration

package sqlserver

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/sprimault/ormeau/internal/calque"
	"github.com/sprimault/ormeau/internal/introspection"
)

// extraireOuEchouer rend le cœur du calque de la base de référence, lu par la
// fonction interne : Extraire refuse tant que le reste du catalogue n'est pas
// lu.
func extraireOuEchouer(t *testing.T, portee introspection.Portee) *calque.Physique {
	t.Helper()

	i := ouvrir(t)
	ctx, annuler := context.WithTimeout(context.Background(), time.Minute)
	defer annuler()

	physique, err := i.(*pilote).extraire(ctx, portee)
	if err != nil {
		t.Fatalf("extraction : %v", err)
	}
	return physique
}

// colonneOuEchouer rend une colonne de ventes, ou arrête le test.
func colonneOuEchouer(t *testing.T, p *calque.Physique, table, colonne string) *calque.Colonne {
	t.Helper()

	c := tableOuEchouer(t, p, table).ColonneParNom(colonne)
	if c == nil {
		t.Fatalf("colonne ventes.%s.%s absente", table, colonne)
	}
	return c
}

// tableOuEchouer rend une table de ventes, ou arrête le test.
func tableOuEchouer(t *testing.T, p *calque.Physique, table string) *calque.Table {
	t.Helper()

	tbl := p.TableParNom("ventes", table)
	if tbl == nil {
		t.Fatalf("table ventes.%s absente", table)
	}
	return tbl
}

// TestExtraireProduitUnCalqueValide : le cœur extrait passe la validation du
// format, clés étrangères résolues comprises.
func TestExtraireProduitUnCalqueValide(t *testing.T) {
	p := extraireOuEchouer(t, introspection.Portee{Schemas: []string{"ventes"}})

	if a := p.Valider(); len(a) != 0 {
		t.Errorf("calque invalide : %+v", a)
	}
	if p.Source.SGBD != "sqlserver" || p.Source.Catalogue != "gescom" || p.Source.Schema != "ventes" {
		t.Errorf("source : %+v", p.Source)
	}
	if p.Source.Version == "" {
		t.Error("version du serveur absente")
	}
	if len(p.Tables) != 15 {
		t.Errorf("%d tables, attendu 15", len(p.Tables))
	}
}

// TestExtraireEstDeterministe : deux extractions de la même base rendent le
// même document, octet pour octet. Les noms de contraintes que le serveur a
// choisis (PK__t_avoir__47184C0E…) en font partie, et ne bougent pas tant que
// la base n'est pas recréée.
func TestExtraireEstDeterministe(t *testing.T) {
	portee := introspection.Portee{Schemas: []string{"ventes"}}
	a, err := calque.Serialiser(extraireOuEchouer(t, portee))
	if err != nil {
		t.Fatalf("serialisation : %v", err)
	}
	b, err := calque.Serialiser(extraireOuEchouer(t, portee))
	if err != nil {
		t.Fatalf("serialisation : %v", err)
	}
	if string(a) != string(b) {
		t.Error("deux extractions produisent des documents differents")
	}
}

// TestExtraireLesTypes : type_brut est celui que le serveur écrit, échelle des
// types temporels comprise, et la longueur se compte en caractères.
func TestExtraireLesTypes(t *testing.T) {
	p := extraireOuEchouer(t, introspection.Portee{Schemas: []string{"ventes"}})

	cas := []struct {
		table, colonne string
		brut           string
		norme          calque.TypeNorm
		longueur       int
	}{
		{"users", "PasswordHash", "binary(64)", calque.TypeBinaire, 64},
		{"users", "nom", "nvarchar(30)", calque.TypeTexte, 30},
		{"users", "session_id", "varchar(255)", calque.TypeTexte, 255},
		{"users", "Salt", "uniqueidentifier", calque.TypeUUID, 0},
		{"users", "mem_montant", "bit", calque.TypeBooleen, 0},
		{"users", "DT_cre_mdp", "datetime", calque.TypeHorodatage, 0},
		{"t_client", "cli_siret", "nchar(14)", calque.TypeTexte, 14},
		{"t_client", "created_at", "datetimeoffset(7)", calque.TypeHorodatage, 0},
		{"t_log_import", "horodatage", "datetime2(7)", calque.TypeHorodatage, 0},
		{"t_log_import", "message", "nvarchar(max)", calque.TypeTexte, 0},
		{"t_facture", "fac_taux", "real", calque.TypeFlottant, 0},
		{"t_facture", "fac_remise", "money", calque.TypeDecimal, 0},
		{"t_commande", "cmd_heure", "time(7)", calque.TypeHeure, 0},
	}
	for _, c := range cas {
		col := colonneOuEchouer(t, p, c.table, c.colonne)
		if col.TypeBrut != c.brut || col.TypeNormalise != c.norme {
			t.Errorf("%s.%s : %q %s, attendu %q %s", c.table, c.colonne, col.TypeBrut, col.TypeNormalise, c.brut, c.norme)
		}
		switch {
		case c.longueur == 0 && col.Longueur != nil:
			t.Errorf("%s.%s : longueur %d, attendue absente", c.table, c.colonne, *col.Longueur)
		case c.longueur != 0 && (col.Longueur == nil || *col.Longueur != c.longueur):
			t.Errorf("%s.%s : longueur %v, attendue %d", c.table, c.colonne, col.Longueur, c.longueur)
		}
	}

	montant := colonneOuEchouer(t, p, "t_client", "cli_ca_ttc")
	if montant.Precision == nil || *montant.Precision != 12 || montant.Echelle == nil || *montant.Echelle != 2 {
		t.Errorf("decimal(12,2) : precision %v, echelle %v", montant.Precision, montant.Echelle)
	}
	if entier := colonneOuEchouer(t, p, "users", "nivhab"); entier.Precision != nil || entier.Echelle != nil {
		t.Error("un smallint ne declare ni precision ni echelle")
	}
}

// TestExtraireIdentitesEtDefauts : IDENTITY refuse une valeur explicite, et
// les trois genres de défaut se distinguent.
func TestExtraireIdentitesEtDefauts(t *testing.T) {
	p := extraireOuEchouer(t, introspection.Portee{Schemas: []string{"ventes"}})

	if c := colonneOuEchouer(t, p, "users", "id_user"); !c.AutoIncrement || c.Identite != calque.IdentiteToujours {
		t.Errorf("IDENTITY : auto_increment %v, identite %q", c.AutoIncrement, c.Identite)
	}
	// Clé par séquence : ni identité, ni défaut ordinaire.
	avoir := colonneOuEchouer(t, p, "t_avoir", "avo_id")
	if avoir.AutoIncrement {
		t.Error("t_avoir.avo_id n'est pas une identite")
	}

	cas := []struct {
		table, colonne string
		genre          calque.GenreDefaut
		valeur         string
	}{
		{"t_avoir", "avo_id", calque.DefautSequence, "(NEXT VALUE FOR [ventes].[sq_avoir])"},
		{"t_client", "cli_statut", calque.DefautLitteral, "ACTIF"},
		{"users", "mem_montant", calque.DefautLitteral, "0"},
		{"users", "Salt", calque.DefautExpression, "(newid())"},
		{"t_client", "created_at", calque.DefautExpression, "(sysdatetimeoffset())"},
	}
	for _, c := range cas {
		d := colonneOuEchouer(t, p, c.table, c.colonne).Defaut
		if d == nil || d.Genre != c.genre || d.Valeur != c.valeur {
			t.Errorf("%s.%s : defaut %+v, attendu %s %q", c.table, c.colonne, d, c.genre, c.valeur)
		}
	}
	if d := colonneOuEchouer(t, p, "users", "nom").Defaut; d != nil {
		t.Errorf("users.nom n'a pas de defaut : %+v", d)
	}
}

// TestExtraireLesCles : ordre des colonnes, actions, auto-référence, table
// sans clé primaire.
func TestExtraireLesCles(t *testing.T) {
	p := extraireOuEchouer(t, introspection.Portee{Schemas: []string{"ventes"}})

	if pk := tableOuEchouer(t, p, "users").ClePrimaire; pk == nil || pk.Nom != "PK_users" {
		t.Errorf("cle primaire de users : %+v", pk)
	}
	if pk := tableOuEchouer(t, p, "t_client_contact").ClePrimaire; pk == nil || !slices.Equal(pk.Colonnes, []string{"cli_id", "ctc_id"}) {
		t.Errorf("cle composite de t_client_contact : %+v", pk)
	}
	if pk := tableOuEchouer(t, p, "t_log_import").ClePrimaire; pk != nil {
		t.Errorf("t_log_import n'a pas de cle primaire : %+v", pk)
	}

	client := tableOuEchouer(t, p, "t_client")
	if len(client.ClesEtrangeres) != 1 {
		t.Fatalf("cles etrangeres de t_client : %+v", client.ClesEtrangeres)
	}
	fk := client.ClesEtrangeres[0]
	if fk.Nom != "fk_client_commercial" || fk.SchemaCible != "ventes" || fk.TableCible != "t_commercial" ||
		!slices.Equal(fk.Colonnes, []string{"cli_com_id"}) || !slices.Equal(fk.ColonnesCibles, []string{"com_id"}) ||
		fk.ALaSuppression != calque.ActionSetNull || fk.ALaMiseAJour != "" {
		t.Errorf("fk_client_commercial : %+v", fk)
	}

	// Deux clés vers deux tables, l'une en cascade.
	tag := tableOuEchouer(t, p, "t_client_tag")
	actions := map[string]calque.Action{}
	for _, fk := range tag.ClesEtrangeres {
		actions[fk.TableCible] = fk.ALaSuppression
	}
	if len(actions) != 2 || actions["t_client"] != calque.ActionCascade || actions["t_tag"] != "" {
		t.Errorf("cles de t_client_tag : %+v", tag.ClesEtrangeres)
	}

	if fks := tableOuEchouer(t, p, "t_categorie").ClesEtrangeres; len(fks) != 1 || fks[0].TableCible != "t_categorie" {
		t.Errorf("auto-reference de t_categorie : %+v", fks)
	}
}

// TestExtraireLesNomsPenibles : accents, mots réservés et espaces traversent
// la description du serveur, composée par QUOTENAME.
func TestExtraireLesNomsPenibles(t *testing.T) {
	p := extraireOuEchouer(t, introspection.Portee{Schemas: []string{"ventes"}})

	for _, nom := range []string{"order", "select", "Libellé", "reprise an"} {
		if c := colonneOuEchouer(t, p, "t_référence", nom); c.TypeBrut == "" {
			t.Errorf("t_référence.%s sans type", nom)
		}
	}
}

// TestExtrairePorteeSansSchema : sans schéma demandé, tous ceux qui portent une
// table sont lus.
func TestExtrairePorteeSansSchema(t *testing.T) {
	p := extraireOuEchouer(t, introspection.Portee{})

	if p.TableParNom("ventes", "t_client") == nil {
		t.Error("ventes.t_client absente d'une extraction sans schema demande")
	}
	if p.Source.Schema != "ventes" {
		t.Errorf("schema de la source : %q", p.Source.Schema)
	}
}

// TestExtrairePorteeFiltrante : la portée retient, et une clé étrangère vers
// une table écartée reste dans le calque.
func TestExtrairePorteeFiltrante(t *testing.T) {
	p := extraireOuEchouer(t, introspection.Portee{
		Schemas:        []string{"ventes"},
		TablesIncluses: []string{"ventes.t_client", "ventes.t_facture"},
	})
	if len(p.Tables) != 2 {
		t.Fatalf("%d tables retenues", len(p.Tables))
	}
	if fks := tableOuEchouer(t, p, "t_client").ClesEtrangeres; len(fks) != 1 {
		t.Errorf("la cle vers t_commercial a disparu : %+v", fks)
	}

	exclu := extraireOuEchouer(t, introspection.Portee{
		Schemas:       []string{"ventes"},
		TablesExclues: []string{"ventes.t_log_import"},
	})
	if exclu.TableParNom("ventes", "t_log_import") != nil {
		t.Error("table exclue presente dans le calque")
	}
}
