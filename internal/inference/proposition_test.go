// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
)

// physiqueDeReference charge un cas pour les tests de propositions.
func physiqueDeReference(t *testing.T, cas string) *calque.Physique {
	t.Helper()

	p, err := calque.LirePhysique(filepath.Join(racineReference, cas, "physique.json"))
	if err != nil {
		t.Fatalf("lecture du physique %s : %v", cas, err)
	}
	return p
}

// TestProposerNAffectePasLeCalque vérifie que suggérer et appliquer restent
// deux choses distinctes.
//
// C'est l'invariant de tout le nommage : les propositions sont un fichier à
// relire, le calque est ce que la génération consomme. Si les deux se
// rejoignaient, l'utilisateur retrouverait des classes qu'il n'a jamais
// demandées.
func TestProposerNAffectePasLeCalque(t *testing.T) {
	t.Parallel()

	p := physiqueDeReference(t, "prefixes")

	logique, _ := Inferer(p, &Decisions{})
	propositions := Proposer(p, &Decisions{})

	nomsDuCalque := map[string]bool{}
	for _, e := range logique.Entites {
		nomsDuCalque[e.Nom] = true
	}

	if !nomsDuCalque["TClients"] {
		t.Error("le calque devrait porter TClients : sans decision, le nom de table est repris tel quel")
	}
	if len(propositions) == 0 {
		t.Fatal("aucune proposition sur une base a prefixe commun")
	}
	for _, prop := range propositions {
		if nomsDuCalque[prop.Nom] {
			t.Errorf("la proposition %s est deja appliquee au calque", prop.Nom)
		}
	}
}

// TestProposerDonneLaRaison vérifie que chaque suggestion dit ce qui a été
// appliqué pour l'obtenir.
//
// Sur quatre cents tables, une liste de renommages sans justification ne se
// relit pas : elle se décommente en bloc, ce qui revient à laisser l'outil
// décider.
func TestProposerDonneLaRaison(t *testing.T) {
	t.Parallel()

	p := physiqueDeReference(t, "prefixes")
	propositions := Proposer(p, &Decisions{})

	attendus := map[string]struct {
		nom    string
		raison string
	}{
		"dbo.T_CLIENTS":   {"Client", "préfixe T_ retiré, singularisé"},
		"dbo.T_COMMANDES": {"Commande", "préfixe T_ retiré, singularisé"},
		"dbo.T_PAYS":      {"Pays", "préfixe T_ retiré"},
	}

	for _, prop := range propositions {
		attendu, connu := attendus[prop.Cible]
		if !connu {
			t.Errorf("proposition inattendue sur %s", prop.Cible)
			continue
		}
		if prop.Nom != attendu.nom {
			t.Errorf("%s : nom %q, attendu %q", prop.Cible, prop.Nom, attendu.nom)
		}
		if prop.Raison != attendu.raison {
			t.Errorf("%s : raison %q, attendue %q", prop.Cible, prop.Raison, attendu.raison)
		}
	}
}

// TestProposerSignaleLAmbiguite vérifie que le cas indécidable porte les deux
// candidats.
//
// categories vaut Category en anglais et Categorie en français. Proposer l'un
// sans nommer l'autre laisserait croire que la question est tranchée.
func TestProposerSignaleLAmbiguite(t *testing.T) {
	t.Parallel()

	p := physiqueDeReference(t, "singularisation")

	for _, prop := range Proposer(p, &Decisions{}) {
		if prop.Cible != "public.categories" {
			continue
		}
		if prop.Nom != "Category" {
			t.Errorf("nom propose %q, attendu Category", prop.Nom)
		}
		if !strings.Contains(prop.Raison, "Categorie") {
			t.Errorf("raison %q : le candidat français devrait y figurer", prop.Raison)
		}
		if prop.Confiance >= 0.6 {
			t.Errorf("confiance %v, attendue basse sur un cas indecidable", prop.Confiance)
		}
		return
	}
	t.Error("aucune proposition sur public.categories")
}

// TestProposerIgnoreCeQuiEstDecide vérifie qu'une table déjà renommée n'est pas
// proposée.
//
// La question est tranchée ; la reposer à chaque écriture du fichier ferait
// remonter du bruit que l'utilisateur a déjà traité.
func TestProposerIgnoreCeQuiEstDecide(t *testing.T) {
	t.Parallel()

	p := physiqueDeReference(t, "prefixes")
	d := &Decisions{Renommages: map[string]string{"dbo.T_CLIENTS": "Acheteur"}}

	for _, prop := range Proposer(p, d) {
		if prop.Cible == "dbo.T_CLIENTS" {
			t.Error("dbo.T_CLIENTS est deja renommee, elle ne devrait pas etre proposee")
		}
	}
}

// TestProposerNeProposePasLIdentique vérifie qu'une table sans rien à changer
// n'apparaît pas.
//
// Une liste où la moitié des lignes redisent le nom courant se relit mal, et
// les quelques-unes qui comptent s'y perdent.
func TestProposerNeProposePasLIdentique(t *testing.T) {
	t.Parallel()

	p := &calque.Physique{
		VersionRI: calque.VersionCourante,
		Tables: []calque.Table{
			{Nom: "client", Schema: "public"},
			{Nom: "adresse", Schema: "public"},
			{Nom: "pays", Schema: "public"},
		},
	}

	if propositions := Proposer(p, &Decisions{}); len(propositions) != 0 {
		t.Errorf("propositions = %#v, attendu aucune sur des noms deja au singulier", propositions)
	}
}

