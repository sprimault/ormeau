// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

//go:build allerretour

package allerretour

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
	"github.com/sprimault/ormeau/internal/diff"
)

// categorie dit pourquoi un écart est toléré.
type categorie string

// Trois raisons de tolérer un écart, et seulement trois.
const (
	// impossible : une limite de Doctrine ou de DBAL, que rien dans le calque
	// ni dans le générateur ne peut contourner.
	impossible categorie = "IMPOSSIBLE"
	// voulu : une décision de l'outil, justifiée par ce que le calque logique
	// ou le générateur dit de l'objet — jamais tolérée sans cette condition.
	voulu categorie = "VOULU"
	// aCombler : un manque connu, que le lot nommé retire. La phase est close
	// quand il n'en reste aucune.
	aCombler categorie = "À COMBLER"
)

// contexte est ce qu'une tolérance peut consulter pour décider.
type contexte struct {
	origine, recree *calque.Physique
	logique         *calque.Logique
	php             sortiePHP
}

// tolerance couvre une famille d'écarts attendus.
type tolerance struct {
	code      string    // stable, snake_case : c'est lui que cite un échec
	categorie categorie // pourquoi l'écart est accepté
	pourquoi  string    // une phrase, ligne de la matrice de fidélité ou raison du générateur
	lot       string    // pour aCombler : le lot de la phase 6 qui la retire
	cibles    []string  // vide : toutes les cibles connues
	couvre    func(e diff.Ecart, c contexte) bool
}

// ciblesConnues sont, par SGBD, celles dont les écarts ont été relevés. Une
// autre cible échoue au lieu de comparer à une liste qui ne la concerne pas.
var ciblesConnues = map[string][]string{
	"postgres":  {"orm2-dbal3", "orm3-dbal4"},
	"sqlserver": {"orm2-dbal3", "orm3-dbal4"},
}

// tolerancesParSGBD associe à chaque SGBD sa liste fermée. Une entrée ne
// passe pas d'une liste à l'autre sans que l'écart ait été constaté des deux
// côtés : sinon elle ne couvrirait rien chez l'un, et le test le refuserait.
var tolerancesParSGBD = map[string][]tolerance{
	"postgres":  tolerances,
	"sqlserver": tolerancesSQLServer,
}

