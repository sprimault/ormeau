// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"sort"
	"strconv"
	"strings"

	"github.com/sprimault/ormeau/internal/calque"
)

// Ce que l'inférence sait suggérer sans se permettre de l'appliquer.
//
// La singularisation et le retrait de préfixe vivent ici, et non dans le
// calque : ce sont des jugements sur le sens d'un nom, pas des traductions de
// sa forme. Appliqués en silence, ils produisent une classe que personne n'a
// demandée ; proposés en commentaire, ils font gagner le même temps sans rien
// décider à la place de qui que ce soit.

// Proposition est un renommage suggéré, avec de quoi le juger.
//
// Raison n'est pas décorative : c'est elle qui permet de décommenter une ligne
// en connaissance de cause, et de repérer celle qui pose problème dans une
// liste de quatre cents.
type Proposition struct {
	Cible     string
	Nom       string
	Raison    string
	Confiance float64
}

// Proposer rend les renommages que l'inférence suggère, triés par cible.
//
// Fonction pure comme le reste du paquet, et sans effet sur le calque : deux
// appels rendent la même liste, et ne rien en faire est une réponse valable.
//
// Une table déjà renommée par une décision n'y figure pas — la question est
// tranchée —, ni celle dont la proposition redirait le nom que l'inférence
// produit déjà.
func Proposer(p *calque.Physique, d *Decisions) []Proposition {
	if d == nil {
		d = &Decisions{}
	}

	prefixes, detecte := prefixesRetenus(p.Tables, d)
	if detecte != "" {
		// Le préfixe repéré sert aux propositions, sans être appliqué au
		// calque : c'est tout l'objet de cette fonction.
		prefixes = []string{detecte}
	}

	var propositions []Proposition
	for i := range p.Tables {
		t := &p.Tables[i]
		cible := t.Schema + "." + t.Nom

		if _, decide := d.Renommages[cible]; decide {
			continue
		}
		if _, decide := d.Renommages[t.Nom]; decide {
			continue
		}

		if prop, utile := proposerPour(t, prefixes, detecte); utile {
			propositions = append(propositions, prop)
		}
	}

	sort.SliceStable(propositions, func(i, j int) bool {
		return propositions[i].Cible < propositions[j].Cible
	})
	return propositions
}

// proposerPour construit la suggestion d'une table.
//
// Le second retour dit s'il y a quelque chose à proposer. Une table déjà au
// singulier et sans préfixe n'appelle aucune ligne : la remplir de propositions
// qui ne changent rien noierait celles qui comptent.
func proposerPour(t *calque.Table, prefixes []string, detecte string) (Proposition, bool) {
	actuel := pascalCase(t.Nom)

	nu := retirerPrefixe(t.Nom, prefixes)
	r := singulariser(nu)
	propose := pascalCase(r.nom)

	if propose == actuel {
		return Proposition{}, false
	}

	var raisons []string
	confiance := 1.0

	if nu != t.Nom && detecte != "" {
		raisons = append(raisons, "préfixe "+detecte+" retiré")
		confiance = 0.8
	}
	switch {
	case r.ambigu:
		// La règle appliquée dépend d'une langue que rien ne permet d'établir.
		// Les deux candidats sont donnés : c'est ce qui rend l'arbitrage
		// possible sans aller lire le code de l'inférence.
		raisons = append(raisons, "singularisé par la règle anglaise ; le français donnerait "+pascalCase(couper(nu, 1)))
		confiance = 0.5
	case r.tranche:
		raisons = append(raisons, "singularisé")
		if confiance > 0.7 {
			confiance = 0.7
		}
	}

	return Proposition{
		Cible:     t.Schema + "." + t.Nom,
		Nom:       propose,
		Raison:    strings.Join(raisons, ", "),
		Confiance: confiance,
	}, true
}

