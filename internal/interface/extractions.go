// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package ihm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/sprimault/ormeau/internal/calque"
	"github.com/sprimault/ormeau/internal/introspection"
)

// maxExtractionsEnCours plafonne les extractions simultanées. Chacune ouvre sa
// propre connexion sur le serveur du client, et une base de production n'a pas
// à encaisser sans limite ce qu'on lui demande depuis un poste. Au-delà, les
// demandes attendent leur tour plutôt que d'être refusées.
const maxExtractionsEnCours = 2

// maxExtractionsFinies borne ce qu'on garde des tâches terminées, échouées ou
// annulées. C'est l'historique d'un lancement, pas un journal : la plus ancienne
// disparaît au-delà.
const maxExtractionsFinies = 20

// delaiExtraction plafonne une extraction, comme en ligne de commande.
const delaiExtraction = 10 * time.Minute

// tailleJournal borne les événements gardés pour la reprise. Un onglet qui
// revient de plus loin reçoit un instantané plutôt qu'une suite trouée.
const tailleJournal = 256

// Noms des événements du flux.
const (
	evenementEtat       = "etat"
	evenementExtraction = "extraction"
	evenementRetrait    = "retrait"
)

// motifBase est ce qu'un nom de base doit respecter pour nommer un fichier.
//
// Refusé plutôt que nettoyé : PostgreSQL accepte « ../x » comme nom de base
// entre guillemets, et un nettoyage se contourne là où un refus tient. L'écran
// d'arbitrage recevra ce même nom du front pour relire le calque, et doit
// appliquer la même règle : une base extraite ici doit pouvoir y être rouverte.
var motifBase = regexp.MustCompile(`^[\p{L}\p{N}_-]+$`)

// baseValide refuse un nom de base qui ne peut pas nommer un fichier du
// répertoire de travail. Rend false quand il a déjà répondu.
func baseValide(w http.ResponseWriter, base string) bool {
	if motifBase.MatchString(base) {
		return true
	}
	repondreErreur(w, http.StatusBadRequest,
		fmt.Sprintf("la base %q ne peut pas nommer un fichier : lettres, chiffres, tiret et souligné seulement", base))
	return false
}

// ErrBaseDejaEnCours signale une deuxième extraction de la même base. Les deux
// écriraient le même fichier, et l'une serait perdue sans trace.
var ErrBaseDejaEnCours = errors.New("une extraction de cette base est deja en cours")

// tache est une extraction et ce qu'il faut pour la mener.
type tache struct {
	Extraction
	// dsn n'existe que le temps d'ouvrir et de tenir la connexion : il est
	// effacé dès que la tâche a fini.
	dsn  string
	sgbd string
	// repertoire est celui du lancement, et non celui du moment où le calque
	// s'écrit : changer de répertoire en cours d'extraction ne doit pas déplacer
	// un fichier que l'écran annonce déjà ailleurs.
	repertoire string
	portee     introspection.Portee
	annuler    context.CancelFunc
}

// evenement est une entrée du flux, sérialisée une fois pour tous les onglets
// qui la liront.
type evenement struct {
	id      uint64
	nom     string
	donnees []byte
}

// extractions tient les tâches d'un lancement.
//
// Au niveau du serveur et non de la session : une extraction continue quand on
// change de base ou qu'on se déconnecte, et ne retient jamais la connexion dont
// l'arbre a besoin pour se déplier.
type extractions struct {
	ctx context.Context
	// ouvrir est introspection.Ouvrir ; les tests y mettent une fabrique qui ne
	// joint aucun serveur.
	ouvrir func(ctx context.Context, sgbd, dsn string) (introspection.Introspecteur, error)

	mu      sync.Mutex
	taches  []*tache
	enCours int
	seq     uint64
	journal []evenement
	// signal est fermé à chaque événement puis remplacé : tous les flux ouverts
	// se réveillent, sans file par abonné à dimensionner.
	signal chan struct{}
	// ferme interdit tout démarrage une fois l'attente finale commencée, faute de
	// quoi un Add pourrait croiser le Wait.
	ferme   bool
	attente sync.WaitGroup
}

// nouvellesExtractions rend un registre vide. Ses tâches ne survivent pas à
// ctx.
//
// Le registre ne connaît aucun répertoire : chaque tâche porte celui de son
// lancement, faute de quoi un changement en cours de route réécrirait ailleurs
// un calque déjà annoncé.
func nouvellesExtractions(ctx context.Context) *extractions {
	return &extractions{
		ctx:    ctx,
		ouvrir: introspection.Ouvrir,
		signal: make(chan struct{}),
	}
}

