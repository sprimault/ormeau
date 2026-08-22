// Package calque définit le format pivot d'Ormeau et sa sérialisation.
package calque

// Physique est le décalque du catalogue. Il ne juge rien et ne perd rien.
type Physique struct {
	VersionRI     int                          `json:"version_ri"`
	Source        Source                       `json:"source"`
	Tables        []Table                      `json:"tables"`
	Sequences     []Sequence                   `json:"sequences,omitempty"`
	TypesEnumeres []TypeEnumere                `json:"types_enumeres,omitempty"`
	Vues          []Vue                        `json:"vues,omitempty"`
	Statistiques  map[string]StatistiquesTable `json:"statistiques,omitempty"`
}

type Source struct {
	SGBD      string `json:"sgbd"`
	Version   string `json:"version"`
	Catalogue string `json:"catalogue"`
	Schema    string `json:"schema"`
	// ExtraitLe et Empreinte sont exclus du calcul de l'empreinte : sans ça,
	// deux extractions identiques produiraient des empreintes différentes.
	ExtraitLe string `json:"extrait_le"`
	Empreinte string `json:"empreinte"`
}

type Table struct {
	Nom            string         `json:"nom"`
	Schema         string         `json:"schema"`
	Commentaire    string         `json:"commentaire,omitempty"`
	Colonnes       []Colonne      `json:"colonnes"`
	ClePrimaire    *ClePrimaire   `json:"cle_primaire,omitempty"`
	ClesEtrangeres []CleEtrangere `json:"cles_etrangeres,omitempty"`
	Unicites       []Contrainte   `json:"unicites,omitempty"`
	Index          []Index        `json:"index,omitempty"`
	Verifications  []Verification `json:"verifications,omitempty"`
	Options        *OptionsTable  `json:"options,omitempty"`
}

type Colonne struct {
	Nom      string `json:"nom"`
	Position int    `json:"position"`
	// TypeBrut est verbatim. Ne jamais le reconstruire à partir des composants :
	// c'est lui qui sauve la mise devant un hierarchyid ou un domaine maison.
	TypeBrut      string   `json:"type_brut"`
	TypeNormalise TypeNorm `json:"type_normalise"`
	// Pointeurs : decimal(10,0) et int doivent rester distinguables, un zéro
	// implicite les confondrait.
	Longueur      *int     `json:"longueur,omitempty"`
	Precision     *int     `json:"precision,omitempty"`
	Echelle       *int     `json:"echelle,omitempty"`
	Nullable      bool     `json:"nullable"`
	AutoIncrement bool     `json:"auto_increment,omitempty"`
	Defaut        *Defaut  `json:"defaut,omitempty"`
	Generee       *Generee `json:"generee,omitempty"`
	Collation     string   `json:"collation,omitempty"`
	TypeEnumere   string   `json:"type_enumere,omitempty"`
	Commentaire   string   `json:"commentaire,omitempty"`
}

// TypeNorm est un vocabulaire fermé. Toute valeur ajoutée incrémente VersionRI.
type TypeNorm string

const (
	TypeEntier     TypeNorm = "entier"
	TypeDecimal    TypeNorm = "decimal"
	TypeFlottant   TypeNorm = "flottant"
	TypeTexte      TypeNorm = "texte"
	TypeBinaire    TypeNorm = "binaire"
	TypeBooleen    TypeNorm = "booleen"
	TypeDate       TypeNorm = "date"
	TypeHeure      TypeNorm = "heure"
	TypeHorodatage TypeNorm = "horodatage"
	TypeIntervalle TypeNorm = "intervalle"
	TypeUUID       TypeNorm = "uuid"
	TypeJSON       TypeNorm = "json"
	TypeXML        TypeNorm = "xml"
	TypeGeometrie  TypeNorm = "geometrie"
	TypeEnumereNorm TypeNorm = "enumere"
	TypeReseau     TypeNorm = "reseau"
	TypeInconnu    TypeNorm = "inconnu"
)

