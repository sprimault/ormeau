// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import "testing"

// TestSingulariserFrancais couvre les règles françaises.
func TestSingulariserFrancais(t *testing.T) {
	t.Parallel()

	cas := []struct {
		entree, attendu string
		tranche         bool
	}{
		{"clients", "client", true},
		{"commandes", "commande", true},
		{"adresses", "adresse", true},
		{"analyses", "analyse", true},
		{"fiches", "fiche", true},
		{"journaux", "journal", true},
		{"chevaux", "cheval", true},
		{"bureaux", "bureau", true},
		{"tableaux", "tableau", true},
		{"choux", "chou", true},
		{"yeux", "oeil", true},

		// Déjà au singulier : rien à faire, et pas de règle appliquée.
		{"client", "client", false},
		{"adresse", "adresse", false},

		// Singuliers en -s : les abîmer donnerait Pay, Prix devenu Pri, Temp.
		{"pays", "pays", false},
		{"prix", "prix", false},
		{"temps", "temps", false},
		{"souris", "souris", false},
		{"cours", "cours", false},
	}

	for _, c := range cas {
		t.Run(c.entree, func(t *testing.T) {
			t.Parallel()

			r := singulariser(c.entree)
			if r.nom != c.attendu {
				t.Errorf("singulariser(%q) = %q, attendu %q", c.entree, r.nom, c.attendu)
			}
			if r.tranche != c.tranche {
				t.Errorf("singulariser(%q) tranche = %v, attendu %v", c.entree, r.tranche, c.tranche)
			}
		})
	}
}

// TestSingulariserAnglais couvre les règles anglaises.
func TestSingulariserAnglais(t *testing.T) {
	t.Parallel()

	cas := []struct {
		entree, attendu string
		tranche         bool
	}{
		{"orders", "order", true},
		{"categories", "category", true},
		{"companies", "company", true},
		{"boxes", "box", true},
		{"dishes", "dish", true},
		{"people", "person", true},
		{"children", "child", true},
		{"indices", "index", true},

		// Pluriels en -es traités comme irréguliers : une règle générale
		// abîmerait les mots français de même terminaison.
		{"buses", "bus", true},
		{"batches", "batch", true},
		{"statuses", "status", true},
		{"processes", "process", true},

		// Le y ne revient que derrière une consonne.
		{"boys", "boy", true},

		// Singuliers qui ressemblent à des pluriels.
		{"status", "status", false},
		{"bus", "bus", false},
		{"series", "series", false},
		{"news", "news", false},

		// Un mot en -ss n'est jamais un pluriel.
		{"address", "address", false},
		{"process", "process", false},
		{"class", "class", false},
	}

	for _, c := range cas {
		t.Run(c.entree, func(t *testing.T) {
			t.Parallel()

			r := singulariser(c.entree)
			if r.nom != c.attendu {
				t.Errorf("singulariser(%q) = %q, attendu %q", c.entree, r.nom, c.attendu)
			}
			if r.tranche != c.tranche {
				t.Errorf("singulariser(%q) tranche = %v, attendu %v", c.entree, r.tranche, c.tranche)
			}
		})
	}
}

// TestSingulariserArbitreLesConflits fige les arbitrages entre les deux langues.
//
// Ces terminaisons appartiennent aux deux, avec des singuliers différents. Le
// français l'emporte parce qu'il est de très loin le plus fréquent sur des noms
// de tables — adresses, fiches, analyses contre buses et batches —, et les
// quelques mots anglais concernés sont traités un par un.
//
// Ce test existe pour que le jour où quelqu'un rajoutera une règle en -ses ou
// en -ches, il voie tout de suite ce qu'elle casse.
func TestSingulariserArbitreLesConflits(t *testing.T) {
	t.Parallel()

	for entree, attendu := range map[string]string{
		"adresses": "adresse", // et non adress
		"analyses": "analyse", // et non analys
		"bases":    "base",    // et non bas
		"caches":   "cache",   // et non cach
		"taches":   "tache",   // et non tach
		"fiches":   "fiche",   // et non fich
		"phrases":  "phrase",
		"reponses": "reponse",
	} {
		if r := singulariser(entree); r.nom != attendu {
			t.Errorf("singulariser(%q) = %q, attendu %q", entree, r.nom, attendu)
		}
	}
}

