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
	"strings"
	"time"

	"github.com/sprimault/ormeau/internal/calque"
	"github.com/sprimault/ormeau/internal/config"
	"github.com/sprimault/ormeau/internal/inference"
	"github.com/sprimault/ormeau/internal/introspection"
)

// delaiConnexion borne l'ouverture et la description d'une base. Un serveur
// injoignable doit rendre la main assez vite pour que l'écran reste utilisable,
// sans être si court qu'une base lente passe pour inaccessible.
const delaiConnexion = 30 * time.Second

// delaiInventaire borne la passe légère qui alimente l'arbre. Une requête de
// catalogue, même sur quatre cents tables, tient largement dedans.
const delaiInventaire = 60 * time.Second

// tailleMaxCorps borne le corps d'une requête. La plupart tiennent en quelques
// centaines d'octets ; une portée qui nomme ses tables en fait quelques
// dizaines de kilo-octets sur une base de plusieurs milliers de tables.
const tailleMaxCorps = 1 << 20

// RequeteConnexion accepte les deux formes de connexion : une chaîne complète,
// ou les composants.
//
// Personne ne tape une URL dans un formulaire, et celui qui découvre une base
// connaît un hôte et un identifiant, pas un DSN. Quand SGBD est vide, le port le
// désigne — c'est l'outil qui aiguille, pas l'utilisateur qui déclare.
type RequeteConnexion struct {
	// Profil désigne une connexion enregistrée. Ce que la requête ne dit pas
	// est repris du profil, mot de passe compris — celui-ci ne descend jamais
	// dans le navigateur, il va du fichier au pilote.
	Profil      string `json:"profil,omitempty"`
	DSN         string `json:"dsn,omitempty"`
	SGBD        string `json:"sgbd,omitempty"`
	Hote        string `json:"hote,omitempty"`
	Port        int    `json:"port,omitempty"`
	Utilisateur string `json:"utilisateur,omitempty"`
	MotDePasse  string `json:"mot_de_passe,omitempty"`
	Base        string `json:"base,omitempty"`
}

// ReponseConnexion décrit le serveur atteint. Le DSN n'y figure sous aucune
// forme, pas même masqué.
type ReponseConnexion struct {
	Session   string   `json:"session"`
	SGBD      string   `json:"sgbd"`
	Version   string   `json:"version"`
	Catalogue string   `json:"catalogue"`
	Schemas   []string `json:"schemas"`
	// BaseImposee dit que la session vient d'un profil qui nomme sa base.
	//
	// L'écran en a besoin pour dire la vérité : sans lui, une liste à une seule
	// entrée serait annoncée comme « ce serveur n'expose qu'une base », alors
	// qu'il en porte vingt et que c'est le profil qui cadre.
	BaseImposee bool `json:"base_imposee,omitempty"`
}

// RequeteFermeture désigne la connexion à refermer.
type RequeteFermeture struct {
	Session string `json:"session"`
}

// ReponseContexte porte ce que l'interface affiche en permanence : le
// répertoire où elle écrira, et la version du binaire.
type ReponseContexte struct {
	Repertoire string `json:"repertoire"`
	Version    string `json:"version"`
}

// RequeteRepertoire change le répertoire où les fichiers du projet s'écrivent.
//
// Le seul chemin de fichier que l'API accepte du navigateur, et il ne désigne
// qu'un répertoire existant : les noms de fichiers restent composés côté
// serveur à partir d'un nom de base validé.
type RequeteRepertoire struct {
	Repertoire string `json:"repertoire"`
}

// ProfilResume est un profil tel que l'écran le reçoit : sans mot de passe,
// pas même chiffré, mais en sachant s'il y en a un.
//
// Le profil est imbriqué et non incorporé : Go aplatirait les champs à la
// sérialisation, mais le générateur de types du front ne sait pas le faire, et
// rendrait un type qui ne décrit pas ce qui passe sur le fil.
type ProfilResume struct {
	Profil               config.Profil `json:"profil"`
	MotDePasseEnregistre bool          `json:"mot_de_passe_enregistre"`
}