// EcrireDecisions rend un fichier de décisions prérempli.
//
// Entièrement en commentaire, sans exception. Un fichier prérempli dont une
// ligne s'applique sans qu'on l'ait lue reproduit ce qu'on refuse à
// l'inférence : décider à la place de l'utilisateur. Tel quel, il ne change
// rien — c'est ce qui permet de l'écrire au premier passage sans rien risquer.
//
// Chaque section dit à quoi elle sert, quand s'en servir, et montre un exemple
// avant les valeurs propres à cette base. Le fichier est la documentation de
// l'arbitrage autant que son support : personne ne va lire docs/ avant de
// corriger un nom de classe.
//
// Le contenu est déterministe : deux appels sur le même calque rendent les
// mêmes octets, propositions triées par cible.
func EcrireDecisions(p *calque.Physique, d *Decisions) []byte {
	var b strings.Builder

	entete(&b, p)
	sectionEspaceDeNoms(&b, d)
	sectionPrefixes(&b, p, d)
	sectionRenommages(&b, Proposer(p, d), d)
	sectionTablesIgnorees(&b, d)
	sectionTypesForces(&b, d)
	sectionRelationsForcees(&b)
	sectionEnumerations(&b)

	return []byte(b.String())
}

// entete annonce ce que le fichier fait, et surtout ce qu'il ne fait pas.
func entete(b *strings.Builder, p *calque.Physique) {
	b.WriteString("# Décisions d'inférence — ")
	b.WriteString(p.Source.Catalogue)
	b.WriteString("\n#\n")
	b.WriteString("# TOUT EST EN COMMENTAIRE : ce fichier ne change rien tant que vous n'avez\n")
	b.WriteString("# rien décommenté. C'est voulu. L'outil propose, il ne décide pas.\n")
	b.WriteString("#\n")
	b.WriteString("# Une décision gagne toujours contre une heuristique, sans discussion. Le\n")
	b.WriteString("# fichier est rejoué à chaque passage : le corriger une fois suffit, et la\n")
	b.WriteString("# régénération de six mois plus tard n'écrasera pas votre arbitrage.\n")
	b.WriteString("#\n")
	b.WriteString("# Une décision qui ne correspond à rien produit un avertissement — c'est le\n")
	b.WriteString("# signal que la base a bougé sous le fichier.\n\n")
}

// sectionEspaceDeNoms écrit l'espace de noms PHP des entités.
func sectionEspaceDeNoms(b *strings.Builder, d *Decisions) {
	b.WriteString("# ── Espace de noms des entités générées ──────────────────────────────\n")
	b.WriteString("#\n")
	b.WriteString("# Défaut : App\\Entity, la disposition d'un projet Symfony standard.\n")
	b.WriteString("#\n")
	b.WriteString("#   espace_de_noms: Gescom\\Domaine\\Entity\n")
	b.WriteString("#\n")
	b.WriteString("#espace_de_noms: ")
	b.WriteString(espaceDeNoms(d))
	b.WriteString("\n\n")
}

// sectionPrefixes écrit les préfixes à retirer, et signale celui que l'outil a
// repéré s'il y en a un.
func sectionPrefixes(b *strings.Builder, p *calque.Physique, d *Decisions) {
	_, detecte := prefixesRetenus(p.Tables, d)

	b.WriteString("# ── Préfixes de tables ───────────────────────────────────────────────\n")
	b.WriteString("#\n")
	b.WriteString("# Une convention de nommage sans valeur métier — T_, tbl_, dbo_ — qu'on ne\n")
	b.WriteString("# veut pas retrouver dans les noms de classes. Avec T_ ci-dessous, la table\n")
	b.WriteString("# T_CLIENTS donne Clients au lieu de TClients.\n")
	b.WriteString("#\n")
	b.WriteString("# Rien n'est retiré sans cette liste : un nom de table est un constat, et\n")
	b.WriteString("# l'amputer change ce que vous lirez dans votre code pendant des années.\n")
	b.WriteString("#\n")
	b.WriteString("#   prefixes_a_retirer:\n#     - T_\n#     - tbl_\n#\n")

	if detecte != "" {
		b.WriteString("# Repéré dans cette base : ")
		b.WriteString(detecte)
		b.WriteString(", commun aux ")
		b.WriteString(strconv.Itoa(len(p.Tables)))
		b.WriteString(" tables.\n")
		b.WriteString("#prefixes_a_retirer:\n#  - ")
		b.WriteString(detecte)
		b.WriteString("\n\n")
		return
	}

	b.WriteString("# Aucun préfixe commun repéré dans cette base.\n")
	b.WriteString("#prefixes_a_retirer: []\n\n")
}