// tolerances est la liste fermée de PostgreSQL. Une entrée s'ajoute avec l'écart qui la
// justifie, et part avec lui. Les raisons renvoient à la matrice de fidélité
// de l'audit d'avant la phase 6 ; la condition de chaque entrée est aussi
// étroite que l'écart qu'elle couvre.
var tolerances = []tolerance{
	// IMPOSSIBLE : ce que Doctrine et DBAL ne savent pas recréer.
	{
		code: "vue_non_recreee", categorie: impossible,
		pourquoi: "Doctrine ne mappe pas les vues",
		couvre: func(e diff.Ecart, _ contexte) bool {
			return e.Objet == diff.ObjetVue && e.Genre == diff.Suppression
		},
	},
	{
		code: "verification_non_recreee", categorie: impossible,
		pourquoi: "Doctrine n'écrit aucune contrainte CHECK, énumération comprise",
		couvre: func(e diff.Ecart, _ contexte) bool {
			return e.Objet == diff.ObjetVerification && e.Genre == diff.Suppression
		},
	},
	{
		code: "type_enumere_natif_en_varchar", categorie: impossible,
		pourquoi: "un type énuméré natif est recréé en VARCHAR(255) : Doctrine n'émet pas de CREATE TYPE",
		couvre: func(e diff.Ecart, c contexte) bool {
			if e.Objet == diff.ObjetTypeEnumere {
				return e.Genre == diff.Suppression
			}
			col := colonneOrigine(c, e)
			return col != nil && col.TypeEnumere != "" && e.Genre == diff.Modification &&
				slices.Contains([]string{"type_brut", "type_normalise", "type_enumere", "longueur", "collation"}, e.Propriete)
		},
	},
	{
		code: "defaut_null_explicite", categorie: impossible,
		pourquoi: "DBAL écrit DEFAULT NULL sur toute colonne nullable sans défaut ; même sens, et le comparateur de Doctrine les tient pour égales",
		couvre: func(e diff.Ecart, c contexte) bool {
			col := colonneOrigine(c, e)
			return col != nil && col.Nullable && col.Defaut == nil && e.Propriete == "defaut" && e.Avant == "" &&
				strings.HasPrefix(e.Apres, string(calque.DefautExpression)+" NULL::")
		},
	},
	{
		code: "precision_horodatage", categorie: impossible,
		pourquoi: "DBAL écrit TIMESTAMP(0) : la précision fractionnaire d'origine est perdue",
		couvre: func(e diff.Ecart, c contexte) bool {
			if e.Objet != diff.ObjetColonne {
				return false
			}
			if e.Propriete == "precision_fractionnaire" {
				col := colonneOrigine(c, e)
				return col != nil && strings.Contains(col.TypeBrut, "time") && e.Apres == "0"
			}
			return e.Propriete == "type_brut" &&
				strings.Contains(e.Avant, "time") && strings.Replace(e.Apres, "(0)", "", 1) == e.Avant
		},
	},
	{
		code: "serial_recree_en_identite", categorie: impossible, cibles: []string{"orm3-dbal4"},
		pourquoi: "une clé serial est rendue IDENTITY sous ORM 3 : le DEFAULT nextval devient une colonne d'identité",
		couvre: func(e diff.Ecart, c contexte) bool {
			col := colonneOrigine(c, e)
			if col == nil || col.Defaut == nil || col.Defaut.Genre != calque.DefautSequence {
				return false
			}
			return (e.Propriete == "auto_increment" && e.Avant == "false" && e.Apres == "true") ||
				(e.Propriete == "identite" && e.Avant == "" && e.Apres == string(calque.IdentiteParDefaut)) ||
				(e.Propriete == "defaut" && e.Apres == "")
		},
	},
	{
		code: "identite_toujours_recreee_par_defaut", categorie: impossible, cibles: []string{"orm3-dbal4"},
		pourquoi: "DBAL 4 crée toute identité en GENERATED BY DEFAULT : une identité ALWAYS accepte ensuite une valeur explicite",
		couvre: func(e diff.Ecart, c contexte) bool {
			col := colonneOrigine(c, e)
			return col != nil && col.Identite == calque.IdentiteToujours && e.Propriete == "identite" &&
				e.Apres == string(calque.IdentiteParDefaut)
		},
	},
	{
		code: "identite_recreee_en_serial", categorie: impossible, cibles: []string{"orm2-dbal3"},
		pourquoi: "DBAL 3 recrée une colonne d'identité en SERIAL : défaut nextval au lieu de l'identité",
		couvre: func(e diff.Ecart, c contexte) bool {
			col := colonneOrigine(c, e)
			if col == nil || !col.AutoIncrement {
				return false
			}
			return (e.Propriete == "auto_increment" && e.Avant == "true" && e.Apres == "false") ||
				(e.Propriete == "identite" && e.Avant == string(col.Identite) && e.Apres == "") ||
				(e.Propriete == "defaut" && e.Avant == "" && strings.HasPrefix(e.Apres, string(calque.DefautSequence)+" nextval("))
		},
	},
	{
		code: "serial_recree_en_sequence", categorie: impossible, cibles: []string{"orm2-dbal3"},
		pourquoi: "une clé serial est rendue SEQUENCE sous ORM 2 : la séquence est créée à part, sans DEFAULT nextval sur la colonne",
		couvre: func(e diff.Ecart, c contexte) bool {
			col := colonneOrigine(c, e)
			return col != nil && col.Defaut != nil && col.Defaut.Genre == calque.DefautSequence &&
				e.Propriete == "defaut" && e.Apres == ""
		},
	},
	{
		code: "bornes_de_sequence_par_defaut", categorie: impossible, cibles: []string{"orm2-dbal3"},
		pourquoi: "Doctrine crée la séquence d'un identifiant sans maximum : celui du type bigint remplace celui d'origine",
		couvre: func(e diff.Ecart, c contexte) bool {
			if e.Objet != diff.ObjetSequence || e.Propriete != "maximum" {
				return false
			}
			nom := e.Schema + "." + e.Nom
			return slices.ContainsFunc(c.logique.Entites, func(en calque.Entite) bool {
				return en.Identifiant != nil && en.Identifiant.Strategie == calque.IdentifiantSequence && en.Identifiant.Sequence == nom
			})
		},
	},
	{
		code: "ordre_des_colonnes_dbal3", categorie: impossible, cibles: []string{"orm2-dbal3"},
		pourquoi: "DBAL 3 rend les colonnes d'une table clé primaire d'abord, puis clés étrangères, puis le reste (Table::getColumns)",
		couvre: func(e diff.Ecart, c contexte) bool {
			if e.Objet != diff.ObjetColonne || e.Propriete != "position" {
				return false
			}
			return ordonneeParDBAL3(c.origine.TableParNom(e.Schema, e.Table), c.recree.TableParNom(e.Schema, e.Table))
		},
	},
	{
		code: "simple_precision_recreee_en_double", categorie: impossible, cibles: []string{"orm2-dbal3"},
		pourquoi: "DBAL 3 n'a pas de type smallfloat : une colonne real est recréée en double precision",
		couvre: func(e diff.Ecart, _ contexte) bool {
			return e.Objet == diff.ObjetColonne && e.Propriete == "type_brut" && e.Avant == "real" && e.Apres == "double precision"
		},
	},
	{
		code: "nom_de_cle_etrangere_genere", categorie: impossible,
		pourquoi: "Doctrine nomme ses clés étrangères FK_ suivi d'un hachage",
		couvre:   cleEtrangereNommeeParDoctrine(nomGenere, "fk_"),
	},
	{
		code: "index_de_cle_etrangere_ajoute", categorie: impossible,
		pourquoi: "DBAL indexe toute clé étrangère qu'aucun index ne couvre déjà",
		couvre:   indexDeCleEtrangereAjoute(nomGenere, "idx_"),
	},
	{
		code: "unicite_recreee_en_index", categorie: impossible,
		pourquoi: "Doctrine crée un index unique, pas une contrainte UNIQUE : la contrainte disparaît, un index apparaît à sa place",
		couvre: func(e diff.Ecart, c contexte) bool {
			switch {
			case e.Objet == diff.ObjetUnicite && e.Genre == diff.Suppression:
				table := c.recree.TableParNom(e.Schema, e.Table)
				return table != nil && slices.ContainsFunc(table.Index, func(idx calque.Index) bool {
					return idx.Unique && idx.Nom == e.Nom
				})
			case e.Objet == diff.ObjetIndex && e.Genre == diff.Ajout:
				// Depuis la version 2, l'origine ne reporte plus l'index qui
				// soutient une contrainte : il n'apparaît que dans la recréée.
				origine, recree := c.origine.TableParNom(e.Schema, e.Table), c.recree.TableParNom(e.Schema, e.Table)
				if origine == nil || recree == nil {
					return false
				}
				i := slices.IndexFunc(recree.Index, func(idx calque.Index) bool { return idx.Nom == e.Nom })
				return i >= 0 && recree.Index[i].Unique && slices.ContainsFunc(origine.Unicites, func(u calque.Contrainte) bool {
					return u.Nom == e.Nom && slices.Equal(u.Colonnes, recree.Index[i].Colonnes)
				})
			}
			return false
		},
	},
	{
		code: "classes_d_operateurs", categorie: impossible,
		pourquoi: "Doctrine ne porte ni méthode ni classe d'opérateurs d'index",
		couvre: func(e diff.Ecart, _ contexte) bool {
			return e.Objet == diff.ObjetIndex && e.Propriete == "operateurs" && e.Apres == ""
		},
	},
	{
		code: "ordres_d_index", categorie: impossible,
		pourquoi: "Doctrine ne porte pas le sens de tri d'une colonne d'index : un index descendant est recréé ascendant",
		couvre: func(e diff.Ecart, _ contexte) bool {
			return e.Objet == diff.ObjetIndex && e.Propriete == "ordres" && e.Apres == ""
		},
	},
	{
		code: "position_identite_derivee", categorie: impossible, cibles: []string{"orm3-dbal4"},
		pourquoi: "SchemaTool place une colonne de jointure sans propriété, clé d'une identité dérivée, après les champs",
		couvre:   positionIdentiteDerivee,
	},

	{
		code: "defaut_calcule_sous_la_forme_de_dbal", categorie: impossible,
		pourquoi: "DBAL écrit un défaut calculé sous sa forme : now() devient CURRENT_TIMESTAMP, LOCALTIME CURRENT_TIME ; même sens",
		couvre: func(e diff.Ecart, c contexte) bool {
			if e.Objet != diff.ObjetColonne || e.Propriete != "defaut" || !strings.HasPrefix(e.Avant, string(calque.DefautExpression)+" ") {
				return false
			}
			p := proprieteLogique(c, e.Schema, e.Table, e.Nom)
			return p != nil && p.DefautExpression != "" && e.Apres == string(calque.DefautExpression)+" "+formesDoctrine[p.DefautExpression]
		},
	},

	// VOULU : décisions de l'outil, tolérées seulement quand elles sont prises.
	{
		code: "table_ecartee_par_le_generateur", categorie: voulu,
		pourquoi: "une entité écartée par le générateur, avec sa raison, ne crée pas de table",
		couvre: func(e diff.Ecart, c contexte) bool {
			if e.Objet != diff.ObjetTable || e.Genre != diff.Suppression {
				return false
			}
			for _, entite := range c.logique.Entites {
				if entite.Table.Schema != e.Schema || entite.Table.Nom != e.Table {
					continue
				}
				for _, ecartee := range c.php.Ecartees {
					if ecartee.Entite == entite.Nom {
						return true
					}
				}
			}
			return false
		},
	},
	{
		code: "commentaire_de_type_immutable", categorie: voulu, cibles: []string{"orm2-dbal3"},
		pourquoi: "l'outil rend les dates en types immutables, que DBAL 3 signale par un commentaire (DC2Type:…)",
		couvre: func(e diff.Ecart, c contexte) bool {
			if e.Objet != diff.ObjetColonne || e.Propriete != "commentaire" {
				return false
			}
			p := proprieteLogique(c, e.Schema, e.Table, e.Nom)
			return p != nil && strings.HasSuffix(p.TypeDoctrine, "_immutable") && e.Apres == e.Avant+"(DC2Type:"+p.TypeDoctrine+")"
		},
	},
	{
		code: "type_non_reconnu_en_chaine", categorie: voulu,
		pourquoi: "un type sans équivalent Doctrine est rendu en chaîne, avec l'avertissement type_non_reconnu",
		couvre: func(e diff.Ecart, c contexte) bool {
			if e.Objet != diff.ObjetColonne || !slices.Contains([]string{"type_brut", "longueur"}, e.Propriete) {
				return false
			}
			cible := e.Schema + "." + e.Table + "." + e.Nom
			return slices.ContainsFunc(c.logique.Avertissements, func(a calque.Avertissement) bool {
				return a.Code == calque.CodeTypeNonReconnu && a.Cible == cible
			})
		},
	},
	{
		code: "uuid_genere_non_reproduit", categorie: voulu,
		pourquoi: "un UUID tiré par la base est reconnu mais non écrit, DBAL n'ayant aucune expression pour lui : l'application fournit la valeur, et le rapport de génération le dit (DefautNonReproduit)",
		couvre: func(e diff.Ecart, c contexte) bool {
			if e.Objet != diff.ObjetColonne || e.Propriete != "defaut" || e.Apres != "" {
				return false
			}
			p := proprieteLogique(c, e.Schema, e.Table, e.Nom)
			return p != nil && p.DefautExpression == calque.DefautUUIDGenere
		},
	},
	{
		code: "colonne_generee_non_recreee", categorie: voulu,
		pourquoi: "une colonne générée est relue après chaque écriture mais recréée ordinaire : seul columnDefinition la recréerait, et l'outil n'écrit pas le DDL d'un dialecte dans une entité",
		couvre: func(e diff.Ecart, c contexte) bool {
			if e.Objet != diff.ObjetColonne || e.Propriete != "generee" || e.Apres != "" {
				return false
			}
			p := proprieteLogique(c, e.Schema, e.Table, e.Nom)
			return p != nil && p.Generee != nil
		},
	},

	{
		code: "collation_hors_pg_catalog_non_recreee", categorie: voulu,
		pourquoi: "une collation hors de pg_catalog n'est pas reportée, avec l'avertissement collation_non_reportee : Doctrine l'écrirait en un seul identifiant, et son nom seul dépendrait du search_path",
		couvre: func(e diff.Ecart, c contexte) bool {
			if e.Objet != diff.ObjetColonne || (e.Propriete != "collation" && e.Propriete != "collation_schema") {
				return false
			}
			cible := e.Schema + "." + e.Table + "." + e.Nom
			return slices.ContainsFunc(c.logique.Avertissements, func(a calque.Avertissement) bool {
				return a.Code == calque.CodeCollationNonReportee && a.Cible == cible
			})
		},
	},

	// À COMBLER : chacune part avec le lot qui la corrige. Aucune ne reste.
}

