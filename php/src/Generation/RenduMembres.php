<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

use Ormeau\Doctrine\Calque\ActionSuppression;
use Ormeau\Doctrine\Calque\Association;
use Ormeau\Doctrine\Calque\ColonneJointure;
use Ormeau\Doctrine\Calque\Enumeration;
use Ormeau\Doctrine\Calque\ExpressionDefaut;
use Ormeau\Doctrine\Calque\GenreAssociation;
use Ormeau\Doctrine\Calque\Identifiant;
use Ormeau\Doctrine\Calque\Propriete;
use Ormeau\Doctrine\Calque\StrategieIdentifiant;
use Ormeau\Doctrine\Calque\TypeSupport;

/**
 * Écrit les propriétés mappées, les associations et leurs accesseurs, pour une
 * classe de base comme pour un trait.
 *
 * Les deux reçoivent exactement le même rendu : une colonne created_at doit
 * s'écrire de la même façon qu'elle vive dans ClientBase ou dans le trait
 * Horodatage, sans quoi sortir des colonnes dans un trait changerait leur
 * mapping.
 *
 * Une colonne de clé étrangère figure deux fois dans le calque : comme
 * propriété et comme colonne de jointure d'une association propriétaire. Les
 * écrire telles quelles donnerait deux écrivains à la même colonne, et
 * Doctrine laisse l'association écraser la propriété sans rien signaler —
 * constaté sous ORM 2.14 et 3.6. Deux traductions, selon la place de la
 * colonne :
 *
 *  - hors clé primaire, la propriété reste, en lecture seule (insertable et
 *    updatable à false) : l'association écrit, la propriété relit la clé sans
 *    charger l'entité liée ;
 *  - dans la clé primaire, #[ORM\Id] passe sur l'association et la propriété
 *    n'est pas écrite : c'est l'identité dérivée de Doctrine, et en garder les
 *    deux obligerait à renseigner deux fois la même valeur.
 *
 * Le rapprochement entre propriété et association se fait par le nom de
 * colonne, jamais par le nom de propriété : commandeId et commande ne se
 * ressemblent que par hasard, commande_id est le même des deux côtés par
 * construction.
 *
 * @internal
 */
final class RenduMembres
{
    /** Action à la suppression, telle que l'attend onDelete. */
    private const ON_DELETE = [
        'cascade' => 'CASCADE',
        'set_null' => 'SET NULL',
        'set_default' => 'SET DEFAULT',
        'restrict' => 'RESTRICT',
    ];

    /**
     * @param Cible                      $cible        détermine le type PHP de chaque propriété
     * @param string                     $espaceDeNoms espace de noms des entités ; les énumérations
     *                                                 vivent dans \Enum sous lui
     * @param array<string, Enumeration> $enumerations énumérations du calque par nom, pour typer une
     *                                                 propriété et retrouver le cas de son défaut
     * @param string                     $sgbd         SGBD du calque, qui décide de la valeur initiale d'une
     *                                                 séquence, de la forme d'un défaut calculé et du schéma
     *                                                 écrit d'une table de jointure
     * @param ClassesUtilisateur|null    $classes      où vivent les classes de l'utilisateur, que les
     *                                                 associations importent ; sans elle, à la racine
     */
    public function __construct(
        private readonly Cible $cible,
        private readonly string $espaceDeNoms,
        private readonly array $enumerations,
        private readonly string $sgbd,
        private readonly ?ClassesUtilisateur $classes = null,
    ) {}

