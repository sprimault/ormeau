// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package introspection

// Les types de données que l'API de l'interface expose, séparés des interfaces
// du paquet : c'est ce fichier, et lui seul, que tygo traduit en TypeScript.
// Une interface Go n'a pas d'équivalent utile de l'autre côté — elle y devient
// un « any » que personne ne peut employer.

// TableSommaire est ce qu'on sait d'une table sans l'avoir introspectée. Aucune
// donnée métier : LignesEstimees vient des statistiques du catalogue, pas d'un
// COUNT.
type TableSommaire struct {
	Schema         string `json:"schema"`
	Nom            string `json:"nom"`
	Commentaire    string `json:"commentaire,omitempty"`
	NbColonnes     int    `json:"nb_colonnes"`
	LignesEstimees int64  `json:"lignes_estimees"`
	ClePrimaire    bool   `json:"cle_primaire"`
	// ReferenceVers liste les tables qualifiées atteintes par clé étrangère.
	// L'interface s'en sert pour proposer les dépendances quand on coche une
	// table, et pour signaler les références qui sortiraient de la sélection.
	ReferenceVers []string `json:"reference_vers,omitempty"`
}

// ColonneSommaire décrit une colonne juste assez pour décider si on la mappe.
//
// Ce n'est pas ce que le calque enregistre : l'extraction en capture bien
// davantage — défaut structuré, collation, expression de colonne générée — et
// elle capture toutes les colonnes, y compris celles qu'on choisit d'écarter de
// l'entité. Ce type sert l'écran de sélection, pas le format.
type ColonneSommaire struct {
	Nom         string `json:"nom"`
	Position    int    `json:"position"`
	TypeBrut    string `json:"type_brut"`
	Nullable    bool   `json:"nullable"`
	ClePrimaire bool   `json:"cle_primaire"`
	Commentaire string `json:"commentaire,omitempty"`
}

// Portee délimite ce qu'une extraction lit. Valeurs zéro : aucune lecture de
// données.
type Portee struct {
	Schemas        []string `json:"schemas,omitempty"`
	TablesIncluses []string `json:"tables_incluses,omitempty"`
	TablesExclues  []string `json:"tables_exclues,omitempty"`
	// Echantillonner est la seule option qui autorise la lecture de données.
	// Elle alimente la détection d'énumérations et des clés étrangères
	// implicites, cas majoritaire sur du legacy.
	Echantillonner bool `json:"echantillonner,omitempty"`
	// CardinaliteMax plafonne l'échantillonnage : au-delà, une colonne produit
	// une statistique, pas un échantillon.
	CardinaliteMax int `json:"cardinalite_max,omitempty"`
}

// Avancement signale la passe qu'une extraction entame.
//
// Rang et Total comptent des passes, pas des tables : chacune interroge le
// catalogue en une requête pour tous les schémas, et rien ne dit où en est le
// serveur à l'intérieur de cette requête. L'interface affiche donc des paliers,
// jamais un pourcentage lissé qui prétendrait le contraire.
type Avancement struct {
	// Etape est l'un des codes Etape ci-dessous, que l'interface traduit.
	Etape string `json:"etape"`
	Rang  int    `json:"rang"`
	Total int    `json:"total"`
}

// Vocabulaire des passes, partagé par tous les pilotes, et stable : l'interface
// traduit le libellé depuis le code, comme pour les avertissements. Un dialecte
// ne déroule que les passes qui existent chez lui — MySQL n'a pas de séquence —
// et le total suit. Une passe propre à un dialecte appelle un code ici, pas une
// chaîne inventée dans le pilote.
//
// Ici et non dans suivi.go : tygo traduit ce fichier, et le front reçoit ainsi
// chaque code généré. Un test y parcourt les constantes Etape et exige un libellé
// dans les deux langues, sans quoi un code ajouté ici s'afficherait brut.

// Codes des passes d'extraction.
const (
	EtapeSource        = "source"
	EtapeTables        = "tables"
	EtapeColonnes      = "colonnes"
	EtapeContraintes   = "contraintes"
	EtapeIndex         = "index"
	EtapeSequences     = "sequences"
	EtapeTypesEnumeres = "types_enumeres"
	EtapeVues          = "vues"
)
