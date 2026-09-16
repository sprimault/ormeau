// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package diff

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
)

// Divergent fait échouer un pipeline. Le faux négatif est le pire cas : CI
// verte alors que la base a bougé sous les entités.
func TestDivergent(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom     string
		ecarts  []Ecart
		attendu bool
	}{
		{"aucun écart", nil, false},
		{"tranche vide", []Ecart{}, false},
		{"un ajout", []Ecart{{Genre: Ajout, Objet: ObjetColonne, Schema: "public", Table: "client", Nom: "email"}}, true},
		{"une suppression", []Ecart{{Genre: Suppression, Objet: ObjetTable, Schema: "public", Table: "facture"}}, true},
		{
			"une modification",
			[]Ecart{{Genre: Modification, Objet: ObjetColonne, Schema: "public", Table: "client", Nom: "id", Propriete: "type_brut", Avant: "integer", Apres: "bigint"}},
			true,
		},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			if obtenu := Divergent(c.ecarts); obtenu != c.attendu {
				t.Errorf("Divergent(%v) = %v, attendu %v", c.ecarts, obtenu, c.attendu)
			}
		})
	}
}

// Chemin se lit sans connaître la structure : schéma, table, objet, nom,
// propriété.
func TestChemin(t *testing.T) {
	t.Parallel()

	cas := map[string]Ecart{
		"public.client colonne email : type_brut": {Objet: ObjetColonne, Schema: "public", Table: "client", Nom: "email", Propriete: "type_brut"},
		"public.facture table":                    {Objet: ObjetTable, Schema: "public", Table: "facture"},
		"public sequence facture_id_seq":          {Objet: ObjetSequence, Schema: "public", Nom: "facture_id_seq"},
	}
	for attendu, e := range cas {
		if obtenu := e.Chemin(); obtenu != attendu {
			t.Errorf("Chemin() = %q, attendu %q", obtenu, attendu)
		}
	}
}

// entierP rend un pointeur vers n, pour écrire les calques de test d'une ligne.
func entierP(n int) *int { return &n }

// entier64P rend un pointeur vers n.
func entier64P(n int64) *int64 { return &n }

// calqueDeBase porte un exemplaire de chaque objet comparé. Chaque cas en part
// et n'y change que ce qu'il éprouve.
func calqueDeBase() *calque.Physique {
	return &calque.Physique{
		VersionRI: 1,
		Tables: []calque.Table{
			{
				Nom: "client", Schema: "public", Commentaire: "Fiche client",
				Colonnes: []calque.Colonne{
					{Nom: "id", Position: 1, TypeBrut: "integer", TypeNormalise: calque.TypeEntier, AutoIncrement: true},
					{Nom: "nom", Position: 2, TypeBrut: "character varying(80)", TypeNormalise: calque.TypeTexte, Longueur: entierP(80), Collation: "default"},
					{Nom: "statut", Position: 3, TypeBrut: "character varying(20)", TypeNormalise: calque.TypeTexte, Nullable: true,
						Defaut: &calque.Defaut{Genre: calque.DefautLitteral, Valeur: "ACTIF"}},
				},
				ClePrimaire: &calque.ClePrimaire{Nom: "client_pkey", Colonnes: []string{"id"}},
				Unicites:    []calque.Contrainte{{Nom: "uq_client_nom", Colonnes: []string{"nom"}}},
				Index: []calque.Index{
					{Nom: "ix_nom", Colonnes: []string{"nom"}, Methode: "btree"},
					{Nom: "ix_nom_prefixe", Colonnes: []string{"nom"}, Methode: "btree", Operateurs: []string{"text_pattern_ops"}},
					{Nom: "ix_actifs", Colonnes: []string{"id"}, Methode: "btree", Predicat: "(statut = 'ACTIF')"},
				},
				Verifications: []calque.Verification{{Nom: "ck_statut", Expression: "CHECK (statut <> '')"}},
			},
			{
				Nom: "commande", Schema: "public",
				Colonnes: []calque.Colonne{
					{Nom: "id", Position: 1, TypeBrut: "integer", TypeNormalise: calque.TypeEntier,
						Defaut: &calque.Defaut{Genre: calque.DefautSequence, Valeur: "nextval('public.commande_id_seq'::regclass)"}},
					{Nom: "client_id", Position: 2, TypeBrut: "integer", TypeNormalise: calque.TypeEntier},
				},
				ClePrimaire: &calque.ClePrimaire{Nom: "commande_pkey", Colonnes: []string{"id"}},
				ClesEtrangeres: []calque.CleEtrangere{{
					Nom: "commande_client_id_fkey", Colonnes: []string{"client_id"},
					TableCible: "client", SchemaCible: "public", ColonnesCibles: []string{"id"},
					ALaSuppression: calque.ActionCascade,
				}},
			},
		},
		Sequences:     []calque.Sequence{{Nom: "commande_id_seq", Schema: "public", Increment: 1, Minimum: entier64P(1), Maximum: entier64P(2147483647)}},
		TypesEnumeres: []calque.TypeEnumere{{Nom: "canal", Schema: "public", Valeurs: []string{"web", "agence"}}},
		Vues:          []calque.Vue{{Nom: "v_actifs", Schema: "public", Definition: " SELECT id FROM public.client;"}},
	}
}