// ReponseProfils liste les connexions enregistrées.
//
// L'avertissement dit ce que le fichier portait d'inutilisable, sans empêcher
// d'ouvrir l'écran : on saisit alors comme avant.
type ReponseProfils struct {
	Profils       []ProfilResume `json:"profils"`
	Avertissement string         `json:"avertissement,omitempty"`
}

// RequeteProfil enregistre une connexion.
//
// Le mot de passe n'est retenu que si EnregistrerMotDePasse est vrai. À faux,
// celui qui avait été enregistré est effacé : croire l'avoir retiré alors qu'il
// reste sur le disque serait le pire des deux.
type RequeteProfil struct {
	Profil config.Profil `json:"profil"`
	// DSN enregistre un profil depuis une chaîne de connexion plutôt que depuis
	// les champs. Le serveur la décompose : le front n'analyse jamais un DSN,
	// et sans cela un profil enregistré depuis ce mode ne retiendrait rien.
	//
	// Le mot de passe qu'elle porte ne compte que si la case est cochée, comme
	// celui du champ.
	DSN                   string `json:"dsn,omitempty"`
	MotDePasse            string `json:"mot_de_passe,omitempty"`
	EnregistrerMotDePasse bool   `json:"enregistrer_mot_de_passe"`
	// Remplacer confirme l'écrasement d'un profil du même nom. Sans lui, un
	// nom déjà pris est refusé avec CodeProfilExistant, et l'écran demande.
	Remplacer bool `json:"remplacer,omitempty"`
}

// ReferenceProfil désigne le profil à supprimer.
type ReferenceProfil struct {
	Nom string `json:"nom"`
}

// ReponseErreur est la forme unique des échecs d'API. Un code HTTP seul
// laisserait le front deviner ce qu'il affiche.
type ReponseErreur struct {
	Erreur string `json:"erreur"`
	// Code n'est posé que sur un refus que le front traite à part ; les autres
	// échecs se distinguent par leur statut.
	Code CodeRefus `json:"code,omitempty"`
}

// CodeRefus distingue les refus qui n'appellent pas la même réaction.
type CodeRefus string

// Codes de refus. Ils disent au front quoi faire, jamais quoi afficher : le
// texte reste celui de l'erreur.
const (
	// CodeCalqueModifie : une extraction a réécrit le calque pendant
	// l'arbitrage, l'écran recharge avant d'aller plus loin.
	CodeCalqueModifie CodeRefus = "calque_modifie"
	// CodeDecisionsModifiees : le fichier de décisions a changé sur disque
	// depuis sa lecture, l'écran le relit.
	CodeDecisionsModifiees CodeRefus = "decisions_modifiees"
	// CodeContenuManuel : le fichier porte un travail humain que la réécriture
	// perdrait, l'écran demande confirmation.
	CodeContenuManuel CodeRefus = "contenu_manuel"
	// CodeProfilExistant : un profil porte déjà ce nom, l'écran demande
	// confirmation avant de l'écraser.
	CodeProfilExistant CodeRefus = "profil_existant"
)

// ReponseBases liste les bases exploitables du serveur atteint.
//
// Les bases système en sont absentes : elles ne produiraient que des calques
// sans intérêt, et template0 refuse même la connexion.
type ReponseBases struct {
	Bases []string `json:"bases"`
}

// RequeteBase demande de basculer la session sur une autre base du même
// serveur.
type RequeteBase struct {
	Session string `json:"session"`
	Base    string `json:"base"`
}

// ReponseColonnes décrit une table dépliée dans l'arbre.
type ReponseColonnes struct {
	Colonnes []introspection.ColonneSommaire `json:"colonnes"`
}

// ReponseInventaire porte l'arbre de sélection.
//
// L'inventaire complet part d'un coup, sans pagination : quelques dizaines de
// kilo-octets pour quatre cents tables, et la recherche reste instantanée côté
// navigateur. Paginer coûterait un aller-retour par frappe pour économiser un
// transfert qui tient dans un paquet réseau.
type ReponseInventaire struct {
	Tables []introspection.TableSommaire `json:"tables"`
}

