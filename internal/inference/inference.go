// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Package inference transforme un calque physique en calque logique.
//
// C'est le cœur de la valeur du projet : tout ce qui distingue Ormeau d'un
// générateur naïf est ici.
//
// # Heuristiques restant à écrire
//
// Nommage — retrait des préfixes, singularisation française et anglaise.
// Associations — client_id devient une propriété client typée Client, avec
// côté propriétaire et côté inverse cohérents. Table de jointure pure, qui
// devient une association et non une entité. Héritage proposé quand la clé
// primaire est aussi une clé étrangère, appliqué sur décision. Énumérations
// depuis un CHECK. Traits pour
// created_at et updated_at.
//
// Chacune produira soit un élément portant son origine, soit un avertissement,
// jamais une invention.
package inference

import (
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/sprimault/ormeau/internal/calque"
)

// Inferer est une fonction pure : pas de réseau, pas d'effet de bord, pas
// d'horloge, pas d'aléa. C'est ce qui permet de corriger une inférence hors
// ligne, de la rejouer, et de la tester sans base de données.
//
// Les avertissements sont une sortie de premier ordre, pas un journal : ce qui
// n'est pas résolu y figure, et n'est jamais inventé ailleurs.
func Inferer(p *calque.Physique, d *Decisions) (*calque.Logique, []calque.Avertissement) {
	if d == nil {
		d = &Decisions{}
	}
	d, avertissements := sansNomsInvalides(d, p.Source.Schema)

	logique := &calque.Logique{
		VersionRI:         calque.VersionCourante,
		EmpreintePhysique: p.Source.Empreinte,
		EspaceDeNoms:      espaceDeNoms(d),
		// Vide et non nulle : le schéma exige une liste, et un calque dont
		// toutes les tables sont écartées s'écrirait sinon "entites": null, que
		// le lecteur PHP prend pour un champ absent.
		Entites: []calque.Entite{},
	}

	ignorees := ensemble(d.TablesIgnorees)
	nomsPris := map[string]string{}

	// Le préfixe se cherche sur l'ensemble des tables, y compris celles qu'une
	// décision écarte : elles suivent la même convention de nommage, et les
	// retirer du calcul ferait dépendre le préfixe trouvé de ce qu'on génère.
	prefixes, detecte := prefixesRetenus(p.Tables, d)
	if detecte != "" {
		avertissements = append(avertissements, calque.Avertissement{
			Code:  calque.CodePrefixeDetecte,
			Cible: p.Source.Schema,
			Message: "préfixe " + detecte + " commun aux " + strconv.Itoa(len(p.Tables)) +
				" tables, conservé ; prefixes_a_retirer le retirerait des noms de classes",
			Resolution: calque.ResolutionAucune,
			Confiance:  0.8,
		})
	}

	// Ce qui se décide à l'échelle du schéma se calcule avant les entités :
	// une table de jointure n'en produit pas, un héritage se lit sur deux
	// tables à la fois, et un type énuméré natif est déclaré à part.
	schema := analyser(p, d, prefixes)
	schema.enumerations = enumerationsDuSchema(p, d)
	avertissements = append(avertissements, verifierRelationsForcees(p, d, schema)...)
	avertissements = append(avertissements, verifierHeritages(d, schema)...)

	for i := range p.Tables {
		t := &p.Tables[i]
		cible := t.Schema + "." + t.Nom

		if ignorees[cible] {
			avertissements = append(avertissements, calque.Avertissement{
				Code:       calque.CodeTableIgnoree,
				Cible:      cible,
				Message:    "table écartée par une décision",
				Resolution: calque.ResolutionIgnoree,
				Confiance:  1,
			})
			continue
		}
		if _, jointure := schema.jointures[cible]; jointure {
			continue
		}

		entite, avs := inferrerEntite(t, d, prefixes, schema, nomsPris)
		avertissements = append(avertissements, avs...)
		logique.Entites = append(logique.Entites, entite)
	}

	ajouterCotesInverses(logique)
	avertissements = append(avertissements, posterJointures(logique, schema)...)

	enumerations, avs := collecterEnumerations(schema, d)
	logique.Enumerations = enumerations
	avertissements = append(avertissements, avs...)

	// En dernier : le trait retire des propriétés aux entités, et tout ce qui
	// les lit — identifiant, associations, index — doit être passé avant.
	avertissements = append(avertissements, extraireTraits(logique)...)

	trierAvertissements(avertissements)
	logique.Avertissements = avertissements
	return logique, avertissements
}