// Chaque objet comparé, dans chaque genre d'écart : ce que le cas change de b,
// et la liste exacte attendue, dans l'ordre de sortie.
func TestComparer(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom      string
		modifier func(b *calque.Physique)
		attendu  []Ecart
	}{
		{"calques identiques", func(*calque.Physique) {}, nil},
		{
			"table ajoutée, sans un écart par colonne",
			func(b *calque.Physique) {
				b.Tables = append(b.Tables, calque.Table{Nom: "facture", Schema: "public", Colonnes: []calque.Colonne{{Nom: "id", Position: 1}}})
			},
			[]Ecart{{Genre: Ajout, Objet: ObjetTable, Schema: "public", Table: "facture"}},
		},
		{
			"table supprimée",
			func(b *calque.Physique) { b.Tables = b.Tables[:1] },
			[]Ecart{{Genre: Suppression, Objet: ObjetTable, Schema: "public", Table: "commande"}},
		},
		{
			"commentaire et options de table",
			func(b *calque.Physique) {
				b.Tables[0].Commentaire = ""
				b.Tables[0].Options = &calque.OptionsTable{Moteur: "InnoDB"}
			},
			[]Ecart{
				{Genre: Modification, Objet: ObjetTable, Schema: "public", Table: "client", Propriete: "commentaire", Avant: "Fiche client"},
				{Genre: Modification, Objet: ObjetTable, Schema: "public", Table: "client", Propriete: "options.moteur", Apres: "InnoDB"},
			},
		},
		{
			"colonne ajoutée et colonne supprimée",
			func(b *calque.Physique) {
				c := &b.Tables[0]
				c.Colonnes = append(c.Colonnes[:2], calque.Colonne{Nom: "email", Position: 3, TypeBrut: "text"})
			},
			[]Ecart{
				{Genre: Ajout, Objet: ObjetColonne, Schema: "public", Table: "client", Nom: "email"},
				{Genre: Suppression, Objet: ObjetColonne, Schema: "public", Table: "client", Nom: "statut"},
			},
		},
		{
			"propriétés de colonne, longueur absente distincte de zéro",
			func(b *calque.Physique) {
				col := &b.Tables[0].Colonnes[1]
				col.TypeBrut = "text"
				col.Longueur = entierP(0)
				b.Tables[0].Colonnes[2].Defaut = &calque.Defaut{Genre: calque.DefautExpression, Valeur: "ACTIF"}
			},
			[]Ecart{
				{Genre: Modification, Objet: ObjetColonne, Schema: "public", Table: "client", Nom: "nom", Propriete: "longueur", Avant: "80", Apres: "0"},
				{Genre: Modification, Objet: ObjetColonne, Schema: "public", Table: "client", Nom: "nom", Propriete: "type_brut", Avant: "character varying(80)", Apres: "text"},
				{Genre: Modification, Objet: ObjetColonne, Schema: "public", Table: "client", Nom: "statut", Propriete: "defaut", Avant: "litteral ACTIF", Apres: "expression ACTIF"},
			},
		},
		{
			"ordre des colonnes de la clé primaire",
			func(b *calque.Physique) {
				b.Tables[0].ClePrimaire = &calque.ClePrimaire{Nom: "client_pkey", Colonnes: []string{"id", "nom"}}
			},
			[]Ecart{{Genre: Modification, Objet: ObjetClePrimaire, Schema: "public", Table: "client", Nom: "client_pkey", Propriete: "colonnes", Avant: `"id"`, Apres: `"id", "nom"`}},
		},
		{
			"clé étrangère renommée : une modification, pas une suppression suivie d'un ajout",
			func(b *calque.Physique) { b.Tables[1].ClesEtrangeres[0].Nom = "FK_6EEAA67D19EB6921" },
			[]Ecart{{Genre: Modification, Objet: ObjetCleEtrangere, Schema: "public", Table: "commande", Nom: "commande_client_id_fkey", Propriete: "nom", Avant: "commande_client_id_fkey", Apres: "FK_6EEAA67D19EB6921"}},
		},
		{
			"action de clé étrangère",
			func(b *calque.Physique) { b.Tables[1].ClesEtrangeres[0].ALaSuppression = "" },
			[]Ecart{{Genre: Modification, Objet: ObjetCleEtrangere, Schema: "public", Table: "commande", Nom: "commande_client_id_fkey", Propriete: "a_la_suppression", Avant: "cascade"}},
		},
		{
			"clé étrangère vers une autre cible",
			func(b *calque.Physique) { b.Tables[1].ClesEtrangeres[0].TableCible = "compte" },
			[]Ecart{
				{Genre: Ajout, Objet: ObjetCleEtrangere, Schema: "public", Table: "commande", Nom: "commande_client_id_fkey"},
				{Genre: Suppression, Objet: ObjetCleEtrangere, Schema: "public", Table: "commande", Nom: "commande_client_id_fkey"},
			},
		},
		{
			"unicité renommée",
			func(b *calque.Physique) { b.Tables[0].Unicites[0].Nom = "UNIQ_C7440455" },
			[]Ecart{{Genre: Modification, Objet: ObjetUnicite, Schema: "public", Table: "client", Nom: "uq_client_nom", Propriete: "nom", Avant: "uq_client_nom", Apres: "UNIQ_C7440455"}},
		},
		{
			"prédicat d'index perdu : une modification",
			func(b *calque.Physique) { b.Tables[0].Index[2].Predicat = "" },
			[]Ecart{{Genre: Modification, Objet: ObjetIndex, Schema: "public", Table: "client", Nom: "ix_actifs", Propriete: "predicat", Avant: "(statut = 'ACTIF')"}},
		},
		{
			"index ajouté par Doctrine sur une clé étrangère",
			func(b *calque.Physique) {
				b.Tables[1].Index = []calque.Index{{Nom: "IDX_6EEAA67D19EB6921", Colonnes: []string{"client_id"}, Methode: "btree"}}
			},
			[]Ecart{{Genre: Ajout, Objet: ObjetIndex, Schema: "public", Table: "commande", Nom: "IDX_6EEAA67D19EB6921"}},
		},
		{
			"deux index sur les mêmes colonnes, l'un renommé",
			func(b *calque.Physique) { b.Tables[0].Index[1].Nom = "IDX_PREFIXE" },
			[]Ecart{{Genre: Modification, Objet: ObjetIndex, Schema: "public", Table: "client", Nom: "ix_nom_prefixe", Propriete: "nom", Avant: "ix_nom_prefixe", Apres: "IDX_PREFIXE"}},
		},
		{
			"deux index sur les mêmes colonnes, classes d'opérateurs échangées",
			func(b *calque.Physique) {
				b.Tables[0].Index[0].Operateurs = []string{"text_pattern_ops"}
				b.Tables[0].Index[1].Operateurs = nil
			},
			[]Ecart{
				{Genre: Modification, Objet: ObjetIndex, Schema: "public", Table: "client", Nom: "ix_nom", Propriete: "operateurs", Apres: `"text_pattern_ops"`},
				{Genre: Modification, Objet: ObjetIndex, Schema: "public", Table: "client", Nom: "ix_nom_prefixe", Propriete: "operateurs", Avant: `"text_pattern_ops"`},
			},
		},
		{
			"vérification dont l'expression change",
			func(b *calque.Physique) { b.Tables[0].Verifications[0].Expression = "CHECK (statut IS NOT NULL)" },
			[]Ecart{
				{Genre: Ajout, Objet: ObjetVerification, Schema: "public", Table: "client", Nom: "ck_statut"},
				{Genre: Suppression, Objet: ObjetVerification, Schema: "public", Table: "client", Nom: "ck_statut"},
			},
		},
		{
			"vérification sans nom, désignée par son expression",
			func(b *calque.Physique) {
				b.Tables[0].Verifications = append(b.Tables[0].Verifications, calque.Verification{Expression: "CHECK (id > 0)"})
			},
			[]Ecart{{Genre: Ajout, Objet: ObjetVerification, Schema: "public", Table: "client", Nom: "(CHECK (id > 0))"}},
		},
		{
			"vue, séquence et type énuméré",
			func(b *calque.Physique) {
				b.Vues[0].Materialisee = true
				b.Sequences[0].Maximum = entier64P(9223372036854775807)
				b.TypesEnumeres[0].Valeurs = []string{"agence", "web"}
			},
			[]Ecart{
				{Genre: Modification, Objet: ObjetVue, Schema: "public", Nom: "v_actifs", Propriete: "materialisee", Avant: "false", Apres: "true"},
				{Genre: Modification, Objet: ObjetSequence, Schema: "public", Nom: "commande_id_seq", Propriete: "maximum", Avant: "2147483647", Apres: "9223372036854775807"},
				{Genre: Modification, Objet: ObjetTypeEnumere, Schema: "public", Nom: "canal", Propriete: "valeurs", Avant: `"web", "agence"`, Apres: `"agence", "web"`},
			},
		},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			b := calqueDeBase()
			c.modifier(b)
			verifierEcarts(t, Comparer(calqueDeBase(), b, Options{}), c.attendu)
		})
	}
}

