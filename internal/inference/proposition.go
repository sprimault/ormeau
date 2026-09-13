// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"cmp"
	"maps"
	"slices"
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
	Cible     string  `json:"cible"`
	Nom       string  `json:"nom"`
	Raison    string  `json:"raison"`
	Confiance float64 `json:"confiance"`
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

// EcrireDecisions rend le fichier de décisions d'une base : documentation,
// décisions, propositions de l'outil.
//
// Sans décision, c'est le fichier prérempli du premier passage, entièrement en
// commentaire. Une ligne qui s'appliquerait sans qu'on l'ait lue reproduirait
// ce qu'on refuse à l'inférence : décider à la place de l'utilisateur. Avec des
// décisions — celles qu'on vient d'arbitrer —, elles s'écrivent hors
// commentaire, et elles seules.
//
// Chaque section dit à quoi elle sert et montre un exemple avant les valeurs
// propres à cette base. Le fichier est la documentation de l'arbitrage autant
// que son support : personne ne va lire docs/ avant de corriger un nom de
// classe.
//
// La détection du contenu manuel repose sur deux règles de forme, que les tests
// gardent : un commentaire généré est toujours séparé d'un bloc de décisions
// par une ligne vide, et aucune clé ne s'écrit sans valeur. Le contenu est
// déterministe : les mêmes entrées rendent les mêmes octets, et un fichier relu
// puis réécrit redonne les siens.
func EcrireDecisions(p *calque.Physique, d *Decisions, base string) []byte {
	var b strings.Builder

	entete(&b, p, d)
	sectionEspaceDeNoms(&b, d)
	sectionPrefixes(&b, p, d)
	sectionRenommages(&b, Proposer(p, d), d)
	sectionTablesIgnorees(&b, d)
	sectionColonnesIgnorees(&b, d)
	sectionTypesForces(&b, d)
	sectionRelationsForcees(&b, d)
	sectionHeritages(&b, p, d)
	sectionEnumerations(&b, d)

	corps := b.String()
	return []byte(lignesEmpreinte(base, corps) + corps)
}

// entete annonce ce que le fichier fait, et surtout ce qu'il ne fait pas.
func entete(b *strings.Builder, p *calque.Physique, d *Decisions) {
	b.WriteString("#\n# Décisions d'inférence — ")
	b.WriteString(p.Source.Catalogue)
	b.WriteString("\n#\n")
	if d.vide() {
		b.WriteString("# TOUT EST EN COMMENTAIRE : ce fichier ne change rien tant que vous n'avez\n")
		b.WriteString("# rien décommenté. C'est voulu. L'outil propose, il ne décide pas.\n")
	} else {
		b.WriteString("# Seules les lignes hors commentaire décident. Le reste documente chaque\n")
		b.WriteString("# section et porte les propositions de l'outil, à décommenter au besoin.\n")
	}
	b.WriteString("#\n")
	b.WriteString("# Une décision gagne toujours contre une heuristique, sans discussion. Le\n")
	b.WriteString("# fichier est rejoué à chaque passage : le corriger une fois suffit, et la\n")
	b.WriteString("# régénération de six mois plus tard n'écrasera pas votre arbitrage.\n")
	b.WriteString("#\n")
	b.WriteString("# Une décision qui ne correspond à rien produit un avertissement — c'est le\n")
	b.WriteString("# signal que la base a bougé sous le fichier.\n")
	b.WriteString("#\n")
	b.WriteString("# L'interface réécrit ce fichier en entier quand elle l'enregistre. Un\n")
	b.WriteString("# commentaire écrit dans le bloc d'une décision — sur sa ligne, ou juste\n")
	b.WriteString("# au-dessus sans ligne vide — lui fait demander confirmation d'abord ; ceux\n")
	b.WriteString("# qu'une ligne vide sépare des décisions sont réécrits sans elle.\n\n")
}