// sectionRenommages écrit les noms de classes : ceux déjà décidés, et ceux que
// l'outil suggère.
func sectionRenommages(b *strings.Builder, propositions []Proposition, d *Decisions) {
	b.WriteString("# ── Noms de classes ─────────────────────────────────────────────────\n")
	b.WriteString("#\n")
	b.WriteString("# Sans entrée ici, le nom de table est repris tel quel, à la casse près :\n")
	b.WriteString("# commandes_clients donne CommandesClients. L'outil ne met pas au singulier\n")
	b.WriteString("# de lui-même — une base ne dit pas sa langue, et categories vaut Category\n")
	b.WriteString("# ou Categorie selon celle qu'on lui prête.\n")
	b.WriteString("#\n")
	b.WriteString("# La clé se qualifie par le schéma quand deux schémas portent la même table.\n")
	b.WriteString("#\n")
	b.WriteString("#   renommages:\n#     dbo.T_CLIENTS: Client\n#     commandes: Commande\n#\n")

	if len(d.Renommages) > 0 {
		b.WriteString("# Déjà décidé :\n")
		for _, cible := range clesTriees(d.Renommages) {
			b.WriteString("#  ")
			b.WriteString(cible)
			b.WriteString(": ")
			b.WriteString(d.Renommages[cible])
			b.WriteString("\n")
		}
		b.WriteString("#\n")
	}

	if len(propositions) == 0 {
		b.WriteString("#renommages: {}\n\n")
		return
	}

	b.WriteString("# Propositions de l'outil pour cette base. La colonne de droite dit ce qui\n")
	b.WriteString("# a été appliqué pour y arriver — lisez-la avant de décommenter.\n")
	b.WriteString("#renommages:\n")

	largeur := 0
	for _, prop := range propositions {
		if n := len(prop.Cible) + len(prop.Nom); n > largeur {
			largeur = n
		}
	}

	for _, prop := range propositions {
		ligne := "#  " + prop.Cible + ": " + prop.Nom
		b.WriteString(ligne)
		b.WriteString(strings.Repeat(" ", largeur+5-len(prop.Cible)-len(prop.Nom)))
		b.WriteString("# ")
		b.WriteString(prop.Raison)
		b.WriteString("\n")
	}
	b.WriteString("\n")
}

// sectionTablesIgnorees écrit les tables à ne pas générer.
func sectionTablesIgnorees(b *strings.Builder, d *Decisions) {
	b.WriteString("# ── Tables à ne pas générer ─────────────────────────────────────────\n")
	b.WriteString("#\n")
	b.WriteString("# Tables techniques, journaux, files d'attente, reliquats de migration : ce\n")
	b.WriteString("# qui existe en base sans avoir de place dans le modèle objet. Chacune\n")
	b.WriteString("# produit un avertissement, pour qu'aucune ne disparaisse en silence.\n")
	b.WriteString("#\n")
	b.WriteString("#   tables_ignorees:\n#     - dbo.T_AUDIT_TECHNIQUE\n#     - public.migrations\n#\n")

	if len(d.TablesIgnorees) > 0 {
		b.WriteString("# Déjà décidé :\n")
		ignorees := append([]string(nil), d.TablesIgnorees...)
		sort.Strings(ignorees)
		for _, t := range ignorees {
			b.WriteString("#  - ")
			b.WriteString(t)
			b.WriteString("\n")
		}
		b.WriteString("#\n")
	}
	b.WriteString("#tables_ignorees: []\n\n")
}

