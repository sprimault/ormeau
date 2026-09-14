// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

// Package postgres introspecte PostgreSQL via pg_catalog.
//
// information_schema serait portable mais perd trop : IDENTITY contre séquence,
// colonnes générées, index partiels, méthodes d'index, commentaires, ordre réel
// des colonnes d'un index. C'est plus de code, et c'est là que se joue la
// qualité de l'outil.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/sprimault/ormeau/internal/introspection"
)

// Importer ce paquet suffit à rendre « postgres » utilisable.
func init() {
	introspection.Enregistrer("postgres", Ouvrir)
}

// delaiRequete plafonne chaque requête de catalogue. Une introspection lancée
// contre une base de production ne doit pas pouvoir s'y installer : mieux vaut
// échouer que rester accroché.
const delaiRequete = 30 * time.Second

// pilote détient la connexion. Non exporté : Ouvrir rend l'interface, et un
// pilote sans connexion ne doit pas pouvoir exister.
type pilote struct {
	conn *pgx.Conn
}

// Ouvrir établit la connexion et la bascule en lecture seule pour toute sa
// durée. L'absence d'écriture est un invariant : la faire tenir par le serveur
// plutôt que par la discipline du code est ce qui la rend vérifiable.
//
// Le DSN ne ressort jamais tel quel d'une erreur, seulement masqué.
func Ouvrir(ctx context.Context, dsn string) (introspection.Introspecteur, error) {
	config, err := pgx.ParseConfig(introspection.NettoyerDSN(dsn))
	if err != nil {
		// L'erreur de pgx n'est ni enveloppée ni citée. Elle recopie la chaîne
		// avec son propre masquage, qui ne reconnaît que « password= » collé :
		// « password = secret » y passe en clair. Filtrer un message qu'on ne
		// contrôle pas reviendrait à attendre la prochaine fuite ; le DSN masqué
		// montre de toute façon la valeur fautive.
		return nil, fmt.Errorf("dsn illisible (%s)", introspection.Masquer(dsn))
	}

	// Posé à la connexion plutôt que par un SET ensuite : il n'existe alors
	// aucune fenêtre pendant laquelle la session serait inscriptible. Vaut pour
	// toute sa durée, échantillonnage compris.
	if config.RuntimeParams == nil {
		config.RuntimeParams = map[string]string{}
	}
	config.RuntimeParams["default_transaction_read_only"] = "on"
	config.RuntimeParams["application_name"] = "ormeau"

	// pg_get_expr, pg_get_constraintdef, pg_get_viewdef et format_type
	// qualifient un nom selon le search_path de la session, qui dépend du rôle
	// et du DSN : deux extractions de la même base divergeraient. Vide, comme
	// pg_dump, tout objet hors de pg_catalog sort qualifié, quelle que soit la
	// connexion. Il remplace celui du DSN, que NettoyerDSN laisse passer pour
	// les autres pilotes. Il ferme aussi CVE-2018-1058 sur nos propres requêtes :
	// un opérateur posé dans public ne peut plus s'y substituer à celui du
	// catalogue.
	config.RuntimeParams["search_path"] = ""

	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("connexion a %s: %w", introspection.Masquer(dsn), err)
	}

	p := &pilote{conn: conn}
	if err := p.controlerSession(ctx); err != nil {
		_ = p.Fermer()
		return nil, err
	}
	return p, nil
}

// controlerSession relit ce que la connexion devait poser. Un paramètre de
// démarrage n'est qu'une demande : un pooler qui ignore ceux qu'il ne connaît
// pas (ignore_startup_parameters de pgbouncer) les perd sans erreur, et la
// session ne serait alors ni en lecture seule ni sous le search_path vide.
// Refuser vaut mieux que d'extraire en croyant tenir l'invariant.
func (p *pilote) controlerSession(ctx context.Context) error {
	ctx, annuler := context.WithTimeout(ctx, delaiRequete)
	defer annuler()

	var chemin, lectureSeule string
	if err := p.conn.QueryRow(ctx, requeteSession).Scan(&chemin, &lectureSeule); err != nil {
		return fmt.Errorf("lecture des parametres de session: %w", err)
	}
	return verifierSession(chemin, lectureSeule)
}

// verifierSession dit ce qui manque à la session, séparée de la lecture pour
// se tester sans serveur.
func verifierSession(chemin, lectureSeule string) error {
	if lectureSeule != "on" {
		return fmt.Errorf("session inscriptible (transaction_read_only = %q) : un intermediaire a ignore default_transaction_read_only, extraction refusee", lectureSeule)
	}
	if chemin != "" {
		return fmt.Errorf("search_path de session %q au lieu d'un chemin vide : un intermediaire a ignore le parametre, extraction refusee", chemin)
	}
	return nil
}

