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
// devient une association et non une entité. Héritage quand la clé primaire
// est aussi une clé étrangère. Énumérations depuis un CHECK. Traits pour
// created_at et updated_at.
//
// Chacune produira soit un élément portant son origine, soit un avertissement,
// jamais une invention.
package inference

import (
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

	logique := &calque.Logique{
		VersionRI:         calque.VersionCourante,
		EmpreintePhysique: p.Source.Empreinte,
		EspaceDeNoms:      espaceDeNoms(d),
	}

	var avertissements []calque.Avertissement
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
	// une table de jointure n'en produit pas, et un héritage se lit sur deux
	// tables à la fois.
	schema := analyser(p, d, prefixes)

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
		Nom:     nom,
		Table:   calque.ReferenceTable{Nom: t.Nom, Schema: t.Schema},
		Origine: origine,
	}

	entite.Proprietes = make([]calque.Propriete, 0, len(t.Colonnes))
	for i := range t.Colonnes {
		propriete, avs := inferrerPropriete(&t.Colonnes[i], cible, d)
		avertissements = append(avertissements, avs...)
		entite.Proprietes = append(entite.Proprietes, propriete)
	}

	// L'index est construit après la boucle, et pas pendant : append réalloue
	// le tableau quand il grandit, et un pointeur pris au passage viserait la
	// zone abandonnée. Les marquages faits ensuite seraient perdus en silence.
	parColonne := make(map[string]*calque.Propriete, len(entite.Proprietes))
	for i := range entite.Proprietes {
		parColonne[entite.Proprietes[i].Colonne] = &entite.Proprietes[i]
	}

	if fk := schema.parents[cible]; fk != nil {
		parent, connu := schema.nomsParTable[fk.SchemaCible+"."+fk.TableCible]
		if connu {
			entite.Heritage = &calque.Heritage{
				Strategie: calque.HeritageJointe,
				Parent:    parent,
				Origine:   calque.OrigineContrainte,
			}
			avertissements = append(avertissements, calque.Avertissement{
				Code:       calque.CodeHeritageDeduit,
				Cible:      cible,
				Message:    "clé primaire portant une clé étrangère vers " + parent + " : héritage par jointure",
				Resolution: calque.ResolutionParDefaut,
				Confiance:  0.7,
			})
		}
	}

	associations, avs := inferrerAssociations(t, schema, parColonne)
	avertissements = append(avertissements, avs...)
	entite.Associations = associations

	identifiant, avs := inferrerIdentifiant(t, cible, parColonne)
	avertissements = append(avertissements, avs...)
	entite.Identifiant = identifiant

	entite.Index = reporterIndex(t)
	marquerUniques(t, parColonne)

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
		avertissements = append(avertissements, calque.Avertissement{
			Code:       calque.CodeTypeNonReconnu,
			Cible:      cible,
			Message:    "type " + c.TypeBrut + " sans correspondance, rendu en chaîne",
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
		Commentaire:  c.Commentaire,
		Origine:      origine,
	}

	// Une colonne générée est calculée par la base : l'écrire depuis PHP
	// échouerait, et Doctrine doit le savoir pour l'exclure des INSERT.
	if c.Generee != nil {
		faux := false
		propriete.Insertable = &faux
		propriete.Modifiable = &faux
	}
	if c.Defaut != nil && c.Defaut.Genre == calque.DefautLitteral {
		propriete.Defaut = c.Defaut.Valeur
	}

	if requalifiee {
		// Longueur, précision, échelle et défaut décrivaient la colonne telle
		// qu'elle était typée. Les reporter sur le nouveau type produirait au
		// mieux du bruit — une longueur sur un booléen —, au pire une entité
		// qui ne compile pas : private bool $actif = 'O'.
		propriete.Longueur, propriete.Precision, propriete.Echelle = nil, nil, nil

		if propriete.Defaut != "" {
			avertissements = append(avertissements, calque.Avertissement{
				Code:       calque.CodeDefautIncompatible,
				Cible:      cible,
				Message:    "défaut " + propriete.Defaut + " écarté, incompatible avec le type " + corr.doctrine + " décidé",
				Resolution: calque.ResolutionForceeParDecision,
				Confiance:  1,
			})
			propriete.Defaut = ""
		}
	}

	return propriete, avertissements
}

// inferrerIdentifiant déduit la clé de l'entité de la clé primaire déclarée.
//
// Une table sans clé primaire n'en reçoit pas d'inventée : Doctrine refusera
// l'entité, et c'est préférable à une clé choisie au hasard qui produirait des
// doublons silencieux.
func inferrerIdentifiant(t *calque.Table, cible string, parColonne map[string]*calque.Propriete) (*calque.Identifiant, []calque.Avertissement) {
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
			identifiant.Strategie = calque.IdentifiantSequence
			identifiant.Sequence = colonnePhysique.Defaut.Valeur
		}
	}

	if len(identifiant.Proprietes) > 1 {
		// Une clé composite est légitime, mais elle interdit un identifiant
		// auto-généré et complique tout le reste : le signaler évite la
		// surprise à la génération.
		identifiant.Strategie = calque.IdentifiantAssignee
		identifiant.Sequence = ""
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
// schéma reste fidèle. Prédicat et classe d'opérateurs n'y survivent pas :
// Doctrine ne sait pas les exprimer.
//
// Les unicités composites y sont jointes : elles ne se rattachent à aucune
// propriété seule, et sans ça elles ne seraient nulle part. Une unicité que le
// catalogue expose déjà comme index — le cas de PostgreSQL, qui en crée un du
// même nom — n'est pas reprise deux fois.
func reporterIndex(t *calque.Table) []calque.IndexEntite {
	var index []calque.IndexEntite
	connus := make(map[string]bool, len(t.Index))

	for _, idx := range t.Index {
		connus[idx.Nom] = true
		index = append(index, calque.IndexEntite{
			Nom:      idx.Nom,
			Colonnes: idx.Colonnes,
			Unique:   idx.Unique,
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
