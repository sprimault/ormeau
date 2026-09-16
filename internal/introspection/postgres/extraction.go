// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/sprimault/ormeau/internal/calque"
	"github.com/sprimault/ormeau/internal/introspection"
)

// Extraire lit le catalogue et rend un calque trié.
//
// Huit passes séquentielles, déroulées par introspection.Derouler qui en
// signale l'avancement. Chacune interroge le catalogue pour tous les schémas à
// la fois. Rien n'est parallélisé — le gain serait marginal devant le risque de
// rendre l'ordre du résultat dépendant de l'ordonnancement.
//
// Ce qui n'est pas capturé ici est perdu : aucune couche en aval ne peut le
// retrouver.
func (p *pilote) Extraire(ctx context.Context, portee introspection.Portee) (*calque.Physique, error) {
	schemas := portee.Schemas
	if len(schemas) == 0 {
		// Tous les schémas de la base, et non « public » seul : une base
		// legacy range rarement dans public, et un balayage de serveur ne peut
		// pas deviner les schémas de chacune.
		var err error
		if schemas, err = p.lireSchemas(ctx); err != nil {
			return nil, err
		}
	}
	if len(schemas) == 0 {
		return nil, errors.New("aucun schema exploitable dans cette base")
	}

	physique := &calque.Physique{VersionRI: calque.VersionCourante}
	var tables *jeuDeTables

	err := introspection.Derouler(ctx,
		introspection.Passe{Etape: introspection.EtapeSource, Lire: func(ctx context.Context) (err error) {
			physique.Source, err = p.lireSource(ctx, schemas)
			return err
		}},
		introspection.Passe{Etape: introspection.EtapeTables, Lire: func(ctx context.Context) (err error) {
			tables, err = p.lireTables(ctx, schemas)
			return err
		}},
		introspection.Passe{Etape: introspection.EtapeColonnes, Lire: func(ctx context.Context) error {
			return p.lireColonnes(ctx, schemas, tables)
		}},
		introspection.Passe{Etape: introspection.EtapeContraintes, Lire: func(ctx context.Context) error {
			return p.lireContraintes(ctx, schemas, tables)
		}},
		introspection.Passe{Etape: introspection.EtapeIndex, Lire: func(ctx context.Context) error {
			return p.lireIndex(ctx, schemas, tables)
		}},
		introspection.Passe{Etape: introspection.EtapeSequences, Lire: func(ctx context.Context) (err error) {
			physique.Sequences, err = p.lireSequences(ctx, schemas)
			return err
		}},
		introspection.Passe{Etape: introspection.EtapeTypesEnumeres, Lire: func(ctx context.Context) (err error) {
			physique.TypesEnumeres, err = p.lireTypesEnumeres(ctx, schemas)
			return err
		}},
		introspection.Passe{Etape: introspection.EtapeVues, Lire: func(ctx context.Context) (err error) {
			physique.Vues, err = p.lireVues(ctx, schemas)
			return err
		}},
	)
	if err != nil {
		return nil, err
	}

	physique.Tables = tables.retenues(portee)
	physique.Trier()
	return physique, nil
}

// lireSchemas rend les schémas de la base, ceux du système exclus.
func (p *pilote) lireSchemas(ctx context.Context) ([]string, error) {
	ctx, annuler := context.WithTimeout(ctx, delaiRequete)
	defer annuler()

	lignes, err := p.conn.Query(ctx, requeteSchemas)
	if err != nil {
		return nil, fmt.Errorf("liste des schemas: %w", err)
	}
	defer lignes.Close()

	var schemas []string
	for lignes.Next() {
		var nom string
		if err := lignes.Scan(&nom); err != nil {
			return nil, fmt.Errorf("lecture d'un schema: %w", err)
		}
		schemas = append(schemas, nom)
	}
	return schemas, lignes.Err()
}

