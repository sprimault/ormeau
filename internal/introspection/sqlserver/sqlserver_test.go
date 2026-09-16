// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package sqlserver

import (
	"strings"
	"testing"
)

// requetes rassemble ce que le pilote envoie au serveur. Toute requête ajoutée
// se déclare ici, faute de quoi le contrôle de lecture seule ne la voit pas.
var requetes = map[string]string{
	"serveur":    requeteServeur,
	"bases":      requeteBases,
	"schemas":    requeteSchemas,
	"inventaire": requeteInventaire,
	"colonnes":   requeteColonnes,
}

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

	// EXEC et sp_ ferment la porte aux procédures : sp_addextendedproperty
	// écrit, et rien ne distingue de l'extérieur celles qui lisent.
	interdits := []string{
		"insert", "update", "delete", "merge", "truncate",
		"drop", "alter", "create", "grant", "revoke", "exec", "sp_",
	}

	for nom, requete := range requetes {
		bas := strings.ToLower(requete)
		for _, verbe := range interdits {
			if strings.Contains(bas, verbe) {
				t.Errorf("requete %s : verbe d'ecriture %q, la lecture seule ne tient plus", nom, verbe)
			}
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
