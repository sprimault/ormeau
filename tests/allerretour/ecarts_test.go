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

// ciblesConnues sont celles dont les écarts ont été relevés. Une autre cible
// échoue au lieu de comparer à une liste qui ne la concerne pas.
var ciblesConnues = []string{"orm3-dbal4"}

// tolerances est la liste fermée. Une entrée s'ajoute avec l'écart qui la
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
		couvre: func(e diff.Ecart, _ contexte) bool {
			return e.Objet == diff.ObjetColonne && e.Propriete == "type_brut" &&
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
				(e.Propriete == "defaut" && e.Apres == "")
		},
	},
	{
		code: "nom_de_cle_etrangere_genere", categorie: impossible,
		pourquoi: "Doctrine nomme ses clés étrangères FK_ suivi d'un hachage",
		couvre: func(e diff.Ecart, _ contexte) bool {
			return e.Objet == diff.ObjetCleEtrangere && e.Propriete == "nom" && nomGenere.MatchString(e.Apres) &&
				strings.HasPrefix(e.Apres, "fk_")
		},
	},
	{
		code: "index_de_cle_etrangere_ajoute", categorie: impossible,
		pourquoi: "DBAL indexe toute clé étrangère qu'aucun index ne couvre déjà",
		couvre: func(e diff.Ecart, c contexte) bool {
			if e.Objet != diff.ObjetIndex || e.Genre != diff.Ajout || !strings.HasPrefix(e.Nom, "idx_") || !nomGenere.MatchString(e.Nom) {
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
		},
	},
	{
		code: "unicite_recreee_en_index", categorie: impossible,
		pourquoi: "Doctrine crée un index unique, pas une contrainte UNIQUE : l'index reste, la contrainte disparaît",
		couvre: func(e diff.Ecart, c contexte) bool {
			if e.Objet != diff.ObjetUnicite || e.Genre != diff.Suppression {
				return false
			}
			table := c.recree.TableParNom(e.Schema, e.Table)
			if table == nil {
				return false
			}
			for _, idx := range table.Index {
				if idx.Unique && idx.Nom == e.Nom {
					return true
				}
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
		code: "position_identite_derivee", categorie: impossible,
		pourquoi: "SchemaTool place une colonne de jointure sans propriété, clé d'une identité dérivée, après les champs",
		couvre: func(e diff.Ecart, c contexte) bool {
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

	// À COMBLER : chacune part avec le lot qui la corrige.
	{
		code: "cle_composite_reordonnee", categorie: aCombler, lot: "6",
		pourquoi: "l'association de clé est rendue après les propriétés : l'ordre de la clé primaire s'inverse",
		couvre: func(e diff.Ecart, _ contexte) bool {
			return e.Objet == diff.ObjetClePrimaire && e.Propriete == "colonnes"
		},
	},
	{
		code: "defaut_par_expression_perdu", categorie: aCombler, lot: "7",
		pourquoi: "un défaut calculé (now()) n'est pas reporté dans le calque logique",
		couvre: func(e diff.Ecart, _ contexte) bool {
			return e.Objet == diff.ObjetColonne && e.Propriete == "defaut" && e.Apres == "" &&
				strings.HasPrefix(e.Avant, string(calque.DefautExpression)+" ")
		},
	},
	{
		code: "predicat_d_index_perdu", categorie: aCombler, lot: "8",
		pourquoi: "le prédicat d'un index partiel n'est pas reporté dans le calque logique",
		couvre: func(e diff.Ecart, _ contexte) bool {
			return e.Objet == diff.ObjetIndex && e.Propriete == "predicat" && e.Apres == ""
		},
	},
	{
		code: "longueur_fixe_perdue", categorie: aCombler, lot: "9",
		pourquoi: "char(n) est rendu en chaîne de longueur variable",
		couvre: func(e diff.Ecart, _ contexte) bool {
			return e.Objet == diff.ObjetColonne && e.Propriete == "type_brut" &&
				strings.HasPrefix(e.Avant, "character(") && strings.HasPrefix(e.Apres, "character varying(")
		},
	},
	{
		code: "colonne_generee_perdue", categorie: aCombler, lot: "10",
		pourquoi: "l'expression d'une colonne générée n'est pas reportée : Doctrine recrée une colonne ordinaire",
		couvre: func(e diff.Ecart, _ contexte) bool {
			return e.Objet == diff.ObjetColonne && e.Propriete == "generee" && e.Apres == ""
		},
	},
}

// nomGenere reconnaît un nom que Doctrine forme d'un préfixe et d'un hachage
// hexadécimal, replié en minuscules par PostgreSQL.
var nomGenere = regexp.MustCompile(`^[a-z]+_[0-9a-f]{16}$`)

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

// confronter vérifie la liste fermée dans les deux sens pour la cible, et rend
// le relevé des écarts avec les entrées qui les couvrent.
func confronter(t *testing.T, cible string, ecarts []diff.Ecart, c contexte) string {
	t.Helper()

	if !slices.Contains(ciblesConnues, cible) {
		t.Fatalf("cible %s : aucun écart relevé pour elle (connues : %s)", cible, strings.Join(ciblesConnues, ", "))
	}

	var applicables []tolerance
	for _, tol := range tolerances {
		if len(tol.cibles) == 0 || slices.Contains(tol.cibles, cible) {
			applicables = append(applicables, tol)
		}
	}

	var releve strings.Builder
	fmt.Fprintf(&releve, "cible %s, ORM %s, DBAL %s : %d écarts\n", cible, c.php.ORM, c.php.DBAL, len(ecarts))
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