// ListerBases rend les bases exploitables du serveur.
//
// La base système postgres est écartée comme les modèles : elle existe sur
// toute installation et ne porte rien de métier. Quelqu'un qui l'introspecte
// vraiment la nomme.
func (p *pilote) ListerBases(ctx context.Context) ([]string, error) {
	ctx, annuler := context.WithTimeout(ctx, delaiRequete)
	defer annuler()

	lignes, err := p.conn.Query(ctx, requeteBases)
	if err != nil {
		return nil, fmt.Errorf("liste des bases: %w", err)
	}
	defer lignes.Close()

	bases := []string{}
	for lignes.Next() {
		var nom string
		if err := lignes.Scan(&nom); err != nil {
			return nil, fmt.Errorf("lecture d'une base: %w", err)
		}
		bases = append(bases, nom)
	}
	return bases, lignes.Err()
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

// ajouter enregistre une table et retient son rang d'arrivée. L'ordre n'est pas
// celui du calque final — Trier s'en charge — mais il rend la collecte
// reproductible, ce qui aide au diagnostic.
func (j *jeuDeTables) ajouter(t *calque.Table) {
	cle := t.Schema + "." + t.Nom
	j.parCle[cle] = t
	j.ordre = append(j.ordre, cle)
}

// trouver rend la table à compléter, ou nil si le catalogue rend une ligne pour
// une table qu'on n'a pas collectée.
func (j *jeuDeTables) trouver(schema, nom string) *calque.Table {
	return j.parCle[schema+"."+nom]
}

// retenues applique la portée. Le filtrage arrive en fin de collecte et non
// dans les requêtes : une clé étrangère vers une table exclue doit rester
// visible dans le calque, c'est la validation qui la signalera.
func (j *jeuDeTables) retenues(portee introspection.Portee) []calque.Table {
	incluses := ensemble(portee.TablesIncluses)
	exclues := ensemble(portee.TablesExclues)

	tables := make([]calque.Table, 0, len(j.ordre))
	for _, cle := range j.ordre {
		if len(incluses) > 0 && !incluses[cle] {
			continue
		}
		if exclues[cle] {
			continue
		}
		tables = append(tables, *j.parCle[cle])
	}
	return tables
}

// ensemble indexe une liste de portée pour l'interroger par appartenance.
func ensemble(valeurs []string) map[string]bool {
	if len(valeurs) == 0 {
		return nil
	}
	e := make(map[string]bool, len(valeurs))
	for _, v := range valeurs {
		e[v] = true
	}
	return e
}

// lireSource renseigne l'en-tête du calque. Ni empreinte ni horodatage ici :
// l'une se calcule à l'écriture, l'autre est posé par la commande.
func (p *pilote) lireSource(ctx context.Context, schemas []string) (calque.Source, error) {
	ctx, annuler := context.WithTimeout(ctx, delaiRequete)
	defer annuler()

	var s calque.Source
	s.SGBD = "postgres"
	// Le schéma du calque en porte un seul : le premier demandé fait foi, les
	// autres restent lisibles table par table.
	s.Schema = schemas[0]

	if err := p.conn.QueryRow(ctx, requeteSource).Scan(&s.Version, &s.Catalogue); err != nil {
		return s, fmt.Errorf("lecture de la source: %w", err)
	}
	return s, nil
}

// lireTables collecte les tables ordinaires et partitionnées. Les vues sont
// lues à part : elles n'ont ni contrainte ni index à compléter.
func (p *pilote) lireTables(ctx context.Context, schemas []string) (*jeuDeTables, error) {
	ctx, annuler := context.WithTimeout(ctx, delaiRequete)
	defer annuler()

	lignes, err := p.conn.Query(ctx, requeteTables, schemas)
	if err != nil {
		return nil, fmt.Errorf("lecture des tables: %w", err)
	}
	defer lignes.Close()

	jeu := nouveauJeu()
	for lignes.Next() {
		t := &calque.Table{}
		var commentaire *string

		if err := lignes.Scan(&t.Schema, &t.Nom, &commentaire); err != nil {
			return nil, fmt.Errorf("lecture d'une table: %w", err)
		}
		if commentaire != nil {
			t.Commentaire = *commentaire
		}
		jeu.ajouter(t)
	}
	if err := lignes.Err(); err != nil {
		return nil, fmt.Errorf("parcours des tables: %w", err)
	}
	return jeu, nil
}

// lireColonnes complète les tables collectées. C'est la passe la plus dense du
// pilote : type verbatim, type normalisé, longueur, précision, échelle,
// nullabilité, identité, défaut structuré, expression de génération, collation
// et commentaire y sont tous décidés.
func (p *pilote) lireColonnes(ctx context.Context, schemas []string, jeu *jeuDeTables) error {
	ctx, annuler := context.WithTimeout(ctx, delaiRequete)
	defer annuler()

	lignes, err := p.conn.Query(ctx, requeteColonnes, schemas)
	if err != nil {
		return fmt.Errorf("lecture des colonnes: %w", err)
	}
	defer lignes.Close()

	for lignes.Next() {
		var tableNom, tableSchema, typeInterne, identite, generee string
		var estEnumere bool
		var longueur, precision, echelle *int
		var defaut, commentaire, collation, collationSchema *string

		c := calque.Colonne{}
		if err := lignes.Scan(
			&tableNom, &tableSchema, &c.Nom, &c.Position, &c.TypeBrut, &typeInterne,
			&estEnumere, &longueur, &precision, &echelle, &c.Nullable,
			&identite, &defaut, &generee, &commentaire, &collation, &collationSchema,
		); err != nil {
			return fmt.Errorf("lecture d'une colonne: %w", err)
		}

		// attidentity vaut 'a' pour GENERATED ALWAYS, 'd' pour BY DEFAULT, et
		// vide hors identité. Une autre valeur arrête l'extraction : la deviner
		// écrirait un calque faux.
		switch identite {
		case "a":
			c.AutoIncrement, c.Identite = true, calque.IdentiteToujours
		case "d":
			c.AutoIncrement, c.Identite = true, calque.IdentiteParDefaut
		case "":
		default:
			return fmt.Errorf("colonne %s.%s.%s: attidentity %q inconnu", tableSchema, tableNom, c.Nom, identite)
		}

		c.TypeNormalise = normaliserType(typeInterne, estEnumere)
		c.Longueur, c.Precision, c.Echelle = longueur, precision, echelle
		// bpchar sans longueur déclarée n'est pas complété d'espaces : seul
		// character(n) est fixe. "char", sur un octet, n'a pas de longueur.
		c.LongueurFixe = typeInterne == "bpchar" && longueur != nil
		if estEnumere {
			c.TypeEnumere = typeInterne
		}
		if commentaire != nil {
			c.Commentaire = *commentaire
		}
		if collation != nil {
			c.Collation = *collation
		}
		// Seule une collation hors de pg_catalog porte son schéma : c'est la
		// règle du search_path vide, sous lequel tout le reste se résout sans
		// chemin, et les calques existants ne changent pas.
		if collationSchema != nil {
			c.CollationSchema = *collationSchema
		}
		if defaut != nil {
			c.Defaut = classerDefaut(*defaut)
		}
		// attgenerated vaut 's' pour une colonne stockée, 'v' pour une colonne
		// calculée à la lecture, et vide sinon. Comparé à la valeur attendue
		// plutôt qu'au vide : le type interne « char » ne se scanne pas en
		// chaîne vide, et tester la non-vacuité effacerait le défaut de toutes
		// les colonnes.
		if generee == "s" || generee == "v" {
			c.Generee = &calque.Generee{Stockee: generee == "s"}
			if defaut != nil {
				c.Generee.Expression = *defaut
			}
			c.Defaut = nil
		}

		table := jeu.trouver(tableSchema, tableNom)
		if table == nil {
			// Une colonne sans table est impossible sauf modification du schéma
			// pendant l'extraction. On l'ignore plutôt que de fabriquer une
			// table partielle.
			continue
		}
		table.Colonnes = append(table.Colonnes, c)
	}
	return lignes.Err()
}

// cleContrainte identifie une contrainte pendant la collecte de ses colonnes.
type cleContrainte struct {
	schema, table, nom string
}

// lireContraintes répartit clés primaires, unicités, clés étrangères et
// vérifications. Les colonnes sont chargées d'abord, en une requête : les
// demander contrainte par contrainte multiplierait les aller-retours sur une
// base qui en compte des milliers.
func (p *pilote) lireContraintes(ctx context.Context, schemas []string, jeu *jeuDeTables) error {
	colonnes, err := p.lireColonnesContraintes(ctx, schemas)
	if err != nil {
		return err
	}

	ctx, annuler := context.WithTimeout(ctx, delaiRequete)
	defer annuler()

	lignes, err := p.conn.Query(ctx, requeteContraintes, schemas)
	if err != nil {
		return fmt.Errorf("lecture des contraintes: %w", err)
	}
	defer lignes.Close()

	for lignes.Next() {
		var schema, table, nom, genre, definition string
		var tableCible, schemaCible *string
		var aLaSuppression, aLaMiseAJour string

		if err := lignes.Scan(&schema, &table, &nom, &genre, &definition,
			&tableCible, &schemaCible, &aLaSuppression, &aLaMiseAJour); err != nil {
			return fmt.Errorf("lecture d'une contrainte: %w", err)
		}

		cible := jeu.trouver(schema, table)
		if cible == nil {
			continue
		}
		cle := cleContrainte{schema, table, nom}

		switch genre {
		case "p":
			cible.ClePrimaire = &calque.ClePrimaire{
				Nom:      nom,
				Colonnes: colonnes[cle]["source"],
			}
		case "u":
			cible.Unicites = append(cible.Unicites, calque.Contrainte{
				Nom:      nom,
				Colonnes: colonnes[cle]["source"],
			})
		case "f":
			fk := calque.CleEtrangere{
				Nom:            nom,
				Colonnes:       colonnes[cle]["source"],
				ColonnesCibles: colonnes[cle]["cible"],
				ALaSuppression: actionReferentielle(aLaSuppression),
				ALaMiseAJour:   actionReferentielle(aLaMiseAJour),
			}
			if tableCible != nil {
				fk.TableCible = *tableCible
			}
			if schemaCible != nil {
				fk.SchemaCible = *schemaCible
			}
			cible.ClesEtrangeres = append(cible.ClesEtrangeres, fk)
		case "c":
			// Verbatim : la syntaxe d'un CHECK dépend du dialecte, et c'est la
			// source d'énumération la plus fiable pour l'inférence.
			cible.Verifications = append(cible.Verifications, calque.Verification{
				Nom:        nom,
				Expression: definition,
			})
		}
	}
	return lignes.Err()
}

// lireColonnesContraintes rend, par contrainte et par côté, les colonnes dans
// l'ordre déclaré. L'ordre est significatif : une clé composite (a, b) n'est
// pas (b, a).
func (p *pilote) lireColonnesContraintes(ctx context.Context, schemas []string) (map[cleContrainte]map[string][]string, error) {
	ctx, annuler := context.WithTimeout(ctx, delaiRequete)
	defer annuler()

	lignes, err := p.conn.Query(ctx, requeteColonnesContrainte, schemas)
	if err != nil {
		return nil, fmt.Errorf("lecture des colonnes de contraintes: %w", err)
	}
	defer lignes.Close()

	parContrainte := map[cleContrainte]map[string][]string{}
	for lignes.Next() {
		var schema, table, contrainte, genre, cote, colonne string
		var ordinalite int64

		if err := lignes.Scan(&schema, &table, &contrainte, &genre, &cote, &colonne, &ordinalite); err != nil {
			return nil, fmt.Errorf("lecture d'une colonne de contrainte: %w", err)
		}

		cle := cleContrainte{schema, table, contrainte}
		if parContrainte[cle] == nil {
			parContrainte[cle] = map[string][]string{}
		}
		parContrainte[cle][cote] = append(parContrainte[cle][cote], colonne)
	}
	return parContrainte, lignes.Err()
}

// cleIndex identifie un index pendant la collecte de ses colonnes. Le nom seul
// ne suffit pas : deux schémas peuvent porter le même.
type cleIndex struct {
	schema, table, nom string
}

// lireIndex collecte les index et leurs classes d'opérateurs. Un index
// d'expression est écarté : il n'a aucune colonne exploitable, et le garder
// produirait une contrainte vide que la validation refuserait.
func (p *pilote) lireIndex(ctx context.Context, schemas []string, jeu *jeuDeTables) error {
	colonnes, operateurs, ordres, err := p.lireColonnesIndex(ctx, schemas)
	if err != nil {
		return err
	}

	ctx, annuler := context.WithTimeout(ctx, delaiRequete)
	defer annuler()

	lignes, err := p.conn.Query(ctx, requeteIndex, schemas)
	if err != nil {
		return fmt.Errorf("lecture des index: %w", err)
	}
	defer lignes.Close()

	for lignes.Next() {
		var schema, table, nom, methode, definition string
		var predicat *string
		var unique bool

		if err := lignes.Scan(&schema, &table, &nom, &unique, &methode, &predicat, &definition); err != nil {
			return fmt.Errorf("lecture d'un index: %w", err)
		}

		cible := jeu.trouver(schema, table)
		if cible == nil {
			continue
		}

		cle := cleIndex{schema, table, nom}
		idx := calque.Index{
			Nom:      nom,
			Colonnes: colonnes[cle],
			Unique:   unique,
			Methode:  methode,
		}
		if predicat != nil {
			idx.Predicat = *predicat
		}
		idx.Operateurs = operateurs[cle]
		idx.Ordres = ordres[cle]
		// Un index d'expression n'a aucune colonne exploitable. Le garder sans
		// colonne produirait une contrainte vide, que la validation refuserait.
		if len(idx.Colonnes) == 0 {
			continue
		}
		cible.Index = append(cible.Index, idx)
	}
	return lignes.Err()
}

// lireColonnesIndex rend les colonnes de chaque index et, séparément, leurs
// classes d'opérateurs et leurs sens de tri.
//
// Les classes ne sont rendues que si l'index en porte au moins une explicite ;
// elles le sont alors toutes, y compris les implicites, pour rester appariées
// aux colonnes rang par rang. Les sens de tri suivent la même règle, sur la
// présence d'une colonne descendante.
func (p *pilote) lireColonnesIndex(ctx context.Context, schemas []string) (colonnes, classes map[cleIndex][]string, ordres map[cleIndex][]calque.OrdreIndex, err error) {
	ctx, annuler := context.WithTimeout(ctx, delaiRequete)
	defer annuler()

	lignes, err := p.conn.Query(ctx, requeteColonnesIndex, schemas)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("lecture des colonnes d'index: %w", err)
	}
	defer lignes.Close()

	colonnes = map[cleIndex][]string{}
	classes = map[cleIndex][]string{}
	ordres = map[cleIndex][]calque.OrdreIndex{}
	explicite := map[cleIndex]bool{}

	for lignes.Next() {
		var schema, table, index string
		var colonne, classe *string
		var parDefaut *bool
		var descendant bool
		var ordinalite int64

		if err := lignes.Scan(&schema, &table, &index, &colonne, &classe, &parDefaut, &descendant, &ordinalite); err != nil {
			return nil, nil, nil, fmt.Errorf("lecture d'une colonne d'index: %w", err)
		}
		if colonne == nil {
			continue
		}

		cle := cleIndex{schema, table, index}
		colonnes[cle] = append(colonnes[cle], *colonne)
		sens := calque.OrdreAscendant
		if descendant {
			sens = calque.OrdreDescendant
		}
		ordres[cle] = append(ordres[cle], sens)

		nom := ""
		if classe != nil {
			nom = *classe
		}
		classes[cle] = append(classes[cle], nom)
		if parDefaut != nil && !*parDefaut {
			explicite[cle] = true
		}
	}
	if err := lignes.Err(); err != nil {
		return nil, nil, nil, err
	}

	for cle := range classes {
		if !explicite[cle] {
			delete(classes, cle)
		}
	}
	for cle, sens := range ordres {
		if !slices.Contains(sens, calque.OrdreDescendant) {
			delete(ordres, cle)
		}
	}
	return colonnes, classes, ordres, nil
}