// espaceDeNoms rend celui des décisions, ou le défaut de Symfony. Une entité
// sans espace de noms ne se génère pas.
func espaceDeNoms(d *Decisions) string {
	if d.EspaceDeNoms != "" {
		return d.EspaceDeNoms
	}
	return `App\Entity`
}

// colonnesEcartees rend les colonnes que l'utilisateur a retirées de l'entité,
// et les refus que cet arbitrage produit.
//
// Deux refus, tous deux signalés plutôt qu'appliqués en silence. Une colonne de
// clé primaire reste : Doctrine refuse une entité sans identifiant, et la
// retirer donnerait un modèle que rien ne peut charger. Une colonne inconnue
// signale que la base a bougé sous le fichier de décisions — c'est précisément
// ce qu'on veut apprendre en régénérant six mois plus tard.
//
// L'avertissement d'une colonne effectivement écartée s'écrit plus tard, par
// signalerColonnesEcartees : il nomme ce qui part avec elle, et ce n'est connu
// qu'une fois les index et les associations passés.
func colonnesEcartees(t *calque.Table, d *Decisions) (map[string]bool, []calque.Avertissement) {
	demandees := d.ColonnesIgnorees[t.Schema+"."+t.Nom]
	if len(demandees) == 0 {
		return nil, nil
	}

	presentes := make(map[string]bool, len(t.Colonnes))
	for i := range t.Colonnes {
		presentes[t.Colonnes[i].Nom] = true
	}

	cible := t.Schema + "." + t.Nom
	ecartees := make(map[string]bool, len(demandees))
	var avertissements []calque.Avertissement

	for _, colonne := range demandees {
		switch {
		case !presentes[colonne]:
			avertissements = append(avertissements, calque.Avertissement{
				Code:       calque.CodeDecisionOrpheline,
				Cible:      cible + "." + colonne,
				Message:    "colonnes_ignorees vise " + colonne + ", absente de la table",
				Resolution: calque.ResolutionAucune,
				Confiance:  1,
			})
		case dansClePrimaire(t, colonne):
			avertissements = append(avertissements, calque.Avertissement{
				Code:       calque.CodeClePrimaireGardee,
				Cible:      cible + "." + colonne,
				Message:    colonne + " appartient à la clé primaire et reste mappée : une entité sans identifiant est inutilisable",
				Resolution: calque.ResolutionAucune,
				Confiance:  1,
			})
		default:
			ecartees[colonne] = true
		}
	}
	return ecartees, avertissements
}

// ecarteeParDecision dit si colonnes_ignorees retire la colonne de son entité,
// selon la même règle que colonnesEcartees : une colonne de clé primaire reste.
func ecarteeParDecision(t *calque.Table, d *Decisions, colonne string) bool {
	return slices.Contains(d.ColonnesIgnorees[t.Schema+"."+t.Nom], colonne) && !dansClePrimaire(t, colonne)
}

// dansClePrimaire dit si la colonne appartient à la clé primaire déclarée.
func dansClePrimaire(t *calque.Table, colonne string) bool {
	return t.ClePrimaire != nil && slices.Contains(t.ClePrimaire.Colonnes, colonne)
}

// signalerColonnesEcartees écrit un avertissement par colonne écartée, qui
// nomme ce qui est parti avec elle.
//
// Le message dit aussi ce que Doctrine en fera : une colonne présente en base
// et absente de l'entité, migrations:diff la propose à la suppression. Ormeau
// ne peut pas l'empêcher, aucun filtre de schéma ne visant une colonne ; le
// taire laisserait appliquer un diff qui efface des données.
func signalerColonnesEcartees(t *calque.Table, ecartees map[string]bool, parties map[string][]string) []calque.Avertissement {
	var avertissements []calque.Avertissement
	for i := range t.Colonnes {
		colonne := t.Colonnes[i].Nom
		if !ecartees[colonne] {
			continue
		}
		message := colonne + " est retirée de l'entité ; le calque physique la garde"
		if len(parties[colonne]) > 0 {
			message += ". Avec elle : " + strings.Join(parties[colonne], ", ")
		}
		message += ". migrations:diff proposera de supprimer la colonne : à retirer du diff avant de l'appliquer"
		avertissements = append(avertissements, calque.Avertissement{
			Code:       calque.CodeColonneIgnoree,
			Cible:      t.Schema + "." + t.Nom + "." + colonne,
			Message:    message,
			Resolution: calque.ResolutionForceeParDecision,
			Confiance:  1,
		})
	}
	return avertissements
}

