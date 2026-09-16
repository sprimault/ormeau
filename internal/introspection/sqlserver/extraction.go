// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package sqlserver

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/sprimault/ormeau/internal/calque"
	"github.com/sprimault/ormeau/internal/introspection"
)

// Extraire lit le catalogue et rend un calque trié.
//
// Sept passes, déroulées par introspection.Derouler : celle des types
// énumérés n'existe pas, SQL Server n'en a pas de nommés. Chaque passe
// interroge le catalogue schéma par schéma, le nom passant en paramètre : SQL
// Server n'a pas de tableau paramétrable, et composer une liste IN reviendrait
// à écrire du SQL avec des noms venus de l'appelant.
//
// Ce qui n'est pas capturé ici est perdu : aucune couche en aval ne peut le
// retrouver.
func (p *pilote) Extraire(ctx context.Context, portee introspection.Portee) (*calque.Physique, error) {
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
				if err := p.lireClesEtUnicites(ctx, schema, jeu); err != nil {
					return err
				}
				if err := p.lireClesEtrangeres(ctx, schema, jeu); err != nil {
					return err
				}
				return p.lireVerifications(ctx, schema, jeu)
			})
		}},
		introspection.Passe{Etape: introspection.EtapeIndex, Lire: func(ctx context.Context) error {
			return parSchema(ctx, schemas, func(ctx context.Context, schema string) error {
				return p.lireIndex(ctx, schema, jeu)
			})
		}},
		introspection.Passe{Etape: introspection.EtapeSequences, Lire: func(ctx context.Context) error {
			return parSchema(ctx, schemas, func(ctx context.Context, schema string) (err error) {
				physique.Sequences, err = p.lireSequences(ctx, schema, physique.Sequences)
				return err
			})
		}},
		introspection.Passe{Etape: introspection.EtapeVues, Lire: func(ctx context.Context) error {
			return parSchema(ctx, schemas, func(ctx context.Context, schema string) (err error) {
				physique.Vues, err = p.lireVues(ctx, schema, physique.Vues)
				return err
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
		// Relu à chaque schéma, comme Derouler le fait à chaque passe : le
		// pilote rend sa propre erreur pour une requête annulée, et l'appelant
		// doit pouvoir reconnaître l'annulation.
		if err := ctx.Err(); err != nil {
			return err
		}
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

// lireTables collecte les tables d'un schéma et leur commentaire.
func (p *pilote) lireTables(ctx context.Context, schema string, jeu *jeuDeTables) error {
	lignes, err := p.db.QueryContext(ctx, requeteTablesExtraction, schema)
	if err != nil {
		return fmt.Errorf("lecture des tables de %s: %w", schema, err)
	}
	defer func() { _ = lignes.Close() }()

	for lignes.Next() {
		t := &calque.Table{Schema: schema}
		var commentaire sql.NullString
		if err := lignes.Scan(&t.Nom, &commentaire); err != nil {
			return fmt.Errorf("lecture d'une table: %w", err)
		}
		t.Commentaire = commentaire.String
		jeu.ajouter(t)
	}
	return lignes.Err()
}

// cleColonne identifie une colonne pendant le rapprochement des types.
type cleColonne struct {
	table, colonne string
}

// lireColonnes complète les tables d'un schéma : type verbatim et normalisé,
// longueur, précision, échelle, nullabilité, identité, défaut, calcul,
// collation et commentaire.
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
		var defaut, calcul, collation, commentaire sql.NullString
		var persistee sql.NullBool

		c := calque.Colonne{}
		if err := lignes.Scan(&table, &c.Nom, &c.Position, &typeSysteme,
			&maxLength, &precision, &echelle, &c.Nullable, &identite, &defaut,
			&calcul, &persistee, &collation, &commentaire); err != nil {
			return fmt.Errorf("lecture d'une colonne: %w", err)
		}

		typeBrut, decrit := types[cleColonne{table, c.Nom}]
		if !decrit {
			return fmt.Errorf("colonne %s.%s.%s: type non decrit par le serveur", schema, table, c.Nom)
		}
		c.TypeBrut = typeBrut
		c.TypeNormalise = normaliserType(typeSysteme)
		c.Longueur = longueur(typeSysteme, maxLength)
		c.LongueurFixe = longueurFixe(typeSysteme)
		c.Fuseau = typeSysteme == "datetimeoffset"
		c.PrecisionFractionnaire = precisionFractionnaire(typeSysteme, echelle)
		c.Precision, c.Echelle = precisionEchelle(typeSysteme, precision, echelle)
		// IDENTITY refuse une valeur explicite sans SET IDENTITY_INSERT, option
		// de session : c'est la nature « toujours » du calque.
		if identite {
			c.AutoIncrement, c.Identite = true, calque.IdentiteToujours
		}
		if defaut.Valid {
			c.Defaut = classerDefaut(defaut.String)
		}
		// Une colonne calculée n'accepte pas de contrainte DEFAULT : le calcul
		// est tout ce qu'elle porte. is_persisted distingue la valeur stockée
		// de celle que le serveur calcule à chaque lecture.
		if calcul.Valid {
			c.Generee = &calque.Generee{Expression: calcul.String, Stockee: persistee.Bool}
		}
		c.Collation = collation.String
		c.Commentaire = commentaire.String

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

// lireClesEtUnicites complète les clés primaires et les contraintes d'unicité
// d'un schéma. Les lignes d'une même contrainte se suivent.
func (p *pilote) lireClesEtUnicites(ctx context.Context, schema string, jeu *jeuDeTables) error {
	lignes, err := p.db.QueryContext(ctx, requeteClesEtUnicites, schema)
	if err != nil {
		return fmt.Errorf("lecture des cles et unicites de %s: %w", schema, err)
	}
	defer func() { _ = lignes.Close() }()

	for lignes.Next() {
		var table, genre, nom, colonne string
		if err := lignes.Scan(&table, &genre, &nom, &colonne); err != nil {
			return fmt.Errorf("lecture d'une cle: %w", err)
		}
		t := jeu.trouver(schema, table)
		if t == nil {
			continue
		}
		if genre == "PK" {
			if t.ClePrimaire == nil {
				t.ClePrimaire = &calque.ClePrimaire{Nom: nom}
			}
			t.ClePrimaire.Colonnes = append(t.ClePrimaire.Colonnes, colonne)
			continue
		}
		dernier := len(t.Unicites) - 1
		if dernier < 0 || t.Unicites[dernier].Nom != nom {
			t.Unicites = append(t.Unicites, calque.Contrainte{Nom: nom})
			dernier++
		}
		t.Unicites[dernier].Colonnes = append(t.Unicites[dernier].Colonnes, colonne)
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

// lireVerifications complète les CHECK d'un schéma, verbatim.
func (p *pilote) lireVerifications(ctx context.Context, schema string, jeu *jeuDeTables) error {
	lignes, err := p.db.QueryContext(ctx, requeteVerifications, schema)
	if err != nil {
		return fmt.Errorf("lecture des verifications de %s: %w", schema, err)
	}
	defer func() { _ = lignes.Close() }()

	for lignes.Next() {
		var table string
		var v calque.Verification
		if err := lignes.Scan(&table, &v.Nom, &v.Expression); err != nil {
			return fmt.Errorf("lecture d'une verification: %w", err)
		}
		if t := jeu.trouver(schema, table); t != nil {
			t.Verifications = append(t.Verifications, v)
		}
	}
	return lignes.Err()
}

// lireIndex complète les index d'un schéma. Les lignes d'un même index se
// suivent, dans l'ordre de ses colonnes.
//
// La méthode est type_desc en minuscules : clustered ou nonclustered, deux
// arbres qui ne rangent pas la table de la même façon. Les sens de tri ne sont
// reportés que si l'un d'eux est descendant, comme les classes d'opérateurs
// sous PostgreSQL.
func (p *pilote) lireIndex(ctx context.Context, schema string, jeu *jeuDeTables) error {
	lignes, err := p.db.QueryContext(ctx, requeteIndex, schema)
	if err != nil {
		return fmt.Errorf("lecture des index de %s: %w", schema, err)
	}
	defer func() { _ = lignes.Close() }()

	for lignes.Next() {
		var table, nom, genre, colonne string
		var unique, desc bool
		var filtre sql.NullString
		if err := lignes.Scan(&table, &nom, &genre, &unique, &filtre, &colonne, &desc); err != nil {
			return fmt.Errorf("lecture d'un index: %w", err)
		}
		t := jeu.trouver(schema, table)
		if t == nil {
			continue
		}

		dernier := len(t.Index) - 1
		if dernier < 0 || t.Index[dernier].Nom != nom {
			t.Index = append(t.Index, calque.Index{
				Nom:      nom,
				Unique:   unique,
				Methode:  strings.ToLower(genre),
				Predicat: filtre.String,
			})
			dernier++
		}
		idx := &t.Index[dernier]
		idx.Colonnes = append(idx.Colonnes, colonne)
		ordre := calque.OrdreAscendant
		if desc {
			ordre = calque.OrdreDescendant
		}
		idx.Ordres = append(idx.Ordres, ordre)
	}
	if err := lignes.Err(); err != nil {
		return err
	}

	for _, cle := range jeu.ordre {
		t := jeu.parCle[cle]
		if t.Schema != schema {
			continue
		}
		for i := range t.Index {
			if !slices.Contains(t.Index[i].Ordres, calque.OrdreDescendant) {
				t.Index[i].Ordres = nil
			}
		}
	}
	return nil
}

// lireSequences ajoute les séquences d'un schéma à celles déjà lues.
func (p *pilote) lireSequences(ctx context.Context, schema string, sequences []calque.Sequence) ([]calque.Sequence, error) {
	lignes, err := p.db.QueryContext(ctx, requeteSequences, schema)
	if err != nil {
		return nil, fmt.Errorf("lecture des sequences de %s: %w", schema, err)
	}
	defer func() { _ = lignes.Close() }()

	for lignes.Next() {
		s := calque.Sequence{Schema: schema}
		var depart, increment, minimum, maximum string
		if err := lignes.Scan(&s.Nom, &depart, &increment, &minimum, &maximum, &s.Cyclique); err != nil {
			return nil, fmt.Errorf("lecture d'une sequence: %w", err)
		}

		valeurs := make([]int64, 4)
		for i, texte := range []string{depart, increment, minimum, maximum} {
			if valeurs[i], err = strconv.ParseInt(texte, 10, 64); err != nil {
				return nil, fmt.Errorf("sequence %s.%s: valeur %s hors d'un entier de 64 bits", schema, s.Nom, texte)
			}
		}
		s.Depart, s.Increment, s.Minimum, s.Maximum = &valeurs[0], valeurs[1], &valeurs[2], &valeurs[3]
		sequences = append(sequences, s)
	}
	return sequences, lignes.Err()
}

// lireVues ajoute les vues d'un schéma à celles déjà lues. Une vue indexée
// n'est pas une vue matérialisée au sens de PostgreSQL — elle se met à jour
// avec ses tables — et reste marquée comme une vue ordinaire.
func (p *pilote) lireVues(ctx context.Context, schema string, vues []calque.Vue) ([]calque.Vue, error) {
	lignes, err := p.db.QueryContext(ctx, requeteVues, schema)
	if err != nil {
		return nil, fmt.Errorf("lecture des vues de %s: %w", schema, err)
	}
	defer func() { _ = lignes.Close() }()

	for lignes.Next() {
		v := calque.Vue{Schema: schema}
		var definition sql.NullString
		if err := lignes.Scan(&v.Nom, &definition); err != nil {
			return nil, fmt.Errorf("lecture d'une vue: %w", err)
		}
		v.Definition = definition.String
		vues = append(vues, v)
	}
	return vues, lignes.Err()
}
