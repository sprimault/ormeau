// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

//go:build integration

package postgres

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sprimault/ormeau/internal/calque"
	"github.com/sprimault/ormeau/internal/introspection"
)

// majAttendus réécrit le calque de référence de gescom au lieu de le comparer.
// Jamais automatique : un attendu régénéré sans être relu ne teste plus rien.
//
//	make maj-calque-gescom
var majAttendus = flag.Bool("maj-attendus", false, "réécrit le calque de référence de gescom")

// cheminCalqueGescom est le physique du cas d'inférence gescom. Le même fichier
// sert d'attendu à l'extraction et d'entrée à l'inférence : deux copies
// finiraient par diverger.
const cheminCalqueGescom = "../../../tests/reference/inference/gescom/physique.json"

// extraireOuEchouer rend le calque du schéma de test. Toutes les assertions du
// fichier partent de là : une extraction en échec arrête le test au lieu de
// produire une cascade d'erreurs sur un calque vide.
func extraireOuEchouer(t *testing.T) *calque.Physique {
	t.Helper()
	return extraireDepuis(t, dsnDeTest())
}

// extraireDepuis rend le calque du schéma de test lu par une connexion
// ouverte sur ce DSN.
func extraireDepuis(t *testing.T, dsn string) *calque.Physique {
	t.Helper()

	p := ouvrirDepuis(t, dsn)
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

// L'extraction de tests/ddl/ est versionnée : c'est le seul calque du dépôt que
// le pilote a réellement produit, et les assertions écrites à la main ne
// couvrent qu'une partie de ce qu'il porte. Une modification du pilote qui
// change un octet se relit dans le diff du fichier, régénéré par
// make maj-calque-gescom.
//
// extrait_le est repris du fichier avant la comparaison : Extraire ne le pose
// pas, la ligne de commande si, et il est exclu de l'empreinte.
func TestExtraireCommeLaReference(t *testing.T) {
	extrait := extraireOuEchouer(t)

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
		t.Fatalf("calque de reference illisible, le produire par make maj-calque-gescom : %v", err)
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
		// L'empreinte, en tête du document, diffère dès qu'autre chose diffère :
		// la ligne à montrer est la suivante, cherchée sans elle des deux côtés.
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
			"si le changement est voulu : make maj-calque-gescom, puis relire le diff",
			d.numero, d.objet, d.attendue, d.obtenue)
	}
}

// difference situe le premier écart entre deux calques sérialisés.
type difference struct {
	numero            int
	attendue, obtenue string
	objet             string // dernier "nom" rencontré avant l'écart
}

// premiereDifference rend la première ligne où deux documents divergent, pour
// que l'échec désigne l'écart au lieu d'imprimer deux calques entiers. Une
// ligne seule dit souvent peu (une accolade, une virgule) : le dernier nom qui
// la précède situe l'objet, table ou colonne.
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

