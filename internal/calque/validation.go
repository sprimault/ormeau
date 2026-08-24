// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package calque

import (
	"fmt"
	"regexp"
	"sort"
)

// Anomalie est un défaut du calque lui-même, pas un jugement sur la base. À ne
// pas confondre avec Avertissement, qui porte une incertitude d'inférence :
// une anomalie signale un document que quelque chose a produit de travers.
type Anomalie struct {
	Code    string `json:"code"`
	Cible   string `json:"cible"`
	Message string `json:"message"`
}

// Codes d'anomalie. Stables entre versions, comme ceux des avertissements :
// ils servent de filtre en CI.
const (
	CodeVersionInconnue        = "version_inconnue"
	CodeSGBDInconnu            = "sgbd_inconnu"
	CodeChampRequisVide        = "champ_requis_vide"
	CodeEmpreinteMalformee     = "empreinte_malformee"
	CodeTypeHorsVocabulaire    = "type_hors_vocabulaire"
	CodeGenreDefautInconnu     = "genre_defaut_inconnu"
	CodeActionInconnue         = "action_inconnue"
	CodePositionInvalide       = "position_invalide"
	CodeTableSansColonne       = "table_sans_colonne"
	CodeTableDupliquee         = "table_dupliquee"
	CodeColonneDupliquee       = "colonne_dupliquee"
	CodeColonneIntrouvable     = "colonne_introuvable"
	CodeTableCibleIntrouvable  = "table_cible_introuvable"
	CodeAriteIncoherente       = "arite_incoherente"
	CodeTypeEnumereIntrouvable = "type_enumere_introuvable"
	CodeStatistiquesOrphelines = "statistiques_orphelines"
)

// Forme attendue d'une empreinte, telle qu'Ecrire la pose.
var motifEmpreinte = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

// Valider contrôle la cohérence du calque et rend toutes les anomalies plutôt
// que de s'arrêter à la première : un calque produit de travers l'est
// rarement en un seul endroit.
//
// Le contrôle va au-delà de ce que le JSON Schema sait exprimer. Une clé
// étrangère qui désigne une table absente du calque, une clé primaire dont une
// colonne n'existe pas, un type énuméré jamais déclaré : ces incohérences
// passent la validation du schéma et cassent la génération plus loin.
//
// Ne modifie pas le calque. L'ordre des anomalies suit celui du document.
func (p *Physique) Valider() []Anomalie {
	var a []Anomalie

	if p.VersionRI < 1 || p.VersionRI > VersionCourante {
		a = append(a, anomalie(CodeVersionInconnue, "version_ri",
			"version %d, cet outil connaît la version %d", p.VersionRI, VersionCourante))
	}
	a = append(a, validerSource(&p.Source)...)

	vues := make(map[string]bool, len(p.Tables))
	for i := range p.Tables {
		t := &p.Tables[i]
		cible := t.Schema + "." + t.Nom

		if vues[cible] {
			a = append(a, anomalie(CodeTableDupliquee, cible, "table déclarée deux fois"))
		}
		vues[cible] = true

		a = append(a, p.validerTable(t)...)
	}

	a = append(a, p.validerTypesEnumeres()...)
	a = append(a, p.validerStatistiques(vues)...)
	return a
}

// validerSource contrôle l'en-tête du calque. Le DSN n'y figure jamais : ce qui
// est attendu, c'est le SGBD, sa version, le catalogue et le schéma.
func validerSource(s *Source) []Anomalie {
	var a []Anomalie

	sgbdConnus := map[string]bool{
		"postgres": true, "mysql": true, "mariadb": true,
		"sqlserver": true, "sqlite": true, "oracle": true,
	}
	if !sgbdConnus[s.SGBD] {
		a = append(a, anomalie(CodeSGBDInconnu, "source.sgbd", "SGBD %q hors du vocabulaire", s.SGBD))
	}
	for champ, valeur := range map[string]string{
		"source.version":   s.Version,
		"source.catalogue": s.Catalogue,
		"source.schema":    s.Schema,
	} {
		if valeur == "" {
			a = append(a, anomalie(CodeChampRequisVide, champ, "champ requis vide"))
		}
	}
	// Le champ est optionnel avant Ecrire, qui le pose ; mal formé, il l'est.
	if s.Empreinte != "" && !motifEmpreinte.MatchString(s.Empreinte) {
		a = append(a, anomalie(CodeEmpreinteMalformee, "source.empreinte",
			"empreinte %q hors du format sha256:<64 hexadécimaux>", s.Empreinte))
	}

	trierParCible(a)
	return a
}

