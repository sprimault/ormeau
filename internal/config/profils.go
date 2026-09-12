// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// maxProfils plafonne le fichier. Personne ne reprend cinquante bases à la
// fois, et un formulaire rejoué en boucle ne doit pas pouvoir le faire enfler
// sans fin.
const maxProfils = 50

// nomProfil accepte ce qui se lit dans une liste : lettres de n'importe quelle
// langue, chiffres, espace et ponctuation courante.
var nomProfil = regexp.MustCompile(`^[\p{L}\p{N} ._-]{1,60}$`)

// nomBase accepte ce qui peut composer un nom de fichier de session sans en
// sortir. Même expression que côté interface, et refusée plutôt que nettoyée :
// un nettoyage se contourne, un refus non.
var nomBase = regexp.MustCompile(`^[\p{L}\p{N}_-]{1,64}$`)

// ErrProfilInconnu signale un nom qui ne figure pas dans le fichier.
var ErrProfilInconnu = errors.New("profil inconnu")

// ErrTropDeProfils signale un fichier plein.
var ErrTropDeProfils = errors.New("trop de profils enregistres")

// ErrProfilExistant signale un nom déjà pris, quand le remplacement n'a pas été
// confirmé.
//
// Enregistrer par-dessus est le geste qu'on fait sans y penser après avoir
// modifié un champ, et écraser en silence un profil de production n'a rien
// d'anodin.
var ErrProfilExistant = errors.New("un profil porte deja ce nom")

// Profil est une connexion enregistrée, telle que l'écran de connexion la
// repropose.
//
// Le mot de passe n'y figure qu'en chiffré, et seulement si la case
// « enregistrer le mot de passe » a été cochée. Décocher l'efface : sinon on
// croirait l'avoir retiré alors qu'il resterait sur le disque.
type Profil struct {
	Nom         string `yaml:"nom" json:"nom"`
	SGBD        string `yaml:"sgbd,omitempty" json:"sgbd,omitempty"`
	Hote        string `yaml:"hote,omitempty" json:"hote,omitempty"`
	Port        int    `yaml:"port,omitempty" json:"port,omitempty"`
	Utilisateur string `yaml:"utilisateur,omitempty" json:"utilisateur,omitempty"`
	Base        string `yaml:"base,omitempty" json:"base,omitempty"`
	// Repertoire est le répertoire de travail associé. Vide, choisir le profil
	// laisse le répertoire courant tel quel : quelqu'un qui n'a jamais réglé de
	// répertoire ne s'attend pas à ce qu'un profil le déplace.
	Repertoire string `yaml:"repertoire,omitempty" json:"repertoire,omitempty"`
	// MotDePasse est le chiffré, et ne sort jamais du paquet : l'API rend
	// MotDePasseEnregistre, un booléen, et le clair ne va qu'au pilote.
	MotDePasse string `yaml:"mot_de_passe,omitempty" json:"-"`
}

// MotDePasseEnregistre dit si le profil en porte un, sans le révéler.
func (p Profil) MotDePasseEnregistre() bool {
	return p.MotDePasse != ""
}

// fichierProfils est la forme du fichier. Une clé de tête plutôt qu'une liste
// nue : ajouter une option plus tard ne cassera pas les fichiers existants.
type fichierProfils struct {
	Profils []Profil `yaml:"profils"`
}

// LireProfils rend les connexions enregistrées, un avertissement, et une
// erreur.
//
// Mêmes règles que les préférences, et pour la même raison : un fichier abîmé
// ne vaut pas un refus de démarrer. L'écran de connexion s'ouvre sans profil,
// et on saisit comme avant.
func (e *Emplacements) LireProfils() ([]Profil, string, error) {
	contenu, err := os.ReadFile(e.FichierProfils())
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil, "", nil
	case errors.Is(err, fs.ErrPermission):
		return nil, "", fmt.Errorf("lecture de %s: %w", e.FichierProfils(), err)
	case err != nil:
		return nil, fmt.Sprintf("%s illisible, aucun profil chargé : %v",
			e.FichierProfils(), err), nil
	}

	var lu fichierProfils
	if err := yaml.Unmarshal(contenu, &lu); err != nil {
		return nil, fmt.Sprintf("%s mal formé, aucun profil chargé : %v",
			e.FichierProfils(), err), nil
	}

	// Un nom hors vocabulaire ne vient pas de l'écran, qui le valide : le
	// fichier a été retouché. On écarte l'entrée plutôt que de la laisser
	// désigner un profil qu'on ne saura ni relire ni supprimer.
	var retenus []Profil
	var ecartes []string
	for _, p := range lu.Profils {
		if !nomProfil.MatchString(p.Nom) {
			ecartes = append(ecartes, p.Nom)
			continue
		}
		retenus = append(retenus, p)
	}

	trierProfils(retenus)

	var avertissements []string
	if len(ecartes) > 0 {
		avertissements = append(avertissements,
			fmt.Sprintf("profils ignorés, nom invalide : %s", strings.Join(ecartes, ", ")))
	}
	if a := e.avertirCleInutilisable(retenus); a != "" {
		avertissements = append(avertissements, a)
	}
	return retenus, strings.Join(avertissements, " ; "), nil
}

