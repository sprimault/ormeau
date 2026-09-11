// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/sprimault/ormeau/internal/calque"
	"github.com/sprimault/ormeau/internal/inference"
)

// empreinteInventee ne correspond à aucun calque.
const empreinteInventee = "sha256:0000000000000000000000000000000000000000000000000000000000000000"

// poster envoie un corps JSON à un point d'entrée.
func poster(s *serveur, routeur http.Handler, chemin string, corps any) *httptest.ResponseRecorder {
	donnees, _ := json.Marshal(corps)
	w := httptest.NewRecorder()
	routeur.ServeHTTP(w, requeteAPI(s, http.MethodPost, chemin, bytes.NewReader(donnees)))
	return w
}

// lire interroge un point d'entrée en lecture.
func lire(s *serveur, routeur http.Handler, chemin string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	routeur.ServeHTTP(w, requeteAPI(s, http.MethodGet, chemin, nil))
	return w
}

// decoderReponse lit le corps d'une réponse, ou arrête le test.
func decoderReponse[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()

	var v T
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatalf("réponse illisible : %v\n%s", err, w.Body.String())
	}
	return v
}

// attendreStatut arrête le test si le statut n'est pas celui attendu.
func attendreStatut(t *testing.T, w *httptest.ResponseRecorder, attendu int) {
	t.Helper()

	if w.Code != attendu {
		t.Fatalf("code %d, attendu %d : %s", w.Code, attendu, w.Body.String())
	}
}

// attendreRefus vérifie un conflit et son code.
func attendreRefus(t *testing.T, w *httptest.ResponseRecorder, code CodeRefus) {
	t.Helper()

	attendreStatut(t, w, http.StatusConflict)
	if refus := decoderReponse[ReponseErreur](t, w); refus.Code != code || refus.Erreur == "" {
		t.Errorf("refus %+v, attendu le code %q avec un message", refus, code)
	}
}

// empreinteDuCalque rend l'empreinte du calque écrit pour une base.
func empreinteDuCalque(t *testing.T, s *serveur, base string) string {
	t.Helper()

	p, err := calque.LirePhysique(filepath.Join(s.repertoire, base+".calque.json"))
	if err != nil {
		t.Fatalf("lecture du calque : %v", err)
	}
	return p.Source.Empreinte
}

// TestInferenceHorsLigneRendUnResume couvre le cas nominal, sans aucune
// connexion ouverte : l'arbitrage n'a besoin que du calque.
func TestInferenceHorsLigneRendUnResume(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	ecrireCalque(t, s, "gescom", physiqueDeTest())

	w := poster(s, routeur, "/api/inference", RequeteInference{
		Base:      "gescom",
		Decisions: inference.Decisions{Renommages: map[string]string{"public.client": "Acheteur"}},
	})
	attendreStatut(t, w, http.StatusOK)
	reponse := decoderReponse[ReponseInference](t, w)

	if reponse.EmpreintePhysique != empreinteDuCalque(t, s, "gescom") {
		t.Errorf("empreinte %q, attendue celle du calque", reponse.EmpreintePhysique)
	}
	if len(reponse.Entites) != 1 || reponse.Entites[0].Nom != "Acheteur" || reponse.Entites[0].NbProprietes != 1 {
		t.Errorf("entités %+v", reponse.Entites)
	}
	if !slices.Contains(reponse.TypesDoctrine, "boolean") || !slices.IsSorted(reponse.TypesDoctrine) {
		t.Errorf("types Doctrine %v, attendus triés et complets", reponse.TypesDoctrine)
	}
	if len(reponse.Avertissements) == 0 {
		t.Error("une table sans clé primaire devrait produire un avertissement")
	}
	if strings.Contains(w.Body.String(), `"proprietes"`) {
		t.Error("le détail des entités part avec l'inférence : sur une grande base, plusieurs mégaoctets à chaque frappe")
	}
	if len(s.registre.parID) != 0 {
		t.Error("une connexion a été ouverte pour arbitrer")
	}
}