    /**
     * Rend les propriétés, les associations et leurs accesseurs, avec les
     * imports qu'ils demandent et les collections à initialiser.
     *
     * Les membres de l'identifiant d'abord, dans l'ordre de la clé primaire —
     * propriété ou association de clé —, puis les autres propriétés, puis les
     * autres associations ; les accesseurs suivent le même ordre. Doctrine
     * forme la clé primaire dans l'ordre de déclaration des membres #[ORM\Id] :
     * une association de clé rendue après les propriétés inverse la clé, et
     * migrations:diff propose alors de la supprimer puis de la recréer.
     * L'ordre de la clé l'emporte sur la position des colonnes, que le
     * comparateur de DBAL ignore.
     *
     * @param list<Propriete>   $proprietes   propriétés à rendre, dans l'ordre du calque
     * @param Identifiant|null  $identifiant  clé de l'entité ; null pour un trait, qui n'en porte pas
     * @param list<Association> $associations associations de l'entité ; aucune pour un trait
     * @param string|null       $generateur   classe qui tire la séquence de la clé, relative à
     *                                        l'espace de noms de la classe de base, quand la clé
     *                                        en exige une ; voir RenduEntite::tireParGenerateur
     *
     * @return array{imports: list<string>, membres: list<string>, accesseurs: list<string>, collections: list<string>}
     */
    public function rendre(array $proprietes, ?Identifiant $identifiant, array $associations = [], ?string $generateur = null): array
    {
        $jointures = self::colonnesDeJointure($associations);
        $clesDerivees = array_flip(self::associationsDeCle($proprietes, $identifiant, $associations));

        // Chaque membre, rangé par sa clé : « p:nom » pour une propriété, « a:nom »
        // pour une association. La propriété d'une colonne qu'une association de
        // clé écrit n'a pas de membre ; son commentaire et son défaut passent sur
        // la colonne de jointure.
        $rendus = [];
        $retirees = [];
        $ordre = [];
        foreach ($proprietes as $propriete) {
            $ecritePar = $jointures[$propriete->colonne] ?? null;
            if ($ecritePar !== null) {
                if (self::dansLaCle($propriete, $identifiant)) {
                    $retirees[$propriete->colonne] = $propriete;
                    continue;
                }
                $propriete = self::enLectureSeule($propriete);
            }

            [$type, $import] = $this->type($propriete, $identifiant);
            $rendus['p:' . $propriete->nom] = [
                'imports' => $import === null ? [] : [$import],
                'membre' => $this->propriete($propriete, $identifiant, $type, $ecritePar, $generateur),
                'accesseurs' => $this->accesseurs($propriete, $identifiant, $type),
            ];
            $ordre[] = 'p:' . $propriete->nom;
        }

        $collections = [];
        foreach ($associations as $association) {
            $imports = [$this->classes?->qualifiee($association->cible) ?? $this->espaceDeNoms . '\\' . $association->cible];
            if (self::estCollection($association)) {
                $imports[] = 'Doctrine\Common\Collections\ArrayCollection';
                $imports[] = 'Doctrine\Common\Collections\Collection';
                $collections[] = $association->nom;
            }
            $rendus['a:' . $association->nom] = [
                'imports' => $imports,
                'membre' => $this->association($association, isset($clesDerivees[$association->nom]), $retirees),
                'accesseurs' => $this->accesseursAssociation($association),
            ];
            $ordre[] = 'a:' . $association->nom;
        }

        $cle = [];
        $parColonne = [];
        foreach ($proprietes as $propriete) {
            $parColonne[$propriete->nom] = $propriete->colonne;
        }
        foreach ($identifiant->proprietes ?? [] as $nom) {
            $colonne = $parColonne[$nom] ?? null;
            $membre = isset($rendus['p:' . $nom]) ? 'p:' . $nom : ($colonne !== null && isset($jointures[$colonne]) ? 'a:' . $jointures[$colonne] : null);
            if ($membre !== null && !in_array($membre, $cle, true)) {
                $cle[] = $membre;
            }
        }

        $resultat = ['imports' => [], 'membres' => [], 'accesseurs' => [], 'collections' => $collections];
        foreach ([...$cle, ...array_values(array_diff($ordre, $cle))] as $membre) {
            array_push($resultat['imports'], ...$rendus[$membre]['imports']);
            $resultat['membres'][] = $rendus[$membre]['membre'];
            $resultat['accesseurs'][] = $rendus[$membre]['accesseurs'];
        }

        return $resultat;
    }

    /**
     * Rend le nom qualifié de la classe PHP d'une énumération du calque.
     */
    public function enumerationQualifiee(string $nom): string
    {
        return $this->espaceDeNoms . '\\Enum\\' . $nom;
    }

    /**
     * Rend les noms des associations qui portent l'identifiant d'une entité :
     * celles dont une colonne de jointure est la colonne d'une propriété de la
     * clé primaire. C'est l'identité dérivée, et la seule définition qu'en ont
     * le rendu, qui y pose #[ORM\Id], et le générateur, qui écarte une chaîne
     * de telles identités.
     *
     * @param list<Propriete>   $proprietes   propriétés de l'entité
     * @param Identifiant|null  $identifiant  clé de l'entité
     * @param list<Association> $associations associations de l'entité
     *
     * @return list<string>
     */
    public static function associationsDeCle(array $proprietes, ?Identifiant $identifiant, array $associations): array
    {
        $jointures = self::colonnesDeJointure($associations);
        $noms = [];
        foreach ($proprietes as $propriete) {
            if (isset($jointures[$propriete->colonne]) && self::dansLaCle($propriete, $identifiant)) {
                $noms[$jointures[$propriete->colonne]] = true;
            }
        }

        return array_keys($noms);
    }

    /**
     * Dit si une propriété fait partie de la clé primaire.
     */
    private static function dansLaCle(Propriete $propriete, ?Identifiant $identifiant): bool
    {
        return $identifiant !== null && in_array($propriete->nom, $identifiant->proprietes, true);
    }

