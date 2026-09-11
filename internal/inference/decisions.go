// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// suffixeDecisions termine le nom d'un fichier de décisions : la base gescom a
// son gescom.decisions.yaml.
const suffixeDecisions = ".decisions.yaml"

// Decisions surcharge l'inférence. Une décision gagne toujours contre une
// heuristique, sans discussion.
//
// Au premier passage, l'outil écrit un fichier prérempli où les inférences de
// confiance moyenne figurent en commentaire. C'est ce qui rend la correction
// humaine praticable sans partir d'une page blanche.
//
// Les balises JSON servent l'interface, qui arbitre un brouillon avant de
// l'écrire ; elles reprennent les noms du fichier, pour qu'une clé se lise de
// la même façon des deux côtés.
type Decisions struct {
	EspaceDeNoms string `json:"espace_de_noms,omitempty" yaml:"espace_de_noms,omitempty"`

	// PrefixesARetirer enlève une convention de nommage des noms de classes :
	// avec T_, la table T_CLIENTS donne Clients au lieu de TClients.
	//
	// Rien n'est retiré sans cette liste, même quand l'outil repère un préfixe
	// commun à toutes les tables. Il le signale par un avertissement et
	// l'inscrit en commentaire dans le fichier prérempli ; l'arbitrage revient
	// à celui qui connaît la base.
	PrefixesARetirer []string `json:"prefixes_a_retirer,omitempty" yaml:"prefixes_a_retirer,omitempty"`

	TablesIgnorees []string `json:"tables_ignorees,omitempty" yaml:"tables_ignorees,omitempty"`

	// ColonnesIgnorees retire des propriétés d'une entité sans rien retirer du
	// calque. Clé : la table qualifiée ; valeurs : les noms de colonnes.
	//
	//	colonnes_ignorees:
	//	  public.clients: [photo, blob_import, champ_libre_12]
	//
	// L'arbitrage est ici et pas à l'extraction, et ce n'est pas un détail de
	// rangement : le calque physique ne perd rien, sans quoi le mode diff
	// signalerait ces colonnes comme disparues à chaque comparaison et
	// l'aller-retour vers le DDL cesserait d'être fidèle. Écarter une colonne de
	// l'entité se rejoue hors ligne, et se défait six mois plus tard sans
	// rouvrir la base.
	//
	// La clé primaire ne s'ignore pas : Doctrine refuse une entité sans
	// identifiant, et la retirer produirait un modèle que rien ne peut charger.
	ColonnesIgnorees map[string][]string `json:"colonnes_ignorees,omitempty" yaml:"colonnes_ignorees,omitempty"`

	// omitempty est répété côté YAML pour tygo, qui retient la balise yaml
	// quand elle existe : sans lui, ces champs sortiraient obligatoires dans les
	// types du front. La lecture n'en tient pas compte.
	Renommages       map[string]string   `json:"renommages,omitempty" yaml:"renommages,omitempty"`
	TypesForces      map[string]string   `json:"types_forces,omitempty" yaml:"types_forces,omitempty"`
	RelationsForcees []RelationForcee    `json:"relations_forcees,omitempty" yaml:"relations_forcees,omitempty"`
	Enumerations     []EnumerationForcee `json:"enumerations,omitempty" yaml:"enumerations,omitempty"`
}

// RelationForcee déclare une association que l'heuristique n'a pas vue — la clé
// étrangère jamais déclarée, que seul l'humain confirme.
//
// Source et Cible sont des colonnes qualifiées, schema.table.colonne : celle
// qui porte la relation, et celle qu'elle désigne. Genre vaut plusieurs_vers_un
// ou un_vers_un, ou reste vide pour laisser l'unicité de la colonne trancher :
// une colonne porte un objet, et le côté collection se déduit sur l'autre
// entité. Nom, vide, suit la règle d'une clé déclarée.
type RelationForcee struct {
	Source string `json:"source" yaml:"source"`
	Cible  string `json:"cible" yaml:"cible"`
	Genre  string `json:"genre" yaml:"genre"`
	Nom    string `json:"nom" yaml:"nom"`
}

// EnumerationForcee impose une énumération. Cas apparie la valeur stockée au
// nom PHP : un O/N en base n'a pas à donner un cas nommé O.
type EnumerationForcee struct {
	Colonne string            `json:"colonne" yaml:"colonne"`
	Nom     string            `json:"nom" yaml:"nom"`
	Cas     map[string]string `json:"cas,omitempty" yaml:"cas,omitempty"`
}

// vide dit si rien n'est décidé : c'est le premier passage, dont le fichier
// reste entièrement en commentaire.
func (d *Decisions) vide() bool {
	return d.EspaceDeNoms == "" &&
		len(d.PrefixesARetirer) == 0 &&
		len(d.TablesIgnorees) == 0 &&
		len(d.ColonnesIgnorees) == 0 &&
		len(d.Renommages) == 0 &&
		len(d.TypesForces) == 0 &&
		len(d.RelationsForcees) == 0 &&
		len(d.Enumerations) == 0
}

// LireDecisions charge le fichier. Un chemin vide rend des décisions vides sans
// erreur : c'est le premier passage.
func LireDecisions(chemin string) (*Decisions, error) {
	if chemin == "" {
		return &Decisions{}, nil
	}

	// Chemin fourni par --decisions : le lire est la fonction de l'option.
	donnees, err := os.ReadFile(chemin) // #nosec G304
	if err != nil {
		return nil, err
	}
	return DecisionsDepuis(donnees)
}

// DecisionsDepuis analyse le contenu d'un fichier de décisions.
//
// Une erreur de syntaxe remonte plutôt que d'être avalée : sinon l'utilisateur
// croit ses arbitrages appliqués alors qu'ils sont ignorés.
func DecisionsDepuis(contenu []byte) (*Decisions, error) {
	var d Decisions
	if err := yaml.Unmarshal(contenu, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// BaseDesDecisions rend le nom de la base qu'un fichier de décisions désigne :
// gescom pour projets/gescom/gescom.decisions.yaml.
//
// Seule source de ce nom, pour la ligne de commande qui le tire du chemin
// qu'on lui donne comme pour l'interface qui compose le chemin à partir du nom.
// Il entre dans l'empreinte du fichier, et deux façons de le calculer finiraient
// par classer manuel un fichier que personne n'a touché. Un nom qui ne suit pas
// la convention perd seulement son extension.
func BaseDesDecisions(chemin string) string {
	nom := filepath.Base(chemin)
	if strings.HasSuffix(nom, suffixeDecisions) {
		return strings.TrimSuffix(nom, suffixeDecisions)
	}
	return strings.TrimSuffix(nom, filepath.Ext(nom))
}

// Reste à écrire avec les heuristiques (phase 3) : la vérification des cibles,
// qui signale les décisions ne correspondant à rien dans le calque physique.
// C'est le signal que la base a bougé sous le fichier — et une vérification qui
// ne rend rien parce qu'elle n'est pas écrite dirait exactement l'inverse.
