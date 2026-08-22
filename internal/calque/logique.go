// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package calque

// Logique est le modèle objet inféré. Contrairement au physique, il n'est pas
// neutre : il parle le vocabulaire de la famille Hibernate. Un ORM qui ne
// partage pas ce modèle consomme le physique.
type Logique struct {
	VersionRI         int             `json:"version_ri"`
	EmpreintePhysique string          `json:"empreinte_physique"`
	EspaceDeNoms      string          `json:"espace_de_noms"`
	Entites           []Entite        `json:"entites"`
	Enumerations      []Enumeration   `json:"enumerations,omitempty"`
	Traits            []Trait         `json:"traits,omitempty"`
	Avertissements    []Avertissement `json:"avertissements,omitempty"`
}

// Origine indique d'où vient une décision. Sans elle, l'outil n'est pas
// auditable et personne ne le lancera sur sa base.
type Origine string

// Origines, de la plus sûre à la plus discutable.
const (
	OrigineContrainte   Origine = "contrainte"
	OrigineVerification Origine = "verification"
	OrigineCardinalite  Origine = "cardinalite"
	OrigineNommage      Origine = "nommage"
	OrigineDecision     Origine = "decision"
)

// Entite est une classe à générer. Toutes les tables n'en produisent pas : une
// table de jointure pure devient une association.
type Entite struct {
	Nom          string         `json:"nom"`
	Table        ReferenceTable `json:"table"`
	Heritage     *Heritage      `json:"heritage,omitempty"`
	Traits       []string       `json:"traits,omitempty"`
	Identifiant  *Identifiant   `json:"identifiant,omitempty"`
	Proprietes   []Propriete    `json:"proprietes"`
	Associations []Association  `json:"associations,omitempty"`
	Index        []IndexEntite  `json:"index,omitempty"`
}

// ReferenceTable qualifie la table d'origine, sous son nom de catalogue.
type ReferenceTable struct {
	Nom    string `json:"nom"`
	Schema string `json:"schema"`
}

// Heritage décrit une hiérarchie déduite du schéma.
type Heritage struct {
	Strategie            StrategieHeritage `json:"strategie"`
	Parent               string            `json:"parent"`
	ColonneDiscriminante string            `json:"colonne_discriminante,omitempty"`
	Origine              Origine           `json:"origine,omitempty"`
}

// StrategieHeritage est la projection de la hiérarchie sur les tables.
type StrategieHeritage string

// Stratégies d'héritage. La table par classe concrète est absente
// volontairement : rien dans un schéma ne permet de la déduire.
const (
	HeritageJointe      StrategieHeritage = "jointe"
	HeritageTableUnique StrategieHeritage = "table_unique"
)

// Identifiant porte la ou les propriétés qui identifient l'entité. Plusieurs
// signifient une clé composite, cas courant sur du legacy.
type Identifiant struct {
	Proprietes []string             `json:"proprietes"`
	Strategie  StrategieIdentifiant `json:"strategie"`
	Sequence   string               `json:"sequence,omitempty"`
}

// StrategieIdentifiant dit qui produit la valeur de la clé.
type StrategieIdentifiant string

// Stratégies d'identifiant. IdentifiantAucune couvre la table sans clé
// primaire : elle se signale, elle ne s'invente pas.
const (
	IdentifiantIdentite StrategieIdentifiant = "identite"
	IdentifiantSequence StrategieIdentifiant = "sequence"
	IdentifiantAucune   StrategieIdentifiant = "aucune"
	IdentifiantAssignee StrategieIdentifiant = "assignee"
)

// Propriete est une colonne devenue attribut. Le type Doctrine apparaît ici, et
// pas dans le physique : il suppose la destination.
type Propriete struct {
	Nom          string `json:"nom"`
	Colonne      string `json:"colonne"`
	TypePHP      string `json:"type_php"`
	TypeDoctrine string `json:"type_doctrine"`
	Nullable     bool   `json:"nullable"`
	Longueur     *int   `json:"longueur,omitempty"`
	Precision    *int   `json:"precision,omitempty"`
	Echelle      *int   `json:"echelle,omitempty"`
	Enumeration  string `json:"enumeration,omitempty"`
	Defaut       string `json:"defaut,omitempty"`
	// Une colonne générée n'est ni insérable ni modifiable. Les pointeurs
	// permettent de ne sérialiser que les cas qui s'écartent du défaut.
	Insertable  *bool   `json:"insertable,omitempty"`
	Modifiable  *bool   `json:"modifiable,omitempty"`
	Unique      bool    `json:"unique,omitempty"`
	Commentaire string  `json:"commentaire,omitempty"`
	Origine     Origine `json:"origine,omitempty"`
}

