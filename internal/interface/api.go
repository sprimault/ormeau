// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/sprimault/ormeau/internal/introspection"
)

// delaiConnexion borne l'ouverture et la description d'une base. Un serveur
// injoignable doit rendre la main assez vite pour que l'écran reste utilisable,
// sans être si court qu'une base lente passe pour inaccessible.
const delaiConnexion = 30 * time.Second

// tailleMaxCorps borne le corps d'une requête. Les corps attendus ici tiennent
// en quelques centaines d'octets.
const tailleMaxCorps = 64 << 10

// RequeteConnexion accepte les deux formes de connexion : une chaîne complète,
// ou les composants.
//
// Personne ne tape une URL dans un formulaire, et celui qui découvre une base
// connaît un hôte et un identifiant, pas un DSN. Quand SGBD est vide, le port le
// désigne — c'est l'outil qui aiguille, pas l'utilisateur qui déclare.
type RequeteConnexion struct {
	DSN         string `json:"dsn,omitempty"`
	SGBD        string `json:"sgbd,omitempty"`
	Hote        string `json:"hote,omitempty"`
	Port        int    `json:"port,omitempty"`
	Utilisateur string `json:"utilisateur,omitempty"`
	MotDePasse  string `json:"mot_de_passe,omitempty"`
	Base        string `json:"base,omitempty"`
}

// ReponseConnexion décrit le serveur atteint. Le DSN n'y figure sous aucune
// forme, pas même masqué.
type ReponseConnexion struct {
	Session   string   `json:"session"`
	SGBD      string   `json:"sgbd"`
	Version   string   `json:"version"`
	Catalogue string   `json:"catalogue"`
	Schemas   []string `json:"schemas"`
}

// RequeteFermeture désigne la connexion à refermer.
type RequeteFermeture struct {
	Session string `json:"session"`
}

// ReponseContexte porte ce que l'interface affiche en permanence : le
// répertoire où elle écrira, et la version du binaire.
type ReponseContexte struct {
	Repertoire string `json:"repertoire"`
	Version    string `json:"version"`
}

// ReponseErreur est la forme unique des échecs d'API. Un code HTTP seul
// laisserait le front deviner ce qu'il affiche.
type ReponseErreur struct {
	Erreur string `json:"erreur"`
}

// contexte rend le répertoire de travail et la version.
func (s *serveur) contexte(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		repondreErreur(w, http.StatusMethodNotAllowed, "méthode non acceptée")
		return
	}
	repondreJSON(w, http.StatusOK, ReponseContexte{Repertoire: s.repertoire, Version: s.version})
}

// connexion ouvre une base ou referme une connexion ouverte.
func (s *serveur) connexion(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.ouvrirConnexion(w, r)
	case http.MethodDelete:
		s.fermerConnexion(w, r)
	default:
		repondreErreur(w, http.StatusMethodNotAllowed, "méthode non acceptée")
	}
}

// ouvrirConnexion joint la base et rend de quoi remplir l'écran suivant.
//
// La connexion reste ouverte au-delà de la requête : l'inventaire et
// l'extraction la réutiliseront, et rouvrir à chaque appel ferait ressaisir le
// mot de passe ou le garderait côté navigateur, ce que l'interface refuse.
func (s *serveur) ouvrirConnexion(w http.ResponseWriter, r *http.Request) {
	var requete RequeteConnexion
	if !decoder(w, r, &requete) {
		return
	}

	dsn, err := requete.composer()
	if err != nil {
		repondreErreur(w, http.StatusBadRequest, err.Error())
		return
	}

	sgbd, err := introspection.SGBDDepuisDSN(dsn)
	if err != nil {
		repondreErreur(w, http.StatusBadRequest, sansDSN(err.Error(), dsn))
		return
	}

	ctx, annuler := context.WithTimeout(r.Context(), delaiConnexion)
	defer annuler()

	pilote, err := introspection.Ouvrir(ctx, sgbd, dsn)
	if err != nil {
		// Le message est nettoyé avant d'aller où que ce soit : ni la réponse ni
		// le journal ne doivent porter le DSN, fût-il masqué.
		message := sansDSN(err.Error(), dsn)
		slog.Warn("connexion refusee", "dbms", sgbd, "error", message)
		repondreErreur(w, http.StatusBadGateway, message)
		return
	}

	descripteur, ok := pilote.(introspection.DescripteurServeur)
	if !ok {
		_ = pilote.Fermer()
		repondreErreur(w, http.StatusNotImplemented, "le pilote "+sgbd+" ne sait pas encore se décrire")
		return
	}

	serveurBase, err := descripteur.Decrire(ctx)
	if err != nil {
		_ = pilote.Fermer()
		repondreErreur(w, http.StatusBadGateway, sansDSN(err.Error(), dsn))
		return
	}

	session, err := s.registre.ajouter(pilote)
	if err != nil {
		_ = pilote.Fermer()
		code := http.StatusInternalServerError
		if errors.Is(err, ErrTropDeConnexions) {
			code = http.StatusConflict
		}
		repondreErreur(w, code, err.Error())
		return
	}

	repondreJSON(w, http.StatusOK, ReponseConnexion{
		Session:   session,
		SGBD:      serveurBase.SGBD,
		Version:   serveurBase.Version,
		Catalogue: serveurBase.Catalogue,
		Schemas:   serveurBase.Schemas,
	})
}