// sectionEspaceDeNoms écrit l'espace de noms PHP des entités.
func sectionEspaceDeNoms(b *strings.Builder, d *Decisions) {
	b.WriteString("# ── Espace de noms des entités générées ──────────────────────────────\n")
	b.WriteString("#\n")
	b.WriteString("# Défaut : App\\Entity, la disposition d'un projet Symfony standard.\n")
	b.WriteString("#\n")
	b.WriteString("#   espace_de_noms: Gescom\\Domaine\\Entity\n")

	if d.EspaceDeNoms != "" {
		b.WriteString("\nespace_de_noms: ")
		b.WriteString(scalaire(d.EspaceDeNoms, false))
		b.WriteString("\n\n")
		return
	}
	b.WriteString("#\n#espace_de_noms: ")
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
	b.WriteString("#   prefixes_a_retirer:\n#     - T_\n#     - tbl_\n")

	if len(d.PrefixesARetirer) > 0 {
		b.WriteString("\nprefixes_a_retirer:\n")
		for _, prefixe := range d.PrefixesARetirer {
			b.WriteString("  - ")
			b.WriteString(scalaire(prefixe, false))
			b.WriteString("\n")
		}
		b.WriteString("\n")
		return
	}

	b.WriteString("#\n")
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

// sectionRenommages écrit les noms de classes : ceux décidés, et ceux que
// l'outil suggère pour les autres tables.
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
	b.WriteString("#   renommages:\n#     dbo.T_CLIENTS: Client\n#     commandes: Commande\n")

	actif := len(d.Renommages) > 0
	if actif {
		b.WriteString("\nrenommages:\n")
		ecrireCorrespondances(b, d.Renommages)
	}

	if len(propositions) == 0 {
		if !actif {
			b.WriteString("#\n#renommages: {}\n")
		}
		b.WriteString("\n")
		return
	}

	// Paragraphe à part : les propositions changent avec le calque, et ne
	// doivent jamais se mêler aux décisions que l'empreinte couvre.
	b.WriteString("\n# Propositions de l'outil pour cette base. La colonne de droite dit ce qui\n")
	b.WriteString("# a été appliqué pour y arriver — lisez-la avant de décommenter.\n")
	if !actif {
		b.WriteString("#renommages:\n")
	}

	lignes := make([]string, len(propositions))
	largeur := 0
	for i, prop := range propositions {
		lignes[i] = "#  " + scalaire(prop.Cible, false) + ": " + scalaire(prop.Nom, false)
		largeur = max(largeur, len(lignes[i]))
	}
	for i, prop := range propositions {
		b.WriteString(lignes[i])
		b.WriteString(strings.Repeat(" ", largeur+2-len(lignes[i])))
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
	b.WriteString("#   tables_ignorees:\n#     - dbo.T_AUDIT_TECHNIQUE\n#     - public.migrations\n")

	if len(d.TablesIgnorees) > 0 {
		b.WriteString("\ntables_ignorees:\n")
		for _, table := range slices.Sorted(slices.Values(d.TablesIgnorees)) {
			b.WriteString("  - ")
			b.WriteString(scalaire(table, false))
			b.WriteString("\n")
		}
		b.WriteString("\n")
		return
	}
	b.WriteString("#\n#tables_ignorees: []\n\n")
}

// sectionColonnesIgnorees écrit les colonnes à retirer des entités.
func sectionColonnesIgnorees(b *strings.Builder, d *Decisions) {
	b.WriteString("# ── Colonnes à ne pas mapper ────────────────────────────────────────\n")
	b.WriteString("#\n")
	b.WriteString("# Une colonne qu'on ne veut pas voir dans l'entité : un blob d'import, un\n")
	b.WriteString("# champ libre laissé par une application morte, une colonne technique.\n")
	b.WriteString("#\n")
	b.WriteString("# L'arbitrage est ici et non à l'extraction : le calque physique garde\n")
	b.WriteString("# toutes les colonnes, sans quoi le mode diff les signalerait comme\n")
	b.WriteString("# disparues à chaque comparaison avec la base. En les écartant ici, on se\n")
	b.WriteString("# ravise six mois plus tard sans rouvrir la connexion.\n")
	b.WriteString("#\n")
	b.WriteString("# Une colonne de clé primaire reste mappée quoi qu'il arrive : Doctrine\n")
	b.WriteString("# refuse une entité sans identifiant.\n")
	b.WriteString("#\n")
	b.WriteString("#   colonnes_ignorees:\n")
	b.WriteString("#     public.clients: [photo, blob_import]\n")
	b.WriteString("#     dbo.T_COMMANDES: [champ_libre_12]\n")

	if len(d.ColonnesIgnorees) > 0 {
		b.WriteString("\ncolonnes_ignorees:\n")
		for _, table := range clesTriees(d.ColonnesIgnorees) {
			colonnes := slices.Sorted(slices.Values(d.ColonnesIgnorees[table]))
			for i, colonne := range colonnes {
				colonnes[i] = scalaire(colonne, true)
			}
			b.WriteString("  ")
			b.WriteString(scalaire(table, false))
			b.WriteString(": [")
			b.WriteString(strings.Join(colonnes, ", "))
			b.WriteString("]\n")
		}
		b.WriteString("\n")
		return
	}
	b.WriteString("#\n#colonnes_ignorees: {}\n\n")
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
	b.WriteString("#   types_forces:\n#     dbo.T_CLIENTS.CLI_ACTIF: boolean\n#     public.client.donnees: json\n")

	if len(d.TypesForces) > 0 {
		b.WriteString("\ntypes_forces:\n")
		ecrireCorrespondances(b, d.TypesForces)
		b.WriteString("\n")
		return
	}
	b.WriteString("#\n#types_forces: {}\n\n")
}

// sectionRelationsForcees écrit les associations que le schéma ne déclare pas.
func sectionRelationsForcees(b *strings.Builder, d *Decisions) {
	b.WriteString("# ── Relations non déclarées ─────────────────────────────────────────\n")
	b.WriteString("#\n")
	b.WriteString("# La clé étrangère que personne n'a jamais créée, cas majoritaire sur du\n")
	b.WriteString("# legacy. L'outil peut la soupçonner en échantillonnant les valeurs, mais\n")
	b.WriteString("# seul quelqu'un qui connaît le métier la confirme.\n")
	b.WriteString("#\n")
	b.WriteString("# source est la colonne qui porte la relation, cible celle qu'elle désigne.\n")
	b.WriteString("# genre : plusieurs_vers_un (plusieurs commandes pour un client) ou\n")
	b.WriteString("# un_vers_un ; laissé vide, l'unicité de la colonne tranche. Le côté\n")
	b.WriteString("# collection, sur l'entité cible, se déduit. Une relation écrite ici gagne\n")
	b.WriteString("# sur une clé étrangère déclarée sur la même colonne.\n")
	b.WriteString("#\n")
	b.WriteString("#   relations_forcees:\n")
	b.WriteString("#     - source: public.commande.client_id\n")
	b.WriteString("#       cible: public.client.id\n")
	b.WriteString("#       genre: plusieurs_vers_un\n")
	b.WriteString("#       nom: client\n")

	if len(d.RelationsForcees) > 0 {
		relations := slices.Clone(d.RelationsForcees)
		slices.SortStableFunc(relations, func(x, y RelationForcee) int {
			return cmp.Or(
				strings.Compare(x.Source, y.Source),
				strings.Compare(x.Cible, y.Cible),
				strings.Compare(x.Nom, y.Nom),
				strings.Compare(x.Genre, y.Genre),
			)
		})

		b.WriteString("\nrelations_forcees:\n")
		for _, r := range relations {
			b.WriteString("  - source: " + scalaire(r.Source, false) + "\n")
			b.WriteString("    cible: " + scalaire(r.Cible, false) + "\n")
			b.WriteString("    genre: " + scalaire(r.Genre, false) + "\n")
			b.WriteString("    nom: " + scalaire(r.Nom, false) + "\n")
		}
		b.WriteString("\n")
		return
	}
	b.WriteString("#\n#relations_forcees: []\n\n")
}

// sectionHeritages écrit les hiérarchies déclarées, et propose celles que le
// schéma autorise.
//
// Une proposition est une racine dont au moins une table a pour clé primaire
// une clé étrangère vers elle, et qu'aucune décision ne déclare encore. Elle
// cite les colonnes candidates, ou dit qu'aucune ne s'y prête : le coût est
// alors dans le schéma, pas dans ce fichier. Les valeurs restent à remplir,
// rien ne permet de les deviner.
func sectionHeritages(b *strings.Builder, p *calque.Physique, d *Decisions) {
	b.WriteString("# ── Héritages ──────────────────────────────────────────────────────\n")
	b.WriteString("#\n")
	b.WriteString("# Une table dont la clé primaire est aussi une clé étrangère peut hériter de\n")
	b.WriteString("# la table qu'elle vise, ou lui être simplement reliée : le schéma autorise\n")
	b.WriteString("# les deux. Sans décision, elle lui est reliée par un-vers-un, ce qui\n")
	b.WriteString("# fonctionne sur la base telle qu'elle est.\n")
	b.WriteString("#\n")
	b.WriteString("# Un héritage Doctrine exige une colonne discriminante sur la table racine,\n")
	b.WriteString("# et une valeur par classe concrète, racine comprise. La clé est la table\n")
	b.WriteString("# racine ; une table enfant absente des valeurs reste reliée par un-vers-un.\n")
	b.WriteString("#\n")
	b.WriteString("#   heritages:\n")
	b.WriteString("#     public.personne:\n")
	b.WriteString("#       colonne_discriminante: nature\n")
	b.WriteString("#       valeurs:\n")
	b.WriteString("#         public.personne: P\n")
	b.WriteString("#         public.salarie: S\n")

	propositions := hierarchiesPossibles(p, d)
	if len(propositions) > 0 {
		b.WriteString("#\n# Propositions pour cette base :\n")
		for _, h := range propositions {
			b.WriteString("#\n#   " + scalaire(h.racine, false) + ":\n")
			switch len(h.candidates) {
			case 0:
				b.WriteString("#     colonne_discriminante: # aucune colonne ne s'y prête, à créer en base\n")
			default:
				b.WriteString("#     colonne_discriminante: " + scalaire(h.candidates[0], false))
				if len(h.candidates) > 1 {
					b.WriteString("  # autres candidates : " + strings.Join(h.candidates[1:], ", "))
				}
				b.WriteString("\n")
			}
			b.WriteString("#     valeurs:\n")
			for _, table := range h.tables {
				b.WriteString("#       " + scalaire(table, false) + ": # à remplir\n")
			}
		}
	}

	if len(d.Heritages) > 0 {
		b.WriteString("\nheritages:\n")
		for _, racine := range clesTriees(d.Heritages) {
			h := d.Heritages[racine]
			b.WriteString("  " + scalaire(racine, false) + ":\n")
			b.WriteString("    colonne_discriminante: " + scalaire(h.ColonneDiscriminante, false) + "\n")
			if len(h.Valeurs) == 0 {
				b.WriteString("    valeurs: {}\n")
				continue
			}
			b.WriteString("    valeurs:\n")
			for _, table := range clesTriees(h.Valeurs) {
				b.WriteString("      " + scalaire(table, false) + ": " + scalaire(h.Valeurs[table], false) + "\n")
			}
		}
		b.WriteString("\n")
		return
	}
	b.WriteString("#\n#heritages: {}\n\n")
}

// hierarchiePossible est une hiérarchie que le schéma autorise.
type hierarchiePossible struct {
	racine     string
	tables     []string
	candidates []string
}

// hierarchiesPossibles rend, triées par racine, les hiérarchies qu'aucune
// décision ne déclare encore : la racine, puis ses descendantes par clé
// primaire étrangère, triées.
func hierarchiesPossibles(p *calque.Physique, d *Decisions) []hierarchiePossible {
	s := analyser(p, d, nil)
	s.enumerations = enumerationsDuSchema(p, d)

	parRacine := map[string][]string{}
	for _, table := range clesTriees(s.parents) {
		if s.tableGeneree(table) == nil {
			continue
		}
		fk := s.parents[table]
		if s.tableGeneree(fk.SchemaCible+"."+fk.TableCible) == nil {
			continue
		}
		racine := racineDe(table, s)
		if _, decidee := d.Heritages[racine]; decidee || racine == table {
			continue
		}
		parRacine[racine] = append(parRacine[racine], table)
	}

	hierarchies := make([]hierarchiePossible, 0, len(parRacine))
	for _, racine := range clesTriees(parRacine) {
		tables := append([]string{racine}, parRacine[racine]...)
		slices.Sort(tables[1:])
		hierarchies = append(hierarchies, hierarchiePossible{
			racine:     racine,
			tables:     tables,
			candidates: candidatesDiscriminantes(s.tables[racine], s),
		})
	}
	return hierarchies
}

// sectionEnumerations écrit les énumérations imposées.
func sectionEnumerations(b *strings.Builder, d *Decisions) {
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

	if len(d.Enumerations) > 0 {
		enumerations := slices.Clone(d.Enumerations)
		slices.SortStableFunc(enumerations, func(x, y EnumerationForcee) int {
			return cmp.Or(strings.Compare(x.Colonne, y.Colonne), strings.Compare(x.Nom, y.Nom))
		})

		b.WriteString("\nenumerations:\n")
		for _, e := range enumerations {
			b.WriteString("  - colonne: " + scalaire(e.Colonne, false) + "\n")
			b.WriteString("    nom: " + scalaire(e.Nom, false) + "\n")
			if len(e.Cas) == 0 {
				b.WriteString("    cas: {}\n")
				continue
			}
			b.WriteString("    cas:\n")
			for _, valeur := range clesTriees(e.Cas) {
				b.WriteString("      " + scalaire(valeur, false) + ": " + scalaire(e.Cas[valeur], false) + "\n")
			}
		}
		return
	}
	b.WriteString("#\n#enumerations: []\n")
}

// ecrireCorrespondances écrit un bloc clé: valeur trié par clé, à deux espaces.
func ecrireCorrespondances(b *strings.Builder, correspondances map[string]string) {
	for _, cle := range clesTriees(correspondances) {
		b.WriteString("  ")
		b.WriteString(scalaire(cle, false))
		b.WriteString(": ")
		b.WriteString(scalaire(correspondances[cle], false))
		b.WriteString("\n")
	}
}

// clesTriees rend les clés d'une map dans un ordre stable.
//
// Le paquet en a besoin sur quatre types de valeurs différents, et l'ordre
// d'itération d'une map Go est volontairement aléatoire : sans ce passage, deux
// inférences du même calque rendraient des octets différents.
func clesTriees[V any](m map[string]V) []string {
	return slices.Sorted(maps.Keys(m))
}
