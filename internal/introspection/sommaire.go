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
