// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package introspection

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
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

// Connexion décrit une connexion par ses composants plutôt que par une URL.
//
// Composer une URL à la main est une source d'erreurs : un mot de passe
// contenant « @ » ou « / » casse la chaîne sans que le message le dise. Et
// personne ne connaît par cœur le DSN d'une base qu'il découvre.
//
// MotDePasse ne vient jamais d'un drapeau : il serait visible dans ps et dans
// l'historique du shell.
type Connexion struct {
	SGBD        string
	Hote        string
	Port        int
	Utilisateur string
	MotDePasse  string
	// Base vide signifie « toutes les bases du serveur ». Un calque décrivant
	// une base et une seule, l'appelant en produira alors plusieurs.
	Base string
}

// portsParDefaut évite d'imposer un port que tout le monde connaît.
var portsParDefaut = map[string]int{
	"postgres":  5432,
	"mysql":     3306,
	"mariadb":   3306,
	"sqlserver": 1433,
	"oracle":    1521,
}

// sgbdParPort est écrite à la main plutôt que dérivée de portsParDefaut :
// 3306 y apparaît deux fois, et parcourir une map pour retrouver une clé par sa
// valeur rendrait « mysql » ou « mariadb » selon l'ordre d'itération. Le même
// port donnerait alors deux pilotes d'une exécution à l'autre.
//
// 3306 désigne donc mysql, qui est le pilote à charger. La variante se lit à la
// connexion et c'est elle qui atterrit dans le calque : quelqu'un qui vise un
// serveur MariaDB obtient « mariadb », qu'il ait nommé le SGBD ou non.
var sgbdParPort = map[int]string{
	5432: "postgres",
	3306: "mysql",
	1433: "sqlserver",
	1521: "oracle",
}

// SGBDDepuisPort rend le SGBD que désigne un port, et une chaîne vide pour un
// port inconnu.
//
// C'est le second niveau d'aiguillage, après le préfixe du DSN : l'utilisateur
// fournit des identifiants, pas une configuration. Rien n'est tenté à l'aveugle
// quand le port ne dit rien — chaque essai enverrait des identifiants, et une
// politique qui compte les échecs d'authentification verrouillerait le compte.
func SGBDDepuisPort(port int) string {
	return sgbdParPort[port]
}

// DSN compose une URL à partir des composants. L'encodage est fait par
// net/url : un mot de passe contenant « @ », « / » ou « : » passe sans que
// l'utilisateur ait à s'en soucier.
//
// Le SGBD peut être omis quand le port le désigne sans ambiguïté : c'est le cas
// nominal dans l'interface, où l'on saisit un hôte et un port sans avoir à
// déclarer ce qui écoute derrière.
func (c Connexion) DSN() (string, error) {
	sgbd := strings.ToLower(c.SGBD)
	if sgbd == "" {
		if sgbd = SGBDDepuisPort(c.Port); sgbd == "" {
			return "", errors.New("sgbd indetermine, preciser --sgbd ou un port connu")
		}
	}
	if _, connu := prefixes[sgbd]; !connu {
		return "", fmt.Errorf("sgbd inconnu: %q", c.SGBD)
	}
	if c.Hote == "" {
		return "", errors.New("--hote est requis avec les drapeaux de connexion")
	}

	port := c.Port
	if port == 0 {
		port = portsParDefaut[sgbd]
	}

	u := url.URL{
		Scheme: sgbd,
		Host:   fmt.Sprintf("%s:%d", c.Hote, port),
		Path:   "/" + c.Base,
	}
	if c.Utilisateur != "" {
		if c.MotDePasse != "" {
			u.User = url.UserPassword(c.Utilisateur, c.MotDePasse)
		} else {
			u.User = url.User(c.Utilisateur)
		}
	}
	return u.String(), nil
}

// ConnexionDepuisDSN décompose un DSN en ses composants.
//
// L'inverse de DSN, et pour un besoin précis : enregistrer un profil alors
// qu'on a saisi une chaîne de connexion. Sans cela, le profil ne garderait rien
// et ne servirait à rien la fois suivante.
//
// Le mot de passe est rendu avec le reste : c'est à l'appelant de décider s'il
// le conserve, et lui seul sait si la case était cochée.
//
// La forme « host=serveur dbname=base » de libpq est acceptée au même titre
// qu'une URL : quelqu'un qui la colle attend qu'on la comprenne.
func ConnexionDepuisDSN(dsn string) (Connexion, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return Connexion{}, errors.New("chaine de connexion vide")
	}

	if estCleValeur(dsn) {
		return connexionDepuisCleValeur(dsn), nil
	}

	u, err := url.Parse(NettoyerDSN(dsn))
	if err != nil {
		// L'erreur de url.Parse cite la chaîne entière, mot de passe compris :
		// elle ne peut ni être enveloppée, ni remonter telle quelle.
		return Connexion{}, errors.New("chaine de connexion illisible")
	}

	c := Connexion{
		SGBD: strings.ToLower(u.Scheme),
		Hote: u.Hostname(),
		Base: strings.TrimPrefix(u.Path, "/"),
	}
	if port, err := strconv.Atoi(u.Port()); err == nil {
		c.Port = port
	}
	if u.User != nil {
		c.Utilisateur = u.User.Username()
		c.MotDePasse, _ = u.User.Password()
	}
	// Le préfixe dit quel pilote charger, mais c'est le nom du SGBD qu'un
	// profil doit porter : postgresql:// et postgres:// désignent le même.
	if normalise, connu := prefixes[c.SGBD]; connu {
		c.SGBD = normalise
	}
	return c, nil
}

