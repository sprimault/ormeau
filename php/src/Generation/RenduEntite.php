<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

use Ormeau\Doctrine\Calque\Entite;
use Ormeau\Doctrine\Calque\Enumeration;
use Ormeau\Doctrine\Calque\StrategieIdentifiant;
use Ormeau\Doctrine\Calque\TraitPartage;

/**
 * Écrit le source des fichiers d'un calque, en mode classe de base : les deux
 * classes de chaque entité, ses énumérations et ses traits.
 *
 * Ce qui va où n'est pas un choix de présentation, c'est ce que Doctrine
 * accepte, constaté sous ORM 2.14 à 3.7 : #[ORM\Table] posé sur un
 * MappedSuperclass est ignoré, et l'héritage y est refusé. Entity et Table vont
 * donc dans la classe de l'utilisateur, écrite une fois ; propriétés,
 * identifiant, accesseurs, index et traits vont dans la classe de base,
 * réécrite à chaque passage. Énumérations et traits appartiennent à l'outil au
 * même titre que la classe de base, et sont réécrits comme elle.
 *
 * Le rendu ne lit rien d'autre que le calque, la cible et l'espace de noms :
 * deux appels identiques rendent deux sources identiques octet pour octet.
 *
 * @internal
 */
final class RenduEntite
{
    /** Écrit propriétés et accesseurs, pour les classes de base comme pour les traits. */
    private readonly RenduMembres $membres;

    /**
     * @param Cible                      $cible        détermine le type PHP de chaque propriété
     * @param string                     $base         base dont vient le calque, nommée dans l'en-tête de
     *                                                 chaque fichier
     * @param string                     $espaceDeNoms espace de noms des classes de l'utilisateur ;
     *                                                 Base, Enum et Trait s'y ajoutent
     * @param string                     $sgbd         SGBD du calque : il décide si le schéma d'une table
     *                                                 s'écrit (SchemaParDefaut), et une clé par séquence ne
     *                                                 se rend pas de la même façon sous chaque plateforme DBAL
     * @param array<string, Enumeration> $enumerations énumérations du calque par nom
     * @param ClassesUtilisateur|null    $classes      où vivent les classes de l'utilisateur ; sans elle,
     *                                                 chacune est supposée à la racine
     */
    public function __construct(
        private readonly Cible $cible,
        private readonly string $base,
        private readonly string $espaceDeNoms,
        private readonly string $sgbd,
        array $enumerations = [],
        private readonly ?ClassesUtilisateur $classes = null,
    ) {
        $this->membres = new RenduMembres($cible, $espaceDeNoms, $enumerations, $sgbd, $classes);
    }

    /**
     * Dit si la clé d'une entité racine tire sa séquence par un générateur
     * produit, écrit dans Base/Generateur.
     *
     * Sous ORM 3 et SQL Server, rien d'autre n'attribue l'identifiant : IDENTITY
     * lit SCOPE_IDENTITY(), qui ne voit pas une valeur tirée d'une séquence, et
     * #[SequenceGenerator] est ignoré sur une classe de base mappée. Sous
     * PostgreSQL, IDENTITY lit LASTVAL() et reste juste ; un générateur y ferait
     * au contraire proposer DROP SEQUENCE … CASCADE, qui passe. Constaté sous
     * ORM 3.7 le 2026-09-16.
     */
    public function tireParGenerateur(Entite $entite): bool
    {
        $identifiant = $entite->identifiant;

        return $this->cible->ormMajeure >= 3
            && $this->sgbd === 'sqlserver'
            && $identifiant !== null
            && count($identifiant->proprietes) === 1
            && $identifiant->strategie === StrategieIdentifiant::Sequence
            && $identifiant->sequence !== null;
    }

    /**
     * Rend le nom court du générateur qui tire la séquence d'une entité.
     */
    public static function nomGenerateur(Entite $entite): string
    {
        return $entite->nom . 'Generateur';
    }