// Defaut est structuré : une chaîne nue rendrait DEFAULT 'now()' et
// DEFAULT now() indistinguables, et les entités produites seraient fausses.
type Defaut struct {
	Genre  GenreDefaut `json:"genre"`
	Valeur string      `json:"valeur"`
}

type GenreDefaut string

const (
	DefautLitteral   GenreDefaut = "litteral"
	DefautExpression GenreDefaut = "expression"
	DefautSequence   GenreDefaut = "sequence"
)

type Generee struct {
	Expression string `json:"expression"`
	Stockee    bool   `json:"stockee"`
}

type ClePrimaire struct {
	Nom      string   `json:"nom,omitempty"`
	Colonnes []string `json:"colonnes"`
}

type CleEtrangere struct {
	Nom             string   `json:"nom,omitempty"`
	Colonnes        []string `json:"colonnes"`
	TableCible      string   `json:"table_cible"`
	SchemaCible     string   `json:"schema_cible"`
	ColonnesCibles  []string `json:"colonnes_cibles"`
	ALaSuppression  Action   `json:"a_la_suppression,omitempty"`
	ALaMiseAJour    Action   `json:"a_la_mise_a_jour,omitempty"`
}

type Action string

const (
	ActionAucune     Action = "aucune"
	ActionCascade    Action = "cascade"
	ActionSetNull    Action = "set_null"
	ActionSetDefault Action = "set_default"
	ActionRestrict   Action = "restrict"
)

type Contrainte struct {
	Nom      string   `json:"nom,omitempty"`
	Colonnes []string `json:"colonnes"`
}

type Index struct {
	Nom      string   `json:"nom,omitempty"`
	Colonnes []string `json:"colonnes"`
	Unique   bool     `json:"unique"`
	Predicat string   `json:"predicat,omitempty"`
	Methode  string   `json:"methode,omitempty"`
}

type Verification struct {
	Nom        string `json:"nom,omitempty"`
	Expression string `json:"expression"`
}

type OptionsTable struct {
	Moteur    string `json:"moteur,omitempty"`
	Collation string `json:"collation,omitempty"`
}

type Sequence struct {
	Nom       string `json:"nom"`
	Schema    string `json:"schema"`
	Increment int64  `json:"increment,omitempty"`
	Minimum   *int64 `json:"minimum,omitempty"`
	Maximum   *int64 `json:"maximum,omitempty"`
	Cyclique  bool   `json:"cyclique,omitempty"`
}

type TypeEnumere struct {
	Nom     string   `json:"nom"`
	Schema  string   `json:"schema"`
	Valeurs []string `json:"valeurs"`
}

type Vue struct {
	Nom          string    `json:"nom"`
	Schema       string    `json:"schema"`
	Definition   string    `json:"definition"`
	Materialisee bool      `json:"materialisee"`
	Colonnes     []Colonne `json:"colonnes,omitempty"`
}

type StatistiquesTable struct {
	LignesEstimees int64                          `json:"lignes_estimees,omitempty"`
	Colonnes       map[string]StatistiquesColonne `json:"colonnes,omitempty"`
}

type StatistiquesColonne struct {
	Distinctes  *int64   `json:"distinctes,omitempty"`
	Nuls        *int64   `json:"nuls,omitempty"`
	Echantillon []string `json:"echantillon,omitempty"`
}

// TableParNom retrouve une table qualifiée. Utilisé par l'inférence pour
// résoudre les cibles de clés étrangères.
func (p *Physique) TableParNom(schema, nom string) *Table {
	for i := range p.Tables {
		if p.Tables[i].Schema == schema && p.Tables[i].Nom == nom {
			return &p.Tables[i]
		}
	}
	return nil
}

func (t *Table) ColonneParNom(nom string) *Colonne {
	for i := range t.Colonnes {
		if t.Colonnes[i].Nom == nom {
			return &t.Colonnes[i]
		}
	}
	return nil
}
