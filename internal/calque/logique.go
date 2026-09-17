// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package calque

// Logique est le modèle objet inféré. Contrairement au physique, il n'est pas
// neutre : il parle le vocabulaire de la famille Hibernate. Un ORM qui ne
// partage pas ce modèle consomme le physique.
type Logique struct {
	VersionRI         int    `json:"version_ri"`
	EmpreintePhysique string `json:"empreinte_physique"`
	// Sgbd recopie celui du physique. Le rendu d'une même entité dépend de la
	// plateforme DBAL, et pas seulement des versions d'ORM et de DBAL : une clé
	// par séquence ne se génère pas de la même façon sous SQL Server et sous
	// PostgreSQL. Absent d'un calque produit avant ce champ : inconnu.
	Sgbd           string          `json:"sgbd,omitempty"`
	EspaceDeNoms   string          `json:"espace_de_noms"`
	Entites        []Entite        `json:"entites"`
	Enumerations   []Enumeration   `json:"enumerations,omitempty"`
	Traits         []Trait         `json:"traits,omitempty"`
	Avertissements []Avertissement `json:"avertissements,omitempty"`
}

// Origine indique d'où vient une décision. Sans elle, l'outil n'est pas
// auditable et personne ne le lancera sur sa base.
type Origine string

// Origines, de la plus sûre à la plus discutable.
const (
	OrigineContrainte   Origine = "contrainte"
	OrigineVerification Origine = "verification"
	// Produite par rien en version 1 : elle attend les énumérations par
	// cardinalité de l'échantillonnage, et se retire au prochain incrément de
	// version_ri si elles ne l'ont pas produite d'ici là.
	OrigineCardinalite Origine = "cardinalite"
	OrigineNommage     Origine = "nommage"
	OrigineDecision    Origine = "decision"
)

// Entite est une classe à générer. Toutes les tables n'en produisent pas : une
// table de jointure pure devient une association.
//
// Origine porte celle du nom de classe, et rien d'autre. L'existence de
// l'entité est un constat — la table est là —, et l'héritage comme les
// associations portent chacun la leur.
type Entite struct {
	Nom      string         `json:"nom"`
	Table    ReferenceTable `json:"table"`
	Heritage *Heritage      `json:"heritage,omitempty"`
	// La valeur que la colonne discriminante prend pour cette classe, racine
	// comprise : c'est l'entité qu'elle identifie, quand la colonne appartient
	// à la relation d'héritage. Absente hors d'une hiérarchie décidée.
	ValeurDiscriminante string        `json:"valeur_discriminante,omitempty"`
	Traits              []string      `json:"traits,omitempty"`
	Identifiant         *Identifiant  `json:"identifiant,omitempty"`
	Proprietes          []Propriete   `json:"proprietes"`
	Associations        []Association `json:"associations,omitempty"`
	Index               []IndexEntite `json:"index,omitempty"`
	// Le commentaire de la table, reporté comme les index : la régénération du
	// schéma le recrée, et le générateur en fait la documentation de la classe.
	Commentaire string  `json:"commentaire,omitempty"`
	Origine     Origine `json:"origine,omitempty"`
}

// ReferenceTable qualifie la table d'origine, sous son nom de catalogue.
type ReferenceTable struct {
	Nom    string `json:"nom"`
	Schema string `json:"schema"`
}

// Heritage décrit une hiérarchie déclarée par décision. Le schéma la rend
// possible — une clé primaire qui est aussi une clé étrangère —, il ne
// l'impose pas : « un salarié est une personne » et « un salarié a une
// personne » sont deux modèles qu'il autorise également.
type Heritage struct {
	Strategie            StrategieHeritage `json:"strategie"`
	Parent               string            `json:"parent"`
	ColonneDiscriminante string            `json:"colonne_discriminante,omitempty"`
	Origine              Origine           `json:"origine,omitempty"`
}

// StrategieHeritage est la projection de la hiérarchie sur les tables.
type StrategieHeritage string

// Stratégies d'héritage. La table par classe concrète est absente
// volontairement : rien dans un schéma ne permet de la déduire.
const (
	HeritageJointe      StrategieHeritage = "jointe"
	HeritageTableUnique StrategieHeritage = "table_unique"
)