// TestInferenceRattacheLesEnumerationsAleursColonnes vérifie ce que l'écran doit
// savoir pour nommer des cas : la colonne qui porte l'énumération.
func TestInferenceRattacheLesEnumerationsAleursColonnes(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	reference, err := os.ReadFile("../../tests/reference/inference/enumerations/physique.json")
	if err != nil {
		t.Fatalf("lecture du cas de référence : %v", err)
	}
	if err := os.WriteFile(filepath.Join(s.repertoire, "gescom.calque.json"), reference, 0o600); err != nil {
		t.Fatalf("écriture : %v", err)
	}

	w := poster(s, routeur, "/api/inference", RequeteInference{Base: "gescom"})
	attendreStatut(t, w, http.StatusOK)
	reponse := decoderReponse[ReponseInference](t, w)

	if len(reponse.Enumerations) == 0 {
		t.Fatal("aucune énumération sur le cas de référence des énumérations")
	}
	for _, e := range reponse.Enumerations {
		if len(e.Colonnes) == 0 || strings.Count(e.Colonnes[0], ".") < 2 {
			t.Errorf("énumération %s : colonnes %q, attendu schema.table.colonne", e.Nom, e.Colonnes)
		}
	}
}

// TestInferenceGardeLeCalqueJuge couvre la réextraction pendant l'arbitrage :
// le calque change sans que sa date ni sa taille ne le disent forcément.
func TestInferenceGardeLeCalqueJuge(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	ecrireCalque(t, s, "gescom", physiqueDeTest())

	premiere := decoderReponse[ReponseInference](t, poster(s, routeur, "/api/inference", RequeteInference{Base: "gescom"}))

	p := physiqueDeTest()
	p.Tables = append(p.Tables, calque.Table{
		Nom: "fournisseur", Schema: "public",
		Colonnes: []calque.Colonne{{Nom: "id", Position: 1, TypeBrut: "integer", TypeNormalise: calque.TypeEntier}},
	})
	ecrireCalque(t, s, "gescom", p)

	attendreRefus(t, poster(s, routeur, "/api/inference", RequeteInference{
		Base: "gescom", EmpreintePhysique: premiere.EmpreintePhysique,
	}), CodeCalqueModifie)

	rechargee := decoderReponse[ReponseInference](t, poster(s, routeur, "/api/inference", RequeteInference{Base: "gescom"}))
	if rechargee.EmpreintePhysique == premiere.EmpreintePhysique || len(rechargee.Entites) != 2 {
		t.Errorf("rechargement : empreinte %q, %d entité(s)", rechargee.EmpreintePhysique, len(rechargee.Entites))
	}
}

// TestInferenceRefuseCeQuiNeDesignePasUnCalque couvre le nom de base qui
// sortirait du répertoire de travail, et le calque absent.
func TestInferenceRefuseCeQuiNeDesignePasUnCalque(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	attendreStatut(t, poster(s, routeur, "/api/inference", RequeteInference{Base: "../gescom"}), http.StatusBadRequest)
	w := poster(s, routeur, "/api/inference", RequeteInference{Base: "gescom"})
	attendreStatut(t, w, http.StatusNotFound)
	if !strings.Contains(w.Body.String(), "gescom.calque.json") {
		t.Errorf("le message ne nomme pas le fichier absent : %s", w.Body.String())
	}
}

// TestEntiteRendLeDetailEtSaTable vérifie la vue côte à côte, et le cas d'une
// table qui ne produit pas d'entité.
func TestEntiteRendLeDetailEtSaTable(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	ecrireCalque(t, s, "gescom", physiqueDeTest())

	w := poster(s, routeur, "/api/inference/entite", RequeteEntite{Base: "gescom", Schema: "public", Table: "client"})
	attendreStatut(t, w, http.StatusOK)
	reponse := decoderReponse[ReponseEntite](t, w)
	if reponse.Entite == nil || len(reponse.Entite.Proprietes) != 1 || reponse.TablePhysique.Nom != "client" {
		t.Errorf("détail %+v", reponse)
	}

	ignoree := decoderReponse[ReponseEntite](t, poster(s, routeur, "/api/inference/entite", RequeteEntite{
		Base: "gescom", Schema: "public", Table: "client",
		Decisions: inference.Decisions{TablesIgnorees: []string{"public.client"}},
	}))
	if ignoree.Entite != nil || ignoree.TablePhysique.Nom != "client" {
		t.Errorf("table ignorée : %+v, attendu la table seule", ignoree)
	}

	attendreStatut(t, poster(s, routeur, "/api/inference/entite", RequeteEntite{
		Base: "gescom", Schema: "public", Table: "absente",
	}), http.StatusNotFound)
}