// avertirCleInutilisable signale des mots de passe enregistrés que ce poste ne
// pourra pas relire, faute d'une clé qui s'ouvre.
//
// Le cas courant est un ménage un peu large dans le répertoire de
// configuration, ou un profils.yaml recopié depuis une autre machine ; plus
// rarement une clé tronquée. Les profils restent utilisables — hôte, port,
// utilisateur, base et répertoire valent encore —, seule la ressaisie du mot
// de passe revient. Dit une fois, au chargement, et non à chaque tentative de
// connexion.
func (e *Emplacements) avertirCleInutilisable(profils []Profil) string {
	if !slices.ContainsFunc(profils, Profil.MotDePasseEnregistre) {
		return ""
	}
	var etat string
	switch _, _, err := e.cle(false); {
	case errors.Is(err, ErrCleAbsente):
		etat = "est absent"
	case errors.Is(err, ErrMotDePasseIndechiffrable):
		etat = "est inutilisable"
	default:
		return ""
	}
	return fmt.Sprintf("mots de passe enregistrés illisibles : %s %s, "+
		"les profils restent utilisables sans eux", e.fichierCle(), etat)
}

// MotDePasseDuProfil rend le mot de passe en clair d'un profil enregistré.
//
// Rendu à part et jamais avec le profil : il ne traverse pas l'API, il ne va
// qu'au pilote au moment d'ouvrir la connexion.
//
// Une clé disparue rend ErrCleAbsente, que l'appelant traite comme une absence
// de mot de passe : le reste du profil vaut encore, seule la ressaisie revient.
func (e *Emplacements) MotDePasseDuProfil(nom string) (string, error) {
	profils, _, err := e.LireProfils()
	if err != nil {
		return "", err
	}

	i := slices.IndexFunc(profils, func(p Profil) bool { return p.Nom == nom })
	if i < 0 {
		return "", ErrProfilInconnu
	}
	if profils[i].MotDePasse == "" {
		return "", nil
	}
	return e.dechiffrer(profils[i].MotDePasse)
}

// EnregistrerProfil ajoute ou remplace un profil, et chiffre le mot de passe
// quand il y en a un.
//
// Le mot de passe vide efface celui qui était enregistré : c'est ce que produit
// une case décochée, et laisser l'ancien en place ferait croire à un retrait
// qui n'a pas eu lieu.
//
// L'avertissement rendu dit qu'une clé inutilisable a été remplacée pour
// chiffrer ce mot de passe. Le chiffrement vient après les refus : une clé ne
// se remplace pas pour un enregistrement qui n'aura pas lieu.
func (e *Emplacements) EnregistrerProfil(p Profil, motDePasse string, remplacer bool) (string, error) {
	if !nomProfil.MatchString(p.Nom) {
		return "", fmt.Errorf("nom de profil invalide: %q", p.Nom)
	}

	profils, _, err := e.LireProfils()
	if err != nil {
		return "", err
	}
	i := slices.IndexFunc(profils, func(q Profil) bool { return q.Nom == p.Nom })
	switch {
	case i >= 0 && !remplacer:
		return "", ErrProfilExistant
	case i < 0 && len(profils) >= maxProfils:
		return "", ErrTropDeProfils
	}

	chiffre, avertissement, err := e.chiffrer(motDePasse)
	if err != nil {
		return "", err
	}
	p.MotDePasse = chiffre

	if i >= 0 {
		profils[i] = p
	} else {
		profils = append(profils, p)
	}
	return avertissement, e.ecrireProfils(profils)
}

// SupprimerProfil retire un profil du fichier.
func (e *Emplacements) SupprimerProfil(nom string) error {
	profils, _, err := e.LireProfils()
	if err != nil {
		return err
	}

	i := slices.IndexFunc(profils, func(p Profil) bool { return p.Nom == nom })
	if i < 0 {
		return ErrProfilInconnu
	}
	return e.ecrireProfils(slices.Delete(profils, i, i+1))
}

// ecrireProfils réécrit le fichier entier.
//
// Entier et non par retouche : le fichier est petit, et une réécriture complète
// évite d'avoir à raisonner sur ce qu'une édition partielle laisse derrière
// elle.
func (e *Emplacements) ecrireProfils(profils []Profil) error {
	trierProfils(profils)

	contenu, err := yaml.Marshal(fichierProfils{Profils: profils})
	if err != nil {
		return fmt.Errorf("serialisation des profils: %w", err)
	}
	if err := os.WriteFile(e.FichierProfils(), contenu, permFichier); err != nil {
		return fmt.Errorf("ecriture de %s: %w", e.FichierProfils(), err)
	}
	return nil
}

// trierProfils range par nom, sans tenir compte de la casse.
//
// L'ordre est celui de la liste affichée : le laisser suivre celui des
// enregistrements rendrait la liste imprévisible à chaque ajout.
func trierProfils(profils []Profil) {
	slices.SortFunc(profils, func(a, b Profil) int {
		return strings.Compare(strings.ToLower(a.Nom), strings.ToLower(b.Nom))
	})
}