// lireSequences collecte les séquences du schéma, y compris celles qu'un SERIAL
// a créées implicitement : c'est ce qui permet à l'inférence de reconnaître une
// stratégie d'identifiant.
func (p *pilote) lireSequences(ctx context.Context, schemas []string) ([]calque.Sequence, error) {
	ctx, annuler := context.WithTimeout(ctx, delaiRequete)
	defer annuler()

	lignes, err := p.conn.Query(ctx, requeteSequences, schemas)
	if err != nil {
		return nil, fmt.Errorf("lecture des sequences: %w", err)
	}
	defer lignes.Close()

	var sequences []calque.Sequence
	for lignes.Next() {
		var s calque.Sequence
		var depart, minimum, maximum int64

		if err := lignes.Scan(&s.Schema, &s.Nom, &s.Increment, &depart, &minimum, &maximum, &s.Cyclique); err != nil {
			return nil, fmt.Errorf("lecture d'une sequence: %w", err)
		}
		s.Depart, s.Minimum, s.Maximum = &depart, &minimum, &maximum
		sequences = append(sequences, s)
	}
	return sequences, lignes.Err()
}

// lireTypesEnumeres collecte les types énumérés natifs. Ce sont les seules
// énumérations certaines : les autres se déduisent d'un CHECK ou d'un
// échantillon, et relèvent de l'inférence.
func (p *pilote) lireTypesEnumeres(ctx context.Context, schemas []string) ([]calque.TypeEnumere, error) {
	ctx, annuler := context.WithTimeout(ctx, delaiRequete)
	defer annuler()

	lignes, err := p.conn.Query(ctx, requeteTypesEnumeres, schemas)
	if err != nil {
		return nil, fmt.Errorf("lecture des types enumeres: %w", err)
	}
	defer lignes.Close()

	// Une ligne par valeur : on regroupe en conservant l'ordre de déclaration,
	// que la requête garantit par enumsortorder.
	var types []calque.TypeEnumere
	index := map[string]int{}

	for lignes.Next() {
		var schema, nom, valeur string
		if err := lignes.Scan(&schema, &nom, &valeur); err != nil {
			return nil, fmt.Errorf("lecture d'un type enumere: %w", err)
		}

		cle := schema + "." + nom
		rang, connu := index[cle]
		if !connu {
			types = append(types, calque.TypeEnumere{Nom: nom, Schema: schema})
			rang = len(types) - 1
			index[cle] = rang
		}
		types[rang].Valeurs = append(types[rang].Valeurs, valeur)
	}
	return types, lignes.Err()
}

// lireVues collecte les vues et vues matérialisées. La définition est reprise
// telle que le catalogue la rend, sans reformatage.
func (p *pilote) lireVues(ctx context.Context, schemas []string) ([]calque.Vue, error) {
	ctx, annuler := context.WithTimeout(ctx, delaiRequete)
	defer annuler()

	lignes, err := p.conn.Query(ctx, requeteVues, schemas)
	if err != nil {
		return nil, fmt.Errorf("lecture des vues: %w", err)
	}
	defer lignes.Close()

	var vues []calque.Vue
	for lignes.Next() {
		var v calque.Vue
		if err := lignes.Scan(&v.Schema, &v.Nom, &v.Definition, &v.Materialisee); err != nil {
			return nil, fmt.Errorf("lecture d'une vue: %w", err)
		}
		vues = append(vues, v)
	}
	return vues, lignes.Err()
}
