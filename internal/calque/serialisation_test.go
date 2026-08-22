// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package calque

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ptrInt : longueur, précision et échelle sont des pointeurs.
func ptrInt(v int) *int { return &v }

// physiqueDeReference est volontairement dans le désordre, colonnes comprises,
// et porte des expressions que l'encodeur JSON échapperait par défaut.
func physiqueDeReference() *Physique {
	return &Physique{
		VersionRI: VersionCourante,
		Source: Source{
			SGBD:      "postgres",
			Version:   "16.2",
			Catalogue: "gescom",
			Schema:    "public",
			ExtraitLe: "2026-08-22T10:00:00Z",
		},
		Tables: []Table{
			{
				Nom:    "commande",
				Schema: "public",
				Colonnes: []Colonne{
					{Nom: "client_id", Position: 2, TypeBrut: "integer", TypeNormalise: TypeEntier},
					{Nom: "id", Position: 1, TypeBrut: "integer", TypeNormalise: TypeEntier},
					{
						Nom: "total", Position: 3, TypeBrut: "numeric(10,0)",
						TypeNormalise: TypeDecimal, Precision: ptrInt(10), Echelle: ptrInt(0),
					},
				},
				ClePrimaire: &ClePrimaire{Nom: "commande_pkey", Colonnes: []string{"id"}},
				ClesEtrangeres: []CleEtrangere{
					{
						Nom: "fk_commande_client", Colonnes: []string{"client_id"},
						TableCible: "client", SchemaCible: "public",
						ColonnesCibles: []string{"id"}, ALaSuppression: ActionRestrict,
					},
				},
				Verifications: []Verification{
					{Nom: "ck_total_positif", Expression: "total > 0 AND total < 1000000"},
					{Nom: "ck_client", Expression: "client_id <> 0"},
				},
				Index: []Index{
					{Nom: "idx_total", Colonnes: []string{"total"}, Unique: false},
					{Nom: "idx_client", Colonnes: []string{"client_id"}, Unique: false},
				},
			},
			{
				Nom:    "client",
				Schema: "public",
				Colonnes: []Colonne{
					{Nom: "id", Position: 1, TypeBrut: "integer", TypeNormalise: TypeEntier, AutoIncrement: true},
					{
						Nom: "actif", Position: 2, TypeBrut: "char(1)", TypeNormalise: TypeTexte,
						Longueur: ptrInt(1), Defaut: &Defaut{Genre: DefautLitteral, Valeur: "O"},
					},
				},
				ClePrimaire: &ClePrimaire{Nom: "client_pkey", Colonnes: []string{"id"}},
			},
			{
				Nom:      "audit",
				Schema:   "archive",
				Colonnes: []Colonne{{Nom: "id", Position: 1, TypeBrut: "bigint", TypeNormalise: TypeEntier}},
			},
		},
		Sequences: []Sequence{
			{Nom: "client_id_seq", Schema: "public", Increment: 1},
			{Nom: "audit_id_seq", Schema: "archive", Increment: 1},
		},
		TypesEnumeres: []TypeEnumere{
			{Nom: "statut", Schema: "public", Valeurs: []string{"ouvert", "ferme"}},
			{Nom: "canal", Schema: "public", Valeurs: []string{"web", "magasin"}},
		},
		Vues: []Vue{
			{Nom: "v_commandes", Schema: "public", Definition: "SELECT * FROM commande WHERE total > 0"},
			{Nom: "v_clients", Schema: "public", Definition: "SELECT * FROM client"},
		},
	}
}

// Le socle du mode diff.
func TestSerialiserEstReproductible(t *testing.T) {
	t.Parallel()

	p := physiqueDeReference()
	premier, err := Serialiser(p)
	if err != nil {
		t.Fatalf("première sérialisation : %v", err)
	}
	second, err := Serialiser(p)
	if err != nil {
		t.Fatalf("seconde sérialisation : %v", err)
	}

	if string(premier) != string(second) {
		t.Error("deux sérialisations du même calque diffèrent")
	}
}

