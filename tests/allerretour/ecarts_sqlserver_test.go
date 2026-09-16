// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

//go:build allerretour

package allerretour

import (
	"regexp"
	"slices"
	"strings"

	"github.com/sprimault/ormeau/internal/calque"
	"github.com/sprimault/ormeau/internal/diff"
)

// tolerancesSQLServer est la liste fermée de SQL Server, sur le modèle de
// celle de PostgreSQL. Un écart de même cause y reprend l'entrée PostgreSQL
// par son code ; les autres tiennent au dialecte, à sa plateforme DBAL, ou à
// une inférence qui ne lit encore que les formes de PostgreSQL.
var tolerancesSQLServer = []tolerance{
	// IMPOSSIBLE : ce que Doctrine et DBAL ne savent pas recréer.
	reprise("vue_non_recreee"),
	reprise("verification_non_recreee"),
	reprise("unicite_recreee_en_index"),
	reprise("ordres_d_index"),
	reprise("position_identite_derivee"),
	reprise("ordre_des_colonnes_dbal3"),
	{
		code: "simple_precision_recreee_en_double", categorie: impossible, cibles: []string{"orm2-dbal3"},
		pourquoi: "DBAL 3 n'a pas de type smallfloat : une colonne real est recréée en float",
		couvre: func(e diff.Ecart, _ contexte) bool {
			return e.Objet == diff.ObjetColonne && e.Propriete == "type_brut" && e.Avant == "real" && e.Apres == "float"
		},
	},
	{
		code: "nom_de_cle_primaire_genere", categorie: impossible,
		pourquoi: "Doctrine ne nomme pas une clé primaire : SQL Server la nomme PK__ suivi d'un suffixe tiré à la création",
		couvre: func(e diff.Ecart, _ contexte) bool {
			return e.Objet == diff.ObjetClePrimaire && e.Propriete == "nom" && nomClePrimaireSQLServer.MatchString(e.Apres)
		},
	},
	{
		code: "nom_de_cle_etrangere_genere", categorie: impossible,
		pourquoi: "Doctrine nomme ses clés étrangères FK_ suivi d'un hachage",
		couvre:   cleEtrangereNommeeParDoctrine(nomGenereSQLServer, "FK_"),
	},
	{
		code: "index_de_cle_etrangere_ajoute", categorie: impossible,
		pourquoi: "DBAL indexe toute clé étrangère qu'aucun index ne couvre déjà",
		couvre:   indexDeCleEtrangereAjoute(nomGenereSQLServer, "IDX_"),
	},
	{
		code: "index_partiel_non_recree", categorie: impossible,
		pourquoi: "la plateforme SQL Server de DBAL n'écrit pas d'index partiel : le prédicat d'un index filtré est perdu, et une unicité filtrée porte ensuite sur toutes les lignes",
		couvre: func(e diff.Ecart, _ contexte) bool {
			return e.Objet == diff.ObjetIndex && e.Propriete == "predicat" && e.Avant != "" &&
				(e.Apres == "" || filtreNonNul.MatchString(e.Apres))
		},
	},
	{
		code: "unique_filtre_sur_non_nul", categorie: impossible,
		pourquoi: "DBAL ajoute WHERE … IS NOT NULL à tout index unique, pour que plusieurs NULL y passent comme sous PostgreSQL",
		couvre: func(e diff.Ecart, c contexte) bool {
			if e.Objet != diff.ObjetIndex || e.Propriete != "predicat" || !filtreNonNul.MatchString(e.Apres) {
				return false
			}
			table := c.recree.TableParNom(e.Schema, e.Table)
			return table != nil && slices.ContainsFunc(table.Index, func(idx calque.Index) bool {
				return idx.Nom == e.Nom && idx.Unique
			})
		},
	},
	{
		code: "decimal_ecrit_numeric", categorie: impossible,
		pourquoi: "DBAL écrit NUMERIC(p, s) pour un décimal, synonyme de DECIMAL sous SQL Server",
		couvre: func(e diff.Ecart, _ contexte) bool {
			return e.Objet == diff.ObjetColonne && e.Propriete == "type_brut" &&
				strings.HasPrefix(e.Avant, "decimal(") && e.Apres == "numeric"+strings.TrimPrefix(e.Avant, "decimal")
		},
	},
	{
		code: "monnaie_recreee_en_numeric", categorie: impossible,
		pourquoi: "Doctrine n'a pas de type monétaire : money et smallmoney sont recréés en NUMERIC de même précision",
		couvre: func(e diff.Ecart, _ contexte) bool {
			return e.Objet == diff.ObjetColonne && e.Propriete == "type_brut" &&
				((e.Avant == "money" && e.Apres == "numeric(19,4)") || (e.Avant == "smallmoney" && e.Apres == "numeric(10,4)"))
		},
	},
	{
		code: "chaine_recreee_en_unicode", categorie: impossible,
		pourquoi: "la plateforme SQL Server de DBAL écrit toute chaîne en NVARCHAR : une colonne VARCHAR change d'encodage",
		couvre: func(e diff.Ecart, _ contexte) bool {
			return e.Objet == diff.ObjetColonne && e.Propriete == "type_brut" &&
				strings.HasPrefix(e.Avant, "varchar(") && e.Apres == "n"+e.Avant
		},
	},
	{
		code: "datetime_recree_en_datetime2", categorie: impossible,
		pourquoi: "DBAL écrit DATETIME2(6) pour une date et heure : datetime et smalldatetime changent de type et de précision",
		couvre: func(e diff.Ecart, c contexte) bool {
			if e.Objet != diff.ObjetColonne {
				return false
			}
			if e.Propriete == "precision_fractionnaire" {
				col := colonneOrigine(c, e)
				return col != nil && (col.TypeBrut == "datetime" || col.TypeBrut == "smalldatetime") && e.Avant == "" && e.Apres == "6"
			}
			return e.Propriete == "type_brut" &&
				(e.Avant == "datetime" || e.Avant == "smalldatetime") && e.Apres == "datetime2(6)"
		},
	},
	{
		code: "tinyint_recree_en_smallint", categorie: impossible,
		pourquoi: "Doctrine n'a pas de type sur un octet : tinyint est recréé en SMALLINT, le plus proche",
		couvre: func(e diff.Ecart, _ contexte) bool {
			return e.Objet == diff.ObjetColonne && e.Propriete == "type_brut" && e.Avant == "tinyint" && e.Apres == "smallint"
		},
	},
	{
		code: "precision_heure", categorie: impossible,
		pourquoi: "DBAL écrit TIME(0) : la précision fractionnaire d'origine est perdue",
		couvre: func(e diff.Ecart, c contexte) bool {
			if e.Objet != diff.ObjetColonne {
				return false
			}
			if e.Propriete == "precision_fractionnaire" {
				col := colonneOrigine(c, e)
				return col != nil && strings.HasPrefix(col.TypeBrut, "time(") && e.Apres == "0"
			}
			return e.Propriete == "type_brut" &&
				strings.HasPrefix(e.Avant, "time(") && e.Apres == "time(0)"
		},
	},

	// VOULU : décisions de l'outil, tolérées seulement quand elles sont prises.
	reprise("table_ecartee_par_le_generateur"),
	reprise("defaut_calcule_non_reconnu"),
	reprise("colonne_generee_non_recreee"),
	reprise("commentaire_de_type_immutable"),
	{
		code: "texte_unicode_recree_en_255", categorie: voulu,
		pourquoi: "un texte Unicode illimité reste en chaîne, avec l'avertissement texte_unicode_sans_equivalent : Doctrine le recrée en NVARCHAR(255), là où text le recréerait en VARCHAR(MAX) et perdrait l'Unicode sans erreur",
		couvre: func(e diff.Ecart, c contexte) bool {
			if e.Objet != diff.ObjetColonne || (e.Propriete != "type_brut" && e.Propriete != "longueur") {
				return false
			}
			cible := e.Schema + "." + e.Table + "." + e.Nom
			return slices.ContainsFunc(c.logique.Avertissements, func(a calque.Avertissement) bool {
				return a.Code == calque.CodeTexteUnicodeSansEquivalent && a.Cible == cible
			}) && ((e.Propriete == "type_brut" && e.Apres == "nvarchar(255)") ||
				(e.Propriete == "longueur" && e.Avant == "" && e.Apres == "255"))
		},
	},
	{
		code: "fuseau_precision_ramenee_a_6", categorie: voulu,
		pourquoi: "un horodatage avec fuseau est rendu en datetimetz, avec l'avertissement fuseau_precision_non_lue au-delà de six décimales : DBAL le recrée en DATETIMEOFFSET(6), là où le rendu sans fuseau écrirait un instant faux",
		couvre: func(e diff.Ecart, c contexte) bool {
			if e.Objet != diff.ObjetColonne {
				return false
			}
			cible := e.Schema + "." + e.Table + "." + e.Nom
			return slices.ContainsFunc(c.logique.Avertissements, func(a calque.Avertissement) bool {
				return a.Code == calque.CodeFuseauPrecisionNonLue && a.Cible == cible
			}) && ((e.Propriete == "type_brut" && strings.HasPrefix(e.Avant, "datetimeoffset(") && e.Apres == "datetimeoffset(6)") ||
				(e.Propriete == "precision_fractionnaire" && e.Apres == "6"))
		},
	},

	// À COMBLER : chacune part avec le lot qui la corrige.
	{
		code: "sequence_next_value_for_non_reconnue", categorie: aCombler, lot: "P7-5c — séquence",
		pourquoi: "une clé par DEFAULT NEXT VALUE FOR n'est pas reconnue : l'identifiant est laissé à l'application, et la séquence n'est pas recréée",
		couvre: func(e diff.Ecart, c contexte) bool {
			if e.Objet == diff.ObjetSequence {
				return e.Genre == diff.Suppression
			}
			col := colonneOrigine(c, e)
			return col != nil && col.Defaut != nil && col.Defaut.Genre == calque.DefautSequence &&
				e.Propriete == "defaut" && e.Apres == ""
		},
	},
	{
		code: "defaut_instant_courant_non_reconnu", categorie: aCombler, lot: "P7-5c — défauts",
		pourquoi: "getdate() et sysdatetimeoffset() ne sont pas reconnus comme l'instant courant : le défaut n'est pas reporté",
		couvre: func(e diff.Ecart, _ contexte) bool {
			return e.Objet == diff.ObjetColonne && e.Propriete == "defaut" && e.Apres == "" &&
				slices.Contains([]string{"expression (getdate())", "expression (sysdatetimeoffset())"}, e.Avant)
		},
	},
}

// nomClePrimaireSQLServer reconnaît le nom qu'SQL Server donne à une clé
// primaire que personne n'a nommée : PK__, le début du nom de table, puis un
// suffixe hexadécimal.
var nomClePrimaireSQLServer = regexp.MustCompile(`^PK__.+__[0-9A-F]{16}$`)

// nomGenereSQLServer reconnaît un nom que Doctrine forme d'un préfixe et d'un
// hachage : SQL Server garde la casse que PostgreSQL replie.
var nomGenereSQLServer = regexp.MustCompile(`^[A-Z]+_[0-9A-F]{16}$`)

// filtreNonNul reconnaît le prédicat que DBAL ajoute à un index unique.
var filtreNonNul = regexp.MustCompile(`^\((\[[^\]]+\] IS NOT NULL)( AND \[[^\]]+\] IS NOT NULL)*\)$`)