// inferrerEntite traduit une table en classe.
func inferrerEntite(t *calque.Table, d *Decisions, prefixes []string, schema *schemaLogique, nomsPris map[string]string) (calque.Entite, []calque.Avertissement) {
	cible := t.Schema + "." + t.Nom

	var avertissements []calque.Avertissement

	nom, origine := nomEntite(t, d, prefixes)
	if precedent, pris := nomsPris[nom]; pris {
		avertissements = append(avertissements, calque.Avertissement{
			Code:       calque.CodeCollision,
			Cible:      cible,
			Message:    "le nom de classe " + nom + " est déjà pris par " + precedent,
			Resolution: calque.ResolutionAucune,
			Confiance:  1,
		})
	}
	nomsPris[nom] = cible

	entite := calque.Entite{
		Nom:         nom,
		Table:       calque.ReferenceTable{Nom: t.Nom, Schema: t.Schema},
		Commentaire: t.Commentaire,
		Origine:     origine,
	}

	ecartees, avs := colonnesEcartees(t, d)
	avertissements = append(avertissements, avs...)

	entite.Proprietes = make([]calque.Propriete, 0, len(t.Colonnes))
	for i := range t.Colonnes {
		if ecartees[t.Colonnes[i].Nom] {
			continue
		}

		propriete, avs := inferrerPropriete(&t.Colonnes[i], cible, d)
		avertissements = append(avertissements, avs...)

		// Le type PHP d'une propriété énumérée est l'enum lui-même, pas la
		// chaîne qu'elle stocke. Le type Doctrine reste string : c'est ce que
		// la colonne contient, et Doctrine hydrate l'un vers l'autre.
		if e, enumeree := schema.enumerations[cible+"."+t.Colonnes[i].Nom]; enumeree {
			propriete.Enumeration = e.nom
			propriete.TypePHP = typeNullable(e.nom, t.Colonnes[i].Nullable)
		}
		entite.Proprietes = append(entite.Proprietes, propriete)
	}

	// L'index est construit après la boucle, et pas pendant : append réalloue
	// le tableau quand il grandit, et un pointeur pris au passage viserait la
	// zone abandonnée. Les marquages faits ensuite seraient perdus en silence.
	parColonne := make(map[string]*calque.Propriete, len(entite.Proprietes))
	for i := range entite.Proprietes {
		parColonne[entite.Proprietes[i].Colonne] = &entite.Proprietes[i]
	}

	// Une clé primaire étrangère ne donne un héritage que sur décision ; sans
	// elle, la clé devient un un-vers-un avec les autres associations, et
	// l'avertissement dit comment déclarer l'héritage. La valeur discriminante
	// va sur chaque classe d'une hiérarchie décidée, racine comprise.
	retenu, decide := schema.heritages[cible]
	if decide {
		entite.ValeurDiscriminante = retenu.valeur
	}
	if fk := schema.parents[cible]; fk != nil {
		parent, connu := schema.nomsParTable[fk.SchemaCible+"."+fk.TableCible]
		switch {
		case decide && !retenu.racine:
			entite.Heritage = &calque.Heritage{
				Strategie:            calque.HeritageJointe,
				Parent:               parent,
				ColonneDiscriminante: retenu.colonne,
				Origine:              calque.OrigineDecision,
			}
		case connu:
			avertissements = append(avertissements, calque.Avertissement{
				Code:       calque.CodeHeritageDeduit,
				Cible:      cible,
				Message:    messageHeritagePropose(cible, parent, schema),
				Resolution: calque.ResolutionParDefaut,
				Confiance:  0.7,
			})
		}
	}

	// Ce qui cite une colonne écartée part avec elle : Doctrine ne sait ni
	// indexer ni joindre une colonne que l'entité ne mappe pas, et une
	// association gardée écrirait la colonne qu'on a demandé d'ignorer.
	parties := map[string][]string{}
	associations, avs := inferrerAssociations(t, schema, parColonne, ecartees, parties)
	avertissements = append(avertissements, avs...)
	entite.Associations = associations

	identifiant, avs := inferrerIdentifiant(t, cible, parColonne, schema.sequences)
	avertissements = append(avertissements, avs...)
	entite.Identifiant = identifiant

	entite.Index = reporterIndex(t, ecartees, parties)
	marquerUniques(t, parColonne)
	avertissements = append(avertissements, signalerColonnesEcartees(t, ecartees, parties)...)

	return entite, avertissements
}

