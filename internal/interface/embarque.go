// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
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

// ancreHTML et ancreTete sont les deux points d'insertion des préférences dans
// la page produite par Vite.
//
// Un remplacement de chaîne sur notre propre sortie de construction, et non sur
// une donnée reçue : ce qui vient du fichier de préférences passe par les
// actions du template, qui les échappe selon leur contexte.
const (
	ancreHTML = `<html lang="fr">`
	ancreTete = `</head>`
)

// pageHTML est ce que le template pose sur la page avant le premier rendu.
//
// Sans cette injection, le front lirait ses préférences après l'hydratation, et
// l'interface s'afficherait en clair puis basculerait, les panneaux à leur
// taille par défaut puis à la bonne. Ce que le port dynamique interdit de
// régler côté navigateur : l'origine change à chaque lancement, et le
// localStorage avec elle.
type pageHTML struct {
	Theme  string
	Langue string
	// ApercuOuvert vaut « 1 », « 0 », ou reste vide quand rien n'a été réglé —
	// l'attribut n'est alors pas écrit et le front garde son défaut.
	ApercuOuvert   string
	LargeurArbre   int
	HauteurApercu  int
	LargeurEntites int
}

// fragmentHTML porte la langue, le thème et le repli de l'aperçu.
const fragmentHTML = `<html lang="{{.Langue}}" data-theme="{{.Theme}}"` +
	`{{if .ApercuOuvert}} data-apercu-ouvert="{{.ApercuOuvert}}"{{end}}>`

// fragmentStyle porte les tailles de panneaux, en variables CSS plutôt qu'en
// données à relire : la mise en page est juste dès le premier octet, sans
// attendre que React s'exécute.
//
// Une taille absente n'écrit pas sa variable, et le composant garde la sienne.
// Les valeurs sont des entiers, donc rien d'injectable, et elles sont déjà
// bornées à la lecture du fichier.
//
// Les espaces autour des accolades du sélecteur ne sont pas cosmétiques :
// collée à une action, une accolade CSS donne « {{{ », que l'analyseur de
// template lit comme une action mal formée.
const fragmentStyle = `<style>:root { ` +
	`{{if .LargeurArbre}}--ormeau-largeur-arbre:{{.LargeurArbre}}px;{{end}}` +
	`{{if .HauteurApercu}}--ormeau-hauteur-apercu:{{.HauteurApercu}}px;{{end}}` +
	`{{if .LargeurEntites}}--ormeau-largeur-entites:{{.LargeurEntites}}px;{{end}}` +
	` }</style></head>`

// frontal sert le bundle, avec repli sur index.html.
//
// Le repli existe pour les écrans à venir : rechargée sur /arbitrage, la page
// doit revenir sur l'application et non sur un 404. Un dernier segment
// contenant un point désigne un fichier et n'est jamais replié, sans quoi une
// ressource absente rendrait du HTML sous un nom de script.
func (s *serveur) frontal() (http.Handler, error) {
	racine, err := fs.Sub(embarque, "embarque")
	if err != nil {
		return nil, fmt.Errorf("lecture du front embarque: %w", err)
	}

	index, err := indexTemplate(racine)
	if err != nil {
		return nil, err
	}

	fichiers := http.FileServer(http.FS(racine))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chemin := strings.TrimPrefix(r.URL.Path, "/")
		if chemin == "" || !strings.Contains(dernierSegment(chemin), ".") {
			s.servirIndex(w, index)
			return
		}
		fichiers.ServeHTTP(w, r)
	}), nil
}

// indexTemplate rend la page d'entrée prête à recevoir les préférences.
//
// Les deux ancres manquantes sont une erreur de démarrage et non un repli
// silencieux : elles viennent de notre propre index.html, et leur disparition
// veut dire que la page a changé sans que l'injection suive. Servir la page
// telle quelle laisserait un thème qui ne tient pas, sans rien pour le
// rattacher à ce changement.
func indexTemplate(racine fs.FS) (*template.Template, error) {
	brut, err := fs.ReadFile(racine, "index.html")
	if err != nil {
		return nil, fmt.Errorf("index.html absent du front embarque: %w", err)
	}

	page := string(brut)
	insertions := []struct{ ancre, remplacement string }{
		{ancreHTML, fragmentHTML},
		{ancreTete, fragmentStyle},
	}
	for _, i := range insertions {
		if !strings.Contains(page, i.ancre) {
			return nil, fmt.Errorf("index.html ne porte pas l'ancre %s", i.ancre)
		}
		page = strings.Replace(page, i.ancre, i.remplacement, 1)
	}

	t, err := template.New("index").Parse(page)
	if err != nil {
		return nil, fmt.Errorf("analyse de index.html: %w", err)
	}
	return t, nil
}

// servirIndex écrit la page d'entrée sans autoriser sa mise en cache : elle
// porte les noms des bundles, qui changent à chaque construction, et désormais
// les préférences, qui changent à chaque réglage.
//
// La page est rendue en mémoire avant d'être écrite : un échec à mi-course
// laisserait sinon un document tronqué derrière un statut 200.
func (s *serveur) servirIndex(w http.ResponseWriter, index *template.Template) {
	var rendue bytes.Buffer
	if err := index.Execute(&rendue, s.pagePreferences()); err != nil {
		repondreErreur(w, http.StatusInternalServerError, "rendu de la page: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(rendue.Bytes())
}

// dernierSegment rend ce qui suit le dernier séparateur, le chemin entier s'il
// n'y en a pas.
func dernierSegment(chemin string) string {
	if i := strings.LastIndex(chemin, "/"); i >= 0 {
		return chemin[i+1:]
	}
	return chemin
}
