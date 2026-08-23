// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

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

// Source identifie la base d'origine. Jamais le DSN : un calque se versionne.
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

// Table porte le nom du catalogue, préfixes compris. Le renommage est ailleurs.
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

// Colonne garde le type verbatim et sa réduction au vocabulaire fermé.
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

// Types normalisés. TypeInconnu est l'issue honnête devant un type non reconnu ;
// TypeBrut reste exploitable dans ce cas.
const (
	TypeEntier      TypeNorm = "entier"
	TypeDecimal     TypeNorm = "decimal"
	TypeFlottant    TypeNorm = "flottant"
	TypeTexte       TypeNorm = "texte"
	TypeBinaire     TypeNorm = "binaire"
	TypeBooleen     TypeNorm = "booleen"
	TypeDate        TypeNorm = "date"
	TypeHeure       TypeNorm = "heure"
	TypeHorodatage  TypeNorm = "horodatage"
	TypeIntervalle  TypeNorm = "intervalle"
	TypeUUID        TypeNorm = "uuid"
	TypeJSON        TypeNorm = "json"
	TypeXML         TypeNorm = "xml"
	TypeGeometrie   TypeNorm = "geometrie"
	TypeEnumereNorm TypeNorm = "enumere"
	TypeReseau      TypeNorm = "reseau"
	TypeInconnu     TypeNorm = "inconnu"
)

// Defaut est structuré : une chaîne nue rendrait DEFAULT 'now()' et
// DEFAULT now() indistinguables, et les entités produites seraient fausses.
type Defaut struct {
	Genre  GenreDefaut `json:"genre"`
	Valeur string      `json:"valeur"`
}

// GenreDefaut dit comment lire la valeur d'un défaut.
type GenreDefaut string

// Genres de valeur par défaut.
const (
	DefautLitteral   GenreDefaut = "litteral"
	DefautExpression GenreDefaut = "expression"
	DefautSequence   GenreDefaut = "sequence"
)

// Generee décrit une colonne calculée. Stockee distingue le calcul matérialisé
// du calcul à la lecture.
type Generee struct {
	Expression string `json:"expression"`
	Stockee    bool   `json:"stockee"`
}

// ClePrimaire garde les colonnes dans l'ordre de la contrainte, qui n'est pas
// toujours celui de la table. Son absence est courante sur du legacy.
type ClePrimaire struct {
	Nom      string   `json:"nom,omitempty"`
	Colonnes []string `json:"colonnes"`
}

// CleEtrangere est une contrainte réellement déclarée. Une clé seulement
// probable est inférée plus loin, et signalée.
type CleEtrangere struct {
	Nom            string   `json:"nom,omitempty"`
	Colonnes       []string `json:"colonnes"`
	TableCible     string   `json:"table_cible"`
	SchemaCible    string   `json:"schema_cible"`
	ColonnesCibles []string `json:"colonnes_cibles"`
	ALaSuppression Action   `json:"a_la_suppression,omitempty"`
	ALaMiseAJour   Action   `json:"a_la_mise_a_jour,omitempty"`
}

// Action est le comportement référentiel d'une clé étrangère.
type Action string

// Actions référentielles.
const (
	ActionAucune     Action = "aucune"
	ActionCascade    Action = "cascade"
	ActionSetNull    Action = "set_null"
	ActionSetDefault Action = "set_default"
	ActionRestrict   Action = "restrict"
)

// Contrainte est un groupe de colonnes nommé, pour les unicités. L'ordre des
// colonnes est significatif.
type Contrainte struct {
	Nom      string   `json:"nom,omitempty"`
	Colonnes []string `json:"colonnes"`
}

