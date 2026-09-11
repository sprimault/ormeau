// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Package ihm sert l'interface locale de sélection et d'arbitrage.
//
// Le nom du paquet diverge de son répertoire, internal/interface, parce que
// « interface » est un mot-clé Go et ne peut pas nommer un paquet. Le répertoire
// garde le nom du domaine, que les règles et le Makefile désignent déjà.
//
// L'interface tourne sur le poste, comme DBeaver ou SSMS, et joint la base par
// le réseau. Elle n'est ni un client SQL ni un explorateur de données : aucune
// chaîne SQL ne transite depuis le navigateur, l'API expose des points d'entrée
// fixes. Et le binaire n'ouvre que deux connexions, la base introspectée et
// 127.0.0.1 — ni télémétrie, ni rapport d'erreur, ni vérification de version.
package ihm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// delaiEnTetes borne la lecture des en-têtes d'une requête. Sans lui, une
// connexion ouverte qui n'envoie rien immobilise un descripteur.
const delaiEnTetes = 10 * time.Second

// delaiArret laisse aux requêtes en cours le temps de finir avant que le
// processus rende la main.
const delaiArret = 5 * time.Second

// Options porte ce que la ligne de commande décide. Le paquet ne lit ni
// drapeau ni variable d'environnement : tout arrive par ici.
type Options struct {
	// Port vaut zéro pour laisser le système en attribuer un libre, ce qui est
	// le mode nominal. Le fixer ne sert qu'à garder un signet stable en
	// développement.
	Port int
	// Repertoire est celui où seront écrits les fichiers de décisions. Il est
	// affiché en permanence par l'interface : on doit savoir où on écrit avant
	// de cliquer.
	Repertoire string
	// Version est celle du binaire, affichée à côté du lien vers les releases.
	Version string
	// SansNavigateur laisse l'utilisateur ouvrir l'URL lui-même.
	SansNavigateur bool
	// Sortie reçoit les lignes de démarrage, et rien d'autre. C'est la seule
	// trace du jeton, qui n'entre ni dans un fichier ni dans un journal.
	Sortie io.Writer
}

// serveur est l'état d'un lancement : deux secrets, les connexions ouvertes, et
// l'origine attendue des appels d'API.
type serveur struct {
	acces       *acces
	registre    *registre
	extractions *extractions
	calques     calquesLus
	// ecriture sérialise les enregistrements de fichiers de décisions.
	ecriture   sync.Mutex
	repertoire string
	version    string
	origine    string
}