// Identifiant porte la ou les propriétés qui identifient l'entité. Plusieurs
// signifient une clé composite, cas courant sur du legacy.
type Identifiant struct {
	Proprietes []string             `json:"proprietes"`
	Strategie  StrategieIdentifiant `json:"strategie"`
	Sequence   string               `json:"sequence,omitempty"`
	// SequenceIncrement et SequenceMinimum recopient la séquence du physique
	// quand le nom lu dans le défaut la désigne sans ambiguïté, et restent
	// absents sinon. Ce sont des faits et non des réglages d'ORM : un incrément
	// de 10 réserve des blocs ou sépare plusieurs nœuds, et la base ne dit pas
	// lequel. Pointeurs, parce qu'un minimum vaut souvent 0.
	SequenceIncrement *int64 `json:"sequence_increment,omitempty"`
	SequenceMinimum   *int64 `json:"sequence_minimum,omitempty"`
	// SequenceDepart est la valeur de départ, que le minimum ne dit pas :
	// SQL Server part de 1 une séquence dont le minimum est celui du type.
	SequenceDepart *int64 `json:"sequence_depart,omitempty"`
}

// StrategieIdentifiant dit qui produit la valeur de la clé.
type StrategieIdentifiant string

// Stratégies d'identifiant. Une table sans clé primaire n'a pas d'Identifiant
// du tout, et un avertissement le signale.
const (
	IdentifiantIdentite StrategieIdentifiant = "identite"
	IdentifiantSequence StrategieIdentifiant = "sequence"
	// Produite par rien : déclarée en version 1, elle se retire au prochain
	// incrément de version_ri, et le générateur écarte d'ici là l'entité qui
	// la porte.
	IdentifiantAucune   StrategieIdentifiant = "aucune"
	IdentifiantAssignee StrategieIdentifiant = "assignee"
)

// Propriete est une colonne devenue attribut. Le type Doctrine apparaît ici, et
// pas dans le physique : il suppose la destination.
//
// Origine porte celle du type, pas celle du nom : la colonne existe, seule sa
// traduction en couple type PHP / type Doctrine est un jugement.
type Propriete struct {
	Nom     string `json:"nom"`
	Colonne string `json:"colonne"`
	// Obsolète, retiré à la prochaine version du format : le type PHP dépend
	// de la version de Doctrine du projet, qu'un calque ne connaît pas. Un
	// générateur ne doit le lire que pour un type Doctrine hors de sa table.
	TypePHP      string `json:"type_php"`
	TypeDoctrine string `json:"type_doctrine"`
	Nullable     bool   `json:"nullable"`
	Longueur     *int   `json:"longueur,omitempty"`
	Precision    *int   `json:"precision,omitempty"`
	Echelle      *int   `json:"echelle,omitempty"`
	// Une chaîne de longueur fixe complète ses valeurs d'espaces, un binaire
	// d'octets nuls, et se compare autrement qu'une colonne variable de même
	// longueur : recréée sans ce fait, la colonne change de comportement. Un
	// fait de colonne, pas un type.
	LongueurFixe bool `json:"longueur_fixe,omitempty"`
	// Collation explicite de la colonne, sous son nom de catalogue. Absente
	// pour la collation par défaut de la base, et pour une collation hors du
	// schéma système, que l'avertissement collation_non_reportee signale.
	Collation   string `json:"collation,omitempty"`
	Enumeration string `json:"enumeration,omitempty"`
	// DEFAULT '' est un défaut : absent et vide doivent rester distinguables,
	// comme pour la longueur.
	Defaut *string `json:"defaut,omitempty"`
	// Le sens d'un défaut calculé, jamais son texte : now() et
	// CURRENT_TIMESTAMP disent la même chose, et c'est ce qu'un générateur
	// doit traduire. Exclusif de Defaut.
	DefautExpression ExpressionDefaut `json:"defaut_expression,omitempty"`
	// Generee reprend du physique l'expression d'une colonne calculée par la
	// base, verbatim. Insertable et Modifiable disent qu'on n'écrit pas la
	// colonne, pas pourquoi : une colonne générée est en plus à relire après
	// chaque écriture, et un générateur ne doit pas le déduire de leur seule
	// absence.
	Generee *Generee `json:"generee,omitempty"`
	// Une colonne générée n'est ni insérable ni modifiable. Les pointeurs
	// permettent de ne sérialiser que les cas qui s'écartent du défaut.
	Insertable  *bool   `json:"insertable,omitempty"`
	Modifiable  *bool   `json:"modifiable,omitempty"`
	Unique      bool    `json:"unique,omitempty"`
	Commentaire string  `json:"commentaire,omitempty"`
	Origine     Origine `json:"origine,omitempty"`
}