    /**
     * Rend le générateur qui tire la séquence de la clé d'une entité.
     *
     * Il appartient à l'outil comme la classe de base, et ne dépend que de
     * Doctrine : rien du bundle n'est chargé à l'exécution. Le nom de la
     * séquence vient du calque et passe par un littéral échappé.
     */
    public function generateur(Entite $entite): string
    {
        $sequence = $entite->identifiant->sequence ?? '';
        $lignes = $this->entete(EnteteOrmeau::outil($this->base), $this->espaceDeNoms . '\\Base\\Generateur', ['Doctrine\ORM\Id\SequenceGenerator']);
        array_push($lignes, ...Emetteur::docblock(
            Emetteur::commentaire(sprintf('Tire la séquence %s pour la clé de %s : ORM 3 ignore #[ORM\SequenceGenerator] sur une classe de base mappée.', $sequence, $entite->nom)),
            [],
            '',
        ));
        $i = Emetteur::INDENTATION;
        array_push(
            $lignes,
            'class ' . self::nomGenerateur($entite) . ' extends SequenceGenerator',
            '{',
            $i . 'public function __construct()',
            $i . '{',
            $i . $i . 'parent::__construct(' . Emetteur::litteral($sequence) . ', 1);',
            $i . '}',
            '}',
        );

        return implode("\n", $lignes) . "\n";
    }

    /**
     * Rend le nom de la classe de base d'une entité.
     */
    public static function nomBase(Entite $entite): string
    {
        return $entite->nom . 'Base';
    }

    /**
     * Rend les arguments attendus de #[ORM\Table] pour une entité. Le contrôle
     * de la classe de l'utilisateur compare à ces valeurs-là, et à rien
     * d'autre.
     *
     * Le commentaire de la table y figure tel que la base le rend : Doctrine
     * l'ignore quand il compare un schéma existant, mais schema:create le
     * recrée, et c'est ce que l'aller-retour vérifie.
     *
     * @return array{name: string, schema: string|null, options: array{comment: string}|null}
     */
    public function argumentsTable(Entite $entite): array
    {
        return [
            ...IdentifiantsSql::table($entite->table->nom, SchemaParDefaut::ecrit($this->sgbd, $entite->table->schema)),
            'options' => $entite->commentaire === null ? null : ['comment' => $entite->commentaire],
        ];
    }

    /**
     * Rend name et schema de #[ORM\Table] cités comme la 0.5.0 les citait.
     *
     * Rétro-compatibilité datée, pour le seul message de divergence : une
     * classe de l'utilisateur créée en 0.5.0 dans un schéma à citer porte des
     * backticks sur le schéma, forme que Doctrine cite deux fois et que
     * PostgreSQL refuse. Reconnaître cette forme permet de dire que la table
     * était inutilisable, et non que la régénération casse ce qui marchait. À
     * retirer quand plus aucun projet ne garde de classe générée en 0.5.0.
     *
     * Le schéma y est écrit selon la règle courante, pas celle de la 0.5.0 :
     * seule la citation distingue les deux formes, et un schéma absent n'a
     * jamais rendu une table inutilisable.
     *
     * @return array{name: string, schema: string|null}
     */
    public function argumentsTableEn050(Entite $entite): array
    {
        $ancien = static fn(string $nom): string => preg_match('/^[a-z_][a-z0-9_]*$/', $nom) === 1 ? $nom : '`' . $nom . '`';
        $schema = SchemaParDefaut::ecrit($this->sgbd, $entite->table->schema);

        return [
            'name' => $ancien($entite->table->nom),
            'schema' => $schema === null ? null : $ancien($schema),
        ];
    }

    /**
     * Rend l'espace de noms complet de la classe de base d'une entité.
     */
    public function classeBaseQualifiee(Entite $entite): string
    {
        return $this->espaceDeNoms . '\\Base\\' . self::nomBase($entite);
    }

