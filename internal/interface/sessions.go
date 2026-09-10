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

// connexion est une base atteinte, gardée ouverte entre deux appels d'API.
//
// Le DSN y figure parce qu'un serveur porte souvent vingt bases : en changer
// demande de rouvrir une connexion, et le redemander obligerait à ressaisir un
// mot de passe qu'on vient de donner. Il vit là et nulle part ailleurs — jamais
// dans une réponse, un journal ou un message d'erreur, ce que les tests
// vérifient.
//
// Le mutex n'est pas une précaution : une connexion de base de données porte un
// protocole à un seul échange à la fois, et pgx refuse net une seconde requête
// sur la même connexion — « conn busy ». Or le navigateur en lance plusieurs de
// front dès le chargement d'un écran, l'arbre et la liste des bases par exemple.
// Sans cette sérialisation, l'une des deux échoue, et laquelle dépend de
// l'ordonnancement.
type connexion struct {
	mu     sync.Mutex
	pilote introspection.Introspecteur
	dsn    string
	sgbd   string
}

// utiliser donne accès au pilote, une requête à la fois.
//
// Le verrou couvre l'appel entier plutôt que l'envoi seul : c'est la lecture du
// résultat qui occupe la connexion, pas la seule émission de la requête.
func (c *connexion) utiliser(action func(introspection.Introspecteur) error) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return action(c.pilote)
}

// registre garde les connexions vivantes de la session d'interface.
type registre struct {
	mu    sync.Mutex
	parID map[string]*connexion
}

// nouveauRegistre rend un registre vide.
func nouveauRegistre() *registre {
	return &registre{parID: map[string]*connexion{}}
}

// ajouter enregistre un pilote ouvert et rend l'identifiant que le front
// portera ensuite.
//
// L'identifiant est tiré comme un secret, pas incrémenté : il désigne une
// connexion utilisable, et deviner celui du voisin reviendrait à emprunter la
// base d'un autre onglet.
func (r *registre) ajouter(pilote introspection.Introspecteur, dsn, sgbd string) (string, error) {
	id, err := genererJeton()
	if err != nil {
		return "", err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.parID) >= maxConnexions {
		return "", ErrTropDeConnexions
	}
	r.parID[id] = &connexion{pilote: pilote, dsn: dsn, sgbd: sgbd}
	return id, nil
}

// trouver rend la connexion d'un identifiant.
func (r *registre) trouver(id string) (*connexion, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.parID[id]
	return c, ok
}

// fermer ferme la connexion et l'oublie. Fermer deux fois n'est pas une erreur :
// un onglet rechargé peut poster la même fermeture.
func (r *registre) fermer(id string) error {
	r.mu.Lock()
	c, ok := r.parID[id]
	delete(r.parID, id)
	r.mu.Unlock()

	if !ok {
		return nil
	}
	// Attend la requête en cours : fermer sous les pieds d'un inventaire lui
	// ferait rendre une erreur de transport plutôt qu'un résultat.
	return c.utiliser(func(pilote introspection.Introspecteur) error { return pilote.Fermer() })
}

// toutFermer libère ce qui reste, à l'arrêt du serveur.
//
// Les erreurs ne sont pas remontées : le processus s'arrête, et une connexion
// qu'on n'a pas su fermer proprement sera coupée de toute façon.
func (r *registre) toutFermer() {
	r.mu.Lock()
	restantes := make([]*connexion, 0, len(r.parID))
	for id, c := range r.parID {
		restantes = append(restantes, c)
		delete(r.parID, id)
	}
	r.mu.Unlock()

	for _, c := range restantes {
		_ = c.utiliser(func(pilote introspection.Introspecteur) error { return pilote.Fermer() })
	}
}