// Inventorier alimente l'arbre de sélection sans introspecter : une requête,
// aucune lecture de données.
func (p *pilote) Inventorier(ctx context.Context, schemas []string) ([]introspection.TableSommaire, error) {
	if len(schemas) == 0 {
		schemas = []string{"public"}
	}

	ctx, annuler := context.WithTimeout(ctx, delaiRequete)
	defer annuler()

	lignes, err := p.conn.Query(ctx, requeteInventaire, schemas)
	if err != nil {
		return nil, fmt.Errorf("inventaire des tables: %w", err)
	}
	defer lignes.Close()

	// Jamais nil : une portée sans table est un résultat vide, pas une absence
	// de résultat.
	sommaires := []introspection.TableSommaire{}
	for lignes.Next() {
		var s introspection.TableSommaire
		var commentaire *string

		if err := lignes.Scan(&s.Schema, &s.Nom, &commentaire, &s.NbColonnes,
			&s.LignesEstimees, &s.ClePrimaire, &s.ReferenceVers); err != nil {
			return nil, fmt.Errorf("lecture de l'inventaire: %w", err)
		}
		if commentaire != nil {
			s.Commentaire = *commentaire
		}
		// reltuples vaut -1 tant que la table n'a jamais été analysée.
		if s.LignesEstimees < 0 {
			s.LignesEstimees = 0
		}
		sommaires = append(sommaires, s)
	}
	if err := lignes.Err(); err != nil {
		return nil, fmt.Errorf("parcours de l'inventaire: %w", err)
	}
	return sommaires, nil
}

// Decrire rend ce que l'écran de connexion affiche en retour : le serveur
// atteint, sa version, et les schémas parmi lesquels choisir.
//
// Deux requêtes de catalogue, aucune donnée lue. Le DSN n'apparaît nulle part
// dans le résultat — la connexion a réussi, c'est tout ce que l'appelant a
// besoin de savoir d'elle.
func (p *pilote) Decrire(ctx context.Context) (introspection.Serveur, error) {
	ctx, annuler := context.WithTimeout(ctx, delaiRequete)
	defer annuler()

	s := introspection.Serveur{SGBD: "postgres"}
	if err := p.conn.QueryRow(ctx, requeteSource).Scan(&s.Version, &s.Catalogue); err != nil {
		return s, fmt.Errorf("lecture de la source: %w", err)
	}

	schemas, err := p.lireSchemas(ctx)
	if err != nil {
		return s, err
	}
	// Jamais nil : une base sans schéma exploitable est un résultat vide, que le
	// front affiche comme tel, pas une absence de réponse.
	s.Schemas = schemas
	if s.Schemas == nil {
		s.Schemas = []string{}
	}
	return s, nil
}

// Colonnes décrit une table pour l'écran de sélection, sans l'introspecter.
//
// Le type est rendu verbatim, comme partout ailleurs : c'est ce qui distingue un
// citext d'un text, et l'utilisateur qui décide d'écarter une colonne veut voir
// ce que sa base contient, pas une traduction.
func (p *pilote) Colonnes(ctx context.Context, schema, table string) ([]introspection.ColonneSommaire, error) {
	ctx, annuler := context.WithTimeout(ctx, delaiRequete)
	defer annuler()

	lignes, err := p.conn.Query(ctx, requeteColonnesSommaire, schema, table)
	if err != nil {
		return nil, fmt.Errorf("colonnes de %s.%s: %w", schema, table, err)
	}
	defer lignes.Close()

	colonnes := []introspection.ColonneSommaire{}
	for lignes.Next() {
		var c introspection.ColonneSommaire
		var commentaire *string

		if err := lignes.Scan(&c.Nom, &c.Position, &c.TypeBrut, &c.Nullable,
			&c.ClePrimaire, &commentaire); err != nil {
			return nil, fmt.Errorf("lecture d'une colonne: %w", err)
		}
		if commentaire != nil {
			c.Commentaire = *commentaire
		}
		colonnes = append(colonnes, c)
	}
	return colonnes, lignes.Err()
}

// Fermer libère la connexion.
func (p *pilote) Fermer() error {
	if p.conn == nil {
		return nil
	}
	// Contexte propre : celui de l'appelant est souvent déjà annulé au moment
	// de fermer, ce qui laisserait la connexion ouverte côté serveur.
	ctx, annuler := context.WithTimeout(context.Background(), 5*time.Second)
	defer annuler()
	return p.conn.Close(ctx)
}