// Une base recréée dans un autre schéma se compare par traduction : noms
// d'objets et cibles de clés étrangères suivent, les expressions restent
// verbatim et sortent en écart.
func TestComparerTraduitLesSchemas(t *testing.T) {
	t.Parallel()

	a := calqueDeBase()
	b := calqueDeBase()
	b.Tables[0].Schema, b.Tables[1].Schema = "gescom", "gescom"
	b.Tables[1].ClesEtrangeres[0].SchemaCible = "gescom"
	b.Sequences[0].Schema, b.TypesEnumeres[0].Schema, b.Vues[0].Schema = "gescom", "gescom", "gescom"

	verifierEcarts(t, Comparer(a, b, Options{Schemas: map[string]string{"public": "gescom"}}), nil)

	b.Tables[1].Colonnes[0].Defaut.Valeur = "nextval('gescom.commande_id_seq'::regclass)"
	verifierEcarts(t, Comparer(a, b, Options{Schemas: map[string]string{"public": "gescom"}}), []Ecart{
		{Genre: Modification, Objet: ObjetColonne, Schema: "gescom", Table: "commande", Nom: "id", Propriete: "defaut",
			Avant: "sequence nextval('public.commande_id_seq'::regclass)", Apres: "sequence nextval('gescom.commande_id_seq'::regclass)"},
	})

	sansTraduction := Comparer(a, b, Options{})
	if len(sansTraduction) != 10 {
		t.Errorf("sans traduction : %d écarts, attendu 10 (deux tables, une vue, une séquence, un type, de chaque côté) :\n%s",
			len(sansTraduction), lister(sansTraduction))
	}
}