// Les expressions sont verbatim : un CHECK (total < 1000000) rendu échappé
// serait illisible en diff comme en fichier de référence.
func TestSerialiserNEchappePasLesCaracteresHTML(t *testing.T) {
	t.Parallel()

	p := physiqueDeReference()
	p.Tables[0].Verifications = append(p.Tables[0].Verifications,
		Verification{Nom: "ck_esperluette", Expression: "flags & 1 = 1"})

	donnees, err := Serialiser(p)
	if err != nil {
		t.Fatalf("sérialisation : %v", err)
	}

	sortie := string(donnees)
	for _, verbatim := range []string{"total > 0 AND total < 1000000", "client_id <> 0", "flags & 1 = 1"} {
		if !strings.Contains(sortie, verbatim) {
			t.Errorf("l'expression %q n'est pas rendue verbatim", verbatim)
		}
	}
	// Les séquences que produirait l'encodeur avec l'échappement HTML actif,
	// écrites sans leur antislash pour rester lisibles telles quelles.
	for _, echappe := range []string{"u003c", "u003e", "u0026"} {
		if strings.Contains(sortie, echappe) {
			t.Errorf("la sortie contient une séquence échappée en %s", echappe)
		}
	}
}

// L'indentation fait partie du format : un calque se versionne dans Git.
func TestSerialiserIndenteDeDeuxEspaces(t *testing.T) {
	t.Parallel()

	donnees, err := Serialiser(physiqueDeReference())
	if err != nil {
		t.Fatalf("sérialisation : %v", err)
	}

	lignes := strings.Split(string(donnees), "\n")
	if len(lignes) < 2 {
		t.Fatal("sortie non indentée")
	}
	if !strings.HasPrefix(lignes[1], `  "`) {
		t.Errorf("indentation de premier niveau inattendue : %q", lignes[1])
	}
}

// Toutes les collections : une seule oubliée fait dériver l'empreinte.
func TestTrierOrdonneToutesLesCollections(t *testing.T) {
	t.Parallel()

	p := physiqueDeReference()
	p.Trier()

	tables := make([]string, 0, len(p.Tables))
	for _, tbl := range p.Tables {
		tables = append(tables, tbl.Schema+"."+tbl.Nom)
	}
	attenduTables := []string{"archive.audit", "public.client", "public.commande"}
	comparerChaines(t, "tables", tables, attenduTables)

	commande := p.TableParNom("public", "commande")
	if commande == nil {
		t.Fatal("table public.commande introuvable après tri")
	}

	positions := make([]int, 0, len(commande.Colonnes))
	for _, c := range commande.Colonnes {
		positions = append(positions, c.Position)
	}
	for i := 1; i < len(positions); i++ {
		if positions[i-1] > positions[i] {
			t.Errorf("colonnes non triées par position : %v", positions)
			break
		}
	}

	comparerChaines(t, "vérifications",
		nomsVerifications(commande.Verifications), []string{"ck_client", "ck_total_positif"})
	comparerChaines(t, "index",
		nomsIndex(commande.Index), []string{"idx_client", "idx_total"})

	sequences := make([]string, 0, len(p.Sequences))
	for _, s := range p.Sequences {
		sequences = append(sequences, s.Schema+"."+s.Nom)
	}
	comparerChaines(t, "séquences", sequences, []string{"archive.audit_id_seq", "public.client_id_seq"})

	comparerChaines(t, "types énumérés",
		[]string{p.TypesEnumeres[0].Nom, p.TypesEnumeres[1].Nom}, []string{"canal", "statut"})
	comparerChaines(t, "vues",
		[]string{p.Vues[0].Nom, p.Vues[1].Nom}, []string{"v_clients", "v_commandes"})
}

// Index et vérifications sans nom sont courants sur du legacy. Pour ces ex
// æquo, seul l'ordre du catalogue départage : le tri doit le préserver.
func TestTrierPreserveLOrdreDesExAequo(t *testing.T) {
	t.Parallel()

	// Au-delà d'une vingtaine d'éléments : en deçà, sort.Slice retombe sur un
	// tri par insertion qui se trouve être stable, et le test ne mordrait pas.
	colonnes := []string{
		"e", "c", "a", "d", "b", "f", "h", "g", "m", "j",
		"p", "l", "n", "i", "k", "o", "t", "q", "s", "r",
		"w", "u", "v", "y", "x",
	}

	p := physiqueDeReference()
	tbl := p.TableParNom("public", "client")
	tbl.Index = nil
	for _, c := range colonnes {
		tbl.Index = append(tbl.Index, Index{Colonnes: []string{c}})
	}

	p.Trier()

	tbl = p.TableParNom("public", "client")
	for i, attendue := range colonnes {
		if tbl.Index[i].Colonnes[0] != attendue {
			t.Errorf("index anonymes réordonnés : position %d porte %q, attendu %q",
				i, tbl.Index[i].Colonnes[0], attendue)
		}
	}
}