// TestDecisionsAbsentes couvre la base jamais arbitrée : un fichier à créer,
// rien de manuel.
func TestDecisionsAbsentes(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)

	w := lire(s, routeur, "/api/decisions?base=gescom")
	attendreStatut(t, w, http.StatusOK)
	if reponse := decoderReponse[ReponseDecisions](t, w); reponse.Existe || reponse.Manuel || reponse.EmpreinteFichier != "" {
		t.Errorf("fichier absent décrit comme %+v", reponse)
	}
	attendreStatut(t, lire(s, routeur, "/api/decisions?base=..%2Fgescom"), http.StatusBadRequest)
}

// TestDecisionsEcritesPuisRelues couvre le cycle d'un enregistrement : le
// fichier écrit se relit avec les mêmes décisions, et n'est pas manuel.
func TestDecisionsEcritesPuisRelues(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	ecrireCalque(t, s, "gescom", physiqueDeTest())
	decisions := inference.Decisions{Renommages: map[string]string{"public.client": "Acheteur"}}

	w := poster(s, routeur, "/api/decisions", RequeteEcritureDecisions{
		Base: "gescom", Decisions: decisions, EmpreintePhysique: empreinteDuCalque(t, s, "gescom"),
	})
	attendreStatut(t, w, http.StatusOK)
	ecrite := decoderReponse[ReponseEcritureDecisions](t, w)

	relue := decoderReponse[ReponseDecisions](t, lire(s, routeur, "/api/decisions?base=gescom"))
	if !relue.Existe || relue.Manuel || relue.EmpreinteFichier != ecrite.EmpreinteFichier {
		t.Errorf("fichier relu %+v, empreinte écrite %q", relue, ecrite.EmpreinteFichier)
	}
	if relue.Decisions.Renommages["public.client"] != "Acheteur" {
		t.Errorf("décisions relues %+v", relue.Decisions)
	}
}

// TestDecisionsRefuseUnFichierModifieSurDisque couvre le fichier retouché dans
// un éditeur pendant que l'écran était ouvert.
func TestDecisionsRefuseUnFichierModifieSurDisque(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	ecrireCalque(t, s, "gescom", physiqueDeTest())
	empreinte := empreinteDuCalque(t, s, "gescom")

	ecrite := decoderReponse[ReponseEcritureDecisions](t, poster(s, routeur, "/api/decisions", RequeteEcritureDecisions{
		Base: "gescom", EmpreintePhysique: empreinte,
	}))

	chemin := filepath.Join(s.repertoire, "gescom.decisions.yaml")
	retouche, err := os.ReadFile(chemin)
	if err != nil {
		t.Fatalf("lecture : %v", err)
	}
	retouche = append(retouche, "\n# retouche dans l'éditeur\n"...)
	if err := os.WriteFile(chemin, retouche, 0o600); err != nil {
		t.Fatalf("écriture : %v", err)
	}

	attendreRefus(t, poster(s, routeur, "/api/decisions", RequeteEcritureDecisions{
		Base:              "gescom",
		Decisions:         inference.Decisions{TablesIgnorees: []string{"public.client"}},
		EmpreintePhysique: empreinte,
		EmpreinteFichier:  ecrite.EmpreinteFichier,
	}), CodeDecisionsModifiees)

	if apres, _ := os.ReadFile(chemin); !bytes.Equal(apres, retouche) {
		t.Error("le fichier retouché a été écrasé malgré le refus")
	}
}