    /**
     * Rend, pour chaque colonne portée par une association propriétaire à
     * colonnes de jointure, le nom de cette association.
     *
     * Un plusieurs-vers-plusieurs n'en porte aucune : ses colonnes sont dans
     * la table de jointure, pas dans celle de l'entité.
     *
     * @param list<Association> $associations associations de l'entité, propriétaires ou non
     *
     * @return array<string, string>
     */
    private static function colonnesDeJointure(array $associations): array
    {
        $colonnes = [];
        foreach ($associations as $association) {
            if (!$association->proprietaire || self::estCollection($association)) {
                continue;
            }
            foreach ($association->jointure as $jointure) {
                $colonnes[$jointure->colonne] = $association->nom;
            }
        }

        return $colonnes;
    }

    /**
     * Rend la même propriété, ni insérable ni modifiable : sa colonne est
     * écrite par une association.
     *
     * Son unicité tombe aussi. Une colonne de jointure unique est ce qui fait
     * un un-vers-un, et Doctrine crée lui-même l'index unique de sa jointure :
     * le redire sur le champ produit deux index de même nom, et un DDL que
     * DBAL refuse de construire — le validateur de mapping, lui, ne dit rien.
     *
     * Le #[ORM\UniqueConstraint] issu de l'index de soutien de la contrainte,
     * lui, reste. DBAL le substitue à l'index qu'il aurait créé pour la
     * jointure : un seul index, qui garde le nom de la base. Sans lui, l'index
     * s'appellerait UNIQ_… et migrations:diff proposerait de renommer celui de
     * la base d'origine.
     */
    private static function enLectureSeule(Propriete $p): Propriete
    {
        return new Propriete(
            $p->nom,
            $p->colonne,
            $p->typeDoctrine,
            $p->nullable,
            $p->longueur,
            $p->precision,
            $p->echelle,
            $p->enumeration,
            $p->defaut,
            false,
            false,
            false,
            $p->commentaire,
            $p->origine,
            $p->defautExpression,
            $p->longueurFixe,
            $p->generee,
            $p->collation,
        );
    }

    /**
     * Dit si une association porte une collection plutôt qu'un objet, et
     * demande donc une initialisation dans le constructeur.
     */
    public static function estCollection(Association $association): bool
    {
        return in_array($association->genre, [GenreAssociation::UnVersPlusieurs, GenreAssociation::PlusieursVersPlusieurs], true);
    }

    /**
     * Dit si la valeur d'une propriété est produite par la base à l'insertion :
     * clé IDENTITY ou séquence, sur une clé à une seule colonne.
     */
    private function estGeneree(Propriete $propriete, ?Identifiant $identifiant): bool
    {
        return $identifiant !== null
            && $identifiant->proprietes === [$propriete->nom]
            && in_array($identifiant->strategie, [StrategieIdentifiant::Identite, StrategieIdentifiant::Sequence], true);
    }

    /**
     * Rend le type déclaré d'une propriété, et la classe à importer pour
     * l'écrire par son nom court.
     *
     * Une propriété énumérée prend le type de l'énumération. Une clé produite
     * par la base est nullable en PHP même quand la colonne ne l'est pas : elle
     * vaut null tant que l'entité n'est pas persistée. Une classe d'un espace
     * de noms s'importe, une classe globale reste qualifiée.
     *
     * @return array{string, string|null}
     */
    private function type(Propriete $propriete, ?Identifiant $identifiant): array
    {
        $nu = $propriete->enumeration !== null
            ? '\\' . $this->enumerationQualifiee($propriete->enumeration)
            : TypesPhp::nu($propriete, $this->cible);

        $nullable = $nu !== 'mixed' && ($propriete->nullable || $this->estGeneree($propriete, $identifiant));

        // \DateTimeImmutable et non un import : c'est aussi courant à la main,
        // mais PHP-CS-Fixer en règles @Symfony requalifie une classe globale
        // importée, et chaque régénération la réimporterait — un diff à chaque
        // passage sur des fichiers que personne n'a touchés.
        $import = null;
        if (str_contains(ltrim($nu, '\\'), '\\')) {
            $import = ltrim($nu, '\\');
            $nu = self::nomCourt($import);
        }

        return [($nullable ? '?' : '') . $nu, $import];
    }

