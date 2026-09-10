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

// delaiInventaire borne la passe légère qui alimente l'arbre. Une requête de
// catalogue, même sur quatre cents tables, tient largement dedans.
const delaiInventaire = 60 * time.Second

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

// ReponseBases liste les bases exploitables du serveur atteint.
//
// Les bases système en sont absentes : elles ne produiraient que des calques
// sans intérêt, et template0 refuse même la connexion.
type ReponseBases struct {
	Bases []string `json:"bases"`
}

// RequeteBase demande de basculer la session sur une autre base du même
// serveur.
type RequeteBase struct {
	Session string `json:"session"`
	Base    string `json:"base"`
}

// ReponseColonnes décrit une table dépliée dans l'arbre.
type ReponseColonnes struct {
	Colonnes []introspection.ColonneSommaire `json:"colonnes"`
}

// ReponseInventaire porte l'arbre de sélection.
//
// L'inventaire complet part d'un coup, sans pagination : quelques dizaines de
// kilo-octets pour quatre cents tables, et la recherche reste instantanée côté
// navigateur. Paginer coûterait un aller-retour par frappe pour économiser un
// transfert qui tient dans un paquet réseau.
type ReponseInventaire struct {
	Tables []introspection.TableSommaire `json:"tables"`
}

// contexte rend le répertoire de travail et la version.
func (s *serveur) contexte(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		repondreErreur(w, http.StatusMethodNotAllowed, "méthode non acceptée")
		return
	}
	repondreJSON(w, http.StatusOK, ReponseContexte{Repertoire: s.repertoire, Version: s.version})
}

// bases rend les bases du serveur atteint par une connexion ouverte.
//
// Un serveur en porte souvent vingt, et personne ne les retient : les faire
// deviner est exactement ce que le projet refuse. L'information est atteignable,
// donc elle ne se demande pas.
func (s *serveur) bases(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		repondreErreur(w, http.StatusMethodNotAllowed, "méthode non acceptée")
		return
	}

	c, ok := s.registre.trouver(r.URL.Query().Get("session"))
	if !ok {
		repondreErreur(w, http.StatusNotFound, "connexion inconnue ou déjà fermée")
		return
	}

	listeur, ok := c.pilote.(introspection.ListeurDeBases)
	if !ok {
		// Un dialecte où la notion n'a pas de sens n'est pas une panne : le
		// front se passe du sélecteur et garde la base saisie.
		repondreJSON(w, http.StatusOK, ReponseBases{Bases: []string{}})
		return
	}

	ctx, annuler := context.WithTimeout(r.Context(), delaiInventaire)
	defer annuler()

	var bases []string
	err := c.utiliser(func(introspection.Introspecteur) error {
		var err error
		bases, err = listeur.ListerBases(ctx)
		return err
	})
	if err != nil {
		repondreErreur(w, http.StatusBadGateway, sansDSN(err.Error(), c.dsn))
		return
	}
	if bases == nil {
		bases = []string{}
	}
	repondreJSON(w, http.StatusOK, ReponseBases{Bases: bases})
}

// basculerBase rouvre la session sur une autre base du même serveur.
//
// Les identifiants sont ceux de la connexion en cours : changer de base ne doit
// pas faire ressaisir un mot de passe qu'on vient de donner. L'ancienne
// connexion est refermée, et la nouvelle reçoit un nouvel identifiant — le
// front repart d'un état propre plutôt que de deviner ce qui a changé sous lui.
func (s *serveur) basculerBase(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		repondreErreur(w, http.StatusMethodNotAllowed, "méthode non acceptée")
		return
	}

	var requete RequeteBase
	if !decoder(w, r, &requete) {
		return
	}

	c, ok := s.registre.trouver(requete.Session)
	if !ok {
		repondreErreur(w, http.StatusNotFound, "connexion inconnue ou déjà fermée")
		return
	}
	if strings.TrimSpace(requete.Base) == "" {
		repondreErreur(w, http.StatusBadRequest, "aucune base demandée")
		return
	}

	ctx, annuler := context.WithTimeout(r.Context(), delaiConnexion)
	defer annuler()

	dsn := introspection.AvecBase(c.dsn, requete.Base)
	pilote, err := introspection.Ouvrir(ctx, c.sgbd, dsn)
	if err != nil {
		message := sansDSN(err.Error(), dsn)
		slog.Warn("bascule de base refusee", "dbms", c.sgbd, "error", message)
		repondreErreur(w, http.StatusBadGateway, message)
		return
	}

	reponse, err := s.enregistrer(ctx, pilote, dsn, c.sgbd)
	if err != nil {
		repondreErreur(w, codeDe(err), sansDSN(err.Error(), dsn))
		return
	}

	// Fermée après coup : si la nouvelle base est inaccessible, l'utilisateur
	// garde celle qu'il avait.
	if err := s.registre.fermer(requete.Session); err != nil {
		slog.Warn("fermeture de l'ancienne connexion", "error", err)
	}
	repondreJSON(w, http.StatusOK, reponse)
}