// CalculerEmpreinte trie, puis Ecrire retrie : non idempotent, l'empreinte
// divergerait du document écrit.
func TestTrierEstIdempotent(t *testing.T) {
	t.Parallel()

	p := physiqueDeReference()
	p.Trier()
	apresUn, err := Serialiser(p)
	if err != nil {
		t.Fatalf("sérialisation : %v", err)
	}

	p.Trier()
	apresDeux, err := Serialiser(p)
	if err != nil {
		t.Fatalf("sérialisation : %v", err)
	}

	if string(apresUn) != string(apresDeux) {
		t.Error("un second tri modifie le document")
	}
}

// Deux extractions de la même base à deux instants différents doivent donner
// la même empreinte, sinon le diff signale une dérive à chaque passage.
func TestEmpreinteExclutExtraitLe(t *testing.T) {
	t.Parallel()

	matin := physiqueDeReference()
	matin.Source.ExtraitLe = "2026-08-22T08:00:00Z"
	soir := physiqueDeReference()
	soir.Source.ExtraitLe = "2026-12-31T23:59:59Z"

	empreinteMatin := empreinteOuEchec(t, matin)
	empreinteSoir := empreinteOuEchec(t, soir)

	if empreinteMatin != empreinteSoir {
		t.Errorf("l'horodatage d'extraction change l'empreinte :\n  %s\n  %s", empreinteMatin, empreinteSoir)
	}
}

// Recalcul sur un calque déjà écrit : sinon on ne peut plus la vérifier.
func TestEmpreinteExclutSaPropreValeur(t *testing.T) {
	t.Parallel()

	p := physiqueDeReference()
	avant := empreinteOuEchec(t, p)

	p.Source.Empreinte = avant
	apres := empreinteOuEchec(t, p)

	if avant != apres {
		t.Errorf("l'empreinte déjà posée modifie le recalcul :\n  %s\n  %s", avant, apres)
	}
}

// Le parallélisme de l'introspection ne doit pas transparaître dans la sortie.
func TestEmpreinteIndependanteDeLOrdreDeCollecte(t *testing.T) {
	t.Parallel()

	direct := physiqueDeReference()
	inverse := physiqueDeReference()
	for i, j := 0, len(inverse.Tables)-1; i < j; i, j = i+1, j-1 {
		inverse.Tables[i], inverse.Tables[j] = inverse.Tables[j], inverse.Tables[i]
	}
	for i, j := 0, len(inverse.Vues)-1; i < j; i, j = i+1, j-1 {
		inverse.Vues[i], inverse.Vues[j] = inverse.Vues[j], inverse.Vues[i]
	}

	if empreinteOuEchec(t, direct) != empreinteOuEchec(t, inverse) {
		t.Error("l'ordre de collecte change l'empreinte")
	}
}

// Le pendant de l'exclusion de l'horodatage : le schéma, lui, doit compter.
func TestEmpreinteChangeAvecLeContenu(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom   string
		muter func(*Physique)
	}{
		{"ajout de colonne", func(p *Physique) {
			tbl := p.TableParNom("public", "client")
			tbl.Colonnes = append(tbl.Colonnes, Colonne{
				Nom: "email", Position: 3, TypeBrut: "text", TypeNormalise: TypeTexte,
			})
		}},
		{"type brut modifié", func(p *Physique) {
			p.TableParNom("public", "client").ColonneParNom("id").TypeBrut = "bigint"
		}},
		{"nullabilité modifiée", func(p *Physique) {
			p.TableParNom("public", "client").ColonneParNom("actif").Nullable = true
		}},
		{"échelle retirée", func(p *Physique) {
			p.TableParNom("public", "commande").ColonneParNom("total").Echelle = nil
		}},
		{"action référentielle modifiée", func(p *Physique) {
			p.TableParNom("public", "commande").ClesEtrangeres[0].ALaSuppression = ActionCascade
		}},
	}

	reference := empreinteOuEchec(t, physiqueDeReference())

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			p := physiqueDeReference()
			c.muter(p)
			if empreinteOuEchec(t, p) == reference {
				t.Error("l'empreinte n'a pas bougé alors que le schéma a changé")
			}
		})
	}
}

