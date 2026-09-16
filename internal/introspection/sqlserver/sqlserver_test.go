// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package sqlserver

import (
	"regexp"
	"strings"
	"testing"
)

// requetes rassemble ce que le pilote envoie au serveur. Toute requête ajoutée
// se déclare ici, faute de quoi le contrôle de lecture seule ne la voit pas.
var requetes = map[string]string{
	"serveur":             requeteServeur,
	"bases":               requeteBases,
	"schemas":             requeteSchemas,
	"inventaire":          requeteInventaire,
	"colonnes":            requeteColonnes,
	"tables extraction":   requeteTablesExtraction,
	"colonnes extraction": requeteColonnesExtraction,
	"types decrits":       requeteTypesDecrits,
	"cles primaires":      requeteClesPrimaires,
	"cles etrangeres":     requeteClesEtrangeres,
}

// verbeDEcriture reconnaît un mot entier, jamais une sous-chaîne : le
// catalogue nomme delete_referential_action_desc et dm_exec_describe_…, qui ne
// modifient rien.
//
// EXEC et sp_ ferment la porte aux procédures : sp_addextendedproperty écrit,
// et rien ne distingue de l'extérieur celles qui lisent. INTO couvre
// SELECT … INTO, qui crée une table.
var verbeDEcriture = regexp.MustCompile(`(?i)\b(insert|update|delete|merge|truncate|drop|alter|create|grant|revoke|deny|exec|execute|into)\b|\bsp_`)

// TestAucuneRequeteNEcrit tient l'invariant « aucune écriture dans la base
// introspectée » là où le serveur ne peut pas le tenir.
//
// PostgreSQL bascule sa session en lecture seule et le pilote le relit ; SQL
// Server n'offre rien d'équivalent — ApplicationIntent=ReadOnly ne vaut que
// pour un réplica de groupe de disponibilité, et il est ignoré ailleurs.
// L'invariant repose donc sur le texte des requêtes, et un invariant qui repose
// sur la discipline doit être vérifié.
func TestAucuneRequeteNEcrit(t *testing.T) {
	t.Parallel()

	for nom, requete := range requetes {
		if verbe := verbeDEcriture.FindString(requete); verbe != "" {
			t.Errorf("requete %s : verbe d'ecriture %q, la lecture seule ne tient plus", nom, verbe)
		}
	}
}

// TestVerbeDEcritureNeRatePasUneEcriture est le contrôle négatif du précédent :
// passer du mot à la sous-chaîne ne doit rien laisser passer d'une écriture,
// ni refuser un nom du catalogue.
func TestVerbeDEcritureNeRatePasUneEcriture(t *testing.T) {
	t.Parallel()

	cas := []struct {
		requete string
		ecrit   bool
	}{
		{"DELETE FROM t", true},
		{"update t set a = 1", true},
		{"EXEC sp_addextendedproperty", true},
		{"EXECUTE dbo.p", true},
		{"select name, sp_x from t", true},
		{"SELECT a INTO copie FROM t", true},
		{"SELECT fk.delete_referential_action_desc, fk.update_referential_action_desc FROM sys.foreign_keys fk", false},
		{"CROSS APPLY sys.dm_exec_describe_first_result_set(N'x', NULL, 0)", false},
		{"SELECT created_at, is_updated FROM t", false},
	}
	for _, c := range cas {
		if ecrit := verbeDEcriture.MatchString(c.requete); ecrit != c.ecrit {
			t.Errorf("%q : ecriture %v, attendu %v", c.requete, ecrit, c.ecrit)
		}
	}
}

// TestAucuneRequeteNEstUnGabarit vérifie qu'aucune requête n'attend un
// fmt.Sprintf : SQL Server n'ayant pas de tableau paramétrable, la tentation
// est de composer une liste IN à partir de noms venus de l'appelant. Le filtre
// par schéma se fait donc en Go, sur le résultat.
//
// Les concaténations SQL, elles, sont permises : reconstruire « nvarchar(30) »
// depuis sys.types n'assemble que des valeurs du catalogue.
func TestAucuneRequeteNEstUnGabarit(t *testing.T) {
	t.Parallel()

	for nom, requete := range requetes {
		for _, verbe := range []string{"%s", "%d", "%q", "%v"} {
			if strings.Contains(requete, verbe) {
				t.Errorf("requete %s : gabarit %q, l'identifiant doit passer en parametre", nom, verbe)
			}
		}
	}
}

// TestEnsembleRetientLesSchemasDemandes : sans schéma demandé, tout est retenu ;
// avec, seuls ceux-là.
func TestEnsembleRetientLesSchemasDemandes(t *testing.T) {
	t.Parallel()

	if ensemble(nil) != nil {
		t.Error("aucun schema demande : tout doit etre retenu")
	}
	if ensemble([]string{}) != nil {
		t.Error("liste vide : tout doit etre retenu")
	}

	e := ensemble([]string{"ventes", "dbo"})
	if !e["ventes"] || !e["dbo"] {
		t.Error("schema demande non retenu")
	}
	if e["autre"] {
		t.Error("schema non demande retenu")
	}
}