// connexionDepuisCleValeur décompose la forme « host=serveur password=secret ».
func connexionDepuisCleValeur(dsn string) Connexion {
	var c Connexion
	for _, champ := range strings.Fields(dsn) {
		cle, valeur, trouve := strings.Cut(champ, "=")
		if !trouve {
			continue
		}
		switch strings.ToLower(cle) {
		case "host":
			c.Hote = valeur
		case "port":
			if port, err := strconv.Atoi(valeur); err == nil {
				c.Port = port
			}
		case "user":
			c.Utilisateur = valeur
		case "password":
			c.MotDePasse = valeur
		case "dbname":
			c.Base = valeur
		}
	}
	// Cette forme n'a pas de préfixe : seul le port peut désigner le SGBD, et
	// c'est déjà ce que fait la composition inverse.
	c.SGBD = SGBDDepuisPort(c.Port)
	return c
}

// BaseDuDSN rend la base désignée par un DSN, ou une chaîne vide s'il n'en
// nomme aucune — auquel cas l'appelant a affaire à un serveur entier.
func BaseDuDSN(dsn string) string {
	if estCleValeur(dsn) {
		for _, champ := range strings.Fields(dsn) {
			if cle, valeur, trouve := strings.Cut(champ, "="); trouve && strings.EqualFold(cle, "dbname") {
				return valeur
			}
		}
		return ""
	}

	u, err := url.Parse(dsn)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(u.Path, "/")
}

// AvecBase rend le même DSN pointant sur une autre base. Sert à réutiliser les
// identifiants d'une connexion pour parcourir les bases d'un serveur.
func AvecBase(dsn, base string) string {
	if estCleValeur(dsn) {
		champs := strings.Fields(dsn)
		for i, champ := range champs {
			if cle, _, trouve := strings.Cut(champ, "="); trouve && strings.EqualFold(cle, "dbname") {
				champs[i] = cle + "=" + base
				return strings.Join(champs, " ")
			}
		}
		return dsn + " dbname=" + base
	}

	u, err := url.Parse(dsn)
	if err != nil {
		return dsn
	}
	u.Path = "/" + base
	return u.String()
}

// parametresDoctrine sont les paramètres que Doctrine DBAL ajoute à un
// DATABASE_URL et qu'aucun SGBD ne connaît.
//
// L'utilisateur d'Ormeau est un développeur Symfony : il colle son
// DATABASE_URL. Les laisser passer fait rejeter la connexion par le serveur —
// « unrecognized configuration parameter » —, message sans rapport visible avec
// ce qu'il vient de faire.
var parametresDoctrine = map[string]bool{
	"serverversion": true,
	"charset":       true,
	"driveroptions": true,
	"defaultdbname": true,
}

// NettoyerDSN retire les paramètres propres à Doctrine et rend le DSN tel
// qu'un pilote l'attend.
//
// Retire seulement ceux-là : sslmode, search_path ou application_name sont
// légitimes et doivent atteindre le serveur. Un filtrage par liste blanche
// casserait les options que ce code ne connaît pas encore.
func NettoyerDSN(dsn string) string {
	if !strings.Contains(dsn, "://") || !strings.Contains(dsn, "?") {
		return dsn
	}

	u, err := url.Parse(dsn)
	if err != nil {
		// Illisible ici, il le sera aussi pour le pilote, dont le message sera
		// plus précis que ce que nous pourrions dire.
		return dsn
	}

	requete := u.Query()
	var retire bool
	for cle := range requete {
		if parametresDoctrine[strings.ToLower(cle)] {
			requete.Del(cle)
			retire = true
		}
	}
	if !retire {
		return dsn
	}

	u.RawQuery = requete.Encode()
	return u.String()
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

	// url.String() encode les astérisques de la partie identifiants, ce qui
	// donne un %2A%2A%2A illisible dans un message d'erreur. Le remplacement
	// porte sur une séquence que l'encodage vient de produire, jamais sur le
	// secret lui-même — celui-ci a déjà disparu.
	return strings.ReplaceAll(u.String(), "%2A%2A%2A", "***")
}

// estCleValeur reconnaît la forme « host=serveur password=secret » de libpq,
// qui n'est pas une URL et que url.Parse accepterait sans rien en tirer.
func estCleValeur(dsn string) bool {
	return !strings.Contains(dsn, "://") && strings.Contains(dsn, "=")
}

// masquerCleValeur remplace la valeur des clés secrètes et laisse le reste
// intact : un DSN masqué doit rester assez lisible pour diagnostiquer.
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

// estCleSecrete reconnaît les clés portant un mot de passe. Les orthographes des
// différents dialectes sont toutes listées : en manquer une laisse fuir un
// secret dans un journal.
func estCleSecrete(cle string) bool {
	switch strings.ToLower(cle) {
	case "password", "passwd", "pwd", "sslpassword":
		return true
	}
	return false
}