// nomEntite dérive le nom de classe du nom de table, et dit d'où il sort.
//
// Deux étapes, pas une de plus : le préfixe qu'une décision demande de retirer,
// puis la casse. Rien d'autre — pas de singularisation, pas de préfixe deviné.
//
// C'est une position, et elle vaut d'être expliquée. Un nom de table est un
// constat, au même titre que dans le calque physique ; le traduire en
// convention PHP ne change que sa forme, alors que le mettre au singulier
// change ce qu'il désigne. Une base ne dit pas sa langue, categories est
// category ou catégorie selon celle qu'on lui prête, et une heuristique qui
// tranche en silence produit une classe que personne n'a demandée — et qui
// changera le jour où la liste de mots grossira.
//
// Ce que l'outil sait proposer, il le propose : Proposer rend les renommages
// que la singularisation suggère, écrits en commentaire dans le fichier de
// décisions. Décommenter une ligne suffit, et l'utilisateur voit laquelle était
// douteuse.
//
// Un renommage décidé gagne sans discussion, et court-circuite le préfixe :
// celui qui écrit Client dans son fichier veut Client. Le nom qualifié est
// essayé avant le nom seul — deux schémas peuvent porter une table de même nom,
// et une décision qualifiée doit pouvoir n'en viser qu'une.
func nomEntite(t *calque.Table, d *Decisions, prefixes []string) (string, calque.Origine) {
	cible := t.Schema + "." + t.Nom

	if force, decide := d.Renommages[cible]; decide {
		return force, calque.OrigineDecision
	}
	if force, decide := d.Renommages[t.Nom]; decide {
		return force, calque.OrigineDecision
	}
	return pascalCase(retirerPrefixe(t.Nom, prefixes)), calque.OrigineNommage
}

// inferrerPropriete traduit une colonne. Le type Doctrine apparaît ici, et pas
// dans le physique : il suppose la destination.
func inferrerPropriete(c *calque.Colonne, cibleTable string, d *Decisions) (calque.Propriete, []calque.Avertissement) {
	cible := cibleTable + "." + c.Nom
	var avertissements []calque.Avertissement

	corr, sur := typerColonne(c)
	origine := calque.OrigineContrainte
	requalifiee := false

	if force, decide := d.TypesForces[cible]; decide {
		avant := corr.php
		corr = forcer(corr, force)
		origine = calque.OrigineDecision

		// Changer de famille PHP invalide ce qui décrivait le type précédent.
		// Forcer varchar en text n'y change rien ; forcer char(1) en booléen,
		// si — et c'est le cas courant du O/N d'une base reprise.
		requalifiee = corr.php != avant
	} else if !sur {
		message := "type " + c.TypeBrut + " sans correspondance, rendu en chaîne"
		if estTableau(c) {
			message = "type " + c.TypeBrut + " : tableau PostgreSQL sans type Doctrine, rendu en chaîne (littéral {…})"
		}
		avertissements = append(avertissements, calque.Avertissement{
			Code:       calque.CodeTypeNonReconnu,
			Cible:      cible,
			Message:    message,
			Resolution: calque.ResolutionParDefaut,
			Confiance:  0.3,
		})
	}

	propriete := calque.Propriete{
		Nom:          camelCase(c.Nom),
		Colonne:      c.Nom,
		TypePHP:      typeNullable(corr.php, c.Nullable),
		TypeDoctrine: corr.doctrine,
		Nullable:     c.Nullable,
		Longueur:     c.Longueur,
		Precision:    c.Precision,
		Echelle:      c.Echelle,
		LongueurFixe: longueurFixe(c),
		Commentaire:  c.Commentaire,
		Origine:      origine,
	}

	// Une colonne générée est calculée par la base : l'écrire depuis PHP
	// échouerait, et Doctrine doit le savoir pour l'exclure des INSERT.
	if c.Generee != nil {
		faux := false
		generee := *c.Generee
		propriete.Generee = &generee
		propriete.Insertable = &faux
		propriete.Modifiable = &faux
	}
	if c.Defaut != nil && c.Defaut.Genre == calque.DefautLitteral {
		valeur := c.Defaut.Valeur
		propriete.Defaut = &valeur
	}
	if c.Defaut != nil && c.Defaut.Genre == calque.DefautExpression && !estDefautNul(c.Defaut.Valeur) {
		if sens, reconnu := sensDuDefaut(c); reconnu {
			propriete.DefautExpression = sens
		} else {
			avertissements = append(avertissements, calque.Avertissement{
				Code:       calque.CodeDefautNonReporte,
				Cible:      cible,
				Message:    "défaut " + c.Defaut.Valeur + " non reconnu, non reporté : la valeur est à fournir par l'application",
				Resolution: calque.ResolutionIgnoree,
				Confiance:  1,
			})
		}
	}

	if requalifiee {
		// Longueur, précision, échelle et défaut décrivaient la colonne telle
		// qu'elle était typée. Les reporter sur le nouveau type produirait au
		// mieux du bruit — une longueur sur un booléen —, au pire une entité
		// qui ne compile pas : private bool $actif = 'O'.
		propriete.Longueur, propriete.Precision, propriete.Echelle = nil, nil, nil
		propriete.LongueurFixe = false

		if propriete.Defaut != nil || propriete.DefautExpression != "" {
			valeur := c.Defaut.Valeur
			if valeur == "" {
				valeur = "''"
			}
			avertissements = append(avertissements, calque.Avertissement{
				Code:       calque.CodeDefautIncompatible,
				Cible:      cible,
				Message:    "défaut " + valeur + " écarté, incompatible avec le type " + corr.doctrine + " décidé",
				Resolution: calque.ResolutionForceeParDecision,
				Confiance:  1,
			})
			propriete.Defaut, propriete.DefautExpression = nil, ""
		}
	}

	return propriete, avertissements
}