// nomGenere reconnaît un nom que Doctrine forme d'un préfixe et d'un hachage
// hexadécimal, replié en minuscules par PostgreSQL.
var nomGenere = regexp.MustCompile(`^[a-z]+_[0-9a-f]{16}$`)

// reprise rend une entrée de la liste PostgreSQL, pour une autre liste où le
// même écart a la même cause : une seule condition, une seule raison.
func reprise(code string) tolerance {
	for _, tol := range tolerances {
		if tol.code == code {
			return tol
		}
	}
	panic("tolérance inconnue : " + code)
}

// cleEtrangereNommeeParDoctrine couvre le nom qu'une clé étrangère reçoit de
// Doctrine, préfixe et hachage, sous la casse que le SGBD lui laisse.
func cleEtrangereNommeeParDoctrine(motif *regexp.Regexp, prefixe string) func(diff.Ecart, contexte) bool {
	return func(e diff.Ecart, _ contexte) bool {
		return e.Objet == diff.ObjetCleEtrangere && e.Propriete == "nom" && motif.MatchString(e.Apres) &&
			strings.HasPrefix(e.Apres, prefixe)
	}
}

// indexDeCleEtrangereAjoute couvre l'index que DBAL ajoute sous une clé
// étrangère qu'aucun index ne couvrait.
func indexDeCleEtrangereAjoute(motif *regexp.Regexp, prefixe string) func(diff.Ecart, contexte) bool {
	return func(e diff.Ecart, c contexte) bool {
		if e.Objet != diff.ObjetIndex || e.Genre != diff.Ajout || !strings.HasPrefix(e.Nom, prefixe) || !motif.MatchString(e.Nom) {
			return false
		}
		table := c.recree.TableParNom(e.Schema, e.Table)
		if table == nil {
			return false
		}
		for _, idx := range table.Index {
			if idx.Nom != e.Nom {
				continue
			}
			for _, fk := range table.ClesEtrangeres {
				if slices.Equal(fk.Colonnes, idx.Colonnes) {
					return true
				}
			}
		}
		return false
	}
}

