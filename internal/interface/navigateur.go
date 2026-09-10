// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"fmt"
	"os/exec"
	"runtime"
)

// lancerNavigateur ouvre l'URL dans le navigateur par défaut du poste.
//
// C'est la seule partie du paquet qui dépende de la plateforme. L'URL est
// composée par le serveur — son origine, et un jeton en base64 sans caractère
// d'échappement — donc jamais une entrée à valider.
//
// Sous Windows, rundll32 plutôt que « cmd /c start » : ce dernier traite « & »
// comme un séparateur de commandes, et l'URL en contient un dès qu'elle porte
// deux paramètres.
func lancerNavigateur(url string) error {
	var commande string
	var args []string

	switch runtime.GOOS {
	case "windows":
		commande, args = "rundll32", []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		commande, args = "open", []string{url}
	default:
		commande, args = "xdg-open", []string{url}
	}

	// Ni la commande ni ses arguments ne viennent de l'extérieur : la première
	// est choisie par un switch sur la plateforme, le second est l'URL que le
	// serveur vient de composer.
	if err := exec.Command(commande, args...).Start(); err != nil { // #nosec G204
		return fmt.Errorf("ouverture de %s: %w", commande, err)
	}
	return nil
}