// ExpressionDefaut est un vocabulaire fermé : le sens d'un défaut calculé que
// l'inférence a reconnu. Toute valeur ajoutée incrémente VersionRI.
type ExpressionDefaut string

// Défauts calculés reconnus. L'instant est celui que la base donne à
// CURRENT_TIMESTAMP, la seule expression que Doctrine écrive : le début de la
// transaction sous PostgreSQL, l'instruction sous SQL Server. Une horloge plus
// précise, clock_timestamp() ou sysdatetime(), n'en fait pas partie.
const (
	DefautHorodatageCourant ExpressionDefaut = "horodatage_courant"
	DefautDateCourante      ExpressionDefaut = "date_courante"
	DefautHeureCourante     ExpressionDefaut = "heure_courante"
)

// Association relie deux entités. Proprietaire décide du côté qui porte la
// colonne de jointure : s'y tromper produit un mapping que Doctrine accepte et
// qui n'écrit rien en base.
//
// OrphelinsSupprimes n'est produit par aucune heuristique ni décision : déclaré
// en version 1, il se retire au prochain incrément de version_ri, et le
// générateur écarte d'ici là l'entité qui le porte.
type Association struct {
	Nom                string            `json:"nom"`
	Genre              GenreAssociation  `json:"genre"`
	Cible              string            `json:"cible"`
	Proprietaire       bool              `json:"proprietaire"`
	InverseePar        string            `json:"inversee_par,omitempty"`
	MappeePar          string            `json:"mappee_par,omitempty"`
	Jointure           []ColonneJointure `json:"jointure,omitempty"`
	TableJointure      *TableJointure    `json:"table_jointure,omitempty"`
	OrphelinsSupprimes bool              `json:"orphelins_supprimes,omitempty"`
	Origine            Origine           `json:"origine"`
}

// GenreAssociation est la cardinalité de la relation.
type GenreAssociation string

// Cardinalités. Un-vers-un se distingue de plusieurs-vers-un par une contrainte
// d'unicité sur la colonne portante, pas par le nommage.
const (
	UnVersUn               GenreAssociation = "un_vers_un"
	PlusieursVersUn        GenreAssociation = "plusieurs_vers_un"
	UnVersPlusieurs        GenreAssociation = "un_vers_plusieurs"
	PlusieursVersPlusieurs GenreAssociation = "plusieurs_vers_plusieurs"
)

// ColonneJointure apparie colonne portante et colonne référencée. La
// nullabilité, reprise du physique, décide si l'association est facultative.
type ColonneJointure struct {
	Colonne           string `json:"colonne"`
	ColonneReferencee string `json:"colonne_referencee"`
	Nullable          bool   `json:"nullable"`
	ALaSuppression    Action `json:"a_la_suppression,omitempty"`
}

// TableJointure décrit la table d'association d'un plusieurs-vers-plusieurs.
// Elle ne devient pas une entité : son commentaire n'a pas d'autre place.
type TableJointure struct {
	Nom             string            `json:"nom"`
	Schema          string            `json:"schema"`
	Jointure        []ColonneJointure `json:"jointure"`
	JointureInverse []ColonneJointure `json:"jointure_inverse"`
	Commentaire     string            `json:"commentaire,omitempty"`
}

// IndexEntite reporte un index du physique. La méthode et la classe
// d'opérateurs n'y survivent pas : Doctrine ne sait pas les exprimer.
type IndexEntite struct {
	Nom      string   `json:"nom,omitempty"`
	Colonnes []string `json:"colonnes"`
	Unique   bool     `json:"unique"`
	// Predicat est la condition d'un index partiel, verbatim et non
	// interprétée. Sans elle, l'index serait recréé complet : une unicité
	// partielle deviendrait plus stricte que la base, et refuserait des lignes
	// qu'elle accepte.
	Predicat string `json:"predicat,omitempty"`
}

// Enumeration est un type PHP à générer. Origine dit d'où elle sort — CHECK,
// type natif ou échantillon — et c'est ce qui permet d'en discuter.
//
// TypeSupport vaut toujours "string" en version 1 : la détection ne reconnaît
// que des littéraux chaîne. "int" est lu et rendu par le générateur, sans
// producteur avant l'échantillonnage.
type Enumeration struct {
	Nom         string           `json:"nom"`
	TypeSupport string           `json:"type_support"`
	Cas         []CasEnumeration `json:"cas"`
	Origine     Origine          `json:"origine"`
}

