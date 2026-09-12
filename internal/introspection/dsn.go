// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package introspection

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
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

// clesAffichees sont les paramètres dont la valeur reste lisible une fois le
// DSN masqué, dans la forme clé/valeur comme dans la requête d'une URL. Toutes
// les autres valeurs sont masquées, qu'on les sache secrètes ou non.
//
// La liste recense ce qui est inoffensif et non ce qui est secret, à cause de
// la façon dont chacune se trompe : une clé secrète oubliée fait fuir un mot de
// passe sans que personne le voie, une clé inoffensive oubliée donne
// « options=*** », que la première relecture signale.
//
// Une entrée ne s'ajoute que si sa valeur ne peut rien porter d'autre que ce
// que son nom annonce. options n'y est pas, et n'y entrera pas : il transmet au
// serveur des réglages arbitraires, donc n'importe quoi.
var clesAffichees = map[string]bool{
	// L'adresse du serveur figure dans toute trace réseau de la connexion.
	"host":     true,
	"hostaddr": true,
	"port":     true,
	// Un nom de base est un nom d'objet, que le calque produit porte de toute
	// façon dans source.catalogue. database est l'orthographe de SQL Server.
	"dbname":   true,
	"database": true,
	// Le compte reste visible : pgx le cite dans chacune de ses erreurs de
	// connexion, et le masquer ici coûterait du diagnostic sans rien cacher.
	"user": true,
	// Vocabulaires fermés : la valeur ne peut être qu'un mot-clé documenté du
	// pilote. encrypt et trustservercertificate sont ceux de SQL Server.
	"sslmode":                true,
	"target_session_attrs":   true,
	"encrypt":                true,
	"trustservercertificate": true,
	// Des durées, donc des nombres.
	"connect_timeout":    true,
	"connection timeout": true,
	// Un libellé que l'application choisit pour se nommer auprès du serveur,
	// et qu'il affiche lui-même dans la liste des sessions.
	"application_name": true,
	"app name":         true,
	// Une liste de noms de schémas.
	"search_path": true,
}

// espaces sont les blancs que la grammaire clé/valeur de libpq sépare, ceux
// d'isspace en C. Un espace insécable n'en fait pas partie : il appartient à
// la valeur, comme chez pgx.
const espaces = " \t\n\v\f\r"