// inventaire rend les tables d'une connexion ouverte.
//
// Une passe légère, jamais une extraction : ce qui alimente l'arbre vient du
// catalogue, et pas une ligne de donnée n'est lue.
func (s *serveur) inventaire(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		repondreErreur(w, http.StatusMethodNotAllowed, "méthode non acceptée")
		return
	}

	c, ok := s.registre.trouver(r.URL.Query().Get("session"))
	if !ok {
		repondreErreur(w, http.StatusNotFound, "connexion inconnue ou déjà fermée")
		return
	}

	ctx, annuler := context.WithTimeout(r.Context(), delaiInventaire)
	defer annuler()

	var tables []introspection.TableSommaire
	err := c.utiliser(func(pilote introspection.Introspecteur) error {
		var err error
		tables, err = pilote.Inventorier(ctx, decouper(r.URL.Query().Get("schemas")))
		return err
	})
	if err != nil {
		slog.Warn("inventaire refuse", "error", err)
		repondreErreur(w, http.StatusBadGateway, err.Error())
		return
	}
	if tables == nil {
		// Une portée sans table est un résultat vide, pas une absence de
		// résultat : le front parcourt la liste sans avoir à distinguer les deux.
		tables = []introspection.TableSommaire{}
	}
	repondreJSON(w, http.StatusOK, ReponseInventaire{Tables: tables})
}

// colonnes décrit une table, à la demande.
//
// Chargée au dépliement plutôt qu'avec l'inventaire : porter les colonnes de
// quatre cents tables coûterait dix fois le transfert pour des lignes que
// personne n'ouvrira.
func (s *serveur) colonnes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		repondreErreur(w, http.StatusMethodNotAllowed, "méthode non acceptée")
		return
	}

	requete := r.URL.Query()
	c, ok := s.registre.trouver(requete.Get("session"))
	if !ok {
		repondreErreur(w, http.StatusNotFound, "connexion inconnue ou déjà fermée")
		return
	}

	schema, table := requete.Get("schema"), requete.Get("table")
	if schema == "" || table == "" {
		repondreErreur(w, http.StatusBadRequest, "schéma et table sont requis")
		return
	}

	listeur, ok := c.pilote.(introspection.ListeurDeColonnes)
	if !ok {
		repondreErreur(w, http.StatusNotImplemented, "ce pilote ne sait pas encore décrire une table")
		return
	}

	ctx, annuler := context.WithTimeout(r.Context(), delaiInventaire)
	defer annuler()

	var colonnes []introspection.ColonneSommaire
	err := c.utiliser(func(introspection.Introspecteur) error {
		var err error
		colonnes, err = listeur.Colonnes(ctx, schema, table)
		return err
	})
	if err != nil {
		slog.Warn("lecture des colonnes refusee", "error", err)
		repondreErreur(w, http.StatusBadGateway, err.Error())
		return
	}
	if colonnes == nil {
		colonnes = []introspection.ColonneSommaire{}
	}
	repondreJSON(w, http.StatusOK, ReponseColonnes{Colonnes: colonnes})
}

// decouper rend les valeurs d'une liste séparée par des virgules. Une entrée
// vide n'est pas un schéma : « public, » n'en demande pas deux.
func decouper(liste string) []string {
	var valeurs []string
	for _, brut := range strings.Split(liste, ",") {
		if valeur := strings.TrimSpace(brut); valeur != "" {
			valeurs = append(valeurs, valeur)
		}
	}
	return valeurs
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

	reponse, err := s.enregistrer(ctx, pilote, dsn, sgbd)
	if err != nil {
		repondreErreur(w, codeDe(err), sansDSN(err.Error(), dsn))
		return
	}
	repondreJSON(w, http.StatusOK, reponse)
}

// errPiloteMuet signale un dialecte dont le pilote ne sait pas encore se
// décrire. C'est un manque annoncé, pas une panne de la base.
var errPiloteMuet = errors.New("ce pilote ne sait pas encore se décrire")

// enregistrer décrit le serveur atteint, garde la connexion ouverte et compose
// ce que le front affiche. Le pilote est refermé dès que l'une des étapes
// échoue : une connexion qu'on n'enregistre pas ne doit pas survivre.
func (s *serveur) enregistrer(
	ctx context.Context,
	pilote introspection.Introspecteur,
	dsn, sgbd string,
) (ReponseConnexion, error) {
	descripteur, ok := pilote.(introspection.DescripteurServeur)
	if !ok {
		_ = pilote.Fermer()
		return ReponseConnexion{}, errPiloteMuet
	}

	serveurBase, err := descripteur.Decrire(ctx)
	if err != nil {
		_ = pilote.Fermer()
		return ReponseConnexion{}, err
	}

	session, err := s.registre.ajouter(pilote, dsn, sgbd)
	if err != nil {
		_ = pilote.Fermer()
		return ReponseConnexion{}, err
	}

	return ReponseConnexion{
		Session:   session,
		SGBD:      serveurBase.SGBD,
		Version:   serveurBase.Version,
		Catalogue: serveurBase.Catalogue,
		Schemas:   serveurBase.Schemas,
	}, nil
}

// codeDe traduit un échec d'enregistrement en code HTTP. Le front distingue ce
// qu'il peut corriger de ce qu'il ne peut pas.
func codeDe(err error) int {
	switch {
	case errors.Is(err, ErrTropDeConnexions):
		return http.StatusConflict
	case errors.Is(err, errPiloteMuet):
		return http.StatusNotImplemented
	default:
		return http.StatusBadGateway
	}
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