// EtatExtraction est l'état d'une tâche d'extraction.
type EtatExtraction string

// États d'une tâche. Les trois derniers sont terminaux : la tâche n'évolue
// plus, elle attend qu'on la retire.
const (
	EtatEnAttente EtatExtraction = "en_attente"
	EtatEnCours   EtatExtraction = "en_cours"
	EtatTerminee  EtatExtraction = "terminee"
	EtatEchouee   EtatExtraction = "echouee"
	EtatAnnulee   EtatExtraction = "annulee"
)

// RequeteExtraction lance l'extraction de la base d'une session. La base n'est
// pas un paramètre : c'est celle de la session.
type RequeteExtraction struct {
	Session string               `json:"session"`
	Portee  introspection.Portee `json:"portee"`
}

// ReferenceExtraction désigne une tâche : celle qu'on annule ou retire, et
// celle qu'un événement « retrait » fait disparaître de l'écran.
type ReferenceExtraction struct {
	ID string `json:"id"`
}

// Extraction est ce que le front sait d'une tâche.
//
// Chaque événement du flux en porte l'état complet, jamais un delta : un
// événement rejoué ne laisse pas l'écran dans un état intermédiaire. Le DSN
// n'y figure sous aucune forme.
type Extraction struct {
	ID      string   `json:"id"`
	Base    string   `json:"base"`
	Fichier string   `json:"fichier"`
	Schemas []string `json:"schemas,omitempty"`
	// NbTables vaut zéro quand la portée ne nomme pas ses tables : toutes
	// celles des schémas partent.
	NbTables   int                       `json:"nb_tables"`
	Etat       EtatExtraction            `json:"etat"`
	Avancement *introspection.Avancement `json:"avancement,omitempty"`
	Debut      string                    `json:"debut,omitempty"`
	Fin        string                    `json:"fin,omitempty"`
	Resultat   *ResultatExtraction       `json:"resultat,omitempty"`
	Erreur     string                    `json:"erreur,omitempty"`
}

// ResultatExtraction résume le calque écrit.
type ResultatExtraction struct {
	Tables    int    `json:"tables"`
	Colonnes  int    `json:"colonnes"`
	Empreinte string `json:"empreinte"`
	// Anomalies sont celles de la validation, qui n'empêchent pas l'écriture.
	// Une clé étrangère vers une table restée hors de la portée en est le cas
	// courant.
	Anomalies []calque.Anomalie `json:"anomalies"`
}

// EtatExtractions est l'instantané de toutes les tâches, envoyé à qui ouvre le
// flux sans pouvoir reprendre là où il en était.
type EtatExtractions struct {
	Extractions []Extraction `json:"extractions"`
}

// ReponseCalque porte le calque d'une base tel que le répertoire de travail le
// contient.
//
// Le contenu est le document sérialisé et non une structure : l'écran l'affiche
// dans l'ordre et l'indentation que le calque impose, ce qu'on retrouve en
// ouvrant le fichier.
type ReponseCalque struct {
	Fichier   string `json:"fichier"`
	ExtraitLe string `json:"extrait_le"`
	Empreinte string `json:"empreinte"`
	Contenu   string `json:"contenu"`
	// StatistiquesRetirees signale un calque échantillonné dont les statistiques
	// n'ont pas été envoyées : l'écran le dit plutôt que de laisser croire qu'il
	// n'y en a pas.
	StatistiquesRetirees bool `json:"statistiques_retirees,omitempty"`
}

// ReponseDecisions décrit le fichier de décisions d'une base.
type ReponseDecisions struct {
	Existe    bool                `json:"existe"`
	Decisions inference.Decisions `json:"decisions"`
	// EmpreinteFichier identifie les octets lus. L'écriture la renvoie, et le
	// serveur refuse d'écraser un fichier qui ne les porte plus.
	EmpreinteFichier string `json:"empreinte_fichier,omitempty"`
	// Manuel annonce qu'une réécriture perdrait un travail humain : l'écran
	// demande confirmation avant d'enregistrer.
	Manuel bool `json:"manuel"`
}