// inferrerIdentifiant déduit la clé de l'entité de la clé primaire déclarée.
//
// Une table sans clé primaire n'en reçoit pas d'inventée : Doctrine refusera
// l'entité, et c'est préférable à une clé choisie au hasard qui produirait des
// doublons silencieux.
func inferrerIdentifiant(t *calque.Table, cible string, parColonne map[string]*calque.Propriete, sequences []calque.Sequence) (*calque.Identifiant, []calque.Avertissement) {
	var avertissements []calque.Avertissement

	if t.ClePrimaire == nil || len(t.ClePrimaire.Colonnes) == 0 {
		return nil, append(avertissements, calque.Avertissement{
			Code:       calque.CodeTableSansClePrimaire,
			Cible:      cible,
			Message:    "aucune clé primaire : Doctrine refusera cette entité en l'état",
			Resolution: calque.ResolutionAucune,
			Confiance:  1,
		})
	}

	identifiant := &calque.Identifiant{Strategie: calque.IdentifiantAssignee}
	for _, colonne := range t.ClePrimaire.Colonnes {
		propriete, connue := parColonne[colonne]
		if !connue {
			continue
		}
		identifiant.Proprietes = append(identifiant.Proprietes, propriete.Nom)

		// L'identité et la séquence se lisent dans le physique : auto_increment
		// pour IDENTITY, un défaut de genre sequence pour un SERIAL.
		colonnePhysique := t.ColonneParNom(colonne)
		switch {
		case colonnePhysique == nil:
		case colonnePhysique.AutoIncrement:
			identifiant.Strategie = calque.IdentifiantIdentite
		case colonnePhysique.Defaut != nil && colonnePhysique.Defaut.Genre == calque.DefautSequence:
			nom, lu := nomDeSequence(colonnePhysique.Defaut.Valeur)
			if !lu {
				// Laisser la stratégie sequence sans nom ferait choisir à
				// Doctrine une séquence <table>_<colonne>_seq qui n'est
				// peut-être pas celle-là.
				avertissements = append(avertissements, calque.Avertissement{
					Code:       calque.CodeSequenceNonReconnue,
					Cible:      cible + "." + colonne,
					Message:    "défaut « " + colonnePhysique.Defaut.Valeur + " » : aucun nom de séquence à lire, l'identifiant est laissé à l'application",
					Resolution: calque.ResolutionParDefaut,
					Confiance:  1,
				})
				continue
			}
			identifiant.Strategie = calque.IdentifiantSequence
			identifiant.Sequence = nom

			ecrit, _ := nomEcritDeSequence(colonnePhysique.Defaut.Valeur)
			if s := rattacherSequence(ecrit, sequences); s != nil {
				if s.Increment != 0 {
					increment := s.Increment
					identifiant.SequenceIncrement = &increment
				}
				if s.Minimum != nil {
					minimum := *s.Minimum
					identifiant.SequenceMinimum = &minimum
				}
			}
		}
	}

	if len(identifiant.Proprietes) > 1 {
		// Une clé composite est légitime, mais elle interdit un identifiant
		// auto-généré et complique tout le reste : le signaler évite la
		// surprise à la génération.
		identifiant.Strategie = calque.IdentifiantAssignee
		identifiant.Sequence = ""
		identifiant.SequenceIncrement, identifiant.SequenceMinimum = nil, nil
		avertissements = append(avertissements, calque.Avertissement{
			Code:       calque.CodeClePrimaireComposite,
			Cible:      cible,
			Message:    "clé primaire sur " + strings.Join(identifiant.Proprietes, ", "),
			Resolution: calque.ResolutionParDefaut,
			Confiance:  1,
		})
	}
	if len(identifiant.Proprietes) == 0 {
		return nil, avertissements
	}
	return identifiant, avertissements
}