// TestSingulariserPreserveLaCasse vérifie qu'un nom en capitales le reste.
//
// Une base SQL Server nomme ses tables en majuscules. Rendre JOURNal au lieu de
// JOURNAL se verrait dans le calque, mais pas forcément avant la génération.
func TestSingulariserPreserveLaCasse(t *testing.T) {
	t.Parallel()

	for entree, attendu := range map[string]string{
		"JOURNAUX":   "JOURNAL",
		"CLIENTS":    "CLIENT",
		"PEOPLE":     "PERSON",
		"People":     "Person",
		"Categories": "Category",
	} {
		if r := singulariser(entree); r.nom != attendu {
			t.Errorf("singulariser(%q) = %q, attendu %q", entree, r.nom, attendu)
		}
	}
}

// TestSingulariserComposes vérifie qu'un nom composé ne singularise que sa tête.
//
// En français, tous les éléments s'accordent : boites_aux_lettres a trois mots
// au pluriel pour un seul objet. Rien ne dit lesquels portent le sens, et
// singulariser le dernier donnerait BoiteAuxLettre.
func TestSingulariserComposes(t *testing.T) {
	t.Parallel()

	for entree, attendu := range map[string]string{
		"boites_aux_lettres": "boite_aux_lettres",
		"lignes_commande":    "ligne_commande",
		"order_items":        "order_items",
	} {
		if r := singulariser(entree); r.nom != attendu {
			t.Errorf("singulariser(%q) = %q, attendu %q", entree, r.nom, attendu)
		}
	}
}

// TestSingulariserSignaleLAmbiguite couvre le cas qu'aucune règle ne peut
// trancher.
//
// categories est le pluriel de category en anglais et de catégorie en français,
// une fois l'accent perdu — ce qu'une base fait très souvent. Les deux
// singuliers diffèrent, et rien dans un catalogue ne dit quelle langue lire.
// L'inférence applique la règle anglaise, la plus productive sur ce suffixe, et
// marque le résultat comme discutable.
func TestSingulariserSignaleLAmbiguite(t *testing.T) {
	t.Parallel()

	for _, mot := range []string{"categories", "companies", "maladies", "technologies"} {
		if r := singulariser(mot); !r.ambigu {
			t.Errorf("singulariser(%q) = %q sans signaler l'ambiguite", mot, r.nom)
		}
	}

	// Ceux-là ne le sont pas : règle française sans équivalent anglais,
	// irrégulier explicite, ou terminaison qu'une seule langue produit.
	for _, mot := range []string{"journaux", "bureaux", "clients", "boxes", "sorties", "cookies"} {
		if r := singulariser(mot); r.ambigu {
			t.Errorf("singulariser(%q) signale une ambiguite qui n'existe pas", mot)
		}
	}
}

// TestSingulariserVoyelleAvantIes vérifie que -ies précédé d'une voyelle échappe
// à la règle anglaise.
//
// L'anglais ne produit -ies que derrière une consonne : boy fait boys. Un -ies
// derrière une voyelle est donc du français — baies, voies, pluies —, que la
// règle générale traite correctement. Sans cette garde, baies donnerait By.
func TestSingulariserVoyelleAvantIes(t *testing.T) {
	t.Parallel()

	for entree, attendu := range map[string]string{
		"baies":  "baie",
		"voies":  "voie",
		"pluies": "pluie",
		"joies":  "joie",
	} {
		r := singulariser(entree)
		if r.nom != attendu {
			t.Errorf("singulariser(%q) = %q, attendu %q", entree, r.nom, attendu)
		}
		if r.ambigu {
			t.Errorf("singulariser(%q) signale une ambiguite : une voyelle exclut l'anglais", entree)
		}
	}
}

