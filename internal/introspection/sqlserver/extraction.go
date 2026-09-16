// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package sqlserver

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sprimault/ormeau/internal/calque"
	"github.com/sprimault/ormeau/internal/introspection"
)

// extractionComplete garde Extraire fermée tant que le calque n'a ni index, ni
// CHECK, ni séquences, ni colonnes calculées, ni commentaires, ni vues : il
// passerait pour complet sans l'être. Les tests d'intégration appellent
// extraire directement en attendant ces lectures.
const extractionComplete = false

// Extraire lit le catalogue et rend un calque trié.
func (p *pilote) Extraire(ctx context.Context, portee introspection.Portee) (*calque.Physique, error) {
	if !extractionComplete {
		return nil, errors.New("extraction SQL Server pas encore ecrite en entier : index, CHECK, sequences, colonnes calculees, commentaires et vues manquent")
	}
	return p.extraire(ctx, portee)
}

// extraire lit le cœur du catalogue : tables, colonnes, défauts, clés
// primaires et étrangères.
//
// Chaque passe interroge le catalogue schéma par schéma, le nom passant en
// paramètre : SQL Server n'a pas de tableau paramétrable, et composer une
// liste IN reviendrait à écrire du SQL avec des noms venus de l'appelant.
func (p *pilote) extraire(ctx context.Context, portee introspection.Portee) (*calque.Physique, error) {
	schemas := portee.Schemas
	if len(schemas) == 0 {
		// Tous les schémas qui portent une table, et non dbo seul : une base
		// reprise range volontiers ailleurs.
		ctxSchemas, annuler := context.WithTimeout(ctx, delaiRequete)
		defer annuler()
		var err error
		if schemas, err = p.schemas(ctxSchemas); err != nil {
			return nil, err
		}
	}
	if len(schemas) == 0 {
		return nil, errors.New("aucun schema exploitable dans cette base")
	}

	physique := &calque.Physique{VersionRI: calque.VersionCourante}
	jeu := nouveauJeu()

	err := introspection.Derouler(ctx,
		introspection.Passe{Etape: introspection.EtapeSource, Lire: func(ctx context.Context) (err error) {
			physique.Source, err = p.lireSource(ctx, schemas[0])
			return err
		}},
		introspection.Passe{Etape: introspection.EtapeTables, Lire: func(ctx context.Context) error {
			return parSchema(ctx, schemas, func(ctx context.Context, schema string) error {
				return p.lireTables(ctx, schema, jeu)
			})
		}},
		introspection.Passe{Etape: introspection.EtapeColonnes, Lire: func(ctx context.Context) error {
			return parSchema(ctx, schemas, func(ctx context.Context, schema string) error {
				return p.lireColonnes(ctx, schema, jeu)
			})
		}},
		introspection.Passe{Etape: introspection.EtapeContraintes, Lire: func(ctx context.Context) error {
			return parSchema(ctx, schemas, func(ctx context.Context, schema string) error {
				if err := p.lireClesPrimaires(ctx, schema, jeu); err != nil {
					return err
				}
				return p.lireClesEtrangeres(ctx, schema, jeu)
			})
		}},
	)
	if err != nil {
		return nil, err
	}

	physique.Tables = jeu.retenues(portee)
	physique.Trier()
	return physique, nil
}

// parSchema répète une lecture pour chaque schéma, chacune sous son propre
// délai : c'est la requête qui est plafonnée, pas la passe, dont la durée
// grandit avec le nombre de schémas.
func parSchema(ctx context.Context, schemas []string, lire func(context.Context, string) error) error {
	for _, schema := range schemas {
		ctxRequete, annuler := context.WithTimeout(ctx, delaiRequete)
		err := lire(ctxRequete, schema)
		annuler()
		if err != nil {
			return err
		}
	}
	return nil
}

// jeuDeTables indexe les tables par leur nom qualifié pendant la collecte. Les
// passes suivantes complètent des pointeurs plutôt que de rechercher
// linéairement à chaque ligne.
type jeuDeTables struct {
	parCle map[string]*calque.Table
	ordre  []string
}

// nouveauJeu rend un jeu vide, prêt pour la passe des tables.
func nouveauJeu() *jeuDeTables {
	return &jeuDeTables{parCle: map[string]*calque.Table{}}
}

// ajouter enregistre une table et retient son rang d'arrivée.
func (j *jeuDeTables) ajouter(t *calque.Table) {
	cle := t.Schema + "." + t.Nom
	j.parCle[cle] = t
	j.ordre = append(j.ordre, cle)
}

// trouver rend la table à compléter, ou nil pour une table apparue pendant
// l'extraction, qu'on ignore plutôt que d'en fabriquer une partielle.
func (j *jeuDeTables) trouver(schema, nom string) *calque.Table {
	return j.parCle[schema+"."+nom]
}