// validerTable contrôle une table et tout ce qu'elle porte. Les contraintes sont
// vérifiées contre les colonnes réellement déclarées : une clé primaire qui
// désigne une colonne absente produit une entité aux propriétés fantômes.
func (p *Physique) validerTable(t *Table) []Anomalie {
	var a []Anomalie
	cible := t.Schema + "." + t.Nom

	if t.Nom == "" || t.Schema == "" {
		a = append(a, anomalie(CodeChampRequisVide, cible, "nom ou schéma de table vide"))
	}
	if len(t.Colonnes) == 0 {
		a = append(a, anomalie(CodeTableSansColonne, cible, "aucune colonne"))
	}

	colonnes := make(map[string]bool, len(t.Colonnes))
	for i := range t.Colonnes {
		c := &t.Colonnes[i]
		cibleColonne := cible + "." + c.Nom

		if colonnes[c.Nom] {
			a = append(a, anomalie(CodeColonneDupliquee, cibleColonne, "colonne déclarée deux fois"))
		}
		colonnes[c.Nom] = true
		a = append(a, p.validerColonne(c, cibleColonne)...)
	}

	if t.ClePrimaire != nil {
		a = append(a, colonnesConnues(t.ClePrimaire.Colonnes, colonnes, cible, "clé primaire")...)
	}
	for _, u := range t.Unicites {
		a = append(a, colonnesConnues(u.Colonnes, colonnes, cible, "unicité "+u.Nom)...)
	}
	for _, idx := range t.Index {
		a = append(a, colonnesConnues(idx.Colonnes, colonnes, cible, "index "+idx.Nom)...)
	}
	for i := range t.ClesEtrangeres {
		a = append(a, p.validerCleEtrangere(&t.ClesEtrangeres[i], colonnes, cible)...)
	}
	return a
}

// validerColonne contrôle une colonne contre les vocabulaires fermés du format.
// Le type brut n'est pas jugé — il est verbatim par construction — mais le type
// normalisé qui l'accompagne doit appartenir au vocabulaire.
func (p *Physique) validerColonne(c *Colonne, cible string) []Anomalie {
	var a []Anomalie

	if c.Nom == "" {
		a = append(a, anomalie(CodeChampRequisVide, cible, "nom de colonne vide"))
	}
	if c.TypeBrut == "" {
		a = append(a, anomalie(CodeChampRequisVide, cible, "type_brut vide"))
	}
	if c.Position < 1 {
		a = append(a, anomalie(CodePositionInvalide, cible, "position %d, attendue >= 1", c.Position))
	}
	if !typesNormalises[c.TypeNormalise] {
		a = append(a, anomalie(CodeTypeHorsVocabulaire, cible,
			"type_normalise %q hors du vocabulaire fermé", c.TypeNormalise))
	}
	if c.Defaut != nil && !genresDefaut[c.Defaut.Genre] {
		a = append(a, anomalie(CodeGenreDefautInconnu, cible,
			"genre de défaut %q hors du vocabulaire", c.Defaut.Genre))
	}
	if c.TypeEnumere != "" && !p.typeEnumereDeclare(c.TypeEnumere) {
		a = append(a, anomalie(CodeTypeEnumereIntrouvable, cible,
			"type énuméré %q absent de types_enumeres", c.TypeEnumere))
	}
	return a
}

// validerCleEtrangere contrôle les deux côtés de la relation. L'arité est
// vérifiée avant la cible : une clé composite mal appariée est un défaut de
// l'extraction, pas une portée trop étroite.
func (p *Physique) validerCleEtrangere(fk *CleEtrangere, colonnes map[string]bool, cible string) []Anomalie {
	var a []Anomalie
	quoi := "clé étrangère " + fk.Nom

	a = append(a, colonnesConnues(fk.Colonnes, colonnes, cible, quoi)...)

	if len(fk.Colonnes) != len(fk.ColonnesCibles) {
		a = append(a, anomalie(CodeAriteIncoherente, cible,
			"%s porte %d colonne(s) pour %d colonne(s) cible(s)",
			quoi, len(fk.Colonnes), len(fk.ColonnesCibles)))
	}
	for _, action := range []Action{fk.ALaSuppression, fk.ALaMiseAJour} {
		if action != "" && !actions[action] {
			a = append(a, anomalie(CodeActionInconnue, cible,
				"%s : action référentielle %q hors du vocabulaire", quoi, action))
		}
	}

	// La table cible peut légitimement manquer quand l'extraction a été limitée
	// à un sous-ensemble de tables. On le signale quand même : c'est une
	// association que la génération ne pourra pas résoudre.
	tableCible := p.TableParNom(fk.SchemaCible, fk.TableCible)
	if tableCible == nil {
		a = append(a, anomalie(CodeTableCibleIntrouvable, cible,
			"%s désigne %s.%s, absente du calque", quoi, fk.SchemaCible, fk.TableCible))
		return a
	}

	cibleConnues := make(map[string]bool, len(tableCible.Colonnes))
	for i := range tableCible.Colonnes {
		cibleConnues[tableCible.Colonnes[i].Nom] = true
	}
	a = append(a, colonnesConnues(fk.ColonnesCibles, cibleConnues,
		fk.SchemaCible+"."+fk.TableCible, quoi+" (côté cible)")...)
	return a
}