// TestEcrireDecisionsEstEntierementCommente est le test qui garde la promesse
// du fichier.
//
// Une seule ligne active suffirait à appliquer un arbitrage que personne n'a
// relu — exactement ce qu'on a retiré de l'inférence pour le mettre ici.
func TestEcrireDecisionsEstEntierementCommente(t *testing.T) {
	t.Parallel()

	fichier := string(EcrireDecisions(physiqueDeReference(t, "prefixes"), &Decisions{}))

	for numero, ligne := range strings.Split(fichier, "\n") {
		nette := strings.TrimSpace(ligne)
		if nette == "" || strings.HasPrefix(nette, "#") {
			continue
		}
		t.Errorf("ligne %d active : %q", numero+1, ligne)
	}
}

// TestEcrireDecisionsSeRelit vérifie que le fichier reste du YAML valide.
//
// Entièrement commenté, il doit se charger sans erreur et ne rien décider : un
// premier passage écrit ce fichier, et le passage suivant le relit tel quel.
func TestEcrireDecisionsSeRelit(t *testing.T) {
	t.Parallel()

	chemin := filepath.Join(t.TempDir(), "decisions.yaml")
	contenu := EcrireDecisions(physiqueDeReference(t, "prefixes"), &Decisions{})
	if err := os.WriteFile(chemin, contenu, 0o600); err != nil {
		t.Fatalf("ecriture : %v", err)
	}

	d, err := LireDecisions(chemin)
	if err != nil {
		t.Fatalf("relecture : %v", err)
	}
	if d.EspaceDeNoms != "" || len(d.Renommages) != 0 || len(d.PrefixesARetirer) != 0 {
		t.Errorf("le fichier prerempli decide quelque chose : %#v", d)
	}
}

// TestEcrireDecisionsRappelleCeQuiEstDecide vérifie que le fichier réécrit
// reprend les arbitrages déjà faits.
//
// Le fichier est régénéré à chaque passage : s'il perdait ce qui a été décidé,
// il faudrait le fusionner à la main, et personne ne le ferait deux fois.
func TestEcrireDecisionsRappelleCeQuiEstDecide(t *testing.T) {
	t.Parallel()

	d := &Decisions{
		EspaceDeNoms:     `Gescom\Domaine\Entity`,
		PrefixesARetirer: []string{"T_"},
		TablesIgnorees:   []string{"dbo.T_PAYS", "dbo.T_AUDIT"},
		Renommages:       map[string]string{"dbo.T_CLIENTS": "Acheteur"},
		TypesForces:      map[string]string{"dbo.T_CLIENTS.CLI_ACTIF": "boolean"},
	}

	fichier := string(EcrireDecisions(physiqueDeReference(t, "prefixes"), d))

	for _, attendu := range []string{
		`Gescom\Domaine\Entity`,
		"dbo.T_CLIENTS: Acheteur",
		"dbo.T_AUDIT",
		"dbo.T_CLIENTS.CLI_ACTIF: boolean",
		"- T_",
	} {
		if !strings.Contains(fichier, attendu) {
			t.Errorf("le fichier reecrit a perdu %q", attendu)
		}
	}

	// Une décision de préfixe éteint la proposition : la question est réglée.
	if strings.Contains(fichier, "Repéré dans cette base") {
		t.Error("un prefixe decide ne devrait plus etre propose")
	}
}

// TestProposerAccepteDesDecisionsAbsentes vérifie l'appel sans fichier de
// décisions, celui du premier passage.
func TestProposerAccepteDesDecisionsAbsentes(t *testing.T) {
	t.Parallel()

	if len(Proposer(physiqueDeReference(t, "prefixes"), nil)) == 0 {
		t.Error("aucune proposition sur une base a prefixe commun")
	}
}

// TestProposerReconnaitUnRenommageNonQualifie vérifie qu'une décision écrite
// sans schéma éteint aussi la proposition.
//
// Une base à schéma unique est le cas courant, et y écrire le schéma partout
// est du bruit : les deux formes doivent valoir.
func TestProposerReconnaitUnRenommageNonQualifie(t *testing.T) {
	t.Parallel()

	d := &Decisions{Renommages: map[string]string{"T_CLIENTS": "Acheteur"}}

	for _, prop := range Proposer(physiqueDeReference(t, "prefixes"), d) {
		if prop.Cible == "dbo.T_CLIENTS" {
			t.Error("le renommage non qualifie devrait eteindre la proposition")
		}
	}
}

// TestEcrireDecisionsEstDeterministe vérifie que deux écritures du même calque
// donnent les mêmes octets.
//
// Le fichier vit dans un dépôt Git à côté du projet de l'utilisateur : un ordre
// qui change à chaque passage produirait un diff sans qu'aucune base n'ait
// bougé.
func TestEcrireDecisionsEstDeterministe(t *testing.T) {
	t.Parallel()

	p := physiqueDeReference(t, "singularisation")
	d := &Decisions{
		Renommages:  map[string]string{"public.pays": "Pays", "public.boxes": "Box", "public.acces": "Acces"},
		TypesForces: map[string]string{"public.client.id": "bigint", "public.boxes.id": "integer"},
	}

	premier := string(EcrireDecisions(p, d))
	for i := 0; i < 5; i++ {
		if suivant := string(EcrireDecisions(p, d)); suivant != premier {
			t.Fatal("deux ecritures du meme calque different")
		}
	}
}
