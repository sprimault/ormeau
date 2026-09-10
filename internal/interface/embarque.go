// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
)

// embarque contient le bundle produit par Vite. Le préfixe « all: » inclut les
// fichiers commençant par un point ou un souligné, que Vite peut produire.
//
// Le répertoire n'est pas versionné : `make web-build` l'écrit, et toutes les
// cibles Go en dépendent. Sans lui la compilation échoue, ce qui est le
// comportement voulu — un embarqué vide passerait la compilation et produirait
// cinq binaires à interface blanche sans qu'aucun avertissement ne le signale.
//
//go:embed all:embarque
var embarque embed.FS

// frontal sert le bundle, avec repli sur index.html.
//
// Le repli existe pour les écrans à venir : rechargée sur /arbitrage, la page
// doit revenir sur l'application et non sur un 404. Un dernier segment
// contenant un point désigne un fichier et n'est jamais replié, sans quoi une
// ressource absente rendrait du HTML sous un nom de script.
func frontal() (http.Handler, error) {
	racine, err := fs.Sub(embarque, "embarque")
	if err != nil {
		return nil, fmt.Errorf("lecture du front embarque: %w", err)
	}

	index, err := fs.ReadFile(racine, "index.html")
	if err != nil {
		return nil, fmt.Errorf("index.html absent du front embarque: %w", err)
	}

	fichiers := http.FileServer(http.FS(racine))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chemin := strings.TrimPrefix(r.URL.Path, "/")
		if chemin == "" || !strings.Contains(dernierSegment(chemin), ".") {
			servirIndex(w, index)
			return
		}
		fichiers.ServeHTTP(w, r)
	}), nil
}

// servirIndex écrit la page d'entrée sans autoriser sa mise en cache : elle
// porte les noms des bundles, qui changent à chaque construction. Un navigateur
// qui garderait l'ancienne demanderait des fichiers qui n'existent plus.
func servirIndex(w http.ResponseWriter, index []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(index)
}

// dernierSegment rend ce qui suit le dernier séparateur, le chemin entier s'il
// n'y en a pas.
func dernierSegment(chemin string) string {
	if i := strings.LastIndex(chemin, "/"); i >= 0 {
		return chemin[i+1:]
	}
	return chemin
}