// validerTypesEnumeres refuse les déclarations vides : un type énuméré sans
// valeur ne produit aucun cas PHP exploitable.
func (p *Physique) validerTypesEnumeres() []Anomalie {
	var a []Anomalie
	for i := range p.TypesEnumeres {
		te := &p.TypesEnumeres[i]
		if te.Nom == "" || te.Schema == "" {
			a = append(a, anomalie(CodeChampRequisVide, te.Schema+"."+te.Nom,
				"nom ou schéma de type énuméré vide"))
		}
		if len(te.Valeurs) == 0 {
			a = append(a, anomalie(CodeChampRequisVide, te.Schema+"."+te.Nom,
				"type énuméré sans valeur"))
		}
	}
	return a
}

// validerStatistiques trie les clés avant de les parcourir : itérer une map en
// ordre d'insertion rendrait la liste d'anomalies non déterministe.
func (p *Physique) validerStatistiques(tables map[string]bool) []Anomalie {
	if len(p.Statistiques) == 0 {
		return nil
	}

	cles := make([]string, 0, len(p.Statistiques))
	for cle := range p.Statistiques {
		cles = append(cles, cle)
	}
	sort.Strings(cles)

	var a []Anomalie
	for _, cle := range cles {
		if !tables[cle] {
			a = append(a, anomalie(CodeStatistiquesOrphelines, cle,
				"statistiques sur une table absente du calque"))
		}
	}
	return a
}

// VerifierEmpreinte recalcule l'empreinte et la compare à celle annoncée. Un
// écart signale un calque retouché après écriture.
//
// Trie le calque, comme CalculerEmpreinte dont elle dépend.
func (p *Physique) VerifierEmpreinte() error {
	if p.Source.Empreinte == "" {
		return fmt.Errorf("aucune empreinte a verifier")
	}

	annoncee := p.Source.Empreinte
	calculee, err := p.CalculerEmpreinte()
	if err != nil {
		return err
	}
	if calculee != annoncee {
		return fmt.Errorf("empreinte %s, calculee %s", annoncee, calculee)
	}
	return nil
}

// typeEnumereDeclare dit si le type est présent dans types_enumeres. Le parcours
// est linéaire : une base en compte quelques dizaines, jamais des milliers.
func (p *Physique) typeEnumereDeclare(nom string) bool {
	for i := range p.TypesEnumeres {
		if p.TypesEnumeres[i].Nom == nom {
			return true
		}
	}
	return false
}

// colonnesConnues vérifie qu'une contrainte ne désigne que des colonnes
// existantes. C'est le contrôle qu'un JSON Schema ne peut pas faire, et celui
// qui évite une entité aux propriétés fantômes.
func colonnesConnues(reference []string, connues map[string]bool, cible, quoi string) []Anomalie {
	var a []Anomalie
	for _, nom := range reference {
		if !connues[nom] {
			a = append(a, anomalie(CodeColonneIntrouvable, cible,
				"%s référence la colonne %q, absente de la table", quoi, nom))
		}
	}
	return a
}

// anomalie assemble le triplet code, cible, message.
func anomalie(code, cible, format string, args ...any) Anomalie {
	return Anomalie{Code: code, Cible: cible, Message: fmt.Sprintf(format, args...)}
}

// trierParCible impose un ordre stable là où les anomalies viennent du parcours
// d'une map. Ailleurs l'ordre du document suffit, et vaut mieux : il place les
// anomalies dans l'ordre où on les lira.
func trierParCible(a []Anomalie) {
	sort.SliceStable(a, func(i, j int) bool {
		if a[i].Cible != a[j].Cible {
			return a[i].Cible < a[j].Cible
		}
		return a[i].Code < a[j].Code
	})
}

// Les trois vocabulaires fermés du format, sous forme interrogeable. Une valeur
// ajoutée ici incrémente version_ri.
var typesNormalises = map[TypeNorm]bool{
	TypeEntier: true, TypeDecimal: true, TypeFlottant: true, TypeTexte: true,
	TypeBinaire: true, TypeBooleen: true, TypeDate: true, TypeHeure: true,
	TypeHorodatage: true, TypeIntervalle: true, TypeUUID: true, TypeJSON: true,
	TypeXML: true, TypeGeometrie: true, TypeEnumereNorm: true, TypeReseau: true,
	TypeInconnu: true,
}

// Genres de défaut reconnus.
var genresDefaut = map[GenreDefaut]bool{
	DefautLitteral: true, DefautExpression: true, DefautSequence: true,
}

// Actions référentielles reconnues.
var actions = map[Action]bool{
	ActionAucune: true, ActionCascade: true, ActionSetNull: true,
	ActionSetDefault: true, ActionRestrict: true,
}
