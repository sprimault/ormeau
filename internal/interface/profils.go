// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/sprimault/ormeau/internal/config"
	"github.com/sprimault/ormeau/internal/introspection"
)

// gererProfils liste, enregistre et supprime les connexions enregistrées.
func (s *serveur) gererProfils(w http.ResponseWriter, r *http.Request) {
	if s.emplacements == nil {
		repondreErreur(w, http.StatusServiceUnavailable,
			"aucun répertoire de configuration : les profils ne peuvent pas être enregistrés")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.listerProfils(w, "")
	case http.MethodPost:
		s.enregistrerProfil(w, r)
	case http.MethodDelete:
		s.supprimerProfil(w, r)
	default:
		repondreErreur(w, http.StatusMethodNotAllowed, "méthode non acceptée")
	}
}

// listerProfils rend les connexions enregistrées.
//
// Sans les mots de passe, pas même chiffrés : le navigateur n'a besoin que de
// savoir s'il y en a un, pour cocher la case et remplir le champ d'un
// substitut. Le clair ne quitte jamais le serveur — il va du fichier au pilote.
//
// evenement porte ce que l'opération qui précède doit dire une seule fois, à
// côté de ce que le chargement dit à chaque fois.
func (s *serveur) listerProfils(w http.ResponseWriter, evenement string) {
	profils, avertissement, err := s.emplacements.LireProfils()
	if err != nil {
		repondreErreur(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if evenement != "" && avertissement != "" {
		avertissement += " ; "
	}
	avertissement += evenement

	resumes := make([]ProfilResume, 0, len(profils))
	for _, p := range profils {
		resumes = append(resumes, ProfilResume{
			Profil:               p,
			MotDePasseEnregistre: p.MotDePasseEnregistre(),
		})
	}
	repondreJSON(w, http.StatusOK, ReponseProfils{
		Profils:       resumes,
		Avertissement: avertissement,
	})
}

// enregistrerProfil ajoute ou remplace une connexion.
//
// Le mot de passe n'est retenu que si la case a été cochée. Décochée, il est
// effacé de ce qui était enregistré : laisser l'ancien en place ferait croire à
// un retrait qui n'a pas eu lieu.
func (s *serveur) enregistrerProfil(w http.ResponseWriter, r *http.Request) {
	var requete RequeteProfil
	if !decoder(w, r, &requete) {
		return
	}

	profil, motDePasse, err := requete.composer()
	if err != nil {
		repondreErreur(w, http.StatusBadRequest, err.Error())
		return
	}
	requete.Profil = profil

	switch avertissement, err := s.emplacements.EnregistrerProfil(requete.Profil, motDePasse, requete.Remplacer); {
	case errors.Is(err, config.ErrProfilExistant):
		repondreRefus(w, CodeProfilExistant, "un profil porte déjà ce nom")
	case errors.Is(err, config.ErrTropDeProfils):
		repondreErreur(w, http.StatusConflict, err.Error())
	case err != nil:
		repondreErreur(w, http.StatusUnprocessableEntity, err.Error())
	default:
		s.listerProfils(w, avertissement)
	}
}

// composer rend le profil à enregistrer et le mot de passe à retenir.
//
// Une chaîne de connexion est décomposée ici : le front ne l'analyse jamais, et
// sans cela un profil enregistré depuis ce mode de saisie ne garderait ni hôte
// ni base, et ne servirait à rien la fois suivante.
//
// Le nom vient toujours du formulaire, jamais du DSN : c'est celui qu'on lit
// dans la liste, et « gescom production » se distingue d'un coup d'œil de
// « gescom recette » là où deux DSN se ressemblent.
func (r RequeteProfil) composer() (config.Profil, string, error) {
	profil, motDePasse := r.Profil, r.MotDePasse

	if r.DSN != "" {
		c, err := introspection.ConnexionDepuisDSN(r.DSN)
		if err != nil {
			return config.Profil{}, "", err
		}
		profil.SGBD, profil.Hote, profil.Port = c.SGBD, c.Hote, c.Port
		profil.Utilisateur, profil.Base, profil.SSLMode = c.Utilisateur, c.Base, c.SSLMode
		if motDePasse == "" {
			motDePasse = c.MotDePasse
		}
	}

	if !r.EnregistrerMotDePasse {
		motDePasse = ""
	}
	return profil, motDePasse, nil
}

// supprimerProfil retire une connexion enregistrée.
func (s *serveur) supprimerProfil(w http.ResponseWriter, r *http.Request) {
	var requete ReferenceProfil
	if !decoder(w, r, &requete) {
		return
	}

	switch err := s.emplacements.SupprimerProfil(requete.Nom); {
	case errors.Is(err, config.ErrProfilInconnu):
		repondreErreur(w, http.StatusNotFound, "profil inconnu")
	case err != nil:
		repondreErreur(w, http.StatusUnprocessableEntity, err.Error())
	default:
		s.listerProfils(w, "")
	}
}

// connexionDuProfil complète une requête de connexion avec ce qu'un profil
// enregistré porte.
//
// C'est le serveur qui relit le profil, pas le navigateur : le mot de passe
// enregistré ne descend jamais dans la page, il va du fichier au pilote. Le
// front n'envoie qu'un nom.
//
// Un mot de passe saisi l'emporte sur celui du profil — on vient de le taper,
// c'est qu'il a changé. Une clé disparue n'est pas une panne : la connexion
// part sans mot de passe et la base la refusera avec son propre message, ce qui
// est plus clair qu'un refus d'ouvrir l'écran.
//
// Le mot de passe enregistré ne part que vers la destination du profil : même
// SGBD, même hôte, même port, même utilisateur. Choisir « production » puis
// corriger l'hôte l'enverrait sinon à un autre serveur, et l'essayer pour un
// autre compte compterait un échec contre lui, jusqu'au verrouillage sous
// SQL Server ou Oracle. Le refus arrive avant tout contact avec la base : une
// connexion tentée sans mot de passe serait une tentative ratée de plus. La
// base, elle, reste libre : le mot de passe est celui du rôle, pas de la base.
//
// Rend vrai quand le profil nomme une base : la session y restera.
func (s *serveur) connexionDuProfil(requete *RequeteConnexion) (bool, error) {
	if requete.Profil == "" || s.emplacements == nil {
		return false, nil
	}
	if strings.TrimSpace(requete.DSN) != "" {
		return false, errProfilEtChaine
	}

	profils, _, err := s.emplacements.LireProfils()
	if err != nil {
		return false, err
	}

	var profil *config.Profil
	for i := range profils {
		if profils[i].Nom == requete.Profil {
			profil = &profils[i]
			break
		}
	}
	if profil == nil {
		return false, config.ErrProfilInconnu
	}

	completer(&requete.SGBD, profil.SGBD)
	completer(&requete.Hote, profil.Hote)
	completer(&requete.Utilisateur, profil.Utilisateur)
	completer(&requete.Base, profil.Base)
	completer(&requete.SSLMode, profil.SSLMode)
	if requete.Port == 0 {
		requete.Port = profil.Port
	}

	if requete.MotDePasse == "" && profil.MotDePasseEnregistre() {
		if champ := champDivergent(*requete, *profil); champ != "" {
			return false, destinationDivergente{profil: profil.Nom, champ: champ}
		}
		clair, err := s.emplacements.MotDePasseDuProfil(profil.Nom)
		switch {
		// Profil par profil, et jamais un rejet global : un mot de passe qu'on
		// ne sait plus lire n'empêche pas de se connecter en le retapant, et
		// les autres profils n'y sont pour rien.
		case errors.Is(err, config.ErrCleAbsente),
			errors.Is(err, config.ErrMotDePasseIndechiffrable):
			slog.Warn("mot de passe enregistre inutilisable",
				"profile", profil.Nom, "error", err.Error())
		case err != nil:
			return false, err
		default:
			requete.MotDePasse = clair
		}
	}

	s.suivreRepertoireDuProfil(profil.Repertoire)

	// Un profil qui nomme une base dit sur quoi l'on travaille ; un profil qui
	// n'en nomme pas sert à parcourir le serveur.
	return profil.Base != "", nil
}

// errProfilEtChaine refuse une requête qui porte un profil et une chaîne de
// connexion : la chaîne décrit à elle seule la connexion, et le profil
// imposerait sa base et son répertoire à un serveur qui n'est peut-être pas le
// sien. Le front n'envoie jamais les deux.
var errProfilEtChaine = errors.New("un profil et une chaîne de connexion à la fois : choisir l'un ou l'autre")

// destinationDivergente refuse d'envoyer le mot de passe d'un profil vers une
// autre destination que la sienne. Le message dit quoi faire, parce que
// l'écran l'affiche tel quel.
type destinationDivergente struct {
	profil, champ string
}

// Error nomme le champ en cause et les deux issues.
func (d destinationDivergente) Error() string {
	return fmt.Sprintf("%s diffère de celui du profil « %s » : ressaisir le mot de passe, ou mettre le profil à jour",
		d.champ, d.profil)
}

// champDivergent rend le champ par lequel la requête vise une autre destination
// que le profil, ou une chaîne vide. La comparaison porte sur ce que les deux
// visent réellement — SGBD déduit du port, port par défaut, hôte sans casse —,
// pas sur le texte saisi.
func champDivergent(requete RequeteConnexion, profil config.Profil) string {
	sgbdR, hoteR, portR := introspection.Connexion{SGBD: requete.SGBD, Hote: requete.Hote, Port: requete.Port}.Destination()
	sgbdP, hoteP, portP := introspection.Connexion{SGBD: profil.SGBD, Hote: profil.Hote, Port: profil.Port}.Destination()
	switch {
	case sgbdR != sgbdP:
		return "le SGBD"
	case hoteR != hoteP:
		return "l'hôte"
	case portR != portP:
		return "le port"
	case requete.Utilisateur != profil.Utilisateur:
		return "l'utilisateur"
	}
	return ""
}

// suivreRepertoireDuProfil bascule le répertoire de travail sur celui que le
// profil a mémorisé.
//
// Un profil sans répertoire laisse le courant tel quel : quelqu'un qui n'en a
// jamais réglé ne s'attend pas à ce qu'un choix de connexion le déplace.
//
// Un répertoire devenu inutilisable — projet déplacé, disque amovible absent —
// ne bloque pas la connexion : on garde le courant, que l'écran affiche, et
// c'est l'écriture du calque qui dira où elle va.
func (s *serveur) suivreRepertoireDuProfil(repertoire string) {
	if repertoire == "" {
		return
	}
	resolu, err := repertoireUtilisable(repertoire)
	if err != nil {
		slog.Warn("repertoire du profil ignore", "error", err.Error())
		return
	}

	s.travail.Lock()
	s.repertoire = resolu
	s.travail.Unlock()
}

// completer pose la valeur du profil quand la requête n'en donne pas.
func completer(champ *string, valeur string) {
	if *champ == "" {
		*champ = valeur
	}
}
