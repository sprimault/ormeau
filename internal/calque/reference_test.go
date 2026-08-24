// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package calque

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

// majAttendus réécrit les fichiers de référence au lieu de les comparer. Jamais
// automatique : un attendu régénéré sans être relu ne teste plus rien.
//
//	go test ./internal/calque/ -maj-attendus
var majAttendus = flag.Bool("maj-attendus", false, "réécrit les fichiers de référence")

// Le calque de référence, seul exemple versionné : il vient de tests/ddl/, pas
// d'une base réelle.
const cheminReference = "../../tests/reference/exemple.calque.json"

// Ce fichier est le pont entre les deux moitiés du projet : la CI le valide
// contre schemas/calque-physique.v1.json, et le lecteur PHP le lira. Un champ
// ajouté aux structures Go sans l'être au schéma se voit ici, pas six mois
// plus tard chez un intégrateur.
func TestCalqueDeReferenceSurDisque(t *testing.T) {
	t.Parallel()

	attendu := physiqueDeReference()
	// L'horodatage est figé : il est exclu de l'empreinte, mais pas du
	// document, et un fichier versionné ne doit pas changer à chaque exécution.
	attendu.Source.ExtraitLe = "2026-01-01T00:00:00Z"
	attendu.Trier()

	empreinte, err := attendu.CalculerEmpreinte()
	if err != nil {
		t.Fatalf("calcul d'empreinte : %v", err)
	}
	attendu.Source.Empreinte = empreinte

	donnees, err := Serialiser(attendu)
	if err != nil {
		t.Fatalf("sérialisation : %v", err)
	}

	if *majAttendus {
		if err := os.MkdirAll(filepath.Dir(cheminReference), 0o750); err != nil {
			t.Fatalf("création du répertoire : %v", err)
		}
		if err := os.WriteFile(cheminReference, donnees, 0o600); err != nil {
			t.Fatalf("écriture : %v", err)
		}
		t.Logf("%s réécrit — à relire avant de commiter", cheminReference)
		return
	}

	surDisque, err := os.ReadFile(cheminReference)
	if err != nil {
		t.Fatalf("lecture de la référence : %v (relancer avec -maj-attendus)", err)
	}
	if string(surDisque) != string(donnees) {
		t.Error("le calque de référence a dérivé des structures Go ; " +
			"relire le diff, puis relancer avec -maj-attendus")
	}
}

// Le fichier versionné doit se relire et passer sa propre validation : c'est ce
// qui en fait un cas de test exploitable et non un simple exemple.
func TestCalqueDeReferenceRelisible(t *testing.T) {
	t.Parallel()

	// Pendant une régénération, le fichier est en train d'être réécrit par le
	// test voisin : le relire n'aurait aucun sens.
	if *majAttendus {
		t.Skip("régénération des attendus en cours")
	}

	p, err := LirePhysique(cheminReference)
	if err != nil {
		t.Fatalf("relecture : %v", err)
	}
	if a := p.Valider(); len(a) != 0 {
		t.Errorf("calque de référence invalide : %+v", a)
	}
	if err := p.VerifierEmpreinte(); err != nil {
		t.Errorf("empreinte de la référence : %v", err)
	}
}