// TestDecisionsProtegeLeContenuManuel couvre le fichier écrit à la main : refus
// sans confirmation, écriture avec.
func TestDecisionsProtegeLeContenuManuel(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	ecrireCalque(t, s, "gescom", physiqueDeTest())
	empreinte := empreinteDuCalque(t, s, "gescom")

	chemin := filepath.Join(s.repertoire, "gescom.decisions.yaml")
	if err := os.WriteFile(chemin, []byte("renommages:\n  public.client: Acheteur  # vu avec le métier\n"), 0o600); err != nil {
		t.Fatalf("écriture : %v", err)
	}

	lu := decoderReponse[ReponseDecisions](t, lire(s, routeur, "/api/decisions?base=gescom"))
	if !lu.Manuel {
		t.Fatal("un fichier écrit à la main n'est pas annoncé manuel")
	}

	ecriture := RequeteEcritureDecisions{
		Base: "gescom", Decisions: lu.Decisions, EmpreintePhysique: empreinte, EmpreinteFichier: lu.EmpreinteFichier,
	}
	attendreRefus(t, poster(s, routeur, "/api/decisions", ecriture), CodeContenuManuel)

	ecriture.EcraserManuel = true
	attendreStatut(t, poster(s, routeur, "/api/decisions", ecriture), http.StatusOK)
	if relu := decoderReponse[ReponseDecisions](t, lire(s, routeur, "/api/decisions?base=gescom")); relu.Manuel {
		t.Error("le fichier réécrit passe encore pour manuel")
	}
}

// TestDecisionsRefuseUnCalqueChangeOuInconnu couvre l'écriture sur un calque
// réécrit entre-temps, ou jamais lu.
func TestDecisionsRefuseUnCalqueChangeOuInconnu(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	ecrireCalque(t, s, "gescom", physiqueDeTest())

	attendreRefus(t, poster(s, routeur, "/api/decisions", RequeteEcritureDecisions{
		Base: "gescom", EmpreintePhysique: empreinteInventee,
	}), CodeCalqueModifie)
	attendreStatut(t, poster(s, routeur, "/api/decisions", RequeteEcritureDecisions{Base: "gescom"}), http.StatusBadRequest)

	if _, err := os.Stat(filepath.Join(s.repertoire, "gescom.decisions.yaml")); err == nil {
		t.Error("un fichier de décisions a été écrit malgré le refus")
	}
}

// TestDecisionsMemeBaseQueLaLigneDeCommande vérifie que l'API et la ligne de
// commande tirent le même nom de base du même fichier : un prérempli écrit par
// ormeau inferer n'est pas manuel pour l'interface, un fichier recopié l'est.
func TestDecisionsMemeBaseQueLaLigneDeCommande(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	p := physiqueDeTest()
	chemin := filepath.Join(s.repertoire, "gescom.decisions.yaml")

	// Ce que fait ormeau inferer au premier passage.
	prerempli := inference.EcrireDecisions(p, &inference.Decisions{}, inference.BaseDesDecisions(chemin))
	if err := os.WriteFile(chemin, prerempli, 0o600); err != nil {
		t.Fatalf("écriture : %v", err)
	}
	if lu := decoderReponse[ReponseDecisions](t, lire(s, routeur, "/api/decisions?base=gescom")); lu.Manuel {
		t.Error("le prérempli de la ligne de commande passe pour manuel dans l'interface")
	}

	recopie := inference.EcrireDecisions(p, &inference.Decisions{TablesIgnorees: []string{"public.client"}}, "paie")
	if err := os.WriteFile(chemin, recopie, 0o600); err != nil {
		t.Fatalf("écriture : %v", err)
	}
	if lu := decoderReponse[ReponseDecisions](t, lire(s, routeur, "/api/decisions?base=gescom")); !lu.Manuel {
		t.Error("un fichier recopié de paie passe pour intact sous gescom")
	}
}

// TestArbitrageEnParallele vérifie que le calque lu, partagé entre les appels,
// supporte plusieurs inférences à la fois : l'écran en relance une à chaque
// modification, et deux onglets peuvent arbitrer ensemble.
func TestArbitrageEnParallele(t *testing.T) {
	t.Parallel()

	s, routeur := serveurDeTest(t)
	ecrireCalque(t, s, "gescom", physiqueDeTest())

	var attente sync.WaitGroup
	for range 8 {
		attente.Add(1)
		go func() {
			defer attente.Done()
			w := poster(s, routeur, "/api/inference", RequeteInference{
				Base:      "gescom",
				Decisions: inference.Decisions{Renommages: map[string]string{"public.client": "Acheteur"}},
			})
			if w.Code != http.StatusOK {
				t.Errorf("code %d : %s", w.Code, w.Body.String())
			}
		}()
	}
	attente.Wait()
}
