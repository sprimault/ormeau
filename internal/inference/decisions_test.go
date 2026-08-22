// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"os"
	"path/filepath"
	"testing"
)

// decisionsCompletes mélange toutes les familles, comme un fichier réel.
const decisionsCompletes = `
espace_de_noms: App\Entity
prefixes_a_retirer:
  - T_
  - tbl_
tables_ignorees:
  - public.migrations
renommages:
  public.T_CLI: Client
types_forces:
  public.T_CLI.actif: boolean
relations_forcees:
  - source: public.T_CMD.cli_id
    cible: public.T_CLI
    genre: plusieurs_vers_un
    nom: client
enumerations:
  - colonne: public.T_CMD.statut
    nom: StatutCommande
    cas:
      O: Ouvert
      F: Ferme
`

// Une clé YAML mal mappée se lit sans erreur et s'ignore en silence.
func TestLireDecisionsFichierComplet(t *testing.T) {
	t.Parallel()

	chemin := filepath.Join(t.TempDir(), "decisions.yaml")
	if err := os.WriteFile(chemin, []byte(decisionsCompletes), 0o600); err != nil {
		t.Fatalf("écriture : %v", err)
	}

	d, err := LireDecisions(chemin)
	if err != nil {
		t.Fatalf("lecture : %v", err)
	}

	if d.EspaceDeNoms != `App\Entity` {
		t.Errorf("espace de noms %q", d.EspaceDeNoms)
	}
	if len(d.PrefixesARetirer) != 2 || d.PrefixesARetirer[0] != "T_" {
		t.Errorf("préfixes %v", d.PrefixesARetirer)
	}
	if len(d.TablesIgnorees) != 1 || d.TablesIgnorees[0] != "public.migrations" {
		t.Errorf("tables ignorées %v", d.TablesIgnorees)
	}
	if d.Renommages["public.T_CLI"] != "Client" {
		t.Errorf("renommages %v", d.Renommages)
	}
	if d.TypesForces["public.T_CLI.actif"] != "boolean" {
		t.Errorf("types forcés %v", d.TypesForces)
	}

	if len(d.RelationsForcees) != 1 {
		t.Fatalf("relations forcées : %d entrées", len(d.RelationsForcees))
	}
	r := d.RelationsForcees[0]
	if r.Source != "public.T_CMD.cli_id" || r.Cible != "public.T_CLI" ||
		r.Genre != "plusieurs_vers_un" || r.Nom != "client" {
		t.Errorf("relation forcée %+v", r)
	}

	if len(d.Enumerations) != 1 {
		t.Fatalf("énumérations : %d entrées", len(d.Enumerations))
	}
	e := d.Enumerations[0]
	if e.Colonne != "public.T_CMD.statut" || e.Nom != "StatutCommande" || e.Cas["O"] != "Ouvert" {
		t.Errorf("énumération %+v", e)
	}
}

// Chemin vide : c'est le premier passage, celui qui produit le prérempli.
func TestLireDecisionsCheminVide(t *testing.T) {
	t.Parallel()

	d, err := LireDecisions("")
	if err != nil {
		t.Fatalf("un chemin vide ne doit pas échouer : %v", err)
	}
	if d == nil {
		t.Fatal("décisions nulles")
	}
	if d.EspaceDeNoms != "" || len(d.PrefixesARetirer) != 0 || len(d.Renommages) != 0 {
		t.Errorf("décisions non vides : %+v", d)
	}
}

// Fichier créé mais pas encore rempli : cas normal avant relecture humaine.
func TestLireDecisionsFichierVide(t *testing.T) {
	t.Parallel()

	chemin := filepath.Join(t.TempDir(), "decisions.yaml")
	if err := os.WriteFile(chemin, nil, 0o600); err != nil {
		t.Fatalf("écriture : %v", err)
	}

	d, err := LireDecisions(chemin)
	if err != nil {
		t.Fatalf("lecture : %v", err)
	}
	if d == nil {
		t.Fatal("décisions nulles")
	}
}

// Un fichier écrit à la main a des fautes de frappe. Les avaler en décisions
// vides ferait croire à des arbitrages appliqués alors qu'ils sont ignorés.
func TestLireDecisionsEchoueProprement(t *testing.T) {
	t.Parallel()

	cas := []struct {
		nom     string
		contenu string
		ecrire  bool
	}{
		{nom: "fichier absent", ecrire: false},
		{nom: "yaml invalide", contenu: "renommages: [ceci: n'est: pas: du yaml", ecrire: true},
		{nom: "type incompatible", contenu: "prefixes_a_retirer: 42", ecrire: true},
		{nom: "racine scalaire", contenu: "une chaîne nue", ecrire: true},
	}

	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			t.Parallel()

			chemin := filepath.Join(t.TempDir(), "decisions.yaml")
			if c.ecrire {
				if err := os.WriteFile(chemin, []byte(c.contenu), 0o600); err != nil {
					t.Fatalf("écriture : %v", err)
				}
			}

			if _, err := LireDecisions(chemin); err == nil {
				t.Error("aucune erreur retournée")
			}
		})
	}
}

// Clés inconnues tolérées : un fichier écrit pour une version ultérieure ne
// bloque pas.
func TestLireDecisionsIgnoreLesClesInconnues(t *testing.T) {
	t.Parallel()

	chemin := filepath.Join(t.TempDir(), "decisions.yaml")
	contenu := "espace_de_noms: App\\Entity\ncle_dune_version_ulterieure: valeur\n"
	if err := os.WriteFile(chemin, []byte(contenu), 0o600); err != nil {
		t.Fatalf("écriture : %v", err)
	}

	d, err := LireDecisions(chemin)
	if err != nil {
		t.Fatalf("lecture : %v", err)
	}
	if d.EspaceDeNoms != `App\Entity` {
		t.Errorf("espace de noms %q", d.EspaceDeNoms)
	}
}