// terminal dit si la tâche a fini d'évoluer.
func (e EtatExtraction) terminal() bool {
	return e == EtatTerminee || e == EtatEchouee || e == EtatAnnulee
}

// lancer inscrit une extraction et la démarre si une place est libre.
func (e *extractions) lancer(base, dsn, sgbd, repertoire string, portee introspection.Portee) (Extraction, error) {
	id, err := genererJeton()
	if err != nil {
		return Extraction{}, err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	for _, t := range e.taches {
		if t.Base == base && !t.Etat.terminal() {
			return Extraction{}, ErrBaseDejaEnCours
		}
	}

	t := &tache{
		Extraction: Extraction{
			ID:       id,
			Base:     base,
			Fichier:  base + ".calque.json",
			Schemas:  portee.Schemas,
			NbTables: len(portee.TablesIncluses),
			Etat:     EtatEnAttente,
		},
		dsn:        dsn,
		sgbd:       sgbd,
		repertoire: repertoire,
		portee:     portee,
	}
	e.taches = append(e.taches, t)
	e.emettre(evenementExtraction, t.Extraction)
	e.demarrer()
	return t.Extraction, nil
}

// demarrer lance les tâches en attente, dans leur ordre d'arrivée, tant qu'une
// place est libre. Verrou pris.
func (e *extractions) demarrer() {
	for _, t := range e.taches {
		if e.ferme || e.enCours >= maxExtractionsEnCours {
			return
		}
		if t.Etat != EtatEnAttente {
			continue
		}

		ctx, annuler := context.WithTimeout(e.ctx, delaiExtraction)
		t.annuler = annuler
		t.Etat = EtatEnCours
		t.Debut = instant()
		e.enCours++
		e.emettre(evenementExtraction, t.Extraction)

		e.attente.Add(1)
		go e.executer(ctx, t)
	}
}

// executer mène une tâche jusqu'à son état final, puis cède sa place.
func (e *extractions) executer(ctx context.Context, t *tache) {
	defer e.attente.Done()
	defer t.annuler()

	resultat, err := e.extraire(ctx, t)

	e.mu.Lock()
	defer e.mu.Unlock()

	t.Fin = instant()
	switch {
	case err == nil:
		t.Etat, t.Resultat = EtatTerminee, resultat
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		t.Etat = EtatEchouee
		t.Erreur = fmt.Sprintf("extraction interrompue au bout de %d minutes", int(delaiExtraction.Minutes()))
	case ctx.Err() != nil:
		t.Etat = EtatAnnulee
	default:
		t.Etat, t.Erreur = EtatEchouee, sansDSN(err.Error(), t.dsn)
		slog.Warn("extraction en echec", "database", t.Base, "error", t.Erreur)
	}
	t.dsn = ""
	e.enCours--

	e.emettre(evenementExtraction, t.Extraction)
	e.oublierFinies()
	e.demarrer()
}

// extraire ouvre la connexion de la tâche, lit le catalogue et écrit le calque.
//
// La connexion est celle de la tâche, jamais celle de la session : une
// extraction de plusieurs minutes n'a pas à bloquer le dépliage d'une table.
func (e *extractions) extraire(ctx context.Context, t *tache) (*ResultatExtraction, error) {
	pilote, err := e.ouvrir(ctx, t.sgbd, t.dsn)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := pilote.Fermer(); err != nil {
			slog.Warn("fermeture de la connexion d'extraction", "database", t.Base, "error", sansDSN(err.Error(), t.dsn))
		}
	}()

	suivi := func(a introspection.Avancement) {
		e.mu.Lock()
		defer e.mu.Unlock()
		t.Avancement = &a
		e.emettre(evenementExtraction, t.Extraction)
	}

	physique, err := pilote.Extraire(introspection.AvecSuivi(ctx, suivi), t.portee)
	if err != nil {
		return nil, err
	}
	// Une annulation arrivée après la dernière passe l'emporte encore : on a
	// demandé de ne rien écrire.
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return e.ecrire(physique, filepath.Join(t.repertoire, t.Fichier))
}