// decimal(10,0) n'est pas int : échelle à zéro et échelle absente diffèrent.
func TestEmpreinteDistingueZeroEtAbsent(t *testing.T) {
	t.Parallel()

	avecZero := physiqueDeReference()
	avecZero.TableParNom("public", "commande").ColonneParNom("total").Echelle = ptrInt(0)

	sansEchelle := physiqueDeReference()
	sansEchelle.TableParNom("public", "commande").ColonneParNom("total").Echelle = nil

	if empreinteOuEchec(t, avecZero) == empreinteOuEchec(t, sansEchelle) {
		t.Error("une échelle à zéro et une échelle absente produisent la même empreinte")
	}
}

// Le JSON Schema contraint la forme de l'empreinte.
func TestEmpreinteRespecteLeFormatDuSchema(t *testing.T) {
	t.Parallel()

	empreinte := empreinteOuEchec(t, physiqueDeReference())
	if !regexp.MustCompile(`^sha256:[0-9a-f]{64}$`).MatchString(empreinte) {
		t.Errorf("empreinte hors du format annoncé par le JSON Schema : %q", empreinte)
	}
}

// L'empreinte écrite est celle du document sans elle, et survit à la relecture.
func TestEcrirePoseLEmpreinteCalculee(t *testing.T) {
	t.Parallel()

	p := physiqueDeReference()
	attendue := empreinteOuEchec(t, physiqueDeReference())

	chemin := filepath.Join(t.TempDir(), "gescom.calque.json")
	if err := p.Ecrire(chemin); err != nil {
		t.Fatalf("écriture : %v", err)
	}

	if p.Source.Empreinte != attendue {
		t.Errorf("empreinte posée %q, attendue %q", p.Source.Empreinte, attendue)
	}

	relu, err := LirePhysique(chemin)
	if err != nil {
		t.Fatalf("relecture : %v", err)
	}
	if relu.Source.Empreinte != attendue {
		t.Errorf("empreinte du fichier %q, attendue %q", relu.Source.Empreinte, attendue)
	}
}

// L'aller-retour octet pour octet. Sans lui, le déterminisme se dégrade sans
// que rien ne le signale.
func TestAllerRetourOctetPourOctet(t *testing.T) {
	t.Parallel()

	repertoire := t.TempDir()
	chemin := filepath.Join(repertoire, "gescom.calque.json")

	if err := physiqueDeReference().Ecrire(chemin); err != nil {
		t.Fatalf("écriture : %v", err)
	}
	surDisque, err := os.ReadFile(chemin)
	if err != nil {
		t.Fatalf("lecture du fichier écrit : %v", err)
	}

	relu, err := LirePhysique(chemin)
	if err != nil {
		t.Fatalf("relecture : %v", err)
	}
	reserialise, err := Serialiser(relu)
	if err != nil {
		t.Fatalf("resérialisation : %v", err)
	}

	if string(surDisque) != string(reserialise) {
		t.Error("la resérialisation d'un calque relu diffère du fichier source")
	}

	// Réécrire depuis le calque relu doit reproduire le même fichier.
	second := filepath.Join(repertoire, "second.calque.json")
	if err := relu.Ecrire(second); err != nil {
		t.Fatalf("seconde écriture : %v", err)
	}
	contenuSecond, err := os.ReadFile(second)
	if err != nil {
		t.Fatalf("lecture de la seconde écriture : %v", err)
	}
	if string(surDisque) != string(contenuSecond) {
		t.Error("deux écritures successives du même calque produisent deux fichiers différents")
	}
}

// Un format plus récent porterait des champs ignorés en silence.
func TestLirePhysiqueRefuseUneVersionPlusRecente(t *testing.T) {
	t.Parallel()

	p := physiqueDeReference()
	p.VersionRI = VersionCourante + 1

	chemin := filepath.Join(t.TempDir(), "futur.calque.json")
	donnees, err := Serialiser(p)
	if err != nil {
		t.Fatalf("sérialisation : %v", err)
	}
	if err := os.WriteFile(chemin, donnees, 0o600); err != nil {
		t.Fatalf("écriture : %v", err)
	}

	if _, err := LirePhysique(chemin); err == nil {
		t.Fatal("une version supérieure a été acceptée")
	}
}

