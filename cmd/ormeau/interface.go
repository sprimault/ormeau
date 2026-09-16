// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	ihm "github.com/sprimault/ormeau/internal/interface"

	// Comme pour extraire : importer le pilote est ce qui le rend disponible.
	_ "github.com/sprimault/ormeau/internal/introspection/postgres"
	_ "github.com/sprimault/ormeau/internal/introspection/sqlserver"
)

// interfaceLocale sert l'interface sur la boucle locale jusqu'à interruption.
//
// Le nom ne peut pas être « interface », qui est un mot-clé.
func interfaceLocale(args []string) error {
	jeu := flag.NewFlagSet("interface", flag.ContinueOnError)
	port := jeu.Int("port", 0, "port d'écoute ; un port libre est choisi par défaut")
	sansNavigateur := jeu.Bool("sans-navigateur", false, "ne pas ouvrir le navigateur au démarrage")
	repertoire := jeu.String("repertoire", "", "répertoire de travail ; celui du lancement par défaut")
	if err := jeu.Parse(args); err != nil {
		return err
	}

	travail, err := resoudreRepertoire(*repertoire)
	if err != nil {
		return err
	}

	emplacements, err := ouvrirConfiguration()
	if err != nil {
		return err
	}

	// Arrêt propre : le processus meurt, le jeton et le cookie avec.
	ctx, arreter := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer arreter()

	return ihm.Servir(ctx, ihm.Options{
		Port:           *port,
		Repertoire:     travail,
		Emplacements:   emplacements,
		Version:        version,
		SansNavigateur: *sansNavigateur,
		Sortie:         os.Stderr,
	})
}

// resoudreRepertoire rend un chemin absolu et vérifié.
//
// Absolu parce que l'interface l'affiche en permanence : un chemin relatif ne
// dirait pas où les fichiers atterrissent, et un decisions.yaml posé dans le
// mauvais projet ne se remarque pas tout de suite.
func resoudreRepertoire(demande string) (string, error) {
	if demande == "" {
		return os.Getwd()
	}

	chemin, err := filepath.Abs(demande)
	if err != nil {
		return "", err
	}
	infos, err := os.Stat(chemin)
	if err != nil {
		return "", fmt.Errorf("repertoire de travail inaccessible: %w", err)
	}
	if !infos.IsDir() {
		return "", fmt.Errorf("%s n'est pas un repertoire", chemin)
	}
	return chemin, nil
}
