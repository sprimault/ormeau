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

const (
	OrigineContrainte   Origine = "contrainte"
	OrigineVerification Origine = "verification"
	OrigineCardinalite  Origine = "cardinalite"
	OrigineNommage      Origine = "nommage"
	OrigineDecision     Origine = "decision"
)

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

type ReferenceTable struct {
	Nom    string `json:"nom"`
	Schema string `json:"schema"`
}

type Heritage struct {
	Strategie           StrategieHeritage `json:"strategie"`
	Parent              string            `json:"parent"`
	ColonneDiscriminante string           `json:"colonne_discriminante,omitempty"`
	Origine             Origine           `json:"origine,omitempty"`
}

type StrategieHeritage string

const (
	HeritageJointe      StrategieHeritage = "jointe"
	HeritageTableUnique StrategieHeritage = "table_unique"
)

type Identifiant struct {
	Proprietes []string             `json:"proprietes"`
	Strategie  StrategieIdentifiant `json:"strategie"`
	Sequence   string               `json:"sequence,omitempty"`
}

type StrategieIdentifiant string

const (
	IdentifiantIdentite StrategieIdentifiant = "identite"
	IdentifiantSequence StrategieIdentifiant = "sequence"
	IdentifiantAucune   StrategieIdentifiant = "aucune"
	IdentifiantAssignee StrategieIdentifiant = "assignee"
)

type Propriete struct {
	Nom           string `json:"nom"`
	Colonne       string `json:"colonne"`
	TypePHP       string `json:"type_php"`
	TypeDoctrine  string `json:"type_doctrine"`
	Nullable      bool   `json:"nullable"`
	Longueur      *int   `json:"longueur,omitempty"`
	Precision     *int   `json:"precision,omitempty"`
	Echelle       *int   `json:"echelle,omitempty"`
	Enumeration   string `json:"enumeration,omitempty"`
	Defaut        string `json:"defaut,omitempty"`
	// Une colonne générée n'est ni insérable ni modifiable. Les pointeurs
	// permettent de ne sérialiser que les cas qui s'écartent du défaut.
	Insertable  *bool   `json:"insertable,omitempty"`
	Modifiable  *bool   `json:"modifiable,omitempty"`
	Unique      bool    `json:"unique,omitempty"`
	Commentaire string  `json:"commentaire,omitempty"`
	Origine     Origine `json:"origine,omitempty"`
}

type Association struct {
	Nom          string           `json:"nom"`
	Genre        GenreAssociation `json:"genre"`
	Cible        string           `json:"cible"`
	Proprietaire bool             `json:"proprietaire"`
	InverseePar  string           `json:"inversee_par,omitempty"`
	MappeePar    string           `json:"mappee_par,omitempty"`
	Jointure     []ColonneJointure `json:"jointure,omitempty"`
	TableJointure *TableJointure   `json:"table_jointure,omitempty"`
	OrphelinsSupprimes bool        `json:"orphelins_supprimes,omitempty"`
	Origine      Origine          `json:"origine"`
}

type GenreAssociation string

const (
	UnVersUn                 GenreAssociation = "un_vers_un"
	PlusieursVersUn          GenreAssociation = "plusieurs_vers_un"
	UnVersPlusieurs          GenreAssociation = "un_vers_plusieurs"
	PlusieursVersPlusieurs   GenreAssociation = "plusieurs_vers_plusieurs"
)

type ColonneJointure struct {
	Colonne           string `json:"colonne"`
	ColonneReferencee string `json:"colonne_referencee"`
	Nullable          bool   `json:"nullable"`
	ALaSuppression    Action `json:"a_la_suppression,omitempty"`
}

type TableJointure struct {
	Nom             string            `json:"nom"`
	Schema          string            `json:"schema"`
	Jointure        []ColonneJointure `json:"jointure"`
	JointureInverse []ColonneJointure `json:"jointure_inverse"`
}

type IndexEntite struct {
	Nom      string   `json:"nom,omitempty"`
	Colonnes []string `json:"colonnes"`
	Unique   bool     `json:"unique"`
}

type Enumeration struct {
	Nom         string          `json:"nom"`
	TypeSupport string          `json:"type_support"`
	Cas         []CasEnumeration `json:"cas"`
	Origine     Origine         `json:"origine"`
}

type CasEnumeration struct {
	Nom    string `json:"nom"`
	Valeur any    `json:"valeur"`
}

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

type Resolution string

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
	CodeFKImplicteProbable   = "fk_implicite_probable"
	CodeTypeNonReconnu       = "type_non_reconnu"
	CodeNomNonSingularisable = "nom_non_singularisable"
	CodeCollision            = "collision_de_nom"
	CodeDecisionOrpheline    = "decision_sans_cible"
)