// reporterIndex recopie les index du physique pour que la régénération du
// schéma reste fidèle. Le prédicat d'un index partiel suit tel quel ; méthode
// et classe d'opérateurs n'y survivent pas, Doctrine ne sait pas les exprimer.
//
// Les unicités composites y sont jointes : elles ne se rattachent à aucune
// propriété seule, et sans ça elles ne seraient nulle part. Une unicité que le
// catalogue expose déjà comme index — le cas de PostgreSQL, qui en crée un du
// même nom — n'est pas reprise deux fois.
//
// Un index qui cite une colonne écartée part entier, noté dans parties pour
// chacune de ses colonnes écartées : lui retirer la seule colonne en changerait
// le sens, et une unicité sur (nom, siret) réduite à nom serait fausse. Son
// prédicat compte aussi : il désignerait une colonne que migrations:diff
// propose de supprimer. Le nom y est cherché comme un mot entier, nu ou entre
// délimiteurs ; le prédicat n'est pas analysé, et un littéral qui porte le nom
// fait partir l'index à tort, ce que l'avertissement montre.
func reporterIndex(t *calque.Table, ecartees map[string]bool, parties map[string][]string) []calque.IndexEntite {
	var index []calque.IndexEntite
	for _, idx := range indexReportables(t) {
		cite := false
		for i := range t.Colonnes {
			colonne := t.Colonnes[i].Nom
			if ecartees[colonne] && (slices.Contains(idx.Colonnes, colonne) || mentionne(idx.Predicat, colonne)) {
				cite = true
				parties[colonne] = append(parties[colonne], "index "+idx.Nom)
			}
		}
		if !cite {
			index = append(index, idx)
		}
	}
	return index
}

// indexReportables rend ce que le physique déclare d'indexable sur la table :
// ses index, puis ses unicités composites qu'il n'expose pas déjà comme index.
func indexReportables(t *calque.Table) []calque.IndexEntite {
	var index []calque.IndexEntite
	connus := make(map[string]bool, len(t.Index))

	for _, idx := range t.Index {
		connus[idx.Nom] = true
		index = append(index, calque.IndexEntite{
			Nom:      idx.Nom,
			Colonnes: idx.Colonnes,
			Unique:   idx.Unique,
			Predicat: idx.Predicat,
		})
	}

	for _, u := range t.Unicites {
		if len(u.Colonnes) < 2 || connus[u.Nom] {
			continue
		}
		index = append(index, calque.IndexEntite{
			Nom:      u.Nom,
			Colonnes: u.Colonnes,
			Unique:   true,
		})
	}
	return index
}

// marquerUniques reporte les contraintes d'unicité mono-colonne sur la
// propriété. Une unicité composite reste un index d'entité : elle ne se
// rattache à aucune propriété seule.
func marquerUniques(t *calque.Table, parColonne map[string]*calque.Propriete) {
	for _, u := range t.Unicites {
		if len(u.Colonnes) != 1 {
			continue
		}
		if propriete, connue := parColonne[u.Colonnes[0]]; connue {
			propriete.Unique = true
		}
	}
}

// ensemble indexe une liste de décisions pour l'interroger par appartenance.
func ensemble(valeurs []string) map[string]bool {
	if len(valeurs) == 0 {
		return nil
	}
	e := make(map[string]bool, len(valeurs))
	for _, v := range valeurs {
		e[v] = true
	}
	return e
}

// trierAvertissements impose un ordre stable : ils servent de filtre en CI, et
// un ordre qui change d'une exécution à l'autre y produirait du bruit.
func trierAvertissements(a []calque.Avertissement) {
	sort.SliceStable(a, func(i, j int) bool {
		if a[i].Cible != a[j].Cible {
			return a[i].Cible < a[j].Cible
		}
		return a[i].Code < a[j].Code
	})
}