    /**
     * Rend une propriété mappée, docblock et attributs compris.
     *
     * Le docblock porte le commentaire de la colonne, ce qui explique une
     * absence — pas de mutateur, parce qu'une association écrit la colonne —
     * et le type des éléments d'un tableau, que l'analyse statique du projet
     * exige dès le niveau 6. Il est sur la propriété et nulle part ailleurs :
     * les méthodes métier de la classe de l'utilisateur y accèdent
     * directement, et le répéter sur chaque accesseur serait de la cérémonie.
     *
     * @param Propriete        $propriete   propriété à rendre
     * @param Identifiant|null $identifiant clé de l'entité
     * @param string           $type        type PHP déclaré
     * @param string|null      $ecritePar   association qui écrit la colonne, quand il y en a une
     * @param string|null      $generateur  classe qui tire la séquence de la clé, quand il en faut une
     */
    private function propriete(Propriete $propriete, ?Identifiant $identifiant, string $type, ?string $ecritePar = null, ?string $generateur = null): string
    {
        $indentation = Emetteur::INDENTATION;
        $texte = $propriete->commentaire === null ? [] : Emetteur::commentaire($propriete->commentaire);
        if ($ecritePar !== null) {
            $note = sprintf('Lecture seule : écrite par l\'association %s.', $ecritePar);
            $texte = $texte === [] ? [$note] : [...$texte, '', $note];
        }
        if ($propriete->generee !== null) {
            // L'expression vient de la base comme un commentaire, et se
            // neutralise de la même façon.
            $note = Emetteur::commentaire('Calculée par la base : ' . $propriete->generee->expression);
            $texte = $texte === [] ? $note : [...$texte, '', ...$note];
        }
        $lignes = Emetteur::docblock($texte, self::estTableau($type) ? ['@var ' . self::typeTableau($type)] : [], $indentation);
        $generee = $this->estGeneree($propriete, $identifiant);

        if ($identifiant !== null && in_array($propriete->nom, $identifiant->proprietes, true)) {
            $lignes[] = Emetteur::attribut('ORM\Id', [], $indentation);
            if ($generee && $generateur !== null) {
                // ORM 3 sous SQL Server : ni IDENTITY, qui lit
                // SCOPE_IDENTITY() et ne voit pas la séquence, ni
                // #[SequenceGenerator], ignoré sur une classe de base mappée.
                // Un générateur CUSTOM, lui, s'hérite. Constaté sous ORM 3.7,
                // contre SQL Server 2022 (essai du 2026-09-16).
                $lignes[] = Emetteur::attribut('ORM\GeneratedValue', ['strategy' => 'CUSTOM'], $indentation);
                $lignes[] = Emetteur::attribut('ORM\CustomIdGenerator', ['class' => new Code($generateur . '::class')], $indentation);
            } elseif ($generee) {
                // La clé est dans une classe de base mappée. ORM 3 y ignore
                // #[SequenceGenerator] et prend <table>_<colonne>_seq, mais son
                // IDENTITY lit LASTVAL(), juste quel que soit le nom de la
                // séquence. ORM 2 respecte le générateur, et son IDENTITY
                // interroge currval sur le nom par défaut — sous SQL Server,
                // il hydrate 0 sans erreur. Constaté sous ORM 2.14 et 3.6,
                // contre PostgreSQL 17, et sous 2.14 contre SQL Server 2022.
                $sequence = $identifiant->strategie === StrategieIdentifiant::Sequence && $this->cible->ormMajeure === 2;
                $lignes[] = Emetteur::attribut('ORM\GeneratedValue', ['strategy' => $sequence ? 'SEQUENCE' : 'IDENTITY'], $indentation);
                if ($sequence && $identifiant->sequence !== null) {
                    // Doctrine ne se sert de la valeur initiale que pour créer
                    // le schéma, jamais pour attribuer un identifiant, mais
                    // schema:update propose un ALTER quand elle diffère de ce
                    // que DBAL 3 relit : le minimum sous PostgreSQL, le départ
                    // sous SQL Server. allocationSize reste à 1, voir
                    // SequenceNonAlignee.
                    $initiale = $this->sgbd === 'sqlserver' ? $identifiant->sequenceDepart : $identifiant->sequenceMinimum;
                    $arguments = ['sequenceName' => $identifiant->sequence];
                    if ($initiale !== null && $initiale !== 1) {
                        $arguments['initialValue'] = $initiale;
                    }
                    $lignes[] = Emetteur::attribut('ORM\SequenceGenerator', $arguments, $indentation);
                }
            }
        }

        $defaut = $this->defaut($propriete);
        $options = $this->optionsDeColonne($propriete);

        $enumType = null;
        if ($propriete->enumeration !== null) {
            $enumType = new Code(self::nomCourt($this->enumerationQualifiee($propriete->enumeration)) . '::class');
        }

        $lignes[] = Emetteur::attribut('ORM\Column', [
            'name' => IdentifiantsSql::colonne($propriete->colonne),
            'type' => $this->typeDeColonne($propriete),
            'length' => $propriete->longueur,
            'precision' => $propriete->precision,
            'scale' => $propriete->echelle,
            'unique' => $propriete->unique ? true : null,
            'nullable' => $propriete->nullable ? true : null,
            'insertable' => $propriete->insertable ? null : false,
            'updatable' => $propriete->modifiable ? null : false,
            // Relue par un SELECT après chaque INSERT et UPDATE, sans quoi elle
            // vaut null en mémoire jusqu'au rechargement ; aucun effet sur le DDL,
            // que DBAL ne sait pas écrire pour une colonne générée.
            // Contrepartie de Doctrine : le
            // flush qui suit une écriture croit la colonne modifiée, n'émet
            // rien, mais appelle #[PreUpdate] une fois.
            'generated' => $propriete->generee === null ? null : 'ALWAYS',
            'enumType' => $enumType,
            'options' => $options === [] ? null : $options,
        ], $indentation);

        $cas = $this->casDuDefaut($propriete);
        $initialisation = match (true) {
            $generee => ' = null',
            $cas !== null => ' = ' . $cas->php,
            $propriete->enumeration === null && $defaut !== null && $type !== 'mixed' => ' = ' . Emetteur::litteral($defaut),
            $propriete->nullable => ' = null',
            default => '',
        };
        $lignes[] = $indentation . 'protected ' . $type . ' $' . $propriete->nom . $initialisation . ';';

        return implode("\n", $lignes);
    }

