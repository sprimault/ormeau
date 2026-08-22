// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Package inference transforme un calque physique en calque logique.
//
// C'est le cœur de la valeur du projet : tout ce qui distingue Ormeau d'un
// générateur naïf est ici.
//
// Le point d'entrée est arrêté, et sa signature ne bougera pas :
//
//	func Inferer(p *calque.Physique, d *Decisions) (*calque.Logique, []calque.Avertissement)
//
// Aucun autre paramètre, aucune variable globale, aucun accès au monde
// extérieur. Cette pureté est ce qui permet de corriger une inférence hors
// ligne, de la rejouer, et de tester sans base de données.
//
// Elle n'est pas déclarée tant qu'elle n'est pas écrite (phase 3) : ne pouvant
// pas signaler son inachèvement par son type de retour, une ébauche rendrait un
// calque logique nul que l'appelant déréférencerait plus loin.
//
// # Heuristiques à écrire, dans l'ordre où elles s'appliquent
//
// Chacune produit soit un élément portant son origine, soit un avertissement —
// jamais une invention.
//
// Table de jointure pure : deux clés étrangères dont la clé primaire est
// exactement l'union des colonnes portantes, et aucune autre colonne. Elle
// devient une association plusieurs-vers-plusieurs, pas une entité. Une colonne
// supplémentaire, même un simple created_at, disqualifie : la table porte alors
// des données propres et mérite son entité.
//
// Héritage par clé primaire : la clé primaire est aussi une clé étrangère vers
// une autre table, d'où un héritage à tables jointes.
//
// Énumération depuis une vérification : les valeurs d'un CHECK ... IN (...).
// L'expression est verbatim et dépend du dialecte, d'où une analyse par SGBD.
//
// Nom d'entité : retrait des préfixes déclarés, singularisation française comme
// anglaise, PascalCase. Un nom non singularisable produit un avertissement
// plutôt qu'une transformation approximative.
//
// Nom de propriété : client_id devient client quand la colonne porte une clé
// étrangère, clientId sinon.
//
// Traits communs : les groupes de colonnes qui méritent un trait plutôt que des
// propriétés recopiées dans chaque entité — created_at/updated_at, deleted_at.
//
// Clés étrangères implicites : les colonnes qui se comportent comme des clés
// étrangères sans contrainte déclarée — nommage compatible, type compatible, et
// toutes les valeurs présentes dans la table cible d'après les statistiques.
// Cas majoritaire sur du legacy. Jamais d'association d'office : un
// avertissement avec sa confiance, que l'utilisateur convertit en décision s'il
// le valide.
package inference
