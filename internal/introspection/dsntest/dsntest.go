// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Package dsntest fournit les chaînes de connexion qu'aucune sortie d'Ormeau
// ne doit laisser fuir.
//
// Une seule table, lue par les tests de chaque couche qui touche un DSN : la
// grammaire et le masquage dans introspection, la sortie d'erreur de
// ormeau extraire, la réponse de l'API de connexion. Un cas ajouté ici est
// vérifié partout. Seuls des tests importent ce paquet, il n'entre pas dans le
// binaire.
package dsntest

// Secret est le mot de passe que porte chaque cas. Aucune sortie ne doit le
// contenir, masquée ou non.
const Secret = "Secret123"

// Cas est une chaîne de connexion piégeuse.
//
// Les hôtes désignent 127.0.0.1, port 1, sauf quand le port est l'objet du
// cas : une chaîne que la lecture accepte part vers le réseau, et le refus
// immédiat d'un port fermé garde les tests rapides sans serveur.
type Cas struct {
	Nom string
	DSN string
	// Lisible dit si Ormeau doit comprendre la chaîne. Une chaîne illisible est
	// refusée avant d'atteindre un pilote, et se masque en entier.
	Lisible bool
}

// Table rassemble les cas. Chacun est là parce qu'une lecture naïve s'y trompe.
var Table = []Cas{
	// Forme clé/valeur de libpq.
	{"cle/valeur", "host=127.0.0.1 port=1 user=u password=Secret123 dbname=gescom", true},
	{"blancs autour du egal", "host=127.0.0.1 port=1 password = Secret123 dbname=gescom", true},
	{"apostrophes et espaces", "host=127.0.0.1 port=1 password='Un Secret123 avec espaces' dbname=gescom", true},
	{"doubles espaces entre apostrophes", "host=127.0.0.1 port=1 password='Secret123  x' dbname=gescom", true},
	{"apostrophe echappee", `host=127.0.0.1 port=1 password='l\'Secret123' dbname=gescom`, true},
	{"barre oblique inverse echappee", `host=127.0.0.1 port=1 password=Secret123\\x dbname=gescom`, true},
	{"espace echappe", `host=127.0.0.1 port=1 password=Un\ Secret123 dbname=gescom`, true},
	{"tabulations et sauts de ligne", "host=127.0.0.1\tport=1\n password=\tSecret123 dbname=gescom", true},
	{"cle repetee", "host=127.0.0.1 port=1 password=x password=Secret123", true},
	{"egal final d'un mot de passe", "host=127.0.0.1 port=1 password=Secret123==", true},
	// pgx lit « password=Secret123 » comme la valeur de user, qu'il affiche.
	{"valeur vide avant le secret", "host=127.0.0.1 port=1 user= password=Secret123", false},
	{"cle en majuscules", "host=127.0.0.1 port=1 PASSWORD=Secret123", true},
	{"sslpassword", "host=127.0.0.1 port=1 sslpassword=Secret123", true},
	{"secret dans une cle inconnue", "host=127.0.0.1 port=1 options='-c Secret123'", true},
	{"sequence de schema dans la valeur", "host=127.0.0.1 port=1 password=Secret123://x dbname=gescom", true},
	{"port illisible", "host=127.0.0.1 password = Secret123 port=abc", true},
	{"apostrophe non fermee", "host=127.0.0.1 port=1 password='Secret123 dbname=gescom", false},
	// pgx panique sur celle-ci au lieu de la refuser.
	{"barre oblique inverse finale entre apostrophes", `host=127.0.0.1 port=1 password='Secret123\`, false},
	{"barre oblique inverse finale", `host=127.0.0.1 port=1 password=Secret123\`, false},
	{"parametre sans egal", "host=127.0.0.1 port=1 password Secret123", false},
	{"nom portant le secret", "host=127.0.0.1 port=1 password Secret123 = x", false},
	{"parametre sans nom", "host=127.0.0.1 port=1 =Secret123", false},
	{"pas un dsn", "pas du tout un dsn Secret123", false},

	// URL.
	{"url", "postgres://u:Secret123@127.0.0.1:1/gescom", true},
	{"url sans base", "postgres://u:Secret123@127.0.0.1:1", true},
	{"url schema en majuscules", "POSTGRES://u:Secret123@127.0.0.1:1/gescom", true},
	{"url mot de passe avec arobase", "postgres://u:Secret123@x@127.0.0.1:1/gescom", true},
	{"url sslpassword en requete", "postgres://u:p@127.0.0.1:1/gescom?sslpassword=Secret123", true},
	{"url secret dans un parametre inconnu", "postgres://u@127.0.0.1:1/gescom?sslmode=disable&options=Secret123", true},
	{"url mot de passe avec barre oblique", "postgres://u:123/Secret123@127.0.0.1:1/gescom", false},
	{"url mot de passe avec diese", "postgres://u:123#Secret123@127.0.0.1:1/gescom", false},
	{"url mot de passe avec point d'interrogation", "postgres://u:123?Secret123@127.0.0.1:1/gescom", false},
	{"url parametre mal encode", "postgres://u@127.0.0.1:1/gescom?sslpassword=%zzSecret123", false},
	{"url parametre sans valeur", "postgres://u@127.0.0.1:1/gescom?Secret123", false},
	{"url caractere de controle", "postgres://u:Secret123@127.0.0.1:1/gescom\x7f", false},
	{"url sqlserver", "sqlserver://sa:Secret123@127.0.0.1:1?database=gescom", true},
	{"url sqlserver mot de passe en requete", "sqlserver://sa@127.0.0.1:1?database=gescom&password=Secret123", true},
	{"url mysql", "mysql://root:Secret123@127.0.0.1:1/gescom", true},
	{"url oracle", "oracle://systeme:Secret123@127.0.0.1:1/ORCL", true},
}
