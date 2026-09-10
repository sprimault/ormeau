// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"errors"
	"sync"

	"github.com/sprimault/ormeau/internal/introspection"
)

// maxConnexions plafonne les connexions simultanées. Chacune est une connexion
// ouverte sur la base d'un client : un formulaire rejoué en boucle ne doit pas
// pouvoir en accumuler des centaines côté serveur.
const maxConnexions = 8

// ErrTropDeConnexions signale un registre plein. L'appelant le traduit en
// message, ce n'est pas une panne.
var ErrTropDeConnexions = errors.New("trop de connexions ouvertes, en fermer une avant d'en ouvrir une autre")

// registre garde les pilotes ouverts entre deux appels d'API.
//
// Le DSN n'y figure pas, et ce n'est pas un oubli : une fois le pilote ouvert,
// plus rien n'en a besoin. Ce qui n'est pas conservé ne peut pas ressortir d'une
// réponse ni d'un journal.
type registre struct {
	mu    sync.Mutex
	parID map[string]introspection.Introspecteur
}

// nouveauRegistre rend un registre vide.
func nouveauRegistre() *registre {
	return &registre{parID: map[string]introspection.Introspecteur{}}
}

// ajouter enregistre un pilote ouvert et rend l'identifiant que le front
// portera ensuite.
//
// L'identifiant est tiré comme un secret, pas incrémenté : il désigne une
// connexion utilisable, et deviner celui du voisin reviendrait à emprunter la
// base d'un autre onglet.
func (r *registre) ajouter(pilote introspection.Introspecteur) (string, error) {
	id, err := genererJeton()
	if err != nil {
		return "", err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.parID) >= maxConnexions {
		return "", ErrTropDeConnexions
	}
	r.parID[id] = pilote
	return id, nil
}

// trouver rend le pilote d'un identifiant.
func (r *registre) trouver(id string) (introspection.Introspecteur, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	pilote, ok := r.parID[id]
	return pilote, ok
}

// fermer ferme la connexion et l'oublie. Fermer deux fois n'est pas une erreur :
// un onglet rechargé peut poster la même fermeture.
func (r *registre) fermer(id string) error {
	r.mu.Lock()
	pilote, ok := r.parID[id]
	delete(r.parID, id)
	r.mu.Unlock()

	if !ok {
		return nil
	}
	return pilote.Fermer()
}

// toutFermer libère ce qui reste, à l'arrêt du serveur.
//
// Les erreurs ne sont pas remontées : le processus s'arrête, et une connexion
// qu'on n'a pas su fermer proprement sera coupée de toute façon.
func (r *registre) toutFermer() {
	r.mu.Lock()
	restants := make([]introspection.Introspecteur, 0, len(r.parID))
	for id, pilote := range r.parID {
		restants = append(restants, pilote)
		delete(r.parID, id)
	}
	r.mu.Unlock()

	for _, pilote := range restants {
		_ = pilote.Fermer()
	}
}
