// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import "strings"

// Singularisation du français et de l'anglais par le même jeu de règles.
//
// La langue d'un nom de table ne se devine pas : rien dans un catalogue ne dit
// si commandes est français ou si orders est anglais, et une base reprise
// mélange souvent les deux dans le même schéma. Deux jeux de règles séparés
// demanderaient donc de trancher une question sans réponse.
//
// Les règles des deux langues se recoupent assez pour cohabiter — le -s final
// domine partout —, et là où elles divergent elles ne se contredisent pas :
// -aux est français, -ies est anglais, aucun mot n'est les deux.

// Formes dont le singulier ne se déduit d'aucune règle. Le pluriel irrégulier
// est rare, mais il tombe sur des noms de tables très courants.
var irreguliers = map[string]string{
	// Anglais
	"people":   "person",
	"children": "child",
	"men":      "man",
	"women":    "woman",
	"teeth":    "tooth",
	"feet":     "foot",
	"mice":     "mouse",
	"geese":    "goose",
	"indices":  "index",
	"matrices": "matrix",
	"vertices": "vertex",
	"criteria": "criterion",
	"media":    "medium",

	// Anglais en -ie, que la règle du -ies vers -y abîmerait : cookies
	// donnerait cooky.
	"cookies":  "cookie",
	"movies":   "movie",
	"calories": "calorie",
	"zombies":  "zombie",

	// Français en -ie dont la forme anglaise en -y n'est pas un mot : sorty et
	// bougy n'existent pas, la règle du -ies n'a donc rien à y arbitrer.
	//
	// Ceux dont les deux langues donnent un mot — categories, maladies,
	// economies, technologies — n'y figurent pas volontairement : y trancher
	// pour le français serait un choix silencieux, et c'est l'avertissement
	// d'ambiguïté qui doit s'en charger.
	"sorties":   "sortie",
	"bougies":   "bougie",
	"garanties": "garantie",

	// Pluriels anglais en -es dont la terminaison est aussi celle d'un mot
	// français au pluriel. Traités un par un plutôt que par une règle, qui
	// abîmerait adresses, fiches et analyses.
	"buses":     "bus",
	"gases":     "gas",
	"statuses":  "status",
	"aliases":   "alias",
	"campuses":  "campus",
	"batches":   "batch",
	"matches":   "match",
	"searches":  "search",
	"branches":  "branch",
	"processes": "process",

	// Français
	"yeux":      "oeil",
	"aieux":     "aieul",
	"messieurs": "monsieur",
}

// Mots identiques au singulier et au pluriel, ou singuliers finissant par s.
//
// C'est la liste qui évite les dégâts : sans elle, la table pays devient Pay,
// status devient Statu, et souris devient Souri. Ce sont des noms de tables
// fréquents, et le nom d'entité produit est ce que l'utilisateur lira dans son
// code pendant des années.
var invariants = map[string]bool{
	// Français
	"pays": true, "souris": true, "temps": true, "corps": true, "cours": true,
	"prix": true, "choix": true, "croix": true, "voix": true, "poids": true,
	"tas": true, "bras": true, "cas": true, "gaz": true, "nez": true,
	"riz": true, "puits": true, "colis": true, "avis": true, "pardis": true,
	"repas": true, "univers": true, "concours": true, "discours": true,
	"acces": true, "proces": true, "succes": true, "progres": true, "exces": true,
	"os": true, "fois": true, "mois": true, "bois": true, "mars": true,

	// Anglais
	"status": true, "bus": true, "alias": true, "campus": true, "bonus": true,
	"virus": true, "series": true, "species": true, "news": true, "analysis": true,
	"basis": true, "census": true, "focus": true, "iris": true, "lens": true,
	"data": true, "info": true, "metadata": true,
}