// positionIdentiteDerivee couvre la colonne de jointure d'une identité
// dérivée, que SchemaTool place après les champs.
func positionIdentiteDerivee(e diff.Ecart, c contexte) bool {
	if e.Objet != diff.ObjetColonne || e.Propriete != "position" {
		return false
	}
	origine := c.origine.TableParNom(e.Schema, e.Table)
	recree := c.recree.TableParNom(e.Schema, e.Table)
	if origine == nil || recree == nil || origine.ClePrimaire == nil || len(recree.Colonnes) == 0 {
		return false
	}
	derniere := recree.Colonnes[len(recree.Colonnes)-1].Nom
	for _, fk := range origine.ClesEtrangeres {
		if slices.Contains(fk.Colonnes, derniere) && slices.Contains(origine.ClePrimaire.Colonnes, derniere) {
			return true
		}
	}
	return false
}

// ordonneeParDBAL3 dit si la table recréée suit exactement l'ordre de DBAL 3 :
// colonnes de la clé primaire, puis colonnes de clé étrangère, puis les autres
// dans l'ordre d'origine. Dans les deux premiers groupes, DBAL garde l'ordre
// d'ajout à la table (Table::filterColumns), où ORM place les champs avant les
// colonnes de jointure : seul l'ensemble de chaque groupe est exigé. Un
// déplacement qui ne s'explique pas ainsi n'est pas couvert.
func ordonneeParDBAL3(origine, recree *calque.Table) bool {
	if origine == nil || recree == nil {
		return false
	}
	noms := make([]string, len(recree.Colonnes))
	for i, col := range recree.Colonnes {
		noms[i] = col.Nom
	}

	var cle []string
	if recree.ClePrimaire != nil {
		cle = recree.ClePrimaire.Colonnes
	}
	if len(noms) < len(cle) {
		return false
	}
	for _, col := range noms[:len(cle)] {
		if !slices.Contains(cle, col) {
			return false
		}
	}

	etrangeres := map[string]bool{}
	for _, fk := range recree.ClesEtrangeres {
		for _, col := range fk.Colonnes {
			if !slices.Contains(cle, col) {
				etrangeres[col] = true
			}
		}
	}
	suite := noms[len(cle):]
	if len(suite) < len(etrangeres) {
		return false
	}
	for _, col := range suite[:len(etrangeres)] {
		if !etrangeres[col] {
			return false
		}
	}

	var reste []string
	for _, col := range origine.Colonnes {
		if !slices.Contains(cle, col.Nom) && !etrangeres[col.Nom] && recree.ColonneParNom(col.Nom) != nil {
			reste = append(reste, col.Nom)
		}
	}
	return slices.Equal(suite[len(etrangeres):], reste)
}

