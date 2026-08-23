// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package introspection

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// Le DSN est le seul secret que l'outil manipule. Il ne transite ni dans les
// journaux, ni dans les messages d'erreur, ni dans le calque : tout ce qui
// pourrait le rendre visible passe par Masquer.

// prefixes associe un préfixe de DSN au SGBD du vocabulaire fermé de
// calque.Source. Plusieurs préfixes peuvent désigner le même : personne
// n'écrit « postgres:// » deux fois de la même façon.
var prefixes = map[string]string{
	"postgres":   "postgres",
	"postgresql": "postgres",
	"mysql":      "mysql",
	"mariadb":    "mariadb",
	"sqlserver":  "sqlserver",
	"mssql":      "sqlserver",
	"oracle":     "oracle",
	"sqlite":     "sqlite",
}

// SGBDDepuisDSN rend le SGBD désigné par le préfixe du DSN.
//
// L'erreur ne cite jamais le DSN, seulement le préfixe : un DSN mal formé
// contient quand même un mot de passe.
func SGBDDepuisDSN(dsn string) (string, error) {
	prefixe, _, trouve := strings.Cut(dsn, "://")
	if !trouve {
		// Forme clé/valeur de libpq : « host=... password=... ». Pas de
		// préfixe, mais pgx la comprend, et elle ne peut être que postgres.
		if estCleValeur(dsn) {
			return "postgres", nil
		}
		return "", errors.New("dsn sans prefixe, attendu la forme sgbd://")
	}

	sgbd, connu := prefixes[strings.ToLower(prefixe)]
	if !connu {
		return "", fmt.Errorf("prefixe de dsn inconnu: %q", prefixe)
	}
	return sgbd, nil
}

// Masquer rend un DSN affichable dans un journal ou un message d'erreur.
//
// En cas de doute, masque tout plutôt que de laisser passer : un DSN qu'on ne
// sait pas analyser est un DSN dont on ne sait pas où est le secret.
func Masquer(dsn string) string {
	if dsn == "" {
		return ""
	}
	if estCleValeur(dsn) {
		return masquerCleValeur(dsn)
	}
	// url.Parse accepte à peu près n'importe quoi sans erreur : sans « :// »,
	// on n'a pas affaire à un DSN qu'on sait analyser, donc on ne montre rien.
	if !strings.Contains(dsn, "://") {
		return "***"
	}

	u, err := url.Parse(dsn)
	if err != nil {
		return "***"
	}
	if u.User != nil {
		if _, avecMotDePasse := u.User.Password(); avecMotDePasse {
			u.User = url.UserPassword(u.User.Username(), "***")
		}
	}

	// Certains paramètres portent aussi un secret : sslpassword côté
	// PostgreSQL, password côté SQL Server quand il est passé en requête.
	if requete := u.Query(); len(requete) > 0 {
		var modifie bool
		for cle := range requete {
			if estCleSecrete(cle) {
				requete.Set(cle, "***")
				modifie = true
			}
		}
		if modifie {
			u.RawQuery = requete.Encode()
		}
	}
	return u.String()
}

// estCleValeur reconnaît la forme « host=serveur password=secret » de libpq,
// qui n'est pas une URL et que url.Parse accepterait sans rien en tirer.
func estCleValeur(dsn string) bool {
	return !strings.Contains(dsn, "://") && strings.Contains(dsn, "=")
}

func masquerCleValeur(dsn string) string {
	champs := strings.Fields(dsn)
	for i, champ := range champs {
		cle, _, trouve := strings.Cut(champ, "=")
		if trouve && estCleSecrete(cle) {
			champs[i] = cle + "=***"
		}
	}
	return strings.Join(champs, " ")
}

func estCleSecrete(cle string) bool {
	switch strings.ToLower(cle) {
	case "password", "passwd", "pwd", "sslpassword":
		return true
	}
	return false
}