// retenues applique la portée en fin de collecte : une clé étrangère vers une
// table exclue reste visible dans le calque, c'est la validation qui la
// signale.
func (j *jeuDeTables) retenues(portee introspection.Portee) []calque.Table {
	incluses := ensemble(portee.TablesIncluses)
	exclues := ensemble(portee.TablesExclues)

	tables := make([]calque.Table, 0, len(j.ordre))
	for _, cle := range j.ordre {
		if incluses != nil && !incluses[cle] {
			continue
		}
		if exclues[cle] {
			continue
		}
		tables = append(tables, *j.parCle[cle])
	}
	return tables
}

// lireSource renseigne l'en-tête du calque. La version seule, sans l'édition
// que Decrire affiche : l'édition dit ce que le serveur permet, pas ce que la
// base contient, et elle n'a pas à entrer dans l'empreinte.
func (p *pilote) lireSource(ctx context.Context, schema string) (calque.Source, error) {
	ctx, annuler := context.WithTimeout(ctx, delaiRequete)
	defer annuler()

	s := calque.Source{SGBD: "sqlserver", Schema: schema}
	var edition string
	if err := p.db.QueryRowContext(ctx, requeteServeur).Scan(&s.Version, &edition, &s.Catalogue); err != nil {
		return s, fmt.Errorf("lecture de la source: %w", err)
	}
	return s, nil
}

// lireTables collecte les tables d'un schéma.
func (p *pilote) lireTables(ctx context.Context, schema string, jeu *jeuDeTables) error {
	lignes, err := p.db.QueryContext(ctx, requeteTablesExtraction, schema)
	if err != nil {
		return fmt.Errorf("lecture des tables de %s: %w", schema, err)
	}
	defer func() { _ = lignes.Close() }()

	for lignes.Next() {
		t := &calque.Table{Schema: schema}
		if err := lignes.Scan(&t.Nom); err != nil {
			return fmt.Errorf("lecture d'une table: %w", err)
		}
		jeu.ajouter(t)
	}
	return lignes.Err()
}

// cleColonne identifie une colonne pendant le rapprochement des types.
type cleColonne struct {
	table, colonne string
}

// lireColonnes complète les tables d'un schéma : type verbatim et normalisé,
// longueur, précision, échelle, nullabilité, identité et défaut.
//
// Le type verbatim vient d'une autre requête, rapproché par nom : sys.columns
// porte toutes les colonnes, la description du serveur seulement celles qu'un
// SELECT * rend. Une colonne sans type décrit arrête l'extraction plutôt que
// de sortir avec un type_brut inventé.
func (p *pilote) lireColonnes(ctx context.Context, schema string, jeu *jeuDeTables) error {
	types, err := p.lireTypesDecrits(ctx, schema)
	if err != nil {
		return err
	}

	lignes, err := p.db.QueryContext(ctx, requeteColonnesExtraction, schema)
	if err != nil {
		return fmt.Errorf("lecture des colonnes de %s: %w", schema, err)
	}
	defer func() { _ = lignes.Close() }()

	for lignes.Next() {
		var table, typeSysteme string
		var maxLength, precision, echelle int
		var identite bool
		var defaut sql.NullString

		c := calque.Colonne{}
		if err := lignes.Scan(&table, &c.Nom, &c.Position, &typeSysteme,
			&maxLength, &precision, &echelle, &c.Nullable, &identite, &defaut); err != nil {
			return fmt.Errorf("lecture d'une colonne: %w", err)
		}

		typeBrut, decrit := types[cleColonne{table, c.Nom}]
		if !decrit {
			return fmt.Errorf("colonne %s.%s.%s: type non decrit par le serveur", schema, table, c.Nom)
		}
		c.TypeBrut = typeBrut
		c.TypeNormalise = normaliserType(typeSysteme)
		c.Longueur = longueur(typeSysteme, maxLength)
		c.Precision, c.Echelle = precisionEchelle(typeSysteme, precision, echelle)
		// IDENTITY refuse une valeur explicite sans SET IDENTITY_INSERT, option
		// de session : c'est la nature « toujours » du calque.
		if identite {
			c.AutoIncrement, c.Identite = true, calque.IdentiteToujours
		}
		if defaut.Valid {
			c.Defaut = classerDefaut(defaut.String)
		}

		t := jeu.trouver(schema, table)
		if t == nil {
			continue
		}
		t.Colonnes = append(t.Colonnes, c)
	}
	return lignes.Err()
}