// fermerConnexion referme et oublie. Fermer une session inconnue n'est pas une
// erreur : un onglet rechargé peut poster la même fermeture.
func (s *serveur) fermerConnexion(w http.ResponseWriter, r *http.Request) {
	var requete RequeteFermeture
	if !decoder(w, r, &requete) {
		return
	}
	if err := s.registre.fermer(requete.Session); err != nil {
		slog.Warn("fermeture de connexion", "error", err)
	}
	w.WriteHeader(http.StatusNoContent)
}

// composer rend le DSN à ouvrir, quelle que soit la forme reçue. La chaîne
// complète l'emporte quand les deux sont là : c'est la plus précise.
func (r RequeteConnexion) composer() (string, error) {
	if strings.TrimSpace(r.DSN) != "" {
		return strings.TrimSpace(r.DSN), nil
	}
	return introspection.Connexion{
		SGBD:        r.SGBD,
		Hote:        r.Hote,
		Port:        r.Port,
		Utilisateur: r.Utilisateur,
		MotDePasse:  r.MotDePasse,
		Base:        r.Base,
	}.DSN()
}

// decoder lit le corps JSON de la requête. Rend false quand il a déjà répondu.
func decoder(w http.ResponseWriter, r *http.Request, cible any) bool {
	decodeur := json.NewDecoder(http.MaxBytesReader(w, r.Body, tailleMaxCorps))
	decodeur.DisallowUnknownFields()
	if err := decodeur.Decode(cible); err != nil {
		// Le corps peut porter un mot de passe : l'erreur de décodage cite le
		// contenu fautif, elle ne sort pas d'ici.
		repondreErreur(w, http.StatusBadRequest, "corps de requête illisible")
		return false
	}
	return true
}

// repondreJSON écrit une réponse sérialisée.
//
// json.Marshal direct, et c'est l'exception assumée à la règle qui fait passer
// toute sérialisation par internal/calque : celle-ci existe pour le déterminisme
// octet pour octet des calques, qu'une réponse HTTP n'a pas à tenir. Le jour où
// l'API rendra un calque, c'est calque qui le sérialisera.
func repondreJSON(w http.ResponseWriter, code int, corps any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(corps); err != nil {
		slog.Error("ecriture de la reponse", "error", err)
	}
}

// repondreErreur écrit un échec sous la forme que le front sait afficher.
func repondreErreur(w http.ResponseWriter, code int, message string) {
	repondreJSON(w, code, ReponseErreur{Erreur: message})
}

// sansDSN retire d'un message toute occurrence de la chaîne de connexion.
//
// Les pilotes masquent déjà le mot de passe, mais un DSN masqué reste un hôte,
// un utilisateur et un nom de base — et la règle est que le DSN ne sorte ni dans
// une réponse, ni dans un journal. Le remplacement porte sur les deux formes
// exactes, celle qu'on a composée et sa version masquée.
func sansDSN(message, dsn string) string {
	for _, forme := range []string{dsn, introspection.Masquer(dsn)} {
		if forme != "" {
			message = strings.ReplaceAll(message, forme, "la base")
		}
	}
	return message
}