    /**
     * Rend une association, attributs de jointure compris.
     *
     * Un objet facultatif — côté inverse, ou jointure dont une colonne accepte
     * null — est nullable et vaut null au départ ; un objet obligatoire reste
     * non initialisé, comme une colonne obligatoire. Une collection porte son
     * type d'éléments en docblock : sans lui, l'analyse statique du projet ne
     * sait pas ce qu'elle contient. Une association de la clé primaire le dit :
     * Doctrine tire l'identifiant de l'entité visée, et un persist() qui la
     * précède échoue.
     *
     * Une colonne de jointure de la clé n'a pas de propriété : son commentaire
     * et son défaut vont en options de la JoinColumn, sans quoi Doctrine la
     * recrée avec les options de la colonne visée.
     *
     * @param Association              $association association à rendre
     * @param bool                     $dansLaCle   l'association porte l'identifiant
     * @param array<string, Propriete> $retirees    propriétés sans membre, par colonne : celles qu'une
     *                                              association de clé écrit
     */
    private function association(Association $association, bool $dansLaCle, array $retirees = []): string
    {
        $i = Emetteur::INDENTATION;
        $cible = new Code($association->cible . '::class');
        $lignes = Emetteur::docblock(
            $dansLaCle ? ['Fait partie de l\'identifiant : à renseigner avant persist().'] : [],
            self::estCollection($association) ? ['@var Collection<int, ' . $association->cible . '>'] : [],
            $i,
        );

        if ($dansLaCle) {
            $lignes[] = Emetteur::attribut('ORM\Id', [], $i);
        }

        $attribut = match ($association->genre) {
            GenreAssociation::UnVersUn => 'ORM\OneToOne',
            GenreAssociation::PlusieursVersUn => 'ORM\ManyToOne',
            GenreAssociation::UnVersPlusieurs => 'ORM\OneToMany',
            GenreAssociation::PlusieursVersPlusieurs => 'ORM\ManyToMany',
        };
        $lignes[] = Emetteur::attribut($attribut, [
            'targetEntity' => $cible,
            'mappedBy' => $association->proprietaire ? null : $association->mappeePar,
            'inversedBy' => $association->proprietaire ? $association->inverseePar : null,
        ], $i);

        if ($association->proprietaire && $association->tableJointure !== null) {
            $table = $association->tableJointure;
            // La table de jointure n'a pas d'entité : son commentaire va sur
            // #[JoinTable], accepté par ORM 2.14 comme par ORM 3.
            $argumentsTable = IdentifiantsSql::table($table->nom, SchemaParDefaut::ecrit($this->sgbd, $table->schema));
            $argumentsTable['options'] = $table->commentaire === null ? null : ['comment' => $table->commentaire];
            $lignes[] = Emetteur::attribut('ORM\JoinTable', $argumentsTable, $i);
            foreach ($table->jointure as $jointure) {
                $lignes[] = Emetteur::attribut('ORM\JoinColumn', self::argumentsJointure($jointure, false), $i);
            }
            foreach ($table->jointureInverse as $jointure) {
                $lignes[] = Emetteur::attribut('ORM\InverseJoinColumn', self::argumentsJointure($jointure, false), $i);
            }
        } elseif ($association->proprietaire) {
            foreach ($association->jointure as $jointure) {
                $arguments = self::argumentsJointure($jointure, !$dansLaCle);
                $retiree = $dansLaCle ? ($retirees[$jointure->colonne] ?? null) : null;
                if ($retiree !== null) {
                    $options = $this->optionsDeColonne($retiree);
                    $arguments['options'] = $options === [] ? null : $options;
                }
                $lignes[] = Emetteur::attribut('ORM\JoinColumn', $arguments, $i);
            }
        }

        $type = self::typeAssociation($association);
        $initialisation = str_starts_with($type, '?') ? ' = null' : '';
        $lignes[] = $i . 'protected ' . $type . ' $' . $association->nom . $initialisation . ';';

        return implode("\n", $lignes);
    }