// proprieteLogique rend la propriété que le calque logique donne à une
// colonne, propriétés des traits de l'entité comprises ; nil si aucune
// propriété ne la porte.
func proprieteLogique(c contexte, schema, table, colonne string) *calque.Propriete {
	for _, entite := range c.logique.Entites {
		if entite.Table.Schema != schema || entite.Table.Nom != table {
			continue
		}
		for i := range entite.Proprietes {
			if entite.Proprietes[i].Colonne == colonne {
				return &entite.Proprietes[i]
			}
		}
		for _, trait := range c.logique.Traits {
			if !slices.Contains(entite.Traits, trait.Nom) {
				continue
			}
			for i := range trait.Proprietes {
				if trait.Proprietes[i].Colonne == colonne {
					return &trait.Proprietes[i]
				}
			}
		}
	}
	return nil
}

// formesDoctrine sont les expressions qu'écrit la plateforme PostgreSQL de
// DBAL pour chaque défaut calculé, chaîne ou objet DefaultExpression.
var formesDoctrine = map[calque.ExpressionDefaut]string{
	calque.DefautHorodatageCourant: "CURRENT_TIMESTAMP",
	calque.DefautDateCourante:      "CURRENT_DATE",
	calque.DefautHeureCourante:     "CURRENT_TIME",
}