// Le pendant du refus.
func TestLirePhysiqueAccepteLaVersionCourante(t *testing.T) {
	t.Parallel()

	chemin := filepath.Join(t.TempDir(), "gescom.calque.json")
	if err := physiqueDeReference().Ecrire(chemin); err != nil {
		t.Fatalf("écriture : %v", err)
	}

	relu, err := LirePhysique(chemin)
	if err != nil {
		t.Fatalf("relecture : %v", err)
	}
	if relu.VersionRI != VersionCourante {
		t.Errorf("version relue %d, attendue %d", relu.VersionRI, VersionCourante)
	}
}

// Jamais un calque vide sur erreur : la suite le prendrait pour une base sans
// table.
func TestLirePhysiqueEchoueProprement(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom     string
		contenu string
		ecrire  bool
	}{
		{nom: "fichier absent", ecrire: false},
		{nom: "json invalide", contenu: "{ceci n'est pas du json", ecrire: true},
		{nom: "racine de mauvais type", contenu: `["pas un objet"]`, ecrire: true},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			chemin := filepath.Join(t.TempDir(), "calque.json")
			if c.ecrire {
				if err := os.WriteFile(chemin, []byte(c.contenu), 0o600); err != nil {
					t.Fatalf("écriture : %v", err)
				}
			}

			if _, err := LirePhysique(chemin); err == nil {
				t.Error("aucune erreur retournée")
			}
		})
	}
}

// Sans omitempty, chaque diff serait noyé sous des nulls et des zéros.
func TestChampsOptionnelsAbsentsDeLaSortie(t *testing.T) {
	t.Parallel()

	minimal := &Physique{
		VersionRI: VersionCourante,
		Source:    Source{SGBD: "postgres", Version: "16.2", Catalogue: "c", Schema: "public"},
		Tables: []Table{{
			Nom:      "t",
			Schema:   "public",
			Colonnes: []Colonne{{Nom: "id", Position: 1, TypeBrut: "integer", TypeNormalise: TypeEntier}},
		}},
	}

	donnees, err := Serialiser(minimal)
	if err != nil {
		t.Fatalf("sérialisation : %v", err)
	}

	sortie := string(donnees)
	for _, absent := range []string{
		"longueur", "precision", "echelle", "defaut", "generee",
		"sequences", "types_enumeres", "vues", "statistiques",
		"cles_etrangeres", "unicites", "index", "verifications",
	} {
		if strings.Contains(sortie, `"`+absent+`"`) {
			t.Errorf("le champ optionnel %q apparaît alors qu'il est vide", absent)
		}
	}
}

// Le lecteur PHP ne partage pas le décodeur Go.
func TestSortieRelisibleParUnDecodeurStandard(t *testing.T) {
	t.Parallel()

	donnees, err := Serialiser(physiqueDeReference())
	if err != nil {
		t.Fatalf("sérialisation : %v", err)
	}

	var generique map[string]any
	if err := json.Unmarshal(donnees, &generique); err != nil {
		t.Fatalf("sortie non relisible : %v", err)
	}
	if _, ok := generique["version_ri"]; !ok {
		t.Error("version_ri absente de la sortie")
	}
}

// comparerChaines tient lieu d'assertion : le projet n'embarque pas de
// bibliothèque pour ça.
func comparerChaines(t *testing.T, quoi string, obtenu, attendu []string) {
	t.Helper()

	if len(obtenu) != len(attendu) {
		t.Errorf("%s : %d éléments, %d attendus (%v)", quoi, len(obtenu), len(attendu), obtenu)
		return
	}
	for i := range attendu {
		if obtenu[i] != attendu[i] {
			t.Errorf("%s : obtenu %v, attendu %v", quoi, obtenu, attendu)
			return
		}
	}
}

// nomsVerifications projette les noms pour comparer un ordre.
func nomsVerifications(v []Verification) []string {
	noms := make([]string, 0, len(v))
	for _, e := range v {
		noms = append(noms, e.Nom)
	}
	return noms
}

// nomsIndex projette les noms pour comparer un ordre.
func nomsIndex(idx []Index) []string {
	noms := make([]string, 0, len(idx))
	for _, e := range idx {
		noms = append(noms, e.Nom)
	}
	return noms
}

// empreinteOuEchec : le calcul ne peut échouer que sur un calque non
// sérialisable, ce qui serait un défaut à part entière.
func empreinteOuEchec(t *testing.T, p *Physique) string {
	t.Helper()

	empreinte, err := p.CalculerEmpreinte()
	if err != nil {
		t.Fatalf("calcul d'empreinte : %v", err)
	}
	return empreinte
}