    /**
     * Rend le source de la classe de l'utilisateur : l'entité, sa table, et
     * pour la racine d'un héritage la stratégie, la colonne discriminante et
     * la carte des classes. Elle n'est écrite qu'une fois, et ce qu'on y
     * ajoute ensuite appartient à l'utilisateur.
     *
     * @param Entite              $entite  entité à rendre
     * @param RacineHeritage|null $racine  ce que déclare la racine d'un héritage ; null pour toute
     *                                     autre entité
     */
    public function classeUtilisateur(Entite $entite, ?RacineHeritage $racine = null): string
    {
        $lignes = [
            ...$this->entete(EnteteOrmeau::utilisateur($this->base), $this->espaceDeNoms, [$this->classeBaseQualifiee($entite), 'Doctrine\ORM\Mapping as ORM']),
            ...Emetteur::docblock($entite->commentaire === null ? [] : Emetteur::commentaire($entite->commentaire), [], ''),
            Emetteur::attribut('ORM\Entity', [], ''),
            Emetteur::attribut('ORM\Table', $this->argumentsTable($entite), ''),
            ...array_values($racine?->attributs() ?? []),
            'class ' . $entite->nom . ' extends ' . self::nomBase($entite) . ' {}',
        ];

        return implode("\n", $lignes) . "\n";
    }

    /**
     * Rend le source de la classe de base : traits, propriétés mappées,
     * associations, index, constructeur des collections, puis accesseurs.
     *
     * Le constructeur n'existe que s'il y a une collection à initialiser. Une
     * classe de l'utilisateur qui déclare le sien doit appeler
     * parent::__construct(), comme pour toute classe parente.
     *
     * Dans une hiérarchie, la classe de base d'une classe fille hérite de la
     * classe de l'utilisateur de son parent — SalarieBase extends Personne —,
     * sans redéclarer l'identifiant, et la racine ne mappe pas sa colonne
     * discriminante : Doctrine refuse l'un et l'autre (voir Hierarchies). Un
     * constructeur appelle celui de l'ancêtre qui en a un, sans quoi les
     * collections de la racine resteraient non initialisées sur un new Salarie.
     */
    public function classeBase(Entite $entite, Hierarchies $hierarchies): string
    {
        $parent = $hierarchies->parent($entite);
        $rendu = $this->membres->rendre(
            $hierarchies->proprietesDeclarees($entite),
            $parent === null ? $entite->identifiant : null,
            $entite->associations,
            $parent === null && $this->tireParGenerateur($entite) ? 'Generateur\\' . self::nomGenerateur($entite) : null,
        );

        $imports = ['Doctrine\ORM\Mapping as ORM', ...$rendu['imports']];
        foreach ($entite->traits as $trait) {
            $imports[] = $this->espaceDeNoms . '\\Trait\\' . $trait;
        }
        if ($parent !== null) {
            $imports[] = $this->classes?->qualifiee($parent->nom) ?? $this->espaceDeNoms . '\\' . $parent->nom;
        }

        $lignes = $this->entete(
            EnteteOrmeau::classeBase($this->base, $entite->nom),
            $this->espaceDeNoms . '\\Base',
            $imports,
        );
        $discriminante = $hierarchies->racine($entite) === $entite ? $hierarchies->colonneDiscriminante($entite) : null;
        if ($discriminante !== null) {
            // Un nom de colonne vient de la base comme un commentaire, et se
            // neutralise de la même façon.
            array_push($lignes, ...Emetteur::docblock(
                Emetteur::commentaire(sprintf('La colonne %s départage la hiérarchie : Doctrine l\'écrit, elle n\'a pas de propriété.', $discriminante)),
                [],
                '',
            ));
        }
        $lignes[] = Emetteur::attribut('ORM\MappedSuperclass', [], '');
        foreach ($entite->index as $index) {
            // Le prédicat part verbatim dans options: where, que DBAL recopie
            // derrière WHERE sans le lire. Il vient du catalogue
            // et passe par un littéral échappé, jamais par du code. Côté SQL,
            // aucun filtre : il finit tel quel dans le DDL de Doctrine, et un
            // calque logique reçu d'ailleurs se relit comme une migration.
            $lignes[] = Emetteur::attribut(
                $index->unique ? 'ORM\UniqueConstraint' : 'ORM\Index',
                [
                    'name' => $index->nom,
                    // Citées comme elles le sont dans #[ORM\Column] : DBAL
                    // recopie ces noms tels quels dans le CREATE INDEX, et un
                    // identifiant nu y serait replié en minuscules par la base,
                    // qui refuserait alors la colonne.
                    'columns' => array_map(IdentifiantsSql::colonne(...), $index->colonnes),
                    'options' => $index->predicat === null ? null : ['where' => $index->predicat],
                ],
                '',
            );
        }
        $lignes[] = 'abstract class ' . self::nomBase($entite) . ($parent === null ? '' : ' extends ' . $parent->nom);
        $lignes[] = '{';

        $blocs = [];
        if ($entite->traits !== []) {
            $blocs[] = implode("\n", array_map(static fn(string $t): string => Emetteur::INDENTATION . 'use ' . $t . ';', $entite->traits));
        }
        if ($rendu['collections'] !== []) {
            $i = Emetteur::INDENTATION;
            $initialisations = array_map(
                static fn(string $nom): string => $i . $i . '$this->' . $nom . ' = new ArrayCollection();',
                $rendu['collections'],
            );
            $aUnConstructeur = static fn(Entite $e): bool => array_filter($e->associations, RenduMembres::estCollection(...)) !== [];
            if ($hierarchies->ancetreAvecConstructeur($entite, $aUnConstructeur)) {
                array_unshift($initialisations, $i . $i . 'parent::__construct();');
            }
            $rendu['membres'][] = implode("\n", [$i . 'public function __construct()', $i . '{', ...$initialisations, $i . '}']);
        }
        $lignes[] = implode("\n\n", [...$blocs, ...$rendu['membres'], ...$rendu['accesseurs']]);
        $lignes[] = '}';

        return implode("\n", $lignes) . "\n";
    }

