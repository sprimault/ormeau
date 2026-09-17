// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

//go:build integration

package sqlserver

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/sprimault/ormeau/internal/calque"
	"github.com/sprimault/ormeau/internal/introspection"
)

// majAttendus réécrit le calque de référence au lieu de le comparer. Jamais
// automatique : un attendu régénéré sans être relu ne teste plus rien.
//
//	make maj-calque-sqlserver
var majAttendus = flag.Bool("maj-attendus", false, "réécrit le calque de référence SQL Server")

// cheminCalqueGescom est l'extraction versionnée de tests/ddl/sqlserver.sql.
// Hors de tests/reference/inference/ : un répertoire là porte un cas complet,
// logique attendu et entités PHP compris, et l'inférence de ce dialecte n'est
// pas encore relue.
const cheminCalqueGescom = "../../../tests/reference/extraction/sqlserver/gescom.calque.json"

// extraireOuEchouer rend le calque de la base de référence.
func extraireOuEchouer(t *testing.T, portee introspection.Portee) *calque.Physique {
	t.Helper()

	i := ouvrir(t)
	ctx, annuler := context.WithTimeout(context.Background(), time.Minute)
	defer annuler()

	physique, err := i.Extraire(ctx, portee)
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
	if p.Source.SGBD != "sqlserver" || p.Source.Catalogue != "gescom" {
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
	if remise := colonneOuEchouer(t, p, "t_facture", "fac_remise"); remise.Precision == nil || *remise.Precision != 19 ||
		remise.Echelle == nil || *remise.Echelle != 4 {
		t.Errorf("money : precision %v, echelle %v, attendu 19, 4", remise.Precision, remise.Echelle)
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
		// Le catalogue réécrit CONVERT(date, getdate()) : l'inférence reconnaît
		// cette forme-ci, pas celle du DDL.
		{"t_facture", "fac_saisie", calque.DefautExpression, "(CONVERT([date],getdate()))"},
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

// TestExtraireLeCatalogue : ce que le cœur ne lisait pas — commentaires,
// colonnes calculées, collations, unicités, CHECK, index, séquences, vues.
func TestExtraireLeCatalogue(t *testing.T) {
	p := extraireOuEchouer(t, introspection.Portee{Schemas: []string{"ventes"}})

	t.Run("commentaires", func(t *testing.T) {
		if c := tableOuEchouer(t, p, "t_client_tag").Commentaire; c != "Étiquettes posées sur un client" {
			t.Errorf("commentaire de table : %q", c)
		}
		if c := colonneOuEchouer(t, p, "t_client", "cli_siret").Commentaire; c != "Nul tant que la fiche n'est pas validée" {
			t.Errorf("commentaire de colonne : %q", c)
		}
	})

	// PERSISTED contre calcul à la lecture : stockee: false n'avait jamais été
	// produit, PostgreSQL ne sachant pas l'exprimer.
	t.Run("colonnes calculees", func(t *testing.T) {
		ht := colonneOuEchouer(t, p, "t_client", "cli_ca_ht")
		if ht.Generee == nil || !ht.Generee.Stockee || ht.Generee.Expression != "([cli_ca_ttc]/(1.2))" {
			t.Errorf("cli_ca_ht : %+v", ht.Generee)
		}
		mdp := colonneOuEchouer(t, p, "users", "DT_Change_mdp")
		if mdp.Generee == nil || mdp.Generee.Stockee || mdp.Generee.Expression != "(dateadd(month,(3),[DT_cre_mdp]))" {
			t.Errorf("DT_Change_mdp : %+v", mdp.Generee)
		}
		if mdp.Defaut != nil {
			t.Errorf("une colonne calculee n'a pas de defaut : %+v", mdp.Defaut)
		}
	})

	// Celle de la base devient default ; les autres gardent leur nom. Un
	// entier n'en a pas.
	t.Run("collations", func(t *testing.T) {
		cas := map[[2]string]string{
			{"t_commercial", "com_nom"}:   "French_CI_AS",
			{"t_tag", "tag_libelle"}:      "Latin1_General_BIN2",
			{"t_commercial", "com_email"}: "default",
			{"t_client", "cli_statut"}:    "default",
			{"t_client", "cli_id"}:        "",
		}
		for cle, attendue := range cas {
			if c := colonneOuEchouer(t, p, cle[0], cle[1]).Collation; c != attendue {
				t.Errorf("%s.%s : collation %q, attendue %q", cle[0], cle[1], c, attendue)
			}
		}
	})

	t.Run("unicites", func(t *testing.T) {
		u := tableOuEchouer(t, p, "t_client").Unicites
		if len(u) != 1 || u[0].Nom != "uq_cli_siret" || !slices.Equal(u[0].Colonnes, []string{"cli_siret"}) {
			t.Errorf("unicites de t_client : %+v", u)
		}
	})

	// Forme réécrite par le serveur : IN devenu OR, dans l'ordre inverse.
	t.Run("verifications", func(t *testing.T) {
		v := tableOuEchouer(t, p, "t_commande").Verifications
		if len(v) != 3 || v[0].Nom != "ck_cmd_canal" ||
			v[0].Expression != "([cmd_canal]='agence' OR [cmd_canal]='telephone' OR [cmd_canal]='web')" {
			t.Errorf("verifications de t_commande : %+v", v)
		}
		// ISJSON(cmd_trace) = 1 se relit sous la forme que l'inférence reconnaît.
		if len(v) == 3 && (v[2].Nom != "ck_cmd_trace" || v[2].Expression != "(isjson([cmd_trace])=(1))") {
			t.Errorf("verification ISJSON : %+v", v[2])
		}
	})

	t.Run("index", func(t *testing.T) {
		par := map[string]calque.Index{}
		for _, idx := range tableOuEchouer(t, p, "t_client").Index {
			par[idx.Nom] = idx
		}
		if idx := par["ix_cli_actifs"]; idx.Predicat != "([cli_statut]='ACTIF')" || idx.Methode != "nonclustered" || idx.Ordres != nil {
			t.Errorf("index filtre : %+v", idx)
		}
		if idx := par["ix_cli_nom_desc"]; !slices.Equal(idx.Ordres, []calque.OrdreIndex{calque.OrdreDescendant}) {
			t.Errorf("index descendant : %+v", idx)
		}
		// L'index qui soutient la contrainte n'y est pas : la contrainte le
		// rend, comme la clé primaire le sien.
		if idx, ok := par["uq_cli_siret"]; ok {
			t.Errorf("index de la contrainte d'unicite reporte deux fois : %+v", idx)
		}
		// Trois, ni la clé primaire ni la contrainte d'unicité n'en font partie.
		if len(par) != 3 {
			t.Errorf("index de t_client : %v", par)
		}
		if ref := tableOuEchouer(t, p, "t_référence").Index; len(ref) != 2 {
			t.Errorf("index de t_référence : %+v", ref)
		}
	})

	// Le départ est ce que le minimum ne dit pas : AS int part de 1, avec un
	// minimum au bas de l'entier.
	t.Run("sequences", func(t *testing.T) {
		if len(p.Sequences) != 1 {
			t.Fatalf("sequences : %+v", p.Sequences)
		}
		s := p.Sequences[0]
		if s.Nom != "sq_avoir" || s.Depart == nil || *s.Depart != 1 || s.Increment != 1 ||
			s.Minimum == nil || *s.Minimum != -2147483648 || s.Maximum == nil || *s.Maximum != 2147483647 || s.Cyclique {
			t.Errorf("sq_avoir : %+v", s)
		}
	})

	t.Run("vues", func(t *testing.T) {
		if len(p.Vues) != 1 || p.Vues[0].Nom != "v_client_actif" ||
			!strings.Contains(p.Vues[0].Definition, "CREATE VIEW ventes.v_client_actif AS") || p.Vues[0].Materialisee {
			t.Errorf("vues : %+v", p.Vues)
		}
	})
}

// TestExtraireCommeLaReference compare au calque versionné, qui couvre ce que
// les assertions écrites à la main ne regardent pas. extrait_le est repris du
// fichier : Extraire ne le pose pas, et il est exclu de l'empreinte.
func TestExtraireCommeLaReference(t *testing.T) {
	extrait := extraireOuEchouer(t, introspection.Portee{Schemas: []string{"ventes"}})

	if *majAttendus {
		extrait.Source.ExtraitLe = time.Now().UTC().Format(time.RFC3339)
		if err := os.MkdirAll(filepath.Dir(cheminCalqueGescom), 0o750); err != nil {
			t.Fatalf("repertoire du calque de reference : %v", err)
		}
		if err := extrait.Ecrire(cheminCalqueGescom); err != nil {
			t.Fatalf("ecriture du calque de reference : %v", err)
		}
		t.Log("calque de reference reecrit, relire le diff")
		return
	}

	attendu, err := os.ReadFile(cheminCalqueGescom)
	if err != nil {
		t.Fatalf("calque de reference illisible, le produire par make maj-calque-sqlserver : %v", err)
	}
	reference, err := calque.LirePhysique(cheminCalqueGescom)
	if err != nil {
		t.Fatalf("calque de reference invalide : %v", err)
	}

	extrait.Source.ExtraitLe = reference.Source.ExtraitLe
	empreinte, err := extrait.CalculerEmpreinte()
	if err != nil {
		t.Fatalf("empreinte : %v", err)
	}
	extrait.Source.Empreinte = empreinte
	obtenu, err := calque.Serialiser(extrait)
	if err != nil {
		t.Fatalf("serialisation : %v", err)
	}

	if !bytes.Equal(obtenu, attendu) {
		reference.Source.Empreinte, extrait.Source.Empreinte = "", ""
		sansA, errA := calque.Serialiser(reference)
		sansB, errB := calque.Serialiser(extrait)
		if errA != nil || errB != nil {
			t.Fatalf("serialisation sans empreinte : %v %v", errA, errB)
		}
		d := premiereDifference(sansA, sansB)
		if d.numero == 0 {
			d = premiereDifference(attendu, obtenu)
		}
		t.Errorf("l'extraction differe du calque de reference, ligne %d, sous %s :\n  attendu %s\n  obtenu  %s\n"+
			"si le changement est voulu : make maj-calque-sqlserver, puis relire le diff",
			d.numero, d.objet, d.attendue, d.obtenue)
	}
}

// difference situe le premier écart entre deux calques sérialisés.
type difference struct {
	numero            int
	attendue, obtenue string
	objet             string // dernier "nom" rencontré avant l'écart
}

// premiereDifference rend la première ligne où deux documents divergent, et
// le dernier nom qui la précède pour situer l'objet.
func premiereDifference(a, b []byte) difference {
	lignesA := strings.Split(string(a), "\n")
	lignesB := strings.Split(string(b), "\n")
	objet := "la source"
	for i := 0; i < len(lignesA) || i < len(lignesB); i++ {
		var la, lb string
		if i < len(lignesA) {
			la = strings.TrimSpace(lignesA[i])
		}
		if i < len(lignesB) {
			lb = strings.TrimSpace(lignesB[i])
		}
		if la != lb {
			return difference{numero: i + 1, attendue: la, obtenue: lb, objet: objet}
		}
		if strings.HasPrefix(la, `"nom":`) {
			objet = strings.TrimSuffix(strings.TrimPrefix(la, `"nom": `), ",")
		}
	}
	return difference{}
}

// TestExtraireSignaleSesPasses : sept passes, dans l'ordre, sans celle des
// types énumérés que SQL Server n'a pas.
func TestExtraireSignaleSesPasses(t *testing.T) {
	i := ouvrir(t)

	ctx, annuler := context.WithTimeout(context.Background(), time.Minute)
	defer annuler()

	var signales []introspection.Avancement
	ctx = introspection.AvecSuivi(ctx, func(a introspection.Avancement) {
		signales = append(signales, a)
	})

	if _, err := i.Extraire(ctx, introspection.Portee{Schemas: []string{"ventes"}}); err != nil {
		t.Fatalf("extraction : %v", err)
	}

	etapes := []string{
		introspection.EtapeSource, introspection.EtapeTables, introspection.EtapeColonnes,
		introspection.EtapeContraintes, introspection.EtapeIndex, introspection.EtapeSequences,
		introspection.EtapeVues,
	}
	if len(signales) != len(etapes) {
		t.Fatalf("%d passes signalees, %d attendues : %+v", len(signales), len(etapes), signales)
	}
	for rang, a := range signales {
		attendu := introspection.Avancement{Etape: etapes[rang], Rang: rang + 1, Total: len(etapes)}
		if a != attendu {
			t.Errorf("passe %d : %+v, attendu %+v", rang+1, a, attendu)
		}
	}
}

// TestExtraireAnnuleeEnCours : une extraction arrêtée depuis l'interface rend
// l'annulation, reconnaissable, et aucun calque.
func TestExtraireAnnuleeEnCours(t *testing.T) {
	i := ouvrir(t)

	ctx, annuler := context.WithCancel(context.Background())
	defer annuler()

	ctx = introspection.AvecSuivi(ctx, func(a introspection.Avancement) {
		if a.Etape == introspection.EtapeColonnes {
			annuler()
		}
	})

	physique, err := i.Extraire(ctx, introspection.Portee{Schemas: []string{"ventes"}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("erreur %v, attendu une annulation", err)
	}
	if physique != nil {
		t.Error("calque rendu malgre l'annulation")
	}
}