// RequeteInference rejoue l'inférence avec le brouillon de décisions de
// l'écran.
type RequeteInference struct {
	Base      string              `json:"base"`
	Decisions inference.Decisions `json:"decisions"`
	// EmpreintePhysique est celle du calque que l'écran juge. Vide au premier
	// chargement, qui l'apprend.
	EmpreintePhysique string `json:"empreinte_physique,omitempty"`
}

// ReponseInference porte ce que l'écran d'arbitrage liste, sans le détail des
// entités.
type ReponseInference struct {
	EmpreintePhysique string                  `json:"empreinte_physique"`
	Avertissements    []calque.Avertissement  `json:"avertissements"`
	Propositions      []inference.Proposition `json:"propositions"`
	Enumerations      []EnumerationInferee    `json:"enumerations"`
	Entites           []ResumeEntite          `json:"entites"`
	// TypesDoctrine sont ceux que l'écran suggère quand on force le type d'une
	// colonne.
	TypesDoctrine []string `json:"types_doctrine"`
}

// ResumeEntite est ce qu'une ligne de la liste montre d'une entité.
type ResumeEntite struct {
	Nom            string                `json:"nom"`
	Table          calque.ReferenceTable `json:"table"`
	NbProprietes   int                   `json:"nb_proprietes"`
	NbAssociations int                   `json:"nb_associations"`
	Heritage       *calque.Heritage      `json:"heritage,omitempty"`
	Traits         []string              `json:"traits,omitempty"`
	Origine        calque.Origine        `json:"origine,omitempty"`
}

// EnumerationInferee est une énumération du calque logique, avec les colonnes
// qui la portent : c'est par la colonne que le fichier de décisions en nomme
// les cas.
type EnumerationInferee struct {
	Nom         string                  `json:"nom"`
	TypeSupport string                  `json:"type_support"`
	Cas         []calque.CasEnumeration `json:"cas"`
	Origine     calque.Origine          `json:"origine"`
	Colonnes    []string                `json:"colonnes"`
}

// RequeteEntite demande le détail d'une entité, calculé avec le brouillon
// courant.
type RequeteEntite struct {
	Base              string              `json:"base"`
	Decisions         inference.Decisions `json:"decisions"`
	EmpreintePhysique string              `json:"empreinte_physique,omitempty"`
	Schema            string              `json:"schema"`
	Table             string              `json:"table"`
}

// ReponseEntite porte une entité inférée et la table physique dont elle vient.
// Sans entité — table ignorée, table de jointure —, la table vient seule.
type ReponseEntite struct {
	Entite        *calque.Entite `json:"entite,omitempty"`
	TablePhysique calque.Table   `json:"table_physique"`
}

// RequeteEcritureDecisions enregistre le brouillon dans le fichier de la base.
type RequeteEcritureDecisions struct {
	Base              string              `json:"base"`
	Decisions         inference.Decisions `json:"decisions"`
	EmpreintePhysique string              `json:"empreinte_physique"`
	// EmpreinteFichier est celle rendue à la lecture, vide quand le fichier
	// n'existait pas.
	EmpreinteFichier string `json:"empreinte_fichier,omitempty"`
	// EcraserManuel confirme la perte d'un contenu écrit à la main.
	EcraserManuel bool `json:"ecraser_manuel,omitempty"`
}

// ReponseEcritureDecisions rend l'empreinte du fichier écrit, que
// l'enregistrement suivant renverra.
type ReponseEcritureDecisions struct {
	EmpreinteFichier string `json:"empreinte_fichier"`
}

// contexte rend le répertoire de travail et la version.
func (s *serveur) contexte(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		repondreErreur(w, http.StatusMethodNotAllowed, "méthode non acceptée")
		return
	}
	repondreJSON(w, http.StatusOK, ReponseContexte{Repertoire: s.repertoireCourant(), Version: s.version})
}