    /**
     * Rend le source d'un trait du calque : ses propriétés mappées et leurs
     * accesseurs, rendus exactement comme dans une classe de base.
     */
    public function traitPartage(TraitPartage $trait): string
    {
        $rendu = $this->membres->rendre($trait->proprietes, null);

        $lignes = $this->entete(EnteteOrmeau::outil($this->base), $this->espaceDeNoms . '\\Trait', ['Doctrine\ORM\Mapping as ORM', ...$rendu['imports']]);
        $lignes[] = 'trait ' . $trait->nom;
        $lignes[] = '{';
        $lignes[] = implode("\n\n", [...$rendu['membres'], ...$rendu['accesseurs']]);
        $lignes[] = '}';

        return implode("\n", $lignes) . "\n";
    }

    /**
     * Rend le source d'une énumération adossée : un cas par valeur, dans
     * l'ordre du calque.
     */
    public function enumeration(Enumeration $enumeration): string
    {
        $lignes = $this->entete(EnteteOrmeau::outil($this->base), $this->espaceDeNoms . '\\Enum', []);
        $lignes[] = 'enum ' . $enumeration->nom . ': ' . $enumeration->typeSupport->value;
        $lignes[] = '{';
        foreach ($enumeration->cas as $cas) {
            $lignes[] = Emetteur::INDENTATION . 'case ' . $cas->nom . ' = ' . Emetteur::litteral($cas->valeur) . ';';
        }
        $lignes[] = '}';

        return implode("\n", $lignes) . "\n";
    }

    /**
     * Rend le début d'un fichier : balise, mention, déclaration stricte,
     * espace de noms et imports triés, suivis d'une ligne vide.
     *
     * @param string       $mention      ce que la ligne d'en-tête dit du fichier
     * @param string       $espaceDeNoms espace de noms du fichier
     * @param list<string> $imports      classes à importer, doublons compris
     *
     * @return list<string>
     */
    private function entete(string $mention, string $espaceDeNoms, array $imports): array
    {
        $lignes = ['<?php', '', '// ' . $mention, '', 'declare(strict_types=1);', '', 'namespace ' . $espaceDeNoms . ';', ''];
        if ($imports === []) {
            return $lignes;
        }

        $imports = array_values(array_unique($imports));
        sort($imports);
        foreach ($imports as $import) {
            $lignes[] = 'use ' . $import . ';';
        }
        $lignes[] = '';

        return $lignes;
    }
}