// Servir écoute sur la boucle locale et rend la main quand le contexte est
// annulé, les connexions refermées.
//
// L'écoute est sur 127.0.0.1 et il n'existe pas d'option pour en changer :
// quelqu'un finirait par exposer sur un réseau une console qui ouvre des
// connexions vers des bases de production.
func Servir(ctx context.Context, o Options) error {
	ecouteur, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", o.Port))
	if err != nil {
		return fmt.Errorf("ecoute sur 127.0.0.1: %w", err)
	}

	// Relu plutôt que repris de o.Port : avec 0, c'est le système qui tranche,
	// et c'est ce numéro-là qu'il faut annoncer et attendre en Origin.
	port := ecouteur.Addr().(*net.TCPAddr).Port

	acces, err := nouvelAcces()
	if err != nil {
		_ = ecouteur.Close()
		return err
	}

	// Les extractions ont leur propre contexte, annulé à la sortie quelle qu'en
	// soit la raison : une panne d'écoute ne doit pas laisser tourner une tâche
	// que plus personne ne peut suivre.
	taches, arreterTaches := context.WithCancel(ctx)

	s := &serveur{
		acces:       acces,
		registre:    nouveauRegistre(),
		extractions: nouvellesExtractions(taches, o.Repertoire),
		repertoire:  o.Repertoire,
		version:     o.Version,
		origine:     fmt.Sprintf("http://127.0.0.1:%d", port),
	}
	defer s.registre.toutFermer()
	defer s.extractions.attendre()
	defer arreterTaches()

	routeur, err := s.routes()
	if err != nil {
		_ = ecouteur.Close()
		return err
	}

	s.annoncer(o, fmt.Sprintf("%s/entrer?jeton=%s", s.origine, acces.jetonURL))

	serveurHTTP := &http.Server{Handler: routeur, ReadHeaderTimeout: delaiEnTetes}

	fini := make(chan error, 1)
	go func() { fini <- serveurHTTP.Serve(ecouteur) }()

	select {
	case err := <-fini:
		if errors.Is(err, http.ErrServerClosed) || errors.Is(err, net.ErrClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		arret, annuler := context.WithTimeout(context.Background(), delaiArret)
		defer annuler()
		return serveurHTTP.Shutdown(arret)
	}
}

// annoncer écrit les lignes de démarrage et ouvre le navigateur.
//
// L'URL jetonnée n'est affichée que si personne ne l'ouvre à la place de
// l'utilisateur : quand le navigateur s'en charge, elle n'a pas à traîner dans
// le terminal. Un navigateur qu'on ne sait pas ouvrir n'est pas une panne, on
// affiche l'URL et on continue de servir.
func (s *serveur) annoncer(o Options, url string) {
	ouverte := false
	if !o.SansNavigateur {
		ouverte = ouvrirNavigateur(url) == nil
	}

	if o.Sortie == nil {
		return
	}
	// Un terminal qui n'accepte plus rien n'empêche pas de servir l'interface :
	// ces trois lignes sont un confort de démarrage, pas une sortie utile.
	_, _ = fmt.Fprintf(o.Sortie, "Interface sur %s\n", s.origine)
	_, _ = fmt.Fprintf(o.Sortie, "Répertoire de travail : %s\n", s.repertoire)
	if !ouverte {
		_, _ = fmt.Fprintf(o.Sortie, "Ouvrir : %s\n", url)
	}
}

// routes construit le routeur. Le front est servi sans jeton — c'est du
// statique, qu'un site tiers ne peut de toute façon pas lire — et l'API ne
// l'est jamais.
func (s *serveur) routes() (http.Handler, error) {
	front, err := frontal()
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/entrer", s.entrer)
	mux.Handle("/api/contexte", s.protegerAPI(http.HandlerFunc(s.contexte)))
	mux.Handle("/api/connexion", s.protegerAPI(http.HandlerFunc(s.connexion)))
	mux.Handle("/api/bases", s.protegerAPI(http.HandlerFunc(s.bases)))
	mux.Handle("/api/base", s.protegerAPI(http.HandlerFunc(s.basculerBase)))
	mux.Handle("/api/inventaire", s.protegerAPI(http.HandlerFunc(s.inventaire)))
	mux.Handle("/api/colonnes", s.protegerAPI(http.HandlerFunc(s.colonnes)))
	mux.Handle("/api/extractions", s.protegerAPI(http.HandlerFunc(s.gererExtractions)))
	mux.Handle("/api/extractions/evenements", s.protegerAPI(http.HandlerFunc(s.suivreExtractions)))
	mux.Handle("/api/calque", s.protegerAPI(http.HandlerFunc(s.calqueDeSession)))
	mux.Handle("/api/decisions", s.protegerAPI(http.HandlerFunc(s.decisions)))
	mux.Handle("/api/inference", s.protegerAPI(http.HandlerFunc(s.inferer)))
	mux.Handle("/api/inference/entite", s.protegerAPI(http.HandlerFunc(s.entite)))
	mux.Handle("/", s.canoniser(front))
	return mux, nil
}

// entrer échange le jeton d'URL contre le cookie de session, puis redirige sur
// la racine pour que le jeton disparaisse de la barre d'adresse.
func (s *serveur) entrer(w http.ResponseWriter, r *http.Request) {
	if !s.acces.consommer(r.URL.Query().Get("jeton")) {
		http.Error(w, "jeton invalide, expiré ou déjà utilisé", http.StatusForbidden)
		return
	}

	// HttpOnly interdit l'accès par script. SameSite=Strict bloque les requêtes
	// déclenchées depuis un autre site, qui est le vrai vecteur ici : une page
	// malveillante ouverte dans le même navigateur peut viser 127.0.0.1.
	//
	// Pas de Secure, et c'est délibéré : l'interface est servie en HTTP sur la
	// boucle locale, un cookie marqué Secure n'y serait jamais renvoyé et la
	// session ne tiendrait pas.
	http.SetCookie(w, &http.Cookie{ // #nosec G124
		Name:     nomCookie,
		Value:    s.acces.jetonSession,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// canoniser renvoie sur 127.0.0.1 celui qui est arrivé par « localhost ».
//
// Les deux désignent la boucle locale, mais l'API n'accepte qu'une origine, et
// une page chargée depuis localhost verrait tous ses appels refusés sans qu'un
// message n'explique pourquoi.
func (s *serveur) canoniser(suivant http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hote, port, err := net.SplitHostPort(r.Host)
		if err != nil || hote == "127.0.0.1" {
			suivant.ServeHTTP(w, r)
			return
		}

		// Composée champ par champ plutôt que concaténée : l'hôte reste le
		// nôtre quoi qu'apporte la requête, et le chemin est échappé par
		// net/url. Seuls le chemin et la requête sont repris de l'appelant.
		cible := url.URL{
			Scheme:   "http",
			Host:     net.JoinHostPort("127.0.0.1", port),
			Path:     r.URL.Path,
			RawQuery: r.URL.RawQuery,
		}
		http.Redirect(w, r, cible.String(), http.StatusSeeOther)
	})
}

// protegerAPI filtre ce qui atteint l'API : origine attendue, puis cookie de
// session.
//
// L'origine est contrôlée avant tout le reste, contre le DNS rebinding — un nom
// de domaine tiers qui résout vers 127.0.0.1 arrive avec son propre Origin, et
// c'est ce qui le démasque. Elle est exigée sur tout ce qui n'est pas une
// lecture : les navigateurs l'omettent sur certains GET de même origine, jamais
// sur un POST.
func (s *serveur) protegerAPI(suivant http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origine := r.Header.Get("Origin")
		lecture := r.Method == http.MethodGet || r.Method == http.MethodHead
		if (origine == "" && !lecture) || (origine != "" && origine != s.origine) {
			repondreErreur(w, http.StatusForbidden, "origine refusée")
			return
		}

		cookie, err := r.Cookie(nomCookie)
		if err != nil || !s.acces.valideSession(cookie.Value) {
			repondreErreur(w, http.StatusUnauthorized, "session absente ou expirée")
			return
		}
		suivant.ServeHTTP(w, r)
	})
}

// ouvrirNavigateur est une variable pour que les tests la remplacent : lancer un
// navigateur depuis une suite de tests n'apprendrait rien et ouvrirait des
// fenêtres.
var ouvrirNavigateur = lancerNavigateur