// bases rend les bases du serveur atteint par une connexion ouverte.
//
// Un serveur en porte souvent vingt, et personne ne les retient : les faire
// deviner est exactement ce que le projet refuse. L'information est atteignable,
// donc elle ne se demande pas.
func (s *serveur) bases(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		repondreErreur(w, http.StatusMethodNotAllowed, "méthode non acceptée")
		return
	}

	c, ok := s.registre.trouver(r.URL.Query().Get("session"))
	if !ok {
		repondreErreur(w, http.StatusNotFound, "connexion inconnue ou déjà fermée")
		return
	}

	// Une session ouverte par un profil qui nomme sa base y reste : l'écran ne
	// propose pas ce sur quoi on n'a pas choisi de travailler.
	//
	// Filtré ici et non dans la page : un tri posé côté navigateur se
	// contournerait en rechargeant, et l'API rendrait quand même la liste
	// entière.
	if c.baseImposee {
		repondreJSON(w, http.StatusOK, ReponseBases{
			Bases: []string{introspection.BaseDuDSN(c.dsn)},
		})
		return
	}

	listeur, ok := c.pilote.(introspection.ListeurDeBases)
	if !ok {
		// Un dialecte où la notion n'a pas de sens n'est pas une panne : le
		// front se passe du sélecteur et garde la base saisie.
		repondreJSON(w, http.StatusOK, ReponseBases{Bases: []string{}})
		return
	}

	ctx, annuler := context.WithTimeout(r.Context(), delaiInventaire)
	defer annuler()

	var bases []string
	err := c.utiliser(func(introspection.Introspecteur) error {
		var err error
		bases, err = listeur.ListerBases(ctx)
		return err
	})
	if err != nil {
		repondreErreur(w, http.StatusBadGateway, sansDSN(err.Error(), c.dsn))
		return
	}
	if bases == nil {
		bases = []string{}
	}
	repondreJSON(w, http.StatusOK, ReponseBases{Bases: bases})
}

// basculerBase rouvre la session sur une autre base du même serveur.
//
// Les identifiants sont ceux de la connexion en cours : changer de base ne doit
// pas faire ressaisir un mot de passe qu'on vient de donner. L'ancienne
// connexion est refermée, et la nouvelle reçoit un nouvel identifiant — le
// front repart d'un état propre plutôt que de deviner ce qui a changé sous lui.
func (s *serveur) basculerBase(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		repondreErreur(w, http.StatusMethodNotAllowed, "méthode non acceptée")
		return
	}

	var requete RequeteBase
	if !decoder(w, r, &requete) {
		return
	}

	c, ok := s.registre.trouver(requete.Session)
	if !ok {
		repondreErreur(w, http.StatusNotFound, "connexion inconnue ou déjà fermée")
		return
	}
	if strings.TrimSpace(requete.Base) == "" {
		repondreErreur(w, http.StatusBadRequest, "aucune base demandée")
		return
	}
	// Refusé, et non ignoré : la session a été ouverte par un profil qui nomme
	// sa base. Se connecter ailleurs demande un autre profil, ou une connexion
	// sans profil.
	if c.baseImposee {
		repondreErreur(w, http.StatusConflict, fmt.Sprintf(
			"cette connexion vient d'un profil qui désigne %s : choisir un autre profil pour une autre base",
			introspection.BaseDuDSN(c.dsn)))
		return
	}

	ctx, annuler := context.WithTimeout(r.Context(), delaiConnexion)
	defer annuler()

	dsn := introspection.AvecBase(c.dsn, requete.Base)
	pilote, err := introspection.Ouvrir(ctx, c.sgbd, dsn)
	if err != nil {
		message := sansDSN(err.Error(), dsn)
		slog.Warn("bascule de base refusee", "dbms", c.sgbd, "error", message)
		repondreErreur(w, http.StatusBadGateway, message)
		return
	}

	// La bascule n'est atteinte que pour une session libre : la nouvelle
	// l'est aussi.
	reponse, err := s.enregistrer(ctx, pilote, dsn, c.sgbd, false)
	if err != nil {
		repondreErreur(w, codeDe(err), sansDSN(err.Error(), dsn))
		return
	}

	// Fermée après coup : si la nouvelle base est inaccessible, l'utilisateur
	// garde celle qu'il avait.
	if err := s.registre.fermer(requete.Session); err != nil {
		slog.Warn("fermeture de l'ancienne connexion", "error", err)
	}
	repondreJSON(w, http.StatusOK, reponse)
}

