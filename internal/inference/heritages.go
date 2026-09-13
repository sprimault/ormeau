// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"strings"

	"github.com/sprimault/ormeau/internal/calque"
)

// L'héritage se propose, il ne se déduit pas. Une clé primaire qui est aussi
// une clé étrangère autorise deux modèles — « un salarié est une personne » et
// « un salarié a une personne » —, et rien dans le schéma ne tranche. Doctrine,
// de son côté, n'a pas d'héritage sans colonne discriminante : il en ajoute une
// quand le mapping n'en déclare pas, et toute requête échoue sur une base qui
// ne la porte pas. Sans décision, la table est donc reliée à son parent par un
// un-vers-un sur la clé partagée, qui fonctionne sur la base telle qu'elle est.

// heritageRetenu est la place d'une table dans une hiérarchie décidée.
type heritageRetenu struct {
	// colonne est la colonne discriminante, portée par la table racine.
	colonne string
	// valeur est celle que la colonne prend pour cette classe.
	valeur string
	// racine distingue la table qui porte la colonne de ses descendantes.
	racine bool
}

// Fragments de nom qui suggèrent une colonne discriminante. Une suggestion,
// jamais une règle : elle alimente le message qui propose l'héritage.
var indicesDiscriminant = []string{"type", "nature", "genre", "categorie", "catégorie", "classe", "discr", "kind"}

// verifierHeritages retient les hiérarchies décidées et signale celles qui ne
// s'appliquent pas.
//
// Une décision s'applique entière ou pas du tout, racine par racine : une
// hiérarchie à moitié appliquée serait incompréhensible, et une erreur sur une
// racine ne bloque pas les autres. Elle se refuse quand la table racine n'est
// pas générée, qu'elle n'a pas la colonne, que valeurs oublie la racine ou ne
// cite aucun enfant, qu'une valeur est donnée deux fois, ou qu'une table citée
// ne descend pas de la racine par sa clé primaire — chaque table citée doit
// avoir pour parent direct la racine ou une autre table citée.
func verifierHeritages(d *Decisions, s *schemaLogique) []calque.Avertissement {
	s.heritages = map[string]heritageRetenu{}
	if len(d.Heritages) == 0 {
		return nil
	}

	var avertissements []calque.Avertissement
	refuser := func(racine, raison string) {
		avertissements = append(avertissements, calque.Avertissement{
			Code:       calque.CodeDecisionOrpheline,
			Cible:      racine,
			Message:    "héritage non appliqué : " + raison,
			Resolution: calque.ResolutionAucune,
			Confiance:  1,
		})
	}

	for _, racine := range clesTriees(d.Heritages) {
		h := d.Heritages[racine]
		if raison := refusHeritage(racine, h, s); raison != "" {
			refuser(racine, raison)
			continue
		}
		for table, valeur := range h.Valeurs {
			s.heritages[table] = heritageRetenu{colonne: h.ColonneDiscriminante, valeur: valeur, racine: table == racine}
		}
	}
	return avertissements
}

// refusHeritage dit pourquoi une hiérarchie décidée ne s'applique pas, ou rend
// la chaîne vide.
func refusHeritage(racine string, h HeritageDecide, s *schemaLogique) string {
	t := s.tableGeneree(racine)
	if t == nil {
		return "la table racine " + racine + " n'est pas une table générée du calque"
	}
	if !contientColonne(t, h.ColonneDiscriminante) {
		return racine + " n'a pas de colonne " + h.ColonneDiscriminante
	}
	if _, donnee := h.Valeurs[racine]; !donnee {
		return "valeurs ne donne pas de valeur à la racine " + racine
	}
	if len(h.Valeurs) < 2 {
		return "valeurs ne cite aucune table enfant"
	}

	prises := map[string]string{}
	for _, table := range clesTriees(h.Valeurs) {
		valeur := h.Valeurs[table]
		if autre, prise := prises[valeur]; prise {
			return "la valeur " + valeur + " est donnée à " + autre + " et à " + table
		}
		prises[valeur] = table

		if _, deja := s.heritages[table]; deja {
			return table + " appartient déjà à une autre hiérarchie"
		}
		if table == racine {
			continue
		}
		if s.tableGeneree(table) == nil {
			return table + " n'est pas une table générée du calque"
		}
		fk := s.parents[table]
		if fk == nil {
			return table + " n'a pas de clé primaire portée par une clé étrangère"
		}
		if _, citee := h.Valeurs[fk.SchemaCible+"."+fk.TableCible]; !citee || racineDe(table, s) != racine {
			return table + " ne descend pas de " + racine + " par des tables citées"
		}
	}
	return ""
}