// Le résultat ne dépend pas de l'ordre des listes : un calque lu d'un pilote
// trie par nom, un calque écrit à la main non.
func TestComparerEstDeterministe(t *testing.T) {
	t.Parallel()

	a := calqueDeBase()
	b := calqueDeBase()
	b.Tables[0].Index[1].Nom = "IDX_PREFIXE"
	b.Tables[0].Colonnes = append(b.Tables[0].Colonnes, calque.Colonne{Nom: "email", Position: 4})
	b.Tables[1].ClesEtrangeres[0].Nom = "FK_X"
	reference := Comparer(a, b, Options{})

	permute := calqueDeBase()
	*permute = *b
	permute.Tables = slices.Clone(b.Tables)
	slices.Reverse(permute.Tables)
	for i := range permute.Tables {
		permute.Tables[i].Colonnes = slices.Clone(permute.Tables[i].Colonnes)
		slices.Reverse(permute.Tables[i].Colonnes)
		permute.Tables[i].Index = slices.Clone(permute.Tables[i].Index)
		slices.Reverse(permute.Tables[i].Index)
	}
	aPermute := calqueDeBase()
	slices.Reverse(aPermute.Tables[0].Index)

	verifierEcarts(t, Comparer(aPermute, permute, Options{}), reference)
}