// TestExtraireEstDeterministe extrait deux fois sous le même DSN, et ne voit
// donc pas ce qui dépend de la session. Ici deux connexions dont le search_path
// diffère : sous gescom, le catalogue écrirait les séquences du schéma sans le
// qualifier. Les octets doivent rester les mêmes, sinon le calque dépend de qui
// l'extrait.
func TestExtraireNeDependPasDuSearchPath(t *testing.T) {
	dsn := dsnDeTest()
	separateur := "?"
	if strings.Contains(dsn, "?") {
		separateur = "&"
	}

	octets := func(p *calque.Physique) string {
		t.Helper()
		o, err := calque.Serialiser(p)
		if err != nil {
			t.Fatalf("serialisation : %v", err)
		}
		return string(o)
	}

	defaut := octets(extraireDepuis(t, dsn))
	sousGescom := octets(extraireDepuis(t, dsn+separateur+"search_path=gescom,public"))
	if defaut != sousGescom {
		t.Error("le search_path du DSN change le calque extrait")
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

	// Le calque n'a pas de notion de tableau : l'élément donne le type
	// normalisé, et seuls les crochets de type_brut disent le reste.
	t.Run("tableau", func(t *testing.T) {
		c := colonneOuEchouer(t, p, "t_commande", "cmd_etiquettes")
		if c.TypeBrut != "text[]" || c.TypeNormalise != calque.TypeTexte {
			t.Errorf("type %q normalise en %q, attendu text[] normalise en texte", c.TypeBrut, c.TypeNormalise)
		}
	})

	// L'expression reste verbatim : c'est l'inférence qui en lit le nom. Le
	// schéma gescom n'étant pas dans le search_path de la session, le
	// catalogue qualifie la séquence.
	t.Run("serial", func(t *testing.T) {
		c := colonneOuEchouer(t, p, "t_avoir", "avo_id")
		if c.AutoIncrement {
			t.Error("un serial n'est pas une colonne IDENTITY")
		}
		attendu := calque.Defaut{Genre: calque.DefautSequence, Valeur: "nextval('gescom.t_avoir_avo_id_seq'::regclass)"}
		if c.Defaut == nil || *c.Defaut != attendu {
			t.Errorf("defaut %+v, attendu %+v", c.Defaut, attendu)
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

	// Une séquence par origine : le serial de t_avoir, l'identité de
	// t_commercial.
	t.Run("sequences", func(t *testing.T) {
		for _, nom := range []string{"t_avoir_avo_id_seq", "t_commercial_com_id_seq"} {
			trouvee := false
			for _, s := range p.Sequences {
				if s.Schema == "gescom" && s.Nom == nom {
					trouvee = true
				}
			}
			if !trouvee {
				t.Errorf("sequence %s absente du calque", nom)
			}
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

// L'avancement suit les huit passes du pilote, dans l'ordre où elles
// s'exécutent : c'est ce que l'interface affiche en paliers.
func TestExtraireSignaleSesPasses(t *testing.T) {
	p := ouvrirOuEchouer(t)

	ctx, annuler := context.WithTimeout(context.Background(), 30*time.Second)
	defer annuler()

	var signales []introspection.Avancement
	ctx = introspection.AvecSuivi(ctx, func(a introspection.Avancement) {
		signales = append(signales, a)
	})

	if _, err := p.Extraire(ctx, introspection.Portee{Schemas: []string{"gescom"}}); err != nil {
		t.Fatalf("extraction : %v", err)
	}

	etapes := []string{
		introspection.EtapeSource, introspection.EtapeTables, introspection.EtapeColonnes,
		introspection.EtapeContraintes, introspection.EtapeIndex, introspection.EtapeSequences,
		introspection.EtapeTypesEnumeres, introspection.EtapeVues,
	}
	if len(signales) != len(etapes) {
		t.Fatalf("%d passes signalees, %d attendues : %+v", len(signales), len(etapes), signales)
	}
	for i, a := range signales {
		attendu := introspection.Avancement{Etape: etapes[i], Rang: i + 1, Total: len(etapes)}
		if a != attendu {
			t.Errorf("passe %d : %+v, attendu %+v", i+1, a, attendu)
		}
	}
}

// Une extraction annulée en cours de route rend l'annulation et aucun calque :
// c'est ce qui permet d'arrêter depuis l'interface une extraction lancée par
// erreur sur une base de production.
func TestExtraireAnnuleeEnCours(t *testing.T) {
	p := ouvrirOuEchouer(t)

	ctx, annuler := context.WithCancel(context.Background())
	defer annuler()

	ctx = introspection.AvecSuivi(ctx, func(a introspection.Avancement) {
		if a.Etape == introspection.EtapeColonnes {
			annuler()
		}
	})

	physique, err := p.Extraire(ctx, introspection.Portee{Schemas: []string{"gescom"}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("erreur %v, attendu une annulation", err)
	}
	if physique != nil {
		t.Error("calque rendu malgre l'annulation")
	}
}

// tableOuEchouer rend une table du schéma de test, ou arrête le test.
func tableOuEchouer(t *testing.T, p *calque.Physique, nom string) *calque.Table {
	t.Helper()

	tbl := p.TableParNom("gescom", nom)
	if tbl == nil {
		t.Fatalf("table %s absente du calque", nom)
	}
	return tbl
}

// colonneOuEchouer rend une colonne du schéma de test, ou arrête le test.
func colonneOuEchouer(t *testing.T, p *calque.Physique, table, colonne string) *calque.Colonne {
	t.Helper()

	c := tableOuEchouer(t, p, table).ColonneParNom(colonne)
	if c == nil {
		t.Fatalf("colonne %s.%s absente", table, colonne)
	}
	return c
}