// inventaire rend les tables d'une connexion ouverte.
//
// Une passe légère, jamais une extraction : ce qui alimente l'arbre vient du
// catalogue, et pas une ligne de donnée n'est lue.
func (s *serveur) inventaire(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		repondreErreur(w, http.StatusMethodNotAllowed, "méthode non acceptée")
		return
	}

	c, ok := s.registre.trouver(r.URL.Query().Get("session"))
	if !ok {
		repondreErreur(w, http.StatusNotFound, "connexion inconnue ou déjà fermée")
		return
	}

	ctx, annuler := context.WithTimeout(r.Context(), delaiInventaire)
	defer annuler()

	var tables []introspection.TableSommaire
	err := c.utiliser(func(pilote introspection.Introspecteur) error {
		var err error
		tables, err = pilote.Inventorier(ctx, decouper(r.URL.Query().Get("schemas")))
		return err
	})
	if err != nil {
		slog.Warn("inventaire refuse", "error", err)
		repondreErreur(w, http.StatusBadGateway, err.Error())
		return
	}
	if tables == nil {
		// Une portée sans table est un résultat vide, pas une absence de
		// résultat : le front parcourt la liste sans avoir à distinguer les deux.
		tables = []introspection.TableSommaire{}
	}
	repondreJSON(w, http.StatusOK, ReponseInventaire{Tables: tables})
}

// colonnes décrit une table, à la demande.
//
// Chargée au dépliement plutôt qu'avec l'inventaire : porter les colonnes de
// quatre cents tables coûterait dix fois le transfert pour des lignes que
// personne n'ouvrira.
func (s *serveur) colonnes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		repondreErreur(w, http.StatusMethodNotAllowed, "méthode non acceptée")
		return
	}

	requete := r.URL.Query()
	c, ok := s.registre.trouver(requete.Get("session"))
	if !ok {
		repondreErreur(w, http.StatusNotFound, "connexion inconnue ou déjà fermée")
		return
	}

	schema, table := requete.Get("schema"), requete.Get("table")
	if schema == "" || table == "" {
		repondreErreur(w, http.StatusBadRequest, "schéma et table sont requis")
		return
	}

	listeur, ok := c.pilote.(introspection.ListeurDeColonnes)
	if !ok {
		repondreErreur(w, http.StatusNotImplemented, "ce pilote ne sait pas encore décrire une table")
		return
	}

	ctx, annuler := context.WithTimeout(r.Context(), delaiInventaire)
	defer annuler()

	var colonnes []introspection.ColonneSommaire
	err := c.utiliser(func(introspection.Introspecteur) error {
		var err error
		colonnes, err = listeur.Colonnes(ctx, schema, table)
		return err
	})
	if err != nil {
		slog.Warn("lecture des colonnes refusee", "error", err)
		repondreErreur(w, http.StatusBadGateway, err.Error())
		return
	}
	if colonnes == nil {
		colonnes = []introspection.ColonneSommaire{}
	}
	repondreJSON(w, http.StatusOK, ReponseColonnes{Colonnes: colonnes})
}

// decouper rend les valeurs d'une liste séparée par des virgules. Une entrée
// vide n'est pas un schéma : « public, » n'en demande pas deux.
func decouper(liste string) []string {
	var valeurs []string
	for _, brut := range strings.Split(liste, ",") {
		if valeur := strings.TrimSpace(brut); valeur != "" {
			valeurs = append(valeurs, valeur)
		}
	}
	return valeurs
}