// SGBDDepuisDSN rend le SGBD désigné par le préfixe du DSN.
//
// C'est la porte d'entrée de toute connexion, et elle refuse ce qu'Ormeau ne
// sait pas lire : une chaîne qu'on ne sait pas lire est une chaîne qu'on ne
// sait pas masquer, et le pilote qui la lirait autrement en citerait des
// morceaux dans ses erreurs.
//
// L'erreur ne cite jamais le DSN, seulement le préfixe : un DSN mal formé
// contient quand même un mot de passe.
func SGBDDepuisDSN(dsn string) (string, error) {
	prefixe, estUneURL := schema(dsn)
	if !estUneURL {
		// Forme clé/valeur de libpq : « host=... password=... ». Pas de
		// préfixe, mais pgx la comprend, et elle ne peut être que postgres.
		if !strings.Contains(dsn, "=") {
			return "", errors.New("dsn sans prefixe, attendu la forme sgbd://")
		}
		if _, err := lireCleValeur(dsn); err != nil {
			return "", fmt.Errorf("forme cle=valeur illisible: %w", err)
		}
		return "postgres", nil
	}

	sgbd, connu := prefixes[strings.ToLower(prefixe)]
	if !connu {
		return "", fmt.Errorf("prefixe de dsn inconnu: %q", prefixe)
	}
	if _, err := lireURL(dsn); err != nil {
		return "", err
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

	if _, estUneURL := schema(dsn); !estUneURL {
		couples, err := lireCleValeur(dsn)
		if err != nil {
			return Connexion{}, fmt.Errorf("chaine de connexion illisible: %w", err)
		}
		return connexionDepuisCouples(couples), nil
	}

	u, err := lireURL(NettoyerDSN(dsn))
	if err != nil {
		return Connexion{}, err
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

// connexionDepuisCouples rassemble les composants de la forme clé/valeur. Les
// clés sont comparées telles quelles, comme pgx le fait : « HOST » n'est pas
// un hôte pour lui, et un profil qui le prendrait pour tel ne décrirait plus la
// même connexion. Une clé répétée garde sa dernière valeur, pour la même
// raison.
func connexionDepuisCouples(couples []couple) Connexion {
	var c Connexion
	for _, p := range couples {
		switch p.cle {
		case "host":
			c.Hote = p.valeur
		case "port":
			if port, err := strconv.Atoi(p.valeur); err == nil {
				c.Port = port
			}
		case "user":
			c.Utilisateur = p.valeur
		case "password":
			c.MotDePasse = p.valeur
		case "dbname":
			c.Base = p.valeur
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
	if _, estUneURL := schema(dsn); !estUneURL {
		couples, _ := lireCleValeur(dsn)
		var base string
		for _, p := range couples {
			if p.cle == "dbname" {
				base = p.valeur
			}
		}
		return base
	}

	u, err := lireURL(dsn)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(u.Path, "/")
}

// AvecBase rend le même DSN pointant sur une autre base. Sert à réutiliser les
// identifiants d'une connexion pour parcourir les bases d'un serveur.
//
// Dans la forme clé/valeur, seule la valeur de dbname est remplacée, sur place :
// réécrire la chaîne entière à partir de sa lecture changerait un mot de passe
// entre apostrophes dont pgx ne relit pas tous les échappements à l'identique.
func AvecBase(dsn, base string) string {
	if _, estUneURL := schema(dsn); !estUneURL {
		couples, err := lireCleValeur(dsn)
		if err != nil {
			return dsn
		}
		nommeLaBase := func(p couple) bool { return p.cle == "dbname" }
		if slices.ContainsFunc(couples, nommeLaBase) {
			return remplacerValeurs(dsn, couples, nommeLaBase, citer(base))
		}
		return dsn + " dbname=" + citer(base)
	}

	u, err := lireURL(dsn)
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
	if _, estUneURL := schema(dsn); !estUneURL || !strings.Contains(dsn, "?") {
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
// sait pas analyser est un DSN dont on ne sait pas où est le secret. Une fois
// analysé, restent lisibles l'hôte, la base, l'utilisateur et les valeurs de
// clesAffichees ; tout le reste est masqué.
func Masquer(dsn string) string {
	if dsn == "" {
		return ""
	}
	if _, estUneURL := schema(dsn); !estUneURL {
		return masquerCleValeur(dsn)
	}

	u, err := lireURL(dsn)
	if err != nil {
		return "***"
	}
	if u.User != nil {
		if _, avecMotDePasse := u.User.Password(); avecMotDePasse {
			u.User = url.UserPassword(u.User.Username(), "***")
		}
	}

	if requete := u.Query(); len(requete) > 0 {
		var modifie bool
		for cle, valeurs := range requete {
			if clesAffichees[strings.ToLower(cle)] {
				continue
			}
			for i := range valeurs {
				valeurs[i] = "***"
			}
			modifie = true
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

// schema rend le préfixe d'une URL, et faux quand la chaîne n'en est pas une.
//
// Une chaîne n'est une URL que si elle commence par un nom de schéma au sens de
// la RFC 3986, suivi de « :// ». Chercher « :// » n'importe où prenait pour une
// URL la forme clé/valeur dont un mot de passe contient la séquence, et le
// « préfixe » cité dans l'erreur portait alors ce mot de passe.
func schema(dsn string) (string, bool) {
	prefixe, _, trouve := strings.Cut(dsn, "://")
	if !trouve || prefixe == "" {
		return "", false
	}
	for i := 0; i < len(prefixe); i++ {
		c := prefixe[i]
		lettre := 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
		suite := '0' <= c && c <= '9' || c == '+' || c == '-' || c == '.'
		if !lettre && (i == 0 || !suite) {
			return "", false
		}
	}
	return prefixe, true
}

// lireURL analyse un DSN de forme URL et refuse ceux dont la lecture est
// ambiguë.
//
// url.Parse accepte presque tout, et c'est le danger : un mot de passe
// contenant « / », « ? » ou « # » coupe la partie identifiants avant
// l'arobase, et sa fin se retrouve dans le chemin, la requête ou le fragment.
// Le pilote lirait alors une base nommée d'après le mot de passe, et la citerait
// dans son erreur de connexion. Un « @ » hors des identifiants suffit à le
// reconnaître.
//
// Les erreurs ne citent jamais la chaîne : celle de url.Parse le fait.
func lireURL(dsn string) (*url.URL, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, errors.New("url de connexion illisible")
	}
	if u.Opaque != "" || strings.Contains(u.Path+u.RawQuery+u.Fragment, "@") {
		return nil, errors.New("url de connexion ambigue, encoder les caracteres reserves du mot de passe")
	}

	// u.Query() écarte en silence une paire mal encodée, que u.String() rend
	// ensuite telle quelle : le masquage ne verrait pas une valeur qui sortirait
	// quand même.
	if _, err := url.ParseQuery(u.RawQuery); err != nil {
		return nil, errors.New("parametres de l'url illisibles")
	}
	for _, paire := range strings.Split(u.RawQuery, "&") {
		if paire == "" {
			continue
		}
		// Une paire sans « = », ou dont le nom contient un espace, est une
		// valeur dont le nom a été oublié : son nom serait affiché, et c'est
		// lui qui porte le secret.
		cle, _, avecEgal := strings.Cut(paire, "=")
		nom, _ := url.QueryUnescape(cle)
		if !avecEgal || strings.ContainsAny(nom, espaces) && !clesAffichees[strings.ToLower(nom)] {
			return nil, errors.New("parametre d'url sans valeur")
		}
	}
	return u, nil
}

// couple est un paramètre de la forme clé/valeur. debut et fin bornent sa
// valeur brute dans la chaîne d'origine, apostrophes comprises.
type couple struct {
	cle, valeur string
	debut, fin  int
}

// lireCleValeur analyse la forme « host=serveur password='un secret' » de
// libpq, selon la grammaire de pgx (parseKeywordValueSettings dans
// pgconn/config.go) : blancs facultatifs autour du « = », valeur entre
// apostrophes, « \' » et « \\ » échappés.
//
// C'est la seule lecture de cette forme dans Ormeau. Un masquage fondé sur une
// autre lecture que celle du pilote masque la mauvaise partie : c'est ce qui a
// laissé passer « password = secret ». Le test de parité avec pgx garde
// l'accord entre les deux.
//
// Elle refuse tout ce que pgx refuse, et deux lectures de plus, où pgx fait
// glisser une valeur dans un nom affiché :
//
//   - un nom contenant un blanc : « password secret = x » donne chez lui la clé
//     « password secret », mot de passe compris ;
//   - une valeur précédée d'un blanc et contenant « = » : dans
//     « user= password=secret », il lit « password=secret » comme le nom
//     d'utilisateur, qu'il cite dans ses erreurs. Entre apostrophes, ou sans
//     blanc après le « = », la valeur reste acceptée.
//
// Les erreurs ne citent jamais la chaîne.
func lireCleValeur(dsn string) ([]couple, error) {
	var couples []couple
	i := passerBlancs(dsn, 0)
	for i < len(dsn) {
		egal := strings.IndexByte(dsn[i:], '=')
		if egal < 0 {
			return nil, errors.New("parametre sans =")
		}
		cle := strings.Trim(dsn[i:i+egal], espaces)
		if cle == "" {
			return nil, errors.New("parametre sans nom")
		}
		if strings.ContainsAny(cle, espaces) {
			return nil, errors.New("nom de parametre contenant un blanc")
		}

		debut := passerBlancs(dsn, i+egal+1)
		fin := debut
		switch {
		case debut == len(dsn):
		case dsn[debut] != '\'':
			for ; fin < len(dsn) && !estBlanc(dsn[fin]); fin++ {
				if dsn[fin] == '\\' {
					if fin++; fin == len(dsn) {
						return nil, errors.New("barre oblique inverse en fin de valeur")
					}
				}
			}
			if debut > i+egal+1 && strings.Contains(dsn[debut:fin], "=") {
				return nil, errors.New("valeur ambigue apres un blanc, la mettre entre apostrophes")
			}
		default:
			// pgx sort de sa boucle au-delà de la chaîne quand elle finit par
			// une barre oblique inverse entre apostrophes, et panique en
			// découpant. Le test fin >= len couvre ce cas.
			for fin = debut + 1; fin < len(dsn) && dsn[fin] != '\''; fin++ {
				if dsn[fin] == '\\' {
					fin++
				}
			}
			if fin >= len(dsn) {
				return nil, errors.New("apostrophe non fermee")
			}
			fin++
		}

		brute := dsn[debut:fin]
		if strings.HasPrefix(brute, "'") {
			brute = brute[1 : len(brute)-1]
		}
		// Le même désechappement que pgx, ordre compris : il n'est pas exact
		// pour toutes les valeurs, mais c'est celui qui décide du mot de passe
		// envoyé au serveur.
		valeur := strings.ReplaceAll(strings.ReplaceAll(brute, `\\`, `\`), `\'`, `'`)

		couples = append(couples, couple{cle: cle, valeur: valeur, debut: debut, fin: fin})
		i = passerBlancs(dsn, fin)
	}
	return couples, nil
}

// estBlanc reconnaît un octet de espaces.
func estBlanc(c byte) bool {
	return strings.IndexByte(espaces, c) >= 0
}

// passerBlancs rend la position du premier octet non blanc à partir de i.
func passerBlancs(s string, i int) int {
	for i < len(s) && estBlanc(s[i]) {
		i++
	}
	return i
}

// remplacerValeurs substitue valeur à la valeur brute de chaque couple retenu,
// et laisse le reste de la chaîne octet pour octet.
func remplacerValeurs(dsn string, couples []couple, retenu func(couple) bool, valeur string) string {
	var b strings.Builder
	precedent := 0
	for _, p := range couples {
		if !retenu(p) {
			continue
		}
		b.WriteString(dsn[precedent:p.debut])
		b.WriteString(valeur)
		precedent = p.fin
	}
	b.WriteString(dsn[precedent:])
	return b.String()
}

// citer rend une valeur écrite pour la forme clé/valeur, entre apostrophes
// quand elle en a besoin.
func citer(valeur string) string {
	if valeur != "" && !strings.ContainsAny(valeur, espaces+`'\`) {
		return valeur
	}
	echappee := strings.ReplaceAll(strings.ReplaceAll(valeur, `\`, `\\`), `'`, `\'`)
	return "'" + echappee + "'"
}

// masquerCleValeur masque la valeur de toute clé absente de clesAffichees et
// laisse le reste intact, blancs compris : un DSN masqué doit rester assez
// lisible pour diagnostiquer. Une valeur vide reste vide, faute de quoi
// « password= » laisserait croire qu'un mot de passe a été donné.
func masquerCleValeur(dsn string) string {
	couples, err := lireCleValeur(dsn)
	if err != nil {
		return "***"
	}
	return remplacerValeurs(dsn, couples, func(p couple) bool {
		return p.debut != p.fin && !clesAffichees[strings.ToLower(p.cle)]
	}, "***")
}