// sectionTypesForces écrit les types Doctrine imposés.
func sectionTypesForces(b *strings.Builder, d *Decisions) {
	b.WriteString("# ── Types imposés ───────────────────────────────────────────────────\n")
	b.WriteString("#\n")
	b.WriteString("# Le cas classique d'une base reprise : un char(1) valant O/N, que le\n")
	b.WriteString("# catalogue déclare en texte et qui est un booléen. L'outil ne peut pas le\n")
	b.WriteString("# deviner sans lire les données ; vous, vous le savez.\n")
	b.WriteString("#\n")
	b.WriteString("# La clé est schema.table.colonne. Le type PHP suit le type Doctrine : forcer\n")
	b.WriteString("# boolean donne un bool, et la longueur comme le défaut de la colonne sont\n")
	b.WriteString("# écartés — ils décrivaient le type d'avant.\n")
	b.WriteString("#\n")
	b.WriteString("#   types_forces:\n#     dbo.T_CLIENTS.CLI_ACTIF: boolean\n#     public.client.donnees: json\n#\n")

	if len(d.TypesForces) > 0 {
		b.WriteString("# Déjà décidé :\n")
		for _, cible := range clesTriees(d.TypesForces) {
			b.WriteString("#  ")
			b.WriteString(cible)
			b.WriteString(": ")
			b.WriteString(d.TypesForces[cible])
			b.WriteString("\n")
		}
		b.WriteString("#\n")
	}
	b.WriteString("#types_forces: {}\n\n")
}

// sectionRelationsForcees écrit les associations que le schéma ne déclare pas.
func sectionRelationsForcees(b *strings.Builder) {
	b.WriteString("# ── Relations non déclarées ─────────────────────────────────────────\n")
	b.WriteString("#\n")
	b.WriteString("# La clé étrangère que personne n'a jamais créée, cas majoritaire sur du\n")
	b.WriteString("# legacy. L'outil peut la soupçonner en échantillonnant les valeurs, mais\n")
	b.WriteString("# seul quelqu'un qui connaît le métier la confirme.\n")
	b.WriteString("#\n")
	b.WriteString("# genre : plusieurs_vers_un, un_vers_un, un_vers_plusieurs,\n")
	b.WriteString("#         plusieurs_vers_plusieurs\n")
	b.WriteString("#\n")
	b.WriteString("#   relations_forcees:\n")
	b.WriteString("#     - source: public.commande.client_id\n")
	b.WriteString("#       cible: public.client.id\n")
	b.WriteString("#       genre: plusieurs_vers_un\n")
	b.WriteString("#       nom: client\n")
	b.WriteString("#\n#relations_forcees: []\n\n")
}

// sectionEnumerations écrit les énumérations imposées.
func sectionEnumerations(b *strings.Builder) {
	b.WriteString("# ── Énumérations ────────────────────────────────────────────────────\n")
	b.WriteString("#\n")
	b.WriteString("# Une colonne à valeurs fermées devient un enum PHP. Les cas apparient la\n")
	b.WriteString("# valeur stockée à un nom lisible : un O/N en base n'a pas à produire un cas\n")
	b.WriteString("# nommé O.\n")
	b.WriteString("#\n")
	b.WriteString("#   enumerations:\n")
	b.WriteString("#     - colonne: dbo.T_CLIENTS.CLI_ETAT\n")
	b.WriteString("#       nom: EtatClient\n")
	b.WriteString("#       cas:\n")
	b.WriteString("#         A: Actif\n")
	b.WriteString("#         S: Suspendu\n")
	b.WriteString("#         R: Radie\n")
	b.WriteString("#\n#enumerations: []\n")
}

// clesTriees rend les clés d'une map dans un ordre stable : itérer une map
// rendrait le fichier différent à chaque écriture.
func clesTriees(m map[string]string) []string {
	cles := make([]string, 0, len(m))
	for cle := range m {
		cles = append(cles, cle)
	}
	sort.Strings(cles)
	return cles
}