// connexion ouvre une base ou referme une connexion ouverte.
func (s *serveur) connexion(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.ouvrirConnexion(w, r)
	case http.MethodDelete:
		s.fermerConnexion(w, r)
	default:
		repondreErreur(w, http.StatusMethodNotAllowed, "méthode non acceptée")
	}
}

// ouvrirConnexion joint la base et rend de quoi remplir l'écran suivant.
//
// La connexion reste ouverte au-delà de la requête : l'inventaire et
// l'extraction la réutiliseront, et rouvrir à chaque appel ferait ressaisir le
// mot de passe ou le garderait côté navigateur, ce que l'interface refuse.
func (s *serveur) ouvrirConnexion(w http.ResponseWriter, r *http.Request) {
	var requete RequeteConnexion
	if !decoder(w, r, &requete) {
		return
	}

	baseImposee, err := s.connexionDuProfil(&requete)
	if err != nil {
		repondreErreur(w, codeDe(err), err.Error())
		return
	}

	dsn, err := requete.composer()
	if err != nil {
		repondreErreur(w, http.StatusBadRequest, err.Error())
		return
	}

	sgbd, err := introspection.SGBDDepuisDSN(dsn)
	if err != nil {
		repondreErreur(w, http.StatusBadRequest, sansDSN(err.Error(), dsn))
		return
	}

	ctx, annuler := context.WithTimeout(r.Context(), delaiConnexion)
	defer annuler()

	pilote, err := introspection.Ouvrir(ctx, sgbd, dsn)
	if err != nil {
		// Le message est nettoyé avant d'aller où que ce soit : ni la réponse ni
		// le journal ne doivent porter le DSN, fût-il masqué.
		message := sansDSN(err.Error(), dsn)
		slog.Warn("connexion refusee", "dbms", sgbd, "error", message)
		repondreErreur(w, http.StatusBadGateway, message)
		return
	}

	reponse, err := s.enregistrer(ctx, pilote, dsn, sgbd, baseImposee)
	if err != nil {
		repondreErreur(w, codeDe(err), sansDSN(err.Error(), dsn))
		return
	}
	repondreJSON(w, http.StatusOK, reponse)
}

// errPiloteMuet signale un dialecte dont le pilote ne sait pas encore se
// décrire. C'est un manque annoncé, pas une panne de la base.
var errPiloteMuet = errors.New("ce pilote ne sait pas encore se décrire")

// enregistrer décrit le serveur atteint, garde la connexion ouverte et compose
// ce que le front affiche. Le pilote est refermé dès que l'une des étapes
// échoue : une connexion qu'on n'enregistre pas ne doit pas survivre.
func (s *serveur) enregistrer(
	ctx context.Context,
	pilote introspection.Introspecteur,
	dsn, sgbd string,
	baseImposee bool,
) (ReponseConnexion, error) {
	descripteur, ok := pilote.(introspection.DescripteurServeur)
	if !ok {
		_ = pilote.Fermer()
		return ReponseConnexion{}, errPiloteMuet
	}

	serveurBase, err := descripteur.Decrire(ctx)
	if err != nil {
		_ = pilote.Fermer()
		return ReponseConnexion{}, err
	}

	// La session nomme toujours sa base, y compris quand la connexion n'en
	// précisait aucune et que le serveur a pris celle par défaut : c'est ce nom
	// qui désigne le calque qu'une extraction écrira.
	if introspection.BaseDuDSN(dsn) == "" {
		dsn = introspection.AvecBase(dsn, serveurBase.Catalogue)
	}

	session, err := s.registre.ajouter(pilote, dsn, sgbd, baseImposee)
	if err != nil {
		_ = pilote.Fermer()
		return ReponseConnexion{}, err
	}

	return ReponseConnexion{
		Session:     session,
		SGBD:        serveurBase.SGBD,
		Version:     serveurBase.Version,
		Catalogue:   serveurBase.Catalogue,
		Schemas:     serveurBase.Schemas,
		BaseImposee: baseImposee,
	}, nil
}

