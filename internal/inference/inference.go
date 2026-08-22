// Package inference transforme un calque physique en calque logique.
//
// C'est le cœur de la valeur du projet : tout ce qui distingue Ormeau d'un
// générateur naïf est ici.
package inference

import (
	"github.com/sprimault/ormeau/internal/calque"
)

// Inferer est une fonction pure : pas de réseau, pas d'effet de bord, pas
// d'horloge, pas d'aléa. Toute heuristique qui aurait besoin d'interroger la
// base est mal placée — ce qu'elle cherche doit être capturé à l'extraction.
//
// Cette pureté est ce qui permet de corriger une inférence hors ligne, de la
// rejouer, et de tester sans base de données.
func Inferer(p *calque.Physique, d *Decisions) (*calque.Logique, []calque.Avertissement) {
	panic("à implémenter")
}

// Les heuristiques ci-dessous sont déclarées dans l'ordre où elles s'appliquent.
// Chacune produit soit un élément avec son origine, soit un avertissement —
// jamais une invention.

// estTableDeJointurePure reconnaît une table à deux clés étrangères dont la clé
// primaire est exactement l'union des colonnes portantes, et qui ne porte
// aucune autre colonne. Elle devient une association plusieurs-vers-plusieurs,
// pas une entité.
//
// Une colonne supplémentaire, même un simple created_at, disqualifie : la table
// porte alors des données propres et mérite son entité.
func estTableDeJointurePure(t *calque.Table) bool {
	panic("à implémenter")
}

// heriteParClePrimaire reconnaît le cas où la clé primaire est aussi une clé
// étrangère vers une autre table : héritage à tables jointes.
func heriteParClePrimaire(t *calque.Table) (parent string, ok bool) {
	panic("à implémenter")
}

// enumerationDepuisVerification extrait les valeurs d'un CHECK ... IN (...).
// L'expression est verbatim et dépend du dialecte, d'où l'analyse par SGBD.
func enumerationDepuisVerification(v calque.Verification, sgbd string) ([]string, bool) {
	panic("à implémenter")
}

// nomEntite retire les préfixes déclarés, singularise en français comme en
// anglais, et passe en PascalCase. Un nom non singularisable produit un
// avertissement plutôt qu'une transformation approximative.
func nomEntite(nomTable string, prefixes []string) (string, bool) {
	panic("à implémenter")
}

// nomPropriete transforme client_id en client quand la colonne porte une clé
// étrangère, et en clientId sinon.
func nomPropriete(nomColonne string, porteFK bool) string {
	panic("à implémenter")
}

// traitsCommuns reconnaît les groupes de colonnes qui méritent un trait plutôt
// que des propriétés recopiées dans chaque entité : created_at/updated_at,
// deleted_at.
func traitsCommuns(t *calque.Table) []string {
	panic("à implémenter")
}

// fkImplicites cherche les colonnes qui se comportent comme des clés étrangères
// sans contrainte déclarée : nommage compatible, type compatible, et toutes les
// valeurs présentes dans la table cible d'après les statistiques.
//
// Cas majoritaire sur du legacy. Ne produit jamais d'association d'office :
// seulement un avertissement avec sa confiance, que l'utilisateur convertit en
// décision s'il valide.
func fkImplicites(p *calque.Physique) []calque.Avertissement {
	panic("à implémenter")
}