// resultatSingulier porte ce que la singularisation a fait, et ce qu'elle en
// pense.
//
// Tranche dit qu'une règle s'est appliquée. Faux ne signifie pas échec : un nom
// déjà au singulier — client, adresse — le rend aussi, et c'est le cas le plus
// courant.
//
// Ambigu dit que la règle appliquée dépendait d'une langue que rien ne permet
// d'établir. Le nom est singularisé quand même — laisser un pluriel serait pire
// —, mais l'appelant doit le signaler.
type resultatSingulier struct {
	nom     string
	tranche bool
	ambigu  bool
}

// singulariser rend le singulier d'un nom de table.
//
// Les mots composés ne sont pas traités pièce par pièce : seul le premier est
// singularisé, parce que la règle française met tous les éléments au pluriel —
// boites aux lettres en a trois pour un seul objet — et qu'on ne sait pas
// lesquels portent le sens.
//
// Souligné, tiret et espace séparent tous les trois. L'espace n'est pas une
// curiosité : SQL Server accepte les identifiants entre crochets, et une base
// reprise en contient parfois. Ne couper que sur le souligné ferait examiner la
// fin de la chaîne — Commandes Clients deviendrait Commandes Client, singularisé
// du mauvais côté.
//
// Le séparateur d'origine est conservé, la casse s'appliquant ensuite.
func singulariser(nom string) resultatSingulier {
	if nom == "" {
		return resultatSingulier{}
	}

	coupe := strings.IndexAny(nom, "_- ")
	if coupe < 0 {
		return singulariserMot(nom)
	}

	r := singulariserMot(nom[:coupe])
	r.nom += nom[coupe:]
	return r
}

// singulariserMot applique les règles à un mot isolé.
//
// L'ordre compte, et il va du plus spécifique au plus général : eaux avant aux,
// aux avant s, sinon bureaux donnerait Bureal.
func singulariserMot(mot string) resultatSingulier {
	bas := strings.ToLower(mot)

	if bas == "" || invariants[bas] {
		return resultatSingulier{nom: mot}
	}
	if singulier, connu := irreguliers[bas]; connu {
		return resultatSingulier{nom: respecterLaCasse(mot, singulier), tranche: true}
	}

	// Un mot en -ss n'est jamais un pluriel : class, process, adresse au
	// pluriel donne adresses et non adressess.
	if strings.HasSuffix(bas, "ss") {
		return resultatSingulier{nom: mot}
	}

	switch {
	// Français : -eaux vers -eau, avant la règle en -aux qui donnerait -al.
	case strings.HasSuffix(bas, "eaux"):
		return resultatSingulier{nom: couper(mot, 1), tranche: true}

	// Français : chevaux vers cheval, journaux vers journal.
	case strings.HasSuffix(bas, "aux"):
		return resultatSingulier{nom: remplacerFin(mot, 3, "al"), tranche: true}

	// Anglais : categories vers category.
	//
	// Ambigu quand rien ne désigne la langue : categories est category au
	// pluriel en anglais, et catégorie au pluriel en français une fois l'accent
	// perdu. Sorties, copies, maladies sont dans le même cas. On applique alors
	// la règle anglaise — la plus productive sur ce suffixe — et on le signale,
	// plutôt que de trancher en silence.
	//
	// Deux indices lèvent l'ambiguïté, et sont testés avant :
	//
	// L'accent. catégories est français, l'anglais n'en porte pas. La règle
	// générale s'applique alors et rend catégorie.
	//
	// La voyelle avant le -ies. L'anglais ne produit cette terminaison que
	// derrière une consonne — boy fait boys —, donc baies, voies et pluies sont
	// français, et la règle générale les traite aussi correctement.
	case strings.HasSuffix(bas, "ies") && len(bas) > 3 &&
		!estVoyelle(bas[len(bas)-4]) && !porteUnAccent(mot):
		return resultatSingulier{nom: remplacerFin(mot, 3, "y"), tranche: true, ambigu: true}

	// Anglais : boxes vers box, dishes vers dish.
	//
	// Seulement ces deux terminaisons. Les autres formes anglaises en -es —
	// -ses, -ches, -zes — sont aussi celles d'un mot français en -e mis au
	// pluriel, et le français l'emporte de loin en fréquence sur des noms de
	// tables : adresses, analyses, fiches, caches, taches contre buses et
	// batches. Les quelques mots anglais concernés sont traités comme des
	// irréguliers, ce qui coûte une ligne et ne casse rien.
	case strings.HasSuffix(bas, "xes"), strings.HasSuffix(bas, "shes"):
		return resultatSingulier{nom: couper(mot, 2), tranche: true}

	// Français : choux, bijoux, genoux. Rare, mais sans coût.
	case strings.HasSuffix(bas, "oux"):
		return resultatSingulier{nom: couper(mot, 1), tranche: true}

	// La règle générale des deux langues.
	case strings.HasSuffix(bas, "s"):
		return resultatSingulier{nom: couper(mot, 1), tranche: true}
	}

	// Ni pluriel reconnu, ni forme suspecte : déjà au singulier.
	return resultatSingulier{nom: mot}
}