// lireTypesDecrits rend le type de chaque colonne tel que le serveur l'écrit.
//
// Le nom d'un type alias ou CLR l'emporte sur le type système : c'est lui que
// la déclaration porte, comme un domaine sous PostgreSQL. Le type système, lui,
// reste lisible par la normalisation et la longueur.
func (p *pilote) lireTypesDecrits(ctx context.Context, schema string) (map[cleColonne]string, error) {
	lignes, err := p.db.QueryContext(ctx, requeteTypesDecrits, schema)
	if err != nil {
		return nil, fmt.Errorf("description des types de %s: %w", schema, err)
	}
	defer func() { _ = lignes.Close() }()

	types := map[cleColonne]string{}
	for lignes.Next() {
		var table string
		var colonne, systeme, utilisateur, message sql.NullString
		var numero sql.NullInt64

		if err := lignes.Scan(&table, &colonne, &systeme, &utilisateur, &numero, &message); err != nil {
			return nil, fmt.Errorf("lecture d'un type decrit: %w", err)
		}
		// La première ligne d'erreur est la cause, les suivantes en découlent
		// (11529 suit toujours l'erreur réelle).
		if numero.Valid {
			return nil, fmt.Errorf("description des types de %s.%s: erreur %d: %s", schema, table, numero.Int64, message.String)
		}
		typeBrut := systeme.String
		if utilisateur.Valid {
			typeBrut = utilisateur.String
		}
		types[cleColonne{table, colonne.String}] = typeBrut
	}
	return types, lignes.Err()
}

// lireClesPrimaires complète les clés primaires d'un schéma.
func (p *pilote) lireClesPrimaires(ctx context.Context, schema string, jeu *jeuDeTables) error {
	lignes, err := p.db.QueryContext(ctx, requeteClesPrimaires, schema)
	if err != nil {
		return fmt.Errorf("lecture des cles primaires de %s: %w", schema, err)
	}
	defer func() { _ = lignes.Close() }()

	for lignes.Next() {
		var table, nom, colonne string
		if err := lignes.Scan(&table, &nom, &colonne); err != nil {
			return fmt.Errorf("lecture d'une cle primaire: %w", err)
		}
		t := jeu.trouver(schema, table)
		if t == nil {
			continue
		}
		if t.ClePrimaire == nil {
			t.ClePrimaire = &calque.ClePrimaire{Nom: nom}
		}
		t.ClePrimaire.Colonnes = append(t.ClePrimaire.Colonnes, colonne)
	}
	return lignes.Err()
}

// lireClesEtrangeres complète les clés étrangères d'un schéma. Les lignes
// d'une même contrainte se suivent, triées par nom puis par rang de colonne.
func (p *pilote) lireClesEtrangeres(ctx context.Context, schema string, jeu *jeuDeTables) error {
	lignes, err := p.db.QueryContext(ctx, requeteClesEtrangeres, schema)
	if err != nil {
		return fmt.Errorf("lecture des cles etrangeres de %s: %w", schema, err)
	}
	defer func() { _ = lignes.Close() }()

	for lignes.Next() {
		var table, nom, schemaCible, tableCible, suppression, miseAJour, colonne, colonneCible string
		if err := lignes.Scan(&table, &nom, &schemaCible, &tableCible,
			&suppression, &miseAJour, &colonne, &colonneCible); err != nil {
			return fmt.Errorf("lecture d'une cle etrangere: %w", err)
		}
		t := jeu.trouver(schema, table)
		if t == nil {
			continue
		}

		dernier := len(t.ClesEtrangeres) - 1
		if dernier < 0 || t.ClesEtrangeres[dernier].Nom != nom {
			fk, err := nouvelleCleEtrangere(nom, schemaCible, tableCible, suppression, miseAJour)
			if err != nil {
				return fmt.Errorf("cle etrangere %s.%s.%s: %w", schema, table, nom, err)
			}
			t.ClesEtrangeres = append(t.ClesEtrangeres, fk)
			dernier++
		}
		fk := &t.ClesEtrangeres[dernier]
		fk.Colonnes = append(fk.Colonnes, colonne)
		fk.ColonnesCibles = append(fk.ColonnesCibles, colonneCible)
	}
	return lignes.Err()
}

// nouvelleCleEtrangere rend une clé sans colonnes. Une action inconnue arrête
// l'extraction : la taire la confondrait avec NO_ACTION.
func nouvelleCleEtrangere(nom, schemaCible, tableCible, suppression, miseAJour string) (calque.CleEtrangere, error) {
	aLaSuppression, connue := actionReferentielle(suppression)
	if !connue {
		return calque.CleEtrangere{}, fmt.Errorf("action %q inconnue", suppression)
	}
	aLaMiseAJour, connue := actionReferentielle(miseAJour)
	if !connue {
		return calque.CleEtrangere{}, fmt.Errorf("action %q inconnue", miseAJour)
	}
	return calque.CleEtrangere{
		Nom:            nom,
		SchemaCible:    schemaCible,
		TableCible:     tableCible,
		ALaSuppression: aLaSuppression,
		ALaMiseAJour:   aLaMiseAJour,
	}, nil
}