// racineDe remonte les clés primaires étrangères jusqu'à la première table
// sans parent généré. Une chaîne qui boucle s'arrête sur la table de départ.
func racineDe(table string, s *schemaLogique) string {
	courante := table
	for range len(s.tables) {
		fk := s.parents[courante]
		if fk == nil {
			return courante
		}
		parent := fk.SchemaCible + "." + fk.TableCible
		if s.tableGeneree(parent) == nil {
			return courante
		}
		courante = parent
	}
	return table
}

// candidatesDiscriminantes rend les colonnes d'une table racine qui pourraient
// porter l'héritage, de la plus probable à la moins probable.
//
// D'abord une énumération reconnue — type énuméré natif ou CHECK IN —, puis un
// texte court dont le nom évoque une nature. Ni la clé primaire ni une clé
// étrangère n'en sont. La cardinalité n'entre pas en compte : sans
// échantillonnage, elle n'est pas connue, et le message ne prétend pas le
// contraire.
func candidatesDiscriminantes(t *calque.Table, s *schemaLogique) []string {
	exclues := map[string]bool{}
	if t.ClePrimaire != nil {
		for _, c := range t.ClePrimaire.Colonnes {
			exclues[c] = true
		}
	}
	for _, fk := range t.ClesEtrangeres {
		for _, c := range fk.Colonnes {
			exclues[c] = true
		}
	}

	var enumerees, nommees []string
	for i := range t.Colonnes {
		c := &t.Colonnes[i]
		if exclues[c.Nom] {
			continue
		}
		if _, reconnue := s.enumerations[t.Schema+"."+t.Nom+"."+c.Nom]; reconnue {
			enumerees = append(enumerees, c.Nom)
			continue
		}
		if c.TypeNormalise == calque.TypeTexte && c.Longueur != nil && *c.Longueur <= 32 && evoqueUneNature(c.Nom) {
			nommees = append(nommees, c.Nom)
		}
	}
	return append(enumerees, nommees...)
}

// evoqueUneNature dit si un nom de colonne contient un des fragments qui
// désignent d'habitude une colonne discriminante.
func evoqueUneNature(nom string) bool {
	bas := strings.ToLower(nom)
	for _, indice := range indicesDiscriminant {
		if strings.Contains(bas, indice) {
			return true
		}
	}
	return false
}

// messageHeritagePropose rend le message de l'avertissement heritage_deduit.
//
// C'est là que l'utilisateur lit la marche à suivre, pas dans les notes de
// version : le message dit ce qui a été fait, ce qu'il faut écrire pour obtenir
// un héritage, et où se trouve le coût — dans le fichier de décisions quand
// une colonne s'y prête, dans le schéma quand aucune ne s'y prête.
func messageHeritagePropose(cible, parent string, s *schemaLogique) string {
	debut := "clé primaire portant une clé étrangère vers " + parent + " : reliée par un-vers-un, sans héritage. "

	racine := racineDe(cible, s)
	if retenu, decidee := s.heritages[racine]; decidee && retenu.racine {
		return debut + "Pour l'inclure dans l'héritage déclaré, ajouter " + cible + " aux valeurs de heritages: " + racine
	}

	t := s.tableGeneree(racine)
	var candidates []string
	if t != nil {
		candidates = candidatesDiscriminantes(t, s)
	}

	message := debut + "Pour un héritage Doctrine, déclarer heritages: " + racine + " dans le fichier de décisions, avec une colonne_discriminante"
	switch len(candidates) {
	case 0:
		return message + " et une valeur par classe, racine comprise ; aucune colonne de " + racine + " ne s'y prête, elle serait à créer en base"
	case 1:
		return message + " (candidate : " + candidates[0] + ") et une valeur par classe, racine comprise"
	default:
		return message + " (candidates : " + strings.Join(candidates, ", ") + ") et une valeur par classe, racine comprise"
	}
}

// contientColonne dit si une table porte une colonne de ce nom.
func contientColonne(t *calque.Table, nom string) bool {
	for i := range t.Colonnes {
		if t.Colonnes[i].Nom == nom {
			return true
		}
	}
	return false
}