// couper retire n caractères de la fin.
func couper(mot string, n int) string {
	if len(mot) <= n {
		return mot
	}
	return mot[:len(mot)-n]
}

// remplacerFin retire n caractères et colle une terminaison, en suivant la
// casse du mot d'origine : JOURNAUX doit donner JOURNAL et non JOURNal.
func remplacerFin(mot string, n int, fin string) string {
	base := couper(mot, n)
	if estMajuscule(mot) {
		return base + strings.ToUpper(fin)
	}
	return base + fin
}

// respecterLaCasse habille un singulier de la casse du pluriel d'origine.
func respecterLaCasse(origine, singulier string) string {
	switch {
	case estMajuscule(origine):
		return strings.ToUpper(singulier)
	case origine != "" && origine[0] >= 'A' && origine[0] <= 'Z':
		return strings.ToUpper(singulier[:1]) + singulier[1:]
	default:
		return singulier
	}
}

// estMajuscule dit si le mot ne porte aucune minuscule. Un mot sans lettre —
// que rien n'interdit dans un nom de colonne — n'en est pas un.
func estMajuscule(mot string) bool {
	lettre := false
	for _, r := range mot {
		if r >= 'a' && r <= 'z' {
			return false
		}
		if r >= 'A' && r <= 'Z' {
			lettre = true
		}
	}
	return lettre
}

// porteUnAccent dit si le mot sort de l'ASCII.
//
// C'est le seul indice de langue qu'un nom de table donne gratuitement, et il
// ne vaut que dans un sens : un accent exclut l'anglais, son absence ne prouve
// rien — beaucoup de bases françaises s'interdisent les accents dans les
// identifiants, par prudence envers les collations ou les vieux clients.
//
// Suffisant là où il sert : sur le seul suffixe où les deux langues se
// disputent le même mot, il tranche chaque fois qu'il est présent.
func porteUnAccent(mot string) bool {
	for _, r := range mot {
		if r > 127 {
			return true
		}
	}
	return false
}

// estVoyelle sert à départager categories de boys, sur la lettre qui précède
// le y. Les accents n'y figurent pas : la règle en -ies est anglaise.
func estVoyelle(r byte) bool {
	switch r {
	case 'a', 'e', 'i', 'o', 'u', 'y':
		return true
	}
	return false
}

// ressembleAUnPluriel dit si un nom qu'aucune règle n'a singularisé en avait
// pourtant l'air.
//
// C'est ce qui distingue l'avertissement du silence : client n'a pas été
// singularisé et personne n'attend qu'il le soit, alors qu'un nom en -s laissé
// intact parce qu'il figure dans les invariants mérite d'être signalé — la
// liste peut avoir tort sur le domaine de cette base.
func ressembleAUnPluriel(nom string) bool {
	bas := strings.ToLower(nom)
	if strings.HasSuffix(bas, "ss") {
		return false
	}
	return strings.HasSuffix(bas, "s") || strings.HasSuffix(bas, "x")
}