// Sur l'extraction réelle de tests/ddl/ : un calque ne diffère pas de lui-même,
// et trois retouches typiques de l'aller-retour sortent exactement.
func TestComparerSurGescom(t *testing.T) {
	t.Parallel()

	const chemin = "../../tests/reference/inference/gescom/physique.json"
	a, err := calque.LirePhysique(chemin)
	if err != nil {
		t.Fatalf("lecture : %v", err)
	}
	b, err := calque.LirePhysique(chemin)
	if err != nil {
		t.Fatalf("lecture : %v", err)
	}
	if ecarts := Comparer(a, b, Options{}); len(ecarts) != 0 {
		t.Fatalf("gescom diffère de lui-même :\n%s", lister(ecarts))
	}

	client := b.TableParNom("gescom", "t_client")
	client.Colonnes = append(client.Colonnes, calque.Colonne{Nom: "cli_email", Position: 10, TypeBrut: "text", TypeNormalise: calque.TypeTexte, Nullable: true})
	for i := range client.Index {
		if client.Index[i].Nom == "ix_cli_actifs" {
			client.Index[i].Predicat = ""
		}
	}
	client.ClesEtrangeres[0].Nom = "FK_C7440455B0B7B3A"

	verifierEcarts(t, Comparer(a, b, Options{}), []Ecart{
		{Genre: Ajout, Objet: ObjetColonne, Schema: "gescom", Table: "t_client", Nom: "cli_email"},
		{Genre: Modification, Objet: ObjetCleEtrangere, Schema: "gescom", Table: "t_client", Nom: "fk_client_commercial", Propriete: "nom", Avant: "fk_client_commercial", Apres: "FK_C7440455B0B7B3A"},
		{Genre: Modification, Objet: ObjetIndex, Schema: "gescom", Table: "t_client", Nom: "ix_cli_actifs", Propriete: "predicat", Avant: "((cli_statut)::text = 'ACTIF'::text)"},
	})
}

