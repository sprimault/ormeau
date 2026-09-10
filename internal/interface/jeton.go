// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

// nomCookie identifie le cookie de session. Une seule valeur dans tout le
// projet : le front n'a pas à le connaître, il est HttpOnly.
const nomCookie = "ormeau_session"

// dureeJetonURL borne la validité du jeton imprimé dans l'URL de démarrage. Le
// navigateur s'ouvre dans la seconde ; au-delà d'une minute, c'est que
// l'ouverture a échoué et que l'URL traîne quelque part.
const dureeJetonURL = time.Minute

// genererJeton tire un secret de 256 bits.
//
// crypto/rand et jamais math/rand : même ensemencé, ce dernier est prédictible,
// et un jeton devinable donne accès à une API qui ouvre des connexions de base
// de données. RawURLEncoding évite « + », « / » et « = », qui demanderaient un
// échappement dans une URL.
func genererJeton() (string, error) {
	octets := make([]byte, 32)
	if _, err := rand.Read(octets); err != nil {
		return "", fmt.Errorf("generation du jeton: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(octets), nil
}

// acces porte les deux secrets du serveur et la règle qui les sépare.
//
// Le jeton d'URL et le cookie de session sont tirés séparément, et c'est ce qui
// rend l'usage unique effectif : réutiliser le premier comme valeur du second
// le laisserait valide après consommation. Une URL fuit — par l'historique, par
// un en-tête Referer, par une capture de la barre d'adresse, et sous Windows par
// la ligne de commande du navigateur, visible dans la liste des processus.
type acces struct {
	mu           sync.Mutex
	jetonURL     string
	jetonSession string
	expiration   time.Time
	consomme     bool
}

// nouvelAcces tire les deux secrets. Ils vivent en mémoire et meurent avec le
// processus : jamais de fichier, jamais de journal.
func nouvelAcces() (*acces, error) {
	url, err := genererJeton()
	if err != nil {
		return nil, err
	}
	session, err := genererJeton()
	if err != nil {
		return nil, err
	}
	return &acces{
		jetonURL:     url,
		jetonSession: session,
		expiration:   time.Now().Add(dureeJetonURL),
	}, nil
}

// consommer valide le jeton d'URL et le retire définitivement.
//
// La comparaison est à temps constant : comparer octet par octet laisserait
// deviner le jeton par la durée des réponses.
func (a *acces) consommer(fourni string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.consomme || time.Now().After(a.expiration) {
		return false
	}
	if subtle.ConstantTimeCompare([]byte(fourni), []byte(a.jetonURL)) != 1 {
		return false
	}
	a.consomme = true
	return true
}

// valideSession dit si la valeur d'un cookie est celle de la session courante.
func (a *acces) valideSession(fourni string) bool {
	return subtle.ConstantTimeCompare([]byte(fourni), []byte(a.jetonSession)) == 1
}
