// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"maps"
	"slices"
	"strings"

	"github.com/sprimault/ormeau/internal/calque"
)

// motsReservesDeClasse sont les noms que PHP refuse pour une classe, une
// énumération ou un trait, en minuscules : il les compare sans la casse.
//
// Relevés en passant chaque mot à php -l sous PHP 8.1 et 8.4, et pas recopiés
// de la documentation : enum, resource et numeric y figurent comme réservés,
// et PHP les accepte pourtant comme noms de classe. __property__ n'est refusé
// qu'à partir de la 8.4 ; il est retenu, le code produit devant se charger sur
// toute la plage.
var motsReservesDeClasse = ensemble([]string{
	"__class__", "__dir__", "__file__", "__function__", "__halt_compiler", "__line__", "__method__",
	"__namespace__", "__property__", "__trait__", "abstract", "and", "array", "as", "bool", "break",
	"callable", "case", "catch", "class", "clone", "const", "continue", "declare", "default", "die", "do",
	"echo", "else", "elseif", "empty", "enddeclare", "endfor", "endforeach", "endif", "endswitch",
	"endwhile", "eval", "exit", "extends", "false", "final", "finally", "float", "fn", "for", "foreach",
	"function", "global", "goto", "if", "implements", "include", "include_once", "instanceof", "insteadof",
	"int", "interface", "isset", "iterable", "list", "match", "mixed", "namespace", "never", "new", "null",
	"object", "or", "parent", "print", "private", "protected", "public", "readonly", "require",
	"require_once", "return", "self", "static", "string", "switch", "throw", "trait", "true", "try",
	"unset", "use", "var", "void", "while", "xor", "yield",
})

// motsReservesDeCas sont les seuls noms qu'un cas d'énumération ne peut pas
// porter ; motsReservesDEspace, les seuls qu'un segment d'espace de noms
// refuse. PHP 8 accepte les autres mots réservés à ces deux places.
var (
	motsReservesDeCas   = ensemble([]string{"class", "__halt_compiler"})
	motsReservesDEspace = ensemble([]string{"namespace", "__halt_compiler"})
)

// estIdentifiantPHP dit si un nom a la forme d'un identifiant PHP.
//
// La règle est celle du lexer, octet par octet : une lettre ASCII, un souligné
// ou un octet non ASCII en tête, les mêmes ou un chiffre ensuite. Un nom
// accentué passe, un point, une barre oblique ou un espace non : c'est ce qui
// empêche un nom de sortir de sa déclaration dans le code produit, ou du
// répertoire des entités dans le chemin du fichier.
func estIdentifiantPHP(nom string) bool {
	if nom == "" {
		return false
	}
	for i := 0; i < len(nom); i++ {
		c := nom[i]
		lettre := c == '_' || c >= 0x80 || (c|0x20 >= 'a' && c|0x20 <= 'z')
		if !lettre && (i == 0 || c < '0' || c > '9') {
			return false
		}
	}
	return true
}

// raisonNomRefuse dit pourquoi PHP refuserait ce nom à cette place, ou rend
// une chaîne vide. reserves est l'ensemble des mots interdits à cette place.
func raisonNomRefuse(nom string, reserves map[string]bool) string {
	if !estIdentifiantPHP(nom) {
		return cite(nom) + " n'est pas un identifiant PHP"
	}
	if reserves[strings.ToLower(nom)] {
		return nom + " est un mot réservé de PHP"
	}
	return ""
}

// raisonEspaceDeNoms dit pourquoi un espace de noms est refusé, ou rend une
// chaîne vide.
func raisonEspaceDeNoms(espace string) string {
	for _, segment := range strings.Split(espace, `\`) {
		if raison := raisonNomRefuse(segment, motsReservesDEspace); raison != "" {
			return cite(espace) + " n'est pas un espace de noms PHP : " + raison
		}
	}
	return ""
}

// cite rend un nom tel que l'utilisateur l'a écrit, entre guillemets
// français, pour qu'un nom vide ou fait d'espaces se voie.
func cite(nom string) string {
	return "« " + nom + " »"
}

// sansNomsInvalides rend les décisions privées de tout nom que PHP refuserait
// là où la génération l'écrit, et un avertissement par nom écarté.
//
// Un fichier de décisions est versionné dans le projet, reçu par une pull
// request, écrit par l'interface : ce qu'il nomme n'est pas plus sûr qu'une
// saisie. Un renommage recopié tel quel deviendrait un nom de classe, une
// ligne de code et un chemin de fichier — ../../public/index écrirait hors du
// répertoire des entités. Le nom est donc refusé ici, au plus près de celui
// qui l'a écrit, et l'inférence garde celui qu'elle aurait produit. Le
// générateur revérifie de son côté : un calque logique se modifie à la main.
//
// Les décisions reçues ne sont pas modifiées : le fichier prérempli les
// réécrit telles quelles, et l'utilisateur retrouve sa ligne à côté de
// l'avertissement.
func sansNomsInvalides(d *Decisions) (*Decisions, []calque.Avertissement) {
	copie := *d
	var avertissements []calque.Avertissement
	refuser := func(cible, message string) {
		avertissements = append(avertissements, calque.Avertissement{
			Code:       calque.CodeDecisionInvalide,
			Cible:      cible,
			Message:    message,
			Resolution: calque.ResolutionIgnoree,
			Confiance:  1,
		})
	}

	if raison := raisonEspaceDeNoms(d.EspaceDeNoms); d.EspaceDeNoms != "" && raison != "" {
		copie.EspaceDeNoms = ""
		refuser("espace_de_noms", "espace_de_noms refusé : "+raison+" ; "+espaceDeNoms(&copie)+" conservé")
	}

	for _, table := range slices.Sorted(maps.Keys(d.Renommages)) {
		raison := raisonNomRefuse(d.Renommages[table], motsReservesDeClasse)
		if raison == "" {
			continue
		}
		if len(copie.Renommages) == len(d.Renommages) {
			copie.Renommages = maps.Clone(d.Renommages)
		}
		delete(copie.Renommages, table)
		refuser(table, "renommage refusé : "+raison+" ; nom inféré conservé")
	}

	copie.RelationsForcees = slices.Clone(d.RelationsForcees)
	for i, r := range copie.RelationsForcees {
		if raison := raisonNomRefuse(r.Nom, nil); r.Nom != "" && raison != "" {
			copie.RelationsForcees[i].Nom = ""
			refuser(r.Source, "nom de relation refusé : "+raison+" ; nom par défaut")
		}
	}

	copie.Enumerations = nil
	for _, e := range d.Enumerations {
		if raison := raisonNomRefuse(e.Nom, motsReservesDeClasse); e.Nom != "" && raison != "" {
			refuser(e.Colonne, "énumération refusée : "+raison)
			continue
		}
		if len(e.Cas) > 0 {
			e.Cas = maps.Clone(e.Cas)
			for _, valeur := range slices.Sorted(maps.Keys(e.Cas)) {
				if raison := raisonNomRefuse(e.Cas[valeur], motsReservesDeCas); raison != "" {
					delete(e.Cas, valeur)
					refuser(e.Colonne, "cas de la valeur "+cite(valeur)+" refusé : "+raison+" ; nom par défaut")
				}
			}
		}
		copie.Enumerations = append(copie.Enumerations, e)
	}

	return &copie, avertissements
}