// ecrire valide, horodate et sérialise. Les anomalies n'empêchent pas
// l'écriture : un calque partiel reste exploitable, et elles accompagnent le
// résultat pour que l'interface les montre.
func (e *extractions) ecrire(physique *calque.Physique, chemin string) (*ResultatExtraction, error) {
	anomalies := physique.Valider()
	if anomalies == nil {
		anomalies = []calque.Anomalie{}
	}

	// Posé au dernier moment, et exclu de l'empreinte : deux extractions de la
	// même base restent identiques.
	physique.Source.ExtraitLe = horodater()
	if err := physique.Ecrire(chemin); err != nil {
		return nil, err
	}

	var colonnes int
	for i := range physique.Tables {
		colonnes += len(physique.Tables[i].Colonnes)
	}
	return &ResultatExtraction{
		Tables:    len(physique.Tables),
		Colonnes:  colonnes,
		Empreinte: physique.Source.Empreinte,
		Anomalies: anomalies,
	}, nil
}

// annuler arrête une tâche en cours ou en attente, et retire une tâche finie.
// Un identifiant inconnu n'est pas une erreur : deux onglets peuvent retirer la
// même.
func (e *extractions) annuler(id string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for i, t := range e.taches {
		if t.ID != id {
			continue
		}
		switch t.Etat {
		case EtatEnAttente:
			t.Etat, t.Fin, t.dsn = EtatAnnulee, instant(), ""
			e.emettre(evenementExtraction, t.Extraction)
			e.oublierFinies()
		case EtatEnCours:
			// L'état final est posé par la tâche quand le pilote a rendu la main :
			// elle ne se dit annulée qu'une fois sa connexion refermée.
			t.annuler()
		default:
			e.taches = slices.Delete(e.taches, i, i+1)
			e.emettre(evenementRetrait, ReferenceExtraction{ID: id})
		}
		return
	}
}

// oublierFinies retire les tâches finies les plus anciennes au-delà du plafond.
// Verrou pris.
func (e *extractions) oublierFinies() {
	var finies int
	for _, t := range e.taches {
		if t.Etat.terminal() {
			finies++
		}
	}

	for i := 0; finies > maxExtractionsFinies && i < len(e.taches); {
		t := e.taches[i]
		if !t.Etat.terminal() {
			i++
			continue
		}
		e.taches = slices.Delete(e.taches, i, i+1)
		e.emettre(evenementRetrait, ReferenceExtraction{ID: t.ID})
		finies--
	}
}

// emettre inscrit un événement au journal et réveille les flux. Verrou pris.
func (e *extractions) emettre(nom string, v any) {
	e.seq++
	e.journal = append(e.journal, evenement{id: e.seq, nom: nom, donnees: serialiserEvenement(v)})
	if len(e.journal) > tailleJournal {
		e.journal = e.journal[len(e.journal)-tailleJournal:]
	}
	close(e.signal)
	e.signal = make(chan struct{})
}

// depuis rend ce qu'un flux doit envoyer après l'événement dernier, et le signal
// qui annoncera le suivant.
//
// Quand la reprise est impossible — premier chargement, identifiant venu d'un
// lancement précédent du binaire, ou trop ancien pour le journal —, c'est un
// instantané de toutes les tâches. Sinon, exactement les événements manqués :
// ni doublon ni trou.
func (e *extractions) depuis(dernier uint64, reprise bool) ([]evenement, chan struct{}) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if reprise && dernier <= e.seq && (len(e.journal) == 0 || dernier+1 >= e.journal[0].id) {
		var manques []evenement
		for _, ev := range e.journal {
			if ev.id > dernier {
				manques = append(manques, ev)
			}
		}
		return manques, e.signal
	}

	etat := EtatExtractions{Extractions: make([]Extraction, 0, len(e.taches))}
	for _, t := range e.taches {
		etat.Extractions = append(etat.Extractions, t.Extraction)
	}
	return []evenement{{id: e.seq, nom: evenementEtat, donnees: serialiserEvenement(etat)}}, e.signal
}

// attendre rend la main quand toutes les tâches ont fini. À appeler une fois
// leur contexte annulé, sans quoi elle attend leur fin normale.
func (e *extractions) attendre() {
	e.mu.Lock()
	e.ferme = true
	e.mu.Unlock()

	e.attente.Wait()
}