// Chaque champ du calque physique est comparé, ou exclu avec sa raison. Un
// champ ajouté au format sans décision sur sa comparaison ferait passer en
// silence l'écart qu'il porte : le faux négatif que ce paquet existe pour
// éviter.
func TestChaqueChampEstCompareOuExclu(t *testing.T) {
	t.Parallel()

	// Les champs qui ne sont pas des propriétés comparées, avec leur raison.
	exclus := map[reflect.Type]map[string]string{
		reflect.TypeFor[calque.Physique](): {
			"version_ri":     "version du format, pas du schéma",
			"source":         "décrit l'extraction, pas le schéma",
			"tables":         "objets rapprochés",
			"sequences":      "objets rapprochés",
			"types_enumeres": "objets rapprochés",
			"vues":           "objets rapprochés",
			"statistiques":   "données, pas schéma",
		},
		reflect.TypeFor[calque.Table](): {
			"nom": "identité", "schema": "identité",
			"colonnes": "objets rapprochés", "cle_primaire": "objet rapproché", "cles_etrangeres": "objets rapprochés",
			"unicites": "objets rapprochés", "index": "objets rapprochés", "verifications": "objets rapprochés",
			"options": "décomposé en options.moteur et options.collation",
		},
		reflect.TypeFor[calque.OptionsTable](): {
			"moteur": "comparé comme options.moteur", "collation": "comparé comme options.collation",
		},
		reflect.TypeFor[calque.Colonne]():     {"nom": "identité"},
		reflect.TypeFor[calque.ClePrimaire](): {},
		reflect.TypeFor[calque.CleEtrangere](): {
			"colonnes": "identité", "table_cible": "identité", "schema_cible": "identité", "colonnes_cibles": "identité",
		},
		reflect.TypeFor[calque.Contrainte]():   {"colonnes": "identité"},
		reflect.TypeFor[calque.Index]():        {"colonnes": "identité"},
		reflect.TypeFor[calque.Verification](): {"expression": "identité"},
		reflect.TypeFor[calque.Vue](): {
			"nom": "identité", "schema": "identité",
			"colonnes": "aucun pilote ne les produit : rien à comparer tant qu'ils ne le font pas",
		},
		reflect.TypeFor[calque.Sequence]():    {"nom": "identité", "schema": "identité"},
		reflect.TypeFor[calque.TypeEnumere](): {"nom": "identité", "schema": "identité"},
		reflect.TypeFor[calque.Defaut](): {
			"genre": "rendu dans defaut", "valeur": "rendu dans defaut", "sequence": "rendu dans defaut",
		},
		reflect.TypeFor[calque.ReferenceSequence](): {"schema": "rendu dans defaut", "nom": "rendu dans defaut"},
		reflect.TypeFor[calque.Generee]():           {"expression": "rendu dans generee", "stockee": "rendu dans generee"},
	}

	compares := map[reflect.Type][]string{
		reflect.TypeFor[calque.Table]():        noms(proprietesTable),
		reflect.TypeFor[calque.Colonne]():      noms(proprietesColonne),
		reflect.TypeFor[calque.ClePrimaire]():  noms(proprietesClePrimaire),
		reflect.TypeFor[calque.CleEtrangere](): noms(proprietesCleEtrangere),
		reflect.TypeFor[calque.Contrainte]():   noms(proprietesUnicite),
		reflect.TypeFor[calque.Index]():        noms(proprietesIndex),
		reflect.TypeFor[calque.Verification](): noms(proprietesVerification),
		reflect.TypeFor[calque.Vue]():          noms(proprietesVue),
		reflect.TypeFor[calque.Sequence]():     noms(proprietesSequence),
		reflect.TypeFor[calque.TypeEnumere]():  noms(proprietesTypeEnumere),
	}

	for typ, raisons := range exclus {
		champs := map[string]bool{}
		for i := range typ.NumField() {
			nom, _, _ := strings.Cut(typ.Field(i).Tag.Get("json"), ",")
			champs[nom] = true
			_, estExclu := raisons[nom]
			estCompare := slices.Contains(compares[typ], nom)
			switch {
			case estExclu && estCompare:
				t.Errorf("%s.%s est à la fois comparé et exclu", typ.Name(), nom)
			case !estExclu && !estCompare:
				t.Errorf("%s.%s n'est ni comparé ni exclu : décider de sa comparaison dans proprietes.go", typ.Name(), nom)
			}
		}
		for nom := range raisons {
			if !champs[nom] {
				t.Errorf("%s.%s exclu, mais le champ n'existe plus", typ.Name(), nom)
			}
		}
		for _, nom := range compares[typ] {
			if !champs[nom] && !strings.HasPrefix(nom, "options.") {
				t.Errorf("%s.%s comparé, mais le champ n'existe pas", typ.Name(), nom)
			}
		}
	}
}

// noms rend les noms d'une liste de propriétés.
func noms[T any](proprietes []propriete[T]) []string {
	var n []string
	for _, p := range proprietes {
		n = append(n, p.nom)
	}
	return n
}

// verifierEcarts compare deux listes d'écarts, ordre compris.
func verifierEcarts(t *testing.T, obtenu, attendu []Ecart) {
	t.Helper()
	if !slices.Equal(obtenu, attendu) {
		t.Errorf("écarts obtenus :\n%s\nattendus :\n%s", lister(obtenu), lister(attendu))
	}
}

// lister met une liste d'écarts en forme pour un message d'échec.
func lister(ecarts []Ecart) string {
	if len(ecarts) == 0 {
		return "  (aucun)"
	}
	var b strings.Builder
	for _, e := range ecarts {
		fmt.Fprintf(&b, "  %s %s [%q -> %q]\n", e.Genre, e.Chemin(), e.Avant, e.Apres)
	}
	return b.String()
}