// codeDe traduit un échec d'enregistrement en code HTTP. Le front distingue ce
// qu'il peut corriger de ce qu'il ne peut pas.
func codeDe(err error) int {
	switch {
	case errors.Is(err, ErrTropDeConnexions):
		return http.StatusConflict
	case errors.Is(err, config.ErrProfilInconnu):
		return http.StatusNotFound
	case errors.Is(err, errPiloteMuet):
		return http.StatusNotImplemented
	default:
		return http.StatusBadGateway
	}
}

// fermerConnexion referme et oublie. Fermer une session inconnue n'est pas une
// erreur : un onglet rechargé peut poster la même fermeture.
func (s *serveur) fermerConnexion(w http.ResponseWriter, r *http.Request) {
	var requete RequeteFermeture
	if !decoder(w, r, &requete) {
		return
	}
	if err := s.registre.fermer(requete.Session); err != nil {
		slog.Warn("fermeture de connexion", "error", err)
	}
	w.WriteHeader(http.StatusNoContent)
}

// composer rend le DSN à ouvrir, quelle que soit la forme reçue. La chaîne
// complète l'emporte quand les deux sont là : c'est la plus précise.
func (r RequeteConnexion) composer() (string, error) {
	if strings.TrimSpace(r.DSN) != "" {
		return strings.TrimSpace(r.DSN), nil
	}
	return introspection.Connexion{
		SGBD:        r.SGBD,
		Hote:        r.Hote,
		Port:        r.Port,
		Utilisateur: r.Utilisateur,
		MotDePasse:  r.MotDePasse,
		Base:        r.Base,
	}.DSN()
}

// decoder lit le corps JSON de la requête. Rend false quand il a déjà répondu.
func decoder(w http.ResponseWriter, r *http.Request, cible any) bool {
	decodeur := json.NewDecoder(http.MaxBytesReader(w, r.Body, tailleMaxCorps))
	decodeur.DisallowUnknownFields()
	if err := decodeur.Decode(cible); err != nil {
		// Le corps peut porter un mot de passe : l'erreur de décodage cite le
		// contenu fautif, elle ne sort pas d'ici.
		repondreErreur(w, http.StatusBadRequest, "corps de requête illisible")
		return false
	}
	return true
}

// repondreJSON écrit une réponse sérialisée.
//
// json.Marshal direct, et c'est l'exception assumée à la règle qui fait passer
// toute sérialisation par internal/calque : celle-ci existe pour le déterminisme
// octet pour octet des calques, qu'une réponse HTTP n'a pas à tenir. Le jour où
// l'API rendra un calque, c'est calque qui le sérialisera.
func repondreJSON(w http.ResponseWriter, code int, corps any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(corps); err != nil {
		slog.Error("ecriture de la reponse", "error", err)
	}
}

// repondreErreur écrit un échec sous la forme que le front sait afficher.
func repondreErreur(w http.ResponseWriter, code int, message string) {
	repondreJSON(w, code, ReponseErreur{Erreur: message})
}

// repondreRefus écrit un refus que le front traite à part : un conflit, avec le
// code qui dit lequel.
func repondreRefus(w http.ResponseWriter, code CodeRefus, message string) {
	repondreJSON(w, http.StatusConflict, ReponseErreur{Erreur: message, Code: code})
}

// sansDSN retire d'un message toute occurrence de la chaîne de connexion.
//
// Les pilotes masquent déjà le mot de passe, mais un DSN masqué reste un hôte,
// un utilisateur et un nom de base — et la règle est que le DSN ne sorte ni dans
// une réponse, ni dans un journal. Le remplacement porte sur les deux formes
// exactes, celle qu'on a composée et sa version masquée.
func sansDSN(message, dsn string) string {
	for _, forme := range []string{dsn, introspection.Masquer(dsn)} {
		if forme != "" {
			message = strings.ReplaceAll(message, forme, "la base")
		}
	}
	return message
}
