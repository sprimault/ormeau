// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package introspection

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

// passeJournalisee rend une passe qui inscrit sa lecture au journal, puis rend
// l'erreur donnée.
func passeJournalisee(etape string, journal *[]string, err error) Passe {
	return Passe{Etape: etape, Lire: func(context.Context) error {
		*journal = append(*journal, "lecture "+etape)
		return err
	}}
}

// suiviJournalise inscrit chaque avancement au même journal que les lectures,
// ce qui rend leur ordre relatif vérifiable.
func suiviJournalise(journal *[]string) Suivi {
	return func(a Avancement) {
		*journal = append(*journal, fmt.Sprintf("signal %s %d/%d", a.Etape, a.Rang, a.Total))
	}
}

// Chaque passe est signalée avant d'être lue : l'interface doit afficher
// « colonnes » pendant que la requête des colonnes tourne, pas après.
func TestDeroulerSignaleAvantDeLire(t *testing.T) {
	t.Parallel()

	var journal []string
	ctx := AvecSuivi(context.Background(), suiviJournalise(&journal))

	err := Derouler(ctx,
		passeJournalisee(EtapeTables, &journal, nil),
		passeJournalisee(EtapeColonnes, &journal, nil),
		passeJournalisee(EtapeIndex, &journal, nil),
	)
	if err != nil {
		t.Fatalf("déroulement : %v", err)
	}

	attendu := []string{
		"signal tables 1/3", "lecture tables",
		"signal colonnes 2/3", "lecture colonnes",
		"signal index 3/3", "lecture index",
	}
	if !reflect.DeepEqual(journal, attendu) {
		t.Errorf("journal %q, attendu %q", journal, attendu)
	}
}

// Sans suivi posé, les passes se lisent quand même : la ligne de commande n'en
// pose aucun.
func TestDeroulerSansSuivi(t *testing.T) {
	t.Parallel()

	var journal []string
	if err := Derouler(context.Background(), passeJournalisee(EtapeTables, &journal, nil)); err != nil {
		t.Fatalf("déroulement : %v", err)
	}
	if !reflect.DeepEqual(journal, []string{"lecture tables"}) {
		t.Errorf("journal %q", journal)
	}
}

// Une passe en échec arrête le déroulement : les suivantes compléteraient un
// calque déjà faux, et l'erreur remonte telle quelle.
func TestDeroulerArreteALaPremiereErreur(t *testing.T) {
	t.Parallel()

	echec := errors.New("lecture des colonnes")
	var journal []string
	ctx := AvecSuivi(context.Background(), suiviJournalise(&journal))

	err := Derouler(ctx,
		passeJournalisee(EtapeTables, &journal, nil),
		passeJournalisee(EtapeColonnes, &journal, echec),
		passeJournalisee(EtapeIndex, &journal, nil),
	)
	if !errors.Is(err, echec) {
		t.Fatalf("erreur %v, attendue %v", err, echec)
	}

	attendu := []string{"signal tables 1/3", "lecture tables", "signal colonnes 2/3", "lecture colonnes"}
	if !reflect.DeepEqual(journal, attendu) {
		t.Errorf("journal %q, attendu %q", journal, attendu)
	}
}

// Une annulation arrivée pendant une passe arrête avant la suivante, y compris
// quand la passe l'a ignorée et rendu nil.
func TestDeroulerArreteSurAnnulation(t *testing.T) {
	t.Parallel()

	ctx, annuler := context.WithCancel(context.Background())
	defer annuler()

	var journal []string
	ctx = AvecSuivi(ctx, suiviJournalise(&journal))

	sourde := Passe{Etape: EtapeTables, Lire: func(context.Context) error {
		journal = append(journal, "lecture tables")
		annuler()
		return nil
	}}

	err := Derouler(ctx, sourde, passeJournalisee(EtapeColonnes, &journal, nil))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("erreur %v, attendu une annulation", err)
	}

	attendu := []string{"signal tables 1/2", "lecture tables"}
	if !reflect.DeepEqual(journal, attendu) {
		t.Errorf("journal %q, attendu %q", journal, attendu)
	}
}
