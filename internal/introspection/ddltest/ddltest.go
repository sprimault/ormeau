// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Package ddltest vérifie qu'une base de test vient bien du DDL du dépôt.
//
// Le conteneur ne charge son DDL qu'à la création : un fichier modifié après
// coup laisse la base sur l'ancien schéma sans rien signaler, et un DSN qui
// vise une autre base ferait tester n'importe quoi. Chaque base porte donc
// l'empreinte du DDL qui l'a créée — commentaire de base sous PostgreSQL,
// propriété étendue sous SQL Server —, et les tests la comparent au fichier
// avant de lancer quoi que ce soit.
//
// Seule la lecture de l'empreinte dépend du dialecte ; la comparaison et ses
// messages sont ici. Seuls des tests importent ce paquet, il n'entre pas dans
// le binaire.
package ddltest

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// prefixe marque une empreinte de DDL, et la distingue d'un commentaire posé
// par quelqu'un d'autre.
const prefixe = "ddl-sha256:"

// Controler compare l'empreinte lue en base, nil si la base n'en porte pas, à
// celle du DDL du dépôt.
func Controler(lue *string, cheminDDL string) error {
	contenu, err := os.ReadFile(cheminDDL) // #nosec G304 — chemin fixé par le test appelant, sous tests/ddl/.
	if err != nil {
		return fmt.Errorf("lecture du DDL de test: %w", err)
	}
	somme := sha256.Sum256(contenu)
	attendue := prefixe + hex.EncodeToString(somme[:])

	switch {
	case lue == nil || !strings.HasPrefix(*lue, prefixe):
		return errors.New("la base visee ne vient pas de tests/ddl/, elle ne porte aucune empreinte de DDL: " +
			"verifier le DSN de test, ou recreer le conteneur par make containers")
	case *lue != attendue:
		return fmt.Errorf("la base de test vient d'un autre DDL que tests/ddl/%s (base %s, fichier %s): "+
			"make containers la recree", filepath.Base(cheminDDL), *lue, attendue)
	}
	return nil
}