// TestSingulariserAccentLeveLAmbiguite vérifie que l'accent tranche la langue.
//
// C'est le seul indice qu'un nom de table donne gratuitement : l'anglais ne
// porte pas d'accent, donc catégories est français et vaut catégorie, là où
// categories sans accent reste indécidable.
func TestSingulariserAccentLeveLAmbiguite(t *testing.T) {
	t.Parallel()

	for entree, attendu := range map[string]string{
		"catégories": "catégorie",
		"économies":  "économie",
		"séries":     "série",
		"règles":     "règle",
		"libellés":   "libellé",
	} {
		r := singulariser(entree)
		if r.nom != attendu {
			t.Errorf("singulariser(%q) = %q, attendu %q", entree, r.nom, attendu)
		}
		if r.ambigu {
			t.Errorf("singulariser(%q) signale une ambiguite : l'accent exclut l'anglais", entree)
		}
	}

	// Sans accent, le même mot redevient indécidable.
	if r := singulariser("categories"); !r.ambigu {
		t.Error("singulariser(categories) devrait signaler l'ambiguite")
	}
}

// TestSingulariserIdentifiantsAEspaces couvre les noms de tables contenant des
// espaces.
//
// SQL Server les accepte entre crochets, et une base reprise en contient
// parfois. Le séparateur doit valoir le souligné, sinon c'est la fin de la
// chaîne qui est examinée et Commandes Clients devient Commandes Client —
// singularisé du mauvais côté.
func TestSingulariserIdentifiantsAEspaces(t *testing.T) {
	t.Parallel()

	for entree, attendu := range map[string]string{
		"Commandes Clients":   "Commande Clients",
		"Lignes de commande":  "Ligne de commande",
		"catégories des noms": "catégorie des noms",
		"Etat Civil":          "Etat Civil",
		"Codes-Postaux":       "Code-Postaux",
	} {
		if r := singulariser(entree); r.nom != attendu {
			t.Errorf("singulariser(%q) = %q, attendu %q", entree, r.nom, attendu)
		}
	}
}

// TestNommageIdentifiantsAEspaces vérifie la chaîne complète sur ces mêmes
// noms : ce qui compte au bout est le nom de classe.
func TestNommageIdentifiantsAEspaces(t *testing.T) {
	t.Parallel()

	for entree, attendu := range map[string]string{
		"Commandes Clients":   "CommandeClients",
		"Lignes de commande":  "LigneDeCommande",
		"catégories des noms": "CatégorieDesNoms",
	} {
		if obtenu := pascalCase(singulariser(entree).nom); obtenu != attendu {
			t.Errorf("nom de classe de %q = %q, attendu %q", entree, obtenu, attendu)
		}
	}

	for entree, attendu := range map[string]string{
		"Nom Client":       "nomClient",
		"CA TTC":           "caTtc",
		"Date de creation": "dateDeCreation",
	} {
		if obtenu := camelCase(entree); obtenu != attendu {
			t.Errorf("nom de propriete de %q = %q, attendu %q", entree, obtenu, attendu)
		}
	}
}

// TestCouperNeVidePasLeMot vérifie la garde du cas dégénéré.
//
// Un identifiant réduit à une lettre existe — une colonne s, une table x — et
// le couper laisserait une classe sans nom.
func TestCouperNeVidePasLeMot(t *testing.T) {
	t.Parallel()

	if obtenu := couper("s", 1); obtenu != "s" {
		t.Errorf("couper(s, 1) = %q, attendu s", obtenu)
	}
	if obtenu := couper("ab", 3); obtenu != "ab" {
		t.Errorf("couper(ab, 3) = %q, attendu ab", obtenu)
	}
}

// TestRessembleAUnPluriel couvre le déclencheur de l'avertissement.
func TestRessembleAUnPluriel(t *testing.T) {
	t.Parallel()

	for nom, attendu := range map[string]bool{
		"pays":    true,
		"status":  true,
		"prix":    true,
		"client":  false,
		"adresse": false,
		"address": false, // -ss n'est pas un pluriel
	} {
		if obtenu := ressembleAUnPluriel(nom); obtenu != attendu {
			t.Errorf("ressembleAUnPluriel(%q) = %v, attendu %v", nom, obtenu, attendu)
		}
	}
}

// TestSingulariserChaineVide vérifie le cas dégénéré.
func TestSingulariserChaineVide(t *testing.T) {
	t.Parallel()

	if r := singulariser(""); r.nom != "" || r.tranche {
		t.Errorf("singulariser(\"\") = %q/%v, attendu \"\"/false", r.nom, r.tranche)
	}
}
