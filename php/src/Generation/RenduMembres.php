<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

use Ormeau\Doctrine\Calque\ActionSuppression;
use Ormeau\Doctrine\Calque\Association;
use Ormeau\Doctrine\Calque\ColonneJointure;
use Ormeau\Doctrine\Calque\Enumeration;
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
     * @param bool                       $avecSchema   écrire le schéma d'une table de jointure
     */
    public function __construct(
        private readonly Cible $cible,
        private readonly string $espaceDeNoms,
        private readonly array $enumerations,
        private readonly bool $avecSchema = false,
    ) {}

    /**
     * Rend les propriétés, les associations et leurs accesseurs, avec les
     * imports qu'ils demandent et les collections à initialiser.
     *
     * Propriétés, puis associations, puis accesseurs dans le même ordre : c'est
     * l'ordre du calque, celui des colonnes de la table.
     *
     * @param list<Propriete>   $proprietes   propriétés à rendre, dans l'ordre du calque
     * @param Identifiant|null  $identifiant  clé de l'entité ; null pour un trait, qui n'en porte pas
     * @param list<Association> $associations associations de l'entité ; aucune pour un trait
     *
     * @return array{imports: list<string>, membres: list<string>, accesseurs: list<string>, collections: list<string>}
     */
    public function rendre(array $proprietes, ?Identifiant $identifiant, array $associations = []): array
    {
        $jointures = self::colonnesDeJointure($associations);
        $clesDerivees = array_flip(self::associationsDeCle($proprietes, $identifiant, $associations));

        $imports = [];
        $membres = [];
        $accesseurs = [];

        foreach ($proprietes as $propriete) {
            $ecritePar = $jointures[$propriete->colonne] ?? null;
            if ($ecritePar !== null) {
                if (self::dansLaCle($propriete, $identifiant)) {
                    continue;
                }
                $propriete = self::enLectureSeule($propriete);
            }

            [$type, $import] = $this->type($propriete, $identifiant);
            if ($import !== null) {
                $imports[] = $import;
            }
            $membres[] = $this->propriete($propriete, $identifiant, $type, $ecritePar);
            $accesseurs[] = $this->accesseurs($propriete, $identifiant, $type);
        }

        $collections = [];
        foreach ($associations as $association) {
            $imports[] = $this->espaceDeNoms . '\\' . $association->cible;
            if (self::estCollection($association)) {
                $imports[] = 'Doctrine\Common\Collections\ArrayCollection';
                $imports[] = 'Doctrine\Common\Collections\Collection';
                $collections[] = $association->nom;
            }
            $membres[] = $this->association($association, isset($clesDerivees[$association->nom]));
            $accesseurs[] = $this->accesseursAssociation($association);
        }

        return ['imports' => $imports, 'membres' => $membres, 'accesseurs' => $accesseurs, 'collections' => $collections];
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
     */
    private static function enLectureSeule(Propriete $p): Propriete
    {
        return new Propriete(
            $p->nom,
            $p->colonne,
            $p->typePhp,
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
     */
    private function propriete(Propriete $propriete, ?Identifiant $identifiant, string $type, ?string $ecritePar = null): string
    {
        $indentation = Emetteur::INDENTATION;
        $texte = $propriete->commentaire === null ? [] : Emetteur::commentaire($propriete->commentaire);
        if ($ecritePar !== null) {
            $note = sprintf('Lecture seule : écrite par l\'association %s.', $ecritePar);
            $texte = $texte === [] ? [$note] : [...$texte, '', $note];
        }
        $lignes = Emetteur::docblock($texte, self::estTableau($type) ? ['@var ' . self::typeTableau($type)] : [], $indentation);
        $generee = $this->estGeneree($propriete, $identifiant);

        if ($identifiant !== null && in_array($propriete->nom, $identifiant->proprietes, true)) {
            $lignes[] = Emetteur::attribut('ORM\Id', [], $indentation);
            if ($generee) {
                $sequence = $identifiant->strategie === StrategieIdentifiant::Sequence;
                $lignes[] = Emetteur::attribut('ORM\GeneratedValue', ['strategy' => $sequence ? 'SEQUENCE' : 'IDENTITY'], $indentation);
                if ($sequence && $identifiant->sequence !== null) {
                    $lignes[] = Emetteur::attribut('ORM\SequenceGenerator', ['sequenceName' => $identifiant->sequence], $indentation);
                }
            }
        }

        $defaut = $this->defaut($propriete);
        $options = array_filter([
            'default' => $propriete->defaut === null ? null : ($defaut ?? $propriete->defaut),
            'comment' => $propriete->commentaire,
        ], static fn($valeur): bool => $valeur !== null);

        $enumType = null;
        if ($propriete->enumeration !== null) {
            $enumType = new Code(self::nomCourt($this->enumerationQualifiee($propriete->enumeration)) . '::class');
        }

        $lignes[] = Emetteur::attribut('ORM\Column', [
            'name' => RenduEntite::identifiantSql($propriete->colonne),
            'type' => $propriete->typeDoctrine,
            'length' => $propriete->longueur,
            'precision' => $propriete->precision,
            'scale' => $propriete->echelle,
            'unique' => $propriete->unique ? true : null,
            'nullable' => $propriete->nullable ? true : null,
            'insertable' => $propriete->insertable ? null : false,
            'updatable' => $propriete->modifiable ? null : false,
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
     */
    private function association(Association $association, bool $dansLaCle): string
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
            'orphanRemoval' => $association->orphelinsSupprimes ? true : null,
        ], $i);

        if ($association->proprietaire && $association->tableJointure !== null) {
            $table = $association->tableJointure;
            $lignes[] = Emetteur::attribut('ORM\JoinTable', [
                'name' => RenduEntite::identifiantSql($table->nom),
                'schema' => $this->avecSchema ? RenduEntite::identifiantSql($table->schema) : null,
            ], $i);
            foreach ($table->jointure as $jointure) {
                $lignes[] = Emetteur::attribut('ORM\JoinColumn', self::argumentsJointure($jointure, false), $i);
            }
            foreach ($table->jointureInverse as $jointure) {
                $lignes[] = Emetteur::attribut('ORM\InverseJoinColumn', self::argumentsJointure($jointure, false), $i);
            }
        } elseif ($association->proprietaire) {
            foreach ($association->jointure as $jointure) {
                $lignes[] = Emetteur::attribut('ORM\JoinColumn', self::argumentsJointure($jointure, true), $i);
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
     * nullable n'est écrit que lorsqu'il vaut false, sur la table de l'entité :
     * Doctrine suppose une jointure facultative, et dans une table de jointure
     * la clé primaire la rend de toute façon obligatoire.
     *
     * @return array<string, mixed>
     */
    private static function argumentsJointure(ColonneJointure $jointure, bool $surLEntite): array
    {
        $action = $jointure->aLaSuppression;

        return [
            'name' => RenduEntite::identifiantSql($jointure->colonne),
            'referencedColumnName' => RenduEntite::identifiantSql($jointure->colonneReferencee),
            'nullable' => $surLEntite && !$jointure->nullable ? false : null,
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