// colonneOrigine rend la colonne d'origine que désigne un écart de colonne.
func colonneOrigine(c contexte, e diff.Ecart) *calque.Colonne {
	if e.Objet != diff.ObjetColonne {
		return nil
	}
	table := c.origine.TableParNom(e.Schema, e.Table)
	if table == nil {
		return nil
	}
	return table.ColonneParNom(e.Nom)
}

// confronter vérifie la liste fermée dans les deux sens pour le SGBD et la
// cible, et rend le relevé des écarts avec les entrées qui les couvrent.
func confronter(t *testing.T, sgbd, cible string, ecarts []diff.Ecart, c contexte) string {
	t.Helper()

	if !slices.Contains(ciblesConnues[sgbd], cible) {
		t.Fatalf("%s, cible %s : aucun écart relevé pour elle (connues : %s)",
			sgbd, cible, strings.Join(ciblesConnues[sgbd], ", "))
	}

	var applicables []tolerance
	for _, tol := range tolerancesParSGBD[sgbd] {
		if len(tol.cibles) == 0 || slices.Contains(tol.cibles, cible) {
			applicables = append(applicables, tol)
		}
	}

	var releve strings.Builder
	fmt.Fprintf(&releve, "%s, cible %s, ORM %s, DBAL %s : %d écarts\n", sgbd, cible, c.php.ORM, c.php.DBAL, len(ecarts))
	utilisees := make([]bool, len(applicables))
	for _, e := range ecarts {
		var codes []string
		for j, tol := range applicables {
			if tol.couvre(e, c) {
				utilisees[j] = true
				codes = append(codes, tol.code)
			}
		}
		ligne := fmt.Sprintf("%s %s [%q -> %q]", e.Genre, e.Chemin(), e.Avant, e.Apres)
		if len(codes) == 0 {
			t.Errorf("écart qu'aucune entrée ne couvre : %s", ligne)
			fmt.Fprintf(&releve, "  NON COUVERT  %s\n", ligne)
			continue
		}
		fmt.Fprintf(&releve, "  %-40s %s\n", strings.Join(codes, ", "), ligne)
	}

	for j, tol := range applicables {
		if !utilisees[j] {
			t.Errorf("entrée %s (%s) : elle ne couvre plus aucun écart, la retirer", tol.code, tol.categorie)
		}
	}
	return releve.String()
}