// CasEnumeration apparie un nom de cas PHP à la valeur stockée. Valeur est un
// any : le type support peut être une chaîne comme un entier.
type CasEnumeration struct {
	Nom    string `json:"nom"`
	Valeur any    `json:"valeur"`
}

// Trait regroupe les propriétés récurrentes — created_at, updated_at,
// deleted_at — plutôt que de les recopier dans chaque entité.
type Trait struct {
	Nom        string      `json:"nom"`
	Proprietes []Propriete `json:"proprietes"`
}

// Avertissement est une sortie de premier ordre, pas un journal : ce qui n'est
// pas résolu est signalé ici, jamais inventé ailleurs.
type Avertissement struct {
	Code       string     `json:"code"`
	Cible      string     `json:"cible"`
	Message    string     `json:"message"`
	Resolution Resolution `json:"resolution"`
	Confiance  float64    `json:"confiance"`
}

// Resolution dit ce que l'inférence a fait de l'incertitude signalée.
type Resolution string

// Résolutions. ResolutionParDefaut annonce un choix appliqué faute de mieux,
// ResolutionAucune un trou laissé ouvert.
const (
	ResolutionIgnoree           Resolution = "ignoree"
	ResolutionAucune            Resolution = "aucune"
	ResolutionParDefaut         Resolution = "par_defaut"
	ResolutionForceeParDecision Resolution = "forcee_par_decision"
)

// Codes d'avertissement. Stables entre versions : ils servent de filtre en CI.
const (
	CodeTableSansClePrimaire = "table_sans_cle_primaire"
	CodeClePrimaireComposite = "cle_primaire_composite"
	CodeTypeNonReconnu       = "type_non_reconnu"
	CodeCollision            = "collision_de_nom"
	CodeDecisionOrpheline    = "decision_sans_cible"
	// Une décision dont le nom serait refusé par PHP là où la génération
	// l'écrit : elle est ignorée, et le nom inféré reste.
	CodeDecisionInvalide     = "decision_invalide"
	CodeTableIgnoree         = "table_ignoree"
	CodeColonneIgnoree       = "colonne_ignoree"
	CodeClePrimaireGardee    = "cle_primaire_non_ignorable"
	CodeDefautIncompatible   = "defaut_incompatible"
	CodePrefixeDetecte       = "prefixe_detecte"
	CodeCibleHorsPortee      = "cible_hors_portee"
	CodeCasEnumerationOpaque = "cas_enumeration_opaque"
	CodeTraitDeduit          = "trait_deduit"
	CodeJointurePure         = "table_de_jointure"
	// Une clé tirée d'une séquence dont le défaut ne nomme pas la séquence :
	// l'identifiant reste à fournir par l'application.
	CodeSequenceNonReconnue = "sequence_non_reconnue"
	// Un texte Unicode illimité de SQL Server, que Doctrine ne recrée qu'en
	// VARCHAR(MAX) : rendu en chaîne plutôt qu'en text, qui perdrait l'Unicode
	// en silence.
	CodeTexteUnicodeSansEquivalent = "texte_unicode_sans_equivalent"
	// Un texte Unicode illimité rendu en json par décision : Doctrine le
	// recrée en VARCHAR(MAX), juste pour ce qu'il écrit, pas pour ce qu'une
	// autre application y écrit en clair.
	CodeJSONSansUnicode = "json_sans_unicode"
	// Un horodatage avec fuseau à plus de six décimales : Doctrine n'en lit
	// que six, et la lecture échoue tant que la colonne n'est pas réduite.
	CodeFuseauPrecisionNonLue = "fuseau_precision_non_lue"
	// Une clé étrangère qui désigne autre chose que la clé primaire de sa
	// cible : Doctrine n'associe que vers l'identifiant, la colonne reste une
	// propriété.
	CodeReferenceHorsIdentifiant = "reference_hors_identifiant"
	// Code conservé, sens inversé : il signalait un héritage appliqué, il
	// signale désormais un héritage possible et non appliqué — l'entité est
	// reliée à son parent par un-vers-un tant qu'aucune décision ne le déclare.
	CodeHeritageDeduit = "heritage_deduit"
	// Un défaut calculé dont le sens n'est pas reconnu : l'entité est générée
	// sans lui, et l'application doit fournir la valeur.
	CodeDefautNonReporte = "defaut_non_reporte"
	// Une collation hors du schéma système : un générateur ne sait pas
	// toujours l'écrire qualifiée, et son nom seul dépendrait du chemin de
	// recherche de l'application. La propriété est rendue sans collation.
	CodeCollationNonReportee = "collation_non_reportee"
)