    /**
     * Rend les arguments d'une colonne de jointure.
     *
     * nullable n'est écrit que lorsqu'il vaut false, et seulement là où il a un
     * sens. Dans une table de jointure, la clé primaire rend la colonne
     * obligatoire de toute façon. Sur une association qui fait partie de
     * l'identifiant, Doctrine l'ignore — la clé primaire s'en charge — et
     * l'écrire est déprécié depuis ORM 3.6, erreur annoncée en 4.0
     * (doctrine/orm#12126). C'est la présence de l'attribut qui déclenche la
     * dépréciation, pas sa valeur. Le DDL est identique avec ou sans, sous
     * ORM 2 comme sous ORM 3 : il n'est donc écrit sous aucune cible, sans
     * règle par version.
     *
     * @param bool $avecNullable faux là où Doctrine ignore nullable : table de jointure, clé
     *
     * @return array<string, mixed>
     */
    private static function argumentsJointure(ColonneJointure $jointure, bool $avecNullable): array
    {
        $action = $jointure->aLaSuppression;

        return [
            'name' => IdentifiantsSql::colonne($jointure->colonne),
            'referencedColumnName' => IdentifiantsSql::colonne($jointure->colonneReferencee),
            'nullable' => $avecNullable && !$jointure->nullable ? false : null,
            'onDelete' => $action === null || $action === ActionSuppression::Aucune ? null : self::ON_DELETE[$action->value],
        ];
    }

    /**
     * Rend le type déclaré d'une association.
     */
    private static function typeAssociation(Association $association): string
    {
        if (self::estCollection($association)) {
            return 'Collection';
        }
        $obligatoire = $association->proprietaire
            && $association->jointure !== []
            && array_filter($association->jointure, static fn(ColonneJointure $j): bool => $j->nullable) === [];

        return ($obligatoire ? '' : '?') . $association->cible;
    }

    /**
     * Rend les accesseurs d'une association.
     *
     * Le côté inverse n'a que son accesseur : Doctrine n'écrit que le côté
     * propriétaire, et un addCommande() sur Client qui ne toucherait pas
     * Commande::$client ajouterait en mémoire sans rien persister. Le tenir à
     * jour depuis la classe de base passerait $this, typé ClientBase, là où
     * setClient() attend un Client — une erreur que l'analyse statique de
     * chaque projet signalerait dans du code généré.
     *
     * Le côté propriétaire d'un objet a son mutateur, celui d'une collection
     * ses add et remove, qui écrivent réellement.
     */
    private function accesseursAssociation(Association $association): string
    {
        $i = Emetteur::INDENTATION;
        $nom = $association->nom;
        $suffixe = ucfirst($nom);
        $type = self::typeAssociation($association);

        $lecture = [
            $i . 'public function get' . $suffixe . '(): ' . $type,
            $i . '{',
            $i . $i . 'return $this->' . $nom . ';',
            $i . '}',
        ];
        if (self::estCollection($association)) {
            array_unshift($lecture, $i . '/** @return Collection<int, ' . $association->cible . '> */');
        }
        $methodes = [implode("\n", $lecture)];

        if (!$association->proprietaire) {
            return $methodes[0];
        }

        if (!self::estCollection($association)) {
            $methodes[] = implode("\n", [
                $i . 'public function set' . $suffixe . '(' . $type . ' $' . $nom . '): static',
                $i . '{',
                $i . $i . '$this->' . $nom . ' = $' . $nom . ';',
                '',
                $i . $i . 'return $this;',
                $i . '}',
            ]);

            return implode("\n\n", $methodes);
        }

        $element = $association->cible;
        $parametre = lcfirst($element);
        $methodes[] = implode("\n", [
            $i . 'public function add' . $suffixe . '(' . $element . ' $' . $parametre . '): static',
            $i . '{',
            $i . $i . 'if (!$this->' . $nom . '->contains($' . $parametre . ')) {',
            $i . $i . $i . '$this->' . $nom . '->add($' . $parametre . ');',
            $i . $i . '}',
            '',
            $i . $i . 'return $this;',
            $i . '}',
        ]);
        $methodes[] = implode("\n", [
            $i . 'public function remove' . $suffixe . '(' . $element . ' $' . $parametre . '): static',
            $i . '{',
            $i . $i . '$this->' . $nom . '->removeElement($' . $parametre . ');',
            '',
            $i . $i . 'return $this;',
            $i . '}',
        ]);

        return implode("\n\n", $methodes);
    }

