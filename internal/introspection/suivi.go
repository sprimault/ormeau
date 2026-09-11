// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package introspection

import "context"

// Codes des passes d'extraction.
//
// Vocabulaire partagé par tous les pilotes, et stable : l'interface traduit le
// libellé depuis le code, comme pour les avertissements. Un dialecte ne déroule
// que les passes qui existent chez lui — MySQL n'a pas de séquence — et le total
// suit. Une passe propre à un dialecte appelle un code ici et sa traduction dans
// les deux langues, pas une chaîne inventée dans le pilote.
const (
	EtapeSource        = "source"
	EtapeTables        = "tables"
	EtapeColonnes      = "colonnes"
	EtapeContraintes   = "contraintes"
	EtapeIndex         = "index"
	EtapeSequences     = "sequences"
	EtapeTypesEnumeres = "types_enumeres"
	EtapeVues          = "vues"
)

// Suivi reçoit l'avancement d'une extraction. Il est appelé depuis la goroutine
// qui extrait, juste avant chaque passe : un suivi qui bloque retarde
// l'extraction d'autant.
type Suivi func(Avancement)

// cleSuivi est la clé de contexte du suivi, non exportée pour qu'aucun autre
// paquet ne puisse l'écraser.
type cleSuivi struct{}

// AvecSuivi rend un contexte qui porte le suivi jusqu'au pilote.
//
// Par le contexte plutôt que par un paramètre d'Extraire, comme httptrace le
// fait pour une requête HTTP : l'avancement est une observation facultative. Le
// mettre dans la signature obligerait la ligne de commande et chaque test à
// passer un suivi dont ils n'ont que faire.
func AvecSuivi(ctx context.Context, suivi Suivi) context.Context {
	return context.WithValue(ctx, cleSuivi{}, suivi)
}

// Passe est une lecture de catalogue, désignée par son code d'étape.
type Passe struct {
	Etape string
	Lire  func(context.Context) error
}

// Derouler exécute les passes dans l'ordre et signale chacune avant de la lire.
//
// Chaque pilote déroule son extraction par ici : le rang et le total s'y
// comptent seuls, et un dialecte ajouté n'a rien à signaler lui-même, donc rien
// à oublier ni à mal compter. Ce n'est pas une couche SQL partagée — rien ici ne
// sait ce qu'une passe interroge.
//
// Le contexte est relu avant chaque passe : une annulation arrivée pendant la
// précédente arrête là, même si le pilote n'a pas su interrompre sa requête.
func Derouler(ctx context.Context, passes ...Passe) error {
	suivi, _ := ctx.Value(cleSuivi{}).(Suivi)
	for i, passe := range passes {
		if err := ctx.Err(); err != nil {
			return err
		}
		if suivi != nil {
			suivi(Avancement{Etape: passe.Etape, Rang: i + 1, Total: len(passes)})
		}
		if err := passe.Lire(ctx); err != nil {
			return err
		}
	}
	return nil
}
