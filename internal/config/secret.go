// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// tailleCle est celle d'AES-256.
const tailleCle = 32

// ErrCleAbsente signale qu'aucune clé n'a encore été tirée, ou qu'elle a
// disparu. L'appelant charge alors le profil sans son mot de passe plutôt que
// d'échouer : le reste vaut encore, seule la ressaisie revient.
var ErrCleAbsente = errors.New("aucune cle de chiffrement sur ce poste")

// ErrMotDePasseIndechiffrable signale un chiffré que la clé de ce poste
// n'ouvre pas.
//
// Deux causes qu'aucun test ne sépare — GCM échoue de la même façon — et que le
// message ne doit donc pas confondre en en accusant une : un profils.yaml
// recopié depuis une autre machine sans sa clé, ce qui est le comportement
// voulu ; ou une clé régénérée, un fichier tronqué, des octets retouchés.
//
// Une clé présente mais inutilisable — illisible, ou d'une autre taille que
// celle d'AES-256 — rend la même erreur : aucun mot de passe ne se relira
// avec elle, et l'appelant doit dégrader exactement comme pour un chiffré
// qu'elle n'ouvre pas.
var ErrMotDePasseIndechiffrable = errors.New("mot de passe enregistre illisible sur ce poste")

// Ce que ce chiffrement protège, et ce qu'il ne protège pas.
//
// Il met un mot de passe de base hors de portée d'une lecture fortuite : un
// fichier ouvert par erreur, une sauvegarde, un partage d'écran, une capture
// jointe à un rapport. C'est ce que font DBeaver et SSMS, et c'est ce qui est
// demandé ici.
//
// Il ne protège pas contre quelqu'un qui a la main sur le compte : la clé vit
// à côté du fichier qu'elle chiffre, en 0600, et qui peut lire l'un peut lire
// l'autre. Le dire ainsi dans SECURITY.md plutôt que de laisser croire à un
// coffre — une promesse de confidentialité démentie coûte plus cher que son
// absence.
//
// GCM et non CBC : il authentifie. Un fichier retouché à la main se détecte au
// déchiffrement, au lieu de rendre des octets quelconques qu'on enverrait
// ensuite comme mot de passe à une base de production.

// chiffrer rend le texte chiffré en base64, prêt à figurer dans profils.yaml,
// et un avertissement quand il a fallu remplacer une clé inutilisable.
func (e *Emplacements) chiffrer(clair string) (chiffre, avertissement string, err error) {
	if clair == "" {
		return "", "", nil
	}

	aead, avertissement, err := e.aead(true)
	if err != nil {
		return "", "", err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", "", fmt.Errorf("tirage du nonce: %w", err)
	}

	// Le nonce est préfixé au message : il n'est pas secret, seulement unique.
	scelle := aead.Seal(nonce, nonce, []byte(clair), nil)
	return base64.StdEncoding.EncodeToString(scelle), avertissement, nil
}

// dechiffrer rend le mot de passe en clair.
func (e *Emplacements) dechiffrer(chiffre string) (string, error) {
	if chiffre == "" {
		return "", nil
	}

	scelle, err := base64.StdEncoding.DecodeString(chiffre)
	if err != nil {
		return "", ErrMotDePasseIndechiffrable
	}

	aead, _, err := e.aead(false)
	if err != nil {
		return "", err
	}
	if len(scelle) < aead.NonceSize() {
		return "", ErrMotDePasseIndechiffrable
	}

	nonce, message := scelle[:aead.NonceSize()], scelle[aead.NonceSize():]
	clair, err := aead.Open(nil, nonce, message, nil)
	if err != nil {
		return "", ErrMotDePasseIndechiffrable
	}
	return string(clair), nil
}

// aead rend le chiffrement de cette installation, en tirant la clé si elle
// manque et que creer le permet.
//
// Au déchiffrement, creer vaut faux : une clé absente veut dire que le mot de
// passe enregistré ne pourra jamais être relu, et en tirer une neuve donnerait
// l'illusion du contraire.
func (e *Emplacements) aead(creer bool) (cipher.AEAD, string, error) {
	cle, avertissement, err := e.cle(creer)
	if err != nil {
		return nil, "", err
	}

	bloc, err := aes.NewCipher(cle)
	if err != nil {
		return nil, "", fmt.Errorf("initialisation du chiffrement: %w", err)
	}
	aead, err := cipher.NewGCM(bloc)
	if err != nil {
		return nil, "", fmt.Errorf("initialisation du chiffrement: %w", err)
	}
	return aead, avertissement, nil
}

// cle lit la clé de cette installation, ou la tire au premier enregistrement.
//
// À l'enregistrement, une clé d'une autre taille est remplacée comme une clé
// absente. Elle n'a jamais pu servir à AES-256 dans cet état : les mots de
// passe chiffrés avant qu'elle s'abîme sont perdus avec elle, et refuser d'en
// tirer une neuve bloquerait tout enregistrement jusqu'à ce que quelqu'un
// supprime le fichier à la main, sans rien sauver de plus.
//
// Mais jamais en silence, et jamais en l'écrasant. Quelqu'un qui restaure
// ensuite cle.bin depuis une sauvegarde relira ses anciens mots de passe et
// perdra le dernier : l'avertissement rendu le prévient au moment où il
// franchit ce point, et l'ancienne clé, mise de côté en cle.bin.invalide,
// laisse de quoi comprendre ce qui s'est passé.
func (e *Emplacements) cle(creer bool) (cle []byte, avertissement string, err error) {
	chemin := e.fichierCle()

	cle, err = os.ReadFile(chemin) // #nosec G304 — chemin composé depuis la racine de configuration.
	switch {
	case err == nil && len(cle) == tailleCle:
		return cle, "", nil
	case err == nil && !creer:
		return nil, "", ErrMotDePasseIndechiffrable
	case err == nil:
		invalide := chemin + ".invalide"
		if err := os.Rename(chemin, invalide); err != nil {
			return nil, "", fmt.Errorf("mise de cote de %s: %w", chemin, err)
		}
		avertissement = fmt.Sprintf("clé de chiffrement inutilisable remplacée : les mots de passe "+
			"enregistrés auparavant sont perdus, l'ancienne clé est conservée dans %s", invalide)
	case !errors.Is(err, fs.ErrNotExist) && !creer:
		return nil, "", fmt.Errorf("%w: lecture de %s: %w", ErrMotDePasseIndechiffrable, chemin, err)
	case !errors.Is(err, fs.ErrNotExist):
		return nil, "", fmt.Errorf("lecture de %s: %w", chemin, err)
	case !creer:
		return nil, "", ErrCleAbsente
	}

	cle = make([]byte, tailleCle)
	if _, err := rand.Read(cle); err != nil {
		return nil, "", fmt.Errorf("tirage de la cle: %w", err)
	}
	// 0600 comme le reste : cette clé vaut les mots de passe qu'elle protège.
	if err := os.WriteFile(chemin, cle, permFichier); err != nil {
		return nil, "", fmt.Errorf("ecriture de %s: %w", chemin, err)
	}
	return cle, avertissement, nil
}

// fichierCle rend le chemin de la clé de chiffrement.
func (e *Emplacements) fichierCle() string {
	return filepath.Join(e.racine, "cle.bin")
}