// Association relie deux entités. Proprietaire décide du côté qui porte la
// colonne de jointure : s'y tromper produit un mapping que Doctrine accepte et
// qui n'écrit rien en base.
type Association struct {
	Nom                string            `json:"nom"`
	Genre              GenreAssociation  `json:"genre"`
	Cible              string            `json:"cible"`
	Proprietaire       bool              `json:"proprietaire"`
	InverseePar        string            `json:"inversee_par,omitempty"`
	MappeePar          string            `json:"mappee_par,omitempty"`
	Jointure           []ColonneJointure `json:"jointure,omitempty"`
	TableJointure      *TableJointure    `json:"table_jointure,omitempty"`
	OrphelinsSupprimes bool              `json:"orphelins_supprimes,omitempty"`
	Origine            Origine           `json:"origine"`
}

// GenreAssociation est la cardinalité de la relation.
type GenreAssociation string

// Cardinalités. Un-vers-un se distingue de plusieurs-vers-un par une contrainte
// d'unicité sur la colonne portante, pas par le nommage.
const (
	UnVersUn               GenreAssociation = "un_vers_un"
	PlusieursVersUn        GenreAssociation = "plusieurs_vers_un"
	UnVersPlusieurs        GenreAssociation = "un_vers_plusieurs"
	PlusieursVersPlusieurs GenreAssociation = "plusieurs_vers_plusieurs"
)

// ColonneJointure apparie colonne portante et colonne référencée. La
// nullabilité, reprise du physique, décide si l'association est facultative.
type ColonneJointure struct {
	Colonne           string `json:"colonne"`
	ColonneReferencee string `json:"colonne_referencee"`
	Nullable          bool   `json:"nullable"`
	ALaSuppression    Action `json:"a_la_suppression,omitempty"`
}

// TableJointure décrit la table d'association d'un plusieurs-vers-plusieurs.
type TableJointure struct {
	Nom             string            `json:"nom"`
	Schema          string            `json:"schema"`
	Jointure        []ColonneJointure `json:"jointure"`
	JointureInverse []ColonneJointure `json:"jointure_inverse"`
}

// IndexEntite reporte un index du physique. Prédicat et méthode n'y survivent
// pas : Doctrine ne sait pas les exprimer.
type IndexEntite struct {
	Nom      string   `json:"nom,omitempty"`
	Colonnes []string `json:"colonnes"`
	Unique   bool     `json:"unique"`
}

// Enumeration est un type PHP à générer. Origine dit d'où elle sort — CHECK,
// type natif ou échantillon — et c'est ce qui permet d'en discuter.
type Enumeration struct {
	Nom         string           `json:"nom"`
	TypeSupport string           `json:"type_support"`
	Cas         []CasEnumeration `json:"cas"`
	Origine     Origine          `json:"origine"`
}

// CasEnumeration apparie un nom de cas PHP à la valeur stockée. Valeur est un
// any : le type support peut être une chaîne comme un entier.
type CasEnumeration struct {
	Nom    string `json:"nom"`
	Valeur any    `json:"valeur"`
}

// Trait regroupe les propriétés récurrentes — created_at, updated_at,
// deleted_at — plutôt que de les recopier dans chaque entité.
type Trait struct {
	Nom        string      `json:"nom"`
	Proprietes []Propriete `json:"proprietes"`
}

// Avertissement est une sortie de premier ordre, pas un journal : ce qui n'est
// pas résolu est signalé ici, jamais inventé ailleurs.
type Avertissement struct {
	Code       string     `json:"code"`
	Cible      string     `json:"cible"`
	Message    string     `json:"message"`
	Resolution Resolution `json:"resolution"`
	Confiance  float64    `json:"confiance"`
}

// Resolution dit ce que l'inférence a fait de l'incertitude signalée.
type Resolution string

// Résolutions. ResolutionParDefaut annonce un choix appliqué faute de mieux,
// ResolutionAucune un trou laissé ouvert.
const (
	ResolutionIgnoree           Resolution = "ignoree"
	ResolutionAucune            Resolution = "aucune"
	ResolutionParDefaut         Resolution = "par_defaut"
	ResolutionForceeParDecision Resolution = "forcee_par_decision"
)

// Codes d'avertissement. Stables entre versions : ils servent de filtre en CI.
const (
	CodeTableSansClePrimaire = "table_sans_cle_primaire"
	CodeClePrimaireComposite = "cle_primaire_composite"
	CodeFKImpliciteProbable  = "fk_implicite_probable"
	CodeTypeNonReconnu       = "type_non_reconnu"
	CodeNomNonSingularisable = "nom_non_singularisable"
	CodeCollision            = "collision_de_nom"
	CodeDecisionOrpheline    = "decision_sans_cible"
)