// Index ne sert pas au modèle objet, mais à régénérer un schéma fidèle.
type Index struct {
	Nom      string   `json:"nom,omitempty"`
	Colonnes []string `json:"colonnes"`
	Unique   bool     `json:"unique"`
	Predicat string   `json:"predicat,omitempty"`
	Methode  string   `json:"methode,omitempty"`
	// Operateurs porte la classe d'opérateurs de chaque colonne, dans le même
	// ordre. Absent quand toutes sont celles par défaut ; complet sinon, une
	// entrée par colonne.
	//
	// Sans lui, deux index HNSW mesurant l'un une distance cosinus et l'autre
	// une distance euclidienne sont indistinguables, et le DDL n'est plus
	// reconstructible. Même chose pour un btree en text_pattern_ops.
	Operateurs []string `json:"operateurs,omitempty"`
}

// Verification est un CHECK, verbatim parce que sa syntaxe dépend du dialecte.
// C'est la source d'énumération la plus fiable, avant tout échantillonnage.
type Verification struct {
	Nom        string `json:"nom,omitempty"`
	Expression string `json:"expression"`
}

// OptionsTable regroupe ce qui n'existe pas partout : moteur et collation de
// table sont des notions MySQL.
type OptionsTable struct {
	Moteur    string `json:"moteur,omitempty"`
	Collation string `json:"collation,omitempty"`
}

// Sequence est un générateur du catalogue, distinct d'une colonne IDENTITY. La
// distinction décide de la stratégie d'identifiant.
type Sequence struct {
	Nom       string `json:"nom"`
	Schema    string `json:"schema"`
	Increment int64  `json:"increment,omitempty"`
	Minimum   *int64 `json:"minimum,omitempty"`
	Maximum   *int64 `json:"maximum,omitempty"`
	Cyclique  bool   `json:"cyclique,omitempty"`
}

// TypeEnumere est un type énuméré natif du SGBD, référencé par Colonne.
// Sans rapport avec les énumérations déduites d'un CHECK.
type TypeEnumere struct {
	Nom     string   `json:"nom"`
	Schema  string   `json:"schema"`
	Valeurs []string `json:"valeurs"`
}

// Vue ne produit pas d'entité, mais sans elle le DDL n'est plus reconstructible.
type Vue struct {
	Nom          string    `json:"nom"`
	Schema       string    `json:"schema"`
	Definition   string    `json:"definition"`
	Materialisee bool      `json:"materialisee"`
	Colonnes     []Colonne `json:"colonnes,omitempty"`
}

// StatistiquesTable n'est renseignée qu'avec --echantillonner. LignesEstimees
// vient du planificateur, jamais d'un COUNT.
type StatistiquesTable struct {
	LignesEstimees int64                          `json:"lignes_estimees,omitempty"`
	Colonnes       map[string]StatistiquesColonne `json:"colonnes,omitempty"`
}

// StatistiquesColonne alimente la détection d'énumérations et des clés
// étrangères implicites. Echantillon reste sous le plafond de cardinalité :
// c'est ce qui empêche le calque de dériver vers le dump.
type StatistiquesColonne struct {
	Distinctes  *int64   `json:"distinctes,omitempty"`
	Nuls        *int64   `json:"nuls,omitempty"`
	Echantillon []string `json:"echantillon,omitempty"`
}

// TableParNom retrouve une table qualifiée. Utilisé par l'inférence pour
// résoudre les cibles de clés étrangères.
//
// Sensible à la casse : sur les SGBD qui l'autorisent, deux tables ne différant
// que par la casse sont deux tables.
func (p *Physique) TableParNom(schema, nom string) *Table {
	for i := range p.Tables {
		if p.Tables[i].Schema == schema && p.Tables[i].Nom == nom {
			return &p.Tables[i]
		}
	}
	return nil
}

// ColonneParNom retrouve une colonne, sensible à la casse comme TableParNom.
func (t *Table) ColonneParNom(nom string) *Colonne {
	for i := range t.Colonnes {
		if t.Colonnes[i].Nom == nom {
			return &t.Colonnes[i]
		}
	}
	return nil
}