// serialiserEvenement rend la ligne de données d'un événement.
//
// json.Marshal direct, comme repondreJSON : ce flux n'est pas un calque. Les
// types sérialisés ici sont ceux de api.go, dont aucun ne peut échouer ; un échec
// serait un défaut de ce fichier, journalisé plutôt que propagé à un flux qui ne
// saurait qu'en faire.
func serialiserEvenement(v any) []byte {
	donnees, err := json.Marshal(v)
	if err != nil {
		slog.Error("serialisation d'un evenement", "error", err)
		return []byte("{}")
	}
	return donnees
}

// horodater rend l'instant présent au format du calque.
func horodater() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// formatInstant est celui des changements d'état d'une tâche : RFC 3339 à la
// milliseconde.
const formatInstant = "2006-01-02T15:04:05.000Z07:00"

// instant horodate un changement d'état d'une tâche.
//
// À la milliseconde, là où le calque s'en tient à la seconde : une extraction
// de quelques tables tient dans la seconde, et une durée arrondie à zéro
// n'apprendrait rien à qui la lit.
func instant() string {
	return time.Now().UTC().Format(formatInstant)
}

// gererExtractions lance une extraction ou en retire une.
func (s *serveur) gererExtractions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.lancerExtraction(w, r)
	case http.MethodDelete:
		s.retirerExtraction(w, r)
	default:
		repondreErreur(w, http.StatusMethodNotAllowed, "méthode non acceptée")
	}
}

// lancerExtraction inscrit une extraction et rend aussitôt la main : la suite
// arrive par le flux.
//
// La base et le fichier ne viennent pas du front : c'est la base de la session,
// et le fichier en porte le nom. Le navigateur ne désigne ni ce qu'on joint ni
// où l'on écrit.
func (s *serveur) lancerExtraction(w http.ResponseWriter, r *http.Request) {
	var requete RequeteExtraction
	if !decoder(w, r, &requete) {
		return
	}

	c, ok := s.registre.trouver(requete.Session)
	if !ok {
		repondreErreur(w, http.StatusNotFound, "connexion inconnue ou déjà fermée")
		return
	}
	// Refusé plutôt qu'ignoré : aucun pilote ne lit encore de données, et une
	// option acceptée sans effet laisserait croire que l'inférence en a profité.
	if requete.Portee.Echantillonner {
		repondreErreur(w, http.StatusNotImplemented, "l'échantillonnage n'est pas encore pris en charge")
		return
	}

	base := introspection.BaseDuDSN(c.dsn)
	if !baseValide(w, base) {
		return
	}

	extraction, err := s.extractions.lancer(base, c.dsn, c.sgbd, s.repertoireCourant(), requete.Portee)
	switch {
	case errors.Is(err, ErrBaseDejaEnCours):
		repondreErreur(w, http.StatusConflict, fmt.Sprintf("une extraction de %s est déjà en cours", base))
	case err != nil:
		repondreErreur(w, http.StatusInternalServerError, err.Error())
	default:
		repondreJSON(w, http.StatusAccepted, extraction)
	}
}

// retirerExtraction annule une tâche en cours ou en attente, retire une tâche
// finie.
func (s *serveur) retirerExtraction(w http.ResponseWriter, r *http.Request) {
	var requete ReferenceExtraction
	if !decoder(w, r, &requete) {
		return
	}
	s.extractions.annuler(requete.ID)
	w.WriteHeader(http.StatusNoContent)
}

// suivreExtractions ouvre le flux des tâches.
//
// Un seul flux pour toutes : un navigateur n'ouvre que six connexions HTTP/1.1
// par origine, et un EventSource par tâche les épuiserait dès trois
// extractions, bloquant les appels de l'arbre sans le moindre message.
//
// Last-Event-ID est posé par EventSource lui-même quand il se reconnecte.
func (s *serveur) suivreExtractions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		repondreErreur(w, http.StatusMethodNotAllowed, "méthode non acceptée")
		return
	}

	dernier, err := strconv.ParseUint(r.Header.Get("Last-Event-ID"), 10, 64)
	evenements, signal := s.extractions.depuis(dernier, err == nil)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	controleur := http.NewResponseController(w)

	for {
		for _, ev := range evenements {
			if _, err := fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", ev.id, ev.nom, ev.donnees); err != nil {
				return
			}
			dernier = ev.id
		}
		// Même sans événement : c'est ce premier envoi qui ouvre le flux côté
		// navigateur.
		if err := controleur.Flush(); err != nil {
			return
		}

		select {
		case <-signal:
		case <-r.Context().Done():
			return
		case <-s.extractions.ctx.Done():
			return
		}
		evenements, signal = s.extractions.depuis(dernier, true)
	}
}
