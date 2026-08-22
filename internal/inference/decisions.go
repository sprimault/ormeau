package inference

import (
	"os"

	"gopkg.in/yaml.v3"

	"github.com/sprimault/ormeau/internal/calque"
)

// Decisions surcharge l'inférence. Une décision gagne toujours contre une
// heuristique, sans discussion.
//
// Au premier passage, l'outil écrit un fichier prérempli où les inférences de
// confiance moyenne figurent en commentaire. C'est ce qui rend la correction
// humaine praticable sans partir d'une page blanche.
type Decisions struct {
	EspaceDeNoms     string             `yaml:"espace_de_noms"`
	PrefixesARetirer []string           `yaml:"prefixes_a_retirer"`
	TablesIgnorees   []string           `yaml:"tables_ignorees"`
	Renommages       map[string]string  `yaml:"renommages"`
	TypesForces      map[string]string  `yaml:"types_forces"`
	RelationsForcees []RelationForcee   `yaml:"relations_forcees"`
	Enumerations     []EnumerationForcee `yaml:"enumerations"`
}

type RelationForcee struct {
	Source string `yaml:"source"`
	Cible  string `yaml:"cible"`
	Genre  string `yaml:"genre"`
	Nom    string `yaml:"nom"`
}

type EnumerationForcee struct {
	Colonne string            `yaml:"colonne"`
	Nom     string            `yaml:"nom"`
	Cas     map[string]string `yaml:"cas"`
}

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

// VerifierCibles signale les décisions qui ne correspondent à rien dans le
// calque physique : c'est le signal que la base a bougé sous le fichier.
func (d *Decisions) VerifierCibles(p *calque.Physique) []calque.Avertissement {
	panic("à implémenter")
}