    /**
     * Rend l'accesseur et, quand la propriété s'écrit depuis PHP, le
     * mutateur.
     *
     * Pas de mutateur pour une clé produite par la base, ni pour une colonne
     * générée ou écrite par une association : Doctrine ne l'écrirait pas, et
     * une méthode qui ne fait rien est pire qu'une méthode absente.
     */
    private function accesseurs(Propriete $propriete, ?Identifiant $identifiant, string $type): string
    {
        $i = Emetteur::INDENTATION;
        $suffixe = ucfirst($propriete->nom);
        $lecture = ltrim($type, '?') === 'bool' ? 'is' . $suffixe : 'get' . $suffixe;

        $tableau = self::estTableau($type);

        $methodes = [implode("\n", [
            ...Emetteur::docblock([], $tableau ? ['@return ' . self::typeTableau($type)] : [], $i),
            $i . 'public function ' . $lecture . '(): ' . $type,
            $i . '{',
            $i . $i . 'return $this->' . $propriete->nom . ';',
            $i . '}',
        ])];

        $ecrivable = $propriete->insertable || $propriete->modifiable;
        if ($ecrivable && !$this->estGeneree($propriete, $identifiant)) {
            $methodes[] = implode("\n", [
                ...Emetteur::docblock([], $tableau ? ['@param ' . self::typeTableau($type) . ' $' . $propriete->nom] : [], $i),
                $i . 'public function set' . $suffixe . '(' . $type . ' $' . $propriete->nom . '): static',
                $i . '{',
                $i . $i . '$this->' . $propriete->nom . ' = $' . $propriete->nom . ';',
                '',
                $i . $i . 'return $this;',
                $i . '}',
            ]);
        }

        return implode("\n\n", $methodes);
    }

    /**
     * Dit si un type déclaré est un tableau, dont l'analyse statique du projet
     * exige le type des éléments dès le niveau 6.
     */
    private static function estTableau(string $type): bool
    {
        return ltrim($type, '?') === 'array';
    }

    /**
     * Rend le type d'un tableau pour un docblock.
     *
     * array<mixed> et pas array<string, mixed> : une colonne JSON contient
     * aussi bien une liste qu'un objet, et le calque ne dit pas lequel.
     */
    private static function typeTableau(string $type): string
    {
        return str_starts_with($type, '?') ? 'array<mixed>|null' : 'array<mixed>';
    }

    /**
     * Rend les options de colonne d'une propriété : son défaut, dans le type
     * de la valeur stockée quand la conversion est certaine, son commentaire,
     * sa longueur fixe, et l'option jsonb quand la cible n'a pas le type. Les
     * mêmes pour une colonne rendue en #[ORM\Column] et pour une colonne de
     * jointure de la clé, qui n'a pas de propriété.
     *
     * Une colonne de jointure hors clé n'en reçoit pas : Doctrine la crée avec
     * les options de la colonne visée, longueur fixe et collation comprises.
     *
     * La collation part telle que le calque la nomme : DBAL la cite en un seul
     * identifiant, ce qui convient à une collation du schéma système, la seule
     * que l'inférence reporte. Un ALTER … TYPE que migrations:diff propose
     * pour une autre raison la perd pourtant : DBAL n'y écrit pas COLLATE.
     *
     * @return array<string, bool|Code|float|int|string>
     */
    public function optionsDeColonne(Propriete $propriete): array
    {
        $defaut = $this->defaut($propriete);

        return array_filter([
            'default' => $propriete->defaut === null ? $this->defautCalcule($propriete) : ($defaut ?? $propriete->defaut),
            'comment' => $propriete->commentaire,
            'fixed' => $propriete->longueurFixe ? true : null,
            'collation' => $propriete->collation,
            'jsonb' => $propriete->typeDoctrine === 'jsonb' && !$this->cible->connaitJsonb() ? true : null,
        ], static fn($valeur): bool => $valeur !== null);
    }

    /**
     * Rend le type Doctrine écrit dans #[ORM\Column], replié sur ce que la
     * cible connaît.
     *
     * jsonb devient json, son option dit le reste : les deux recréent JSONB et
     * se relisent sans écart, sous DBAL 3.10 comme 4.2.
     * smallfloat devient float, qui recrée DOUBLE PRECISION : aucune forme ne
     * décrit la simple précision avant DBAL 4.1, et migrations:diff ne propose
     * rien face à la base d'origine, DBAL 3 relisant real en float.
     */
    private function typeDeColonne(Propriete $propriete): string
    {
        return match (true) {
            $propriete->typeDoctrine === 'jsonb' && !$this->cible->connaitJsonb() => 'json',
            $propriete->typeDoctrine === 'smallfloat' && !$this->cible->connaitSmallfloat() => 'float',
            default => $propriete->typeDoctrine,
        };
    }

