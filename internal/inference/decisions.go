// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Decisions surcharge l'inférence. Une décision gagne toujours contre une
// heuristique, sans discussion.
//
// Au premier passage, l'outil écrit un fichier prérempli où les inférences de
// confiance moyenne figurent en commentaire. C'est ce qui rend la correction
// humaine praticable sans partir d'une page blanche.
type Decisions struct {
	EspaceDeNoms     string              `yaml:"espace_de_noms"`
	PrefixesARetirer []string            `yaml:"prefixes_a_retirer"`
	TablesIgnorees   []string            `yaml:"tables_ignorees"`
	Renommages       map[string]string   `yaml:"renommages"`
	TypesForces      map[string]string   `yaml:"types_forces"`
	RelationsForcees []RelationForcee    `yaml:"relations_forcees"`
	Enumerations     []EnumerationForcee `yaml:"enumerations"`
}

// RelationForcee déclare une association que l'heuristique n'a pas vue — la clé
// étrangère jamais déclarée, que seul l'humain confirme.
type RelationForcee struct {
	Source string `yaml:"source"`
	Cible  string `yaml:"cible"`
	Genre  string `yaml:"genre"`
	Nom    string `yaml:"nom"`
}

// EnumerationForcee impose une énumération. Cas apparie la valeur stockée au
// nom PHP : un O/N en base n'a pas à donner un cas nommé O.
type EnumerationForcee struct {
	Colonne string            `yaml:"colonne"`
	Nom     string            `yaml:"nom"`
	Cas     map[string]string `yaml:"cas"`
}

// LireDecisions charge le fichier. Un chemin vide rend des décisions vides sans
// erreur : c'est le premier passage.
//
// Une erreur de syntaxe remonte plutôt que d'être avalée : sinon l'utilisateur
// croit ses arbitrages appliqués alors qu'ils sont ignorés.
func LireDecisions(chemin string) (*Decisions, error) {
	if chemin == "" {
		return &Decisions{}, nil
	}

	donnees, err := os.ReadFile(chemin)
	if err != nil {
		return nil, err
	}

	var d Decisions
	if err := yaml.Unmarshal(donnees, &d); err != nil {
		return nil, err
	}
	return &d, nil
}

// Reste à écrire avec les heuristiques (phase 3) : la vérification des cibles,
// qui signale les décisions ne correspondant à rien dans le calque physique.
// C'est le signal que la base a bougé sous le fichier — et une vérification qui
// ne rend rien parce qu'elle n'est pas écrite dirait exactement l'inverse.