    /**
     * Rend le défaut calculé d'une propriété dans la forme que la cible lit.
     *
     * DBAL 4.4 le décrit par un objet et déprécie la chaîne 'CURRENT_TIMESTAMP'
     * (dbal#7195, signalé par le SchemaTool d'ORM 3.7), qui reste la seule
     * forme avant : la classe n'existe pas sous DBAL 3. Les deux écrivent le
     * même DDL. Aucune ne reproduit un now() d'origine, que migrations:diff
     * propose de réécrire en CURRENT_TIMESTAMP : même sens, texte différent.
     *
     * Sous SQL Server, la date et l'heure du jour prennent l'instant courant :
     * la base le convertit dans le type de la colonne, et DBAL relit getdate()
     * en CurrentTimestamp quel que soit ce type. CurrentDate s'y écrit
     * CONVERT(date, GETDATE()), relu en chaîne, et migrations:diff
     * proposerait de le recréer à chaque passage. Avant DBAL 4.4, la chaîne
     * 'CURRENT_TIMESTAMP' n'est reconnue que sur un horodatage : sur une date,
     * DBAL l'écrirait en littéral, que toute insertion refuse. Seul CONVERT
     * s'écrit, avec ce diff perpétuel, que le rapport signale
     * (DefautRepropose).
     *
     * La propriété n'est pas initialisée : la valeur n'est connue que de la
     * base, à l'insertion.
     */
    private function defautCalcule(Propriete $propriete): Code|string|null
    {
        // Aucune forme DBAL : une chaîne serait écrite en littéral, que
        // PostgreSQL refuse et que SQL Server stocke en texte. Le rapport le
        // dit (DefautNonReproduit).
        if ($propriete->defautExpression === null || $propriete->defautExpression === ExpressionDefaut::UuidGenere) {
            return null;
        }
        if ($this->sgbd === 'sqlserver') {
            return $this->cible->defautParExpression()
                ? new Code('new \Doctrine\DBAL\Schema\DefaultExpression\CurrentTimestamp()')
                : self::defautConvertiSqlServer($propriete->defautExpression) ?? 'CURRENT_TIMESTAMP';
        }
        if (!$this->cible->defautParExpression()) {
            return match ($propriete->defautExpression) {
                ExpressionDefaut::HorodatageCourant => 'CURRENT_TIMESTAMP',
                ExpressionDefaut::DateCourante => 'CURRENT_DATE',
                ExpressionDefaut::HeureCourante => 'CURRENT_TIME',
            };
        }

        $classe = match ($propriete->defautExpression) {
            ExpressionDefaut::HorodatageCourant => 'CurrentTimestamp',
            ExpressionDefaut::DateCourante => 'CurrentDate',
            ExpressionDefaut::HeureCourante => 'CurrentTime',
        };

        return new Code('new \Doctrine\DBAL\Schema\DefaultExpression\\' . $classe . '()');
    }

    /**
     * Rend la forme que la plateforme SQL Server de DBAL avant 4.4 reconnaît
     * pour la date ou l'heure du jour, null pour l'instant courant, qui
     * s'écrit CURRENT_TIMESTAMP.
     *
     * C'est la chaîne exacte de getCurrentDateSQL et getCurrentTimeSQL : DBAL
     * ne l'écrit en expression que si elle lui est identique.
     */
    public static function defautConvertiSqlServer(ExpressionDefaut $expression): ?string
    {
        return match ($expression) {
            ExpressionDefaut::HorodatageCourant, ExpressionDefaut::UuidGenere => null,
            ExpressionDefaut::DateCourante => 'CONVERT(date, GETDATE())',
            ExpressionDefaut::HeureCourante => 'CONVERT(time, GETDATE())',
        };
    }

    /**
     * Convertit le défaut littéral du calque dans le type de la valeur stockée.
     *
     * Rend null quand la conversion n'est pas certaine — une date, un booléen
     * écrit « O » — : le défaut reste alors une option de colonne, dans sa
     * forme textuelle, sans initialiser la propriété.
     */
    private function defaut(Propriete $propriete): bool|float|int|string|null
    {
        $valeur = $propriete->defaut;
        if ($valeur === null) {
            return null;
        }

        return match (TypesPhp::nu($propriete, $this->cible)) {
            'int' => preg_match('/^-?\d+$/', $valeur) === 1 ? (int) $valeur : null,
            'float' => is_numeric($valeur) ? (float) $valeur : null,
            'bool' => match (strtolower($valeur)) {
                'true', 't', '1' => true,
                'false', 'f', '0' => false,
                default => null,
            },
            'string' => $valeur,
            default => null,
        };
    }

    /**
     * Rend le cas d'énumération qui initialise une propriété énumérée avec un
     * défaut, ou null.
     *
     * Le défaut est comparé à la valeur stockée de chaque cas, dans le type
     * support de l'énumération. Un défaut qu'aucun cas ne porte n'initialise
     * rien : inventer un cas serait décider à la place du calque.
     */
    private function casDuDefaut(Propriete $propriete): ?Code
    {
        if ($propriete->enumeration === null || $propriete->defaut === null) {
            return null;
        }
        $enumeration = $this->enumerations[$propriete->enumeration] ?? null;
        if ($enumeration === null) {
            return null;
        }

        foreach ($enumeration->cas as $cas) {
            $egal = $enumeration->typeSupport === TypeSupport::Entier
                ? (string) $cas->valeur === $propriete->defaut
                : $cas->valeur === $propriete->defaut;
            if ($egal) {
                return new Code($enumeration->nom . '::' . $cas->nom);
            }
        }

        return null;
    }

    /**
     * Rend le dernier segment d'un nom qualifié.
     */
    private static function nomCourt(string $qualifie): string
    {
        $segments = explode('\\', $qualifie);

        return end($segments);
    }
}
