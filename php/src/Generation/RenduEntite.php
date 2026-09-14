<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

use Ormeau\Doctrine\Calque\Entite;
use Ormeau\Doctrine\Calque\Enumeration;
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
    /**
     * En-tête des fichiers qui appartiennent à l'outil : énumérations et
     * traits. Sans version ni date, qui changeraient chaque fichier à chaque
     * mise à jour de l'outil.
     */
    private const MENTION_OUTIL = 'Généré par Ormeau et réécrit à chaque génération.';

    /**
     * En-tête de la classe de l'utilisateur, écrite une fois : c'est la
     * garantie du projet, dite là où on la lit.
     */
    private const MENTION_UTILISATEUR = 'Créé par Ormeau, jamais réécrit : ce fichier appartient au projet.';

    /** Écrit propriétés et accesseurs, pour les classes de base comme pour les traits. */
    private readonly RenduMembres $membres;

    /**
     * @param Cible                      $cible        détermine le type PHP de chaque propriété
     * @param string                     $espaceDeNoms espace de noms des classes de l'utilisateur ;
     *                                                 Base, Enum et Trait s'y ajoutent
     * @param bool                       $avecSchema   écrire le schéma dans #[ORM\Table] ; vrai quand
     *                                                 les entités du calque viennent de plusieurs
     *                                                 schémas, faute de quoi deux tables de même nom
     *                                                 se confondraient
     * @param array<string, Enumeration> $enumerations énumérations du calque par nom
     * @param ClassesUtilisateur|null    $classes      où vivent les classes de l'utilisateur ; sans elle,
     *                                                 chacune est supposée à la racine
     */
    public function __construct(
        Cible $cible,
        private readonly string $espaceDeNoms,
        private readonly bool $avecSchema,
        array $enumerations = [],
        private readonly ?ClassesUtilisateur $classes = null,
    ) {
        $this->membres = new RenduMembres($cible, $espaceDeNoms, $enumerations, $avecSchema, $classes);
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
            ...IdentifiantsSql::table($entite->table->nom, $this->avecSchema ? $entite->table->schema : null),
            'options' => $entite->commentaire === null ? null : ['comment' => $entite->commentaire],
        ];
    }

    /**
     * Rend name et schema de #[ORM\Table] tels que la 0.5.0 les écrivait.
     *
     * Rétro-compatibilité datée, pour le seul message de divergence : une
     * classe de l'utilisateur créée en 0.5.0 dans un schéma à citer porte des
     * backticks sur le schéma, forme que Doctrine cite deux fois et que
     * PostgreSQL refuse. Reconnaître cette forme permet de dire que la table
     * était inutilisable, et non que la régénération casse ce qui marchait. À
     * retirer quand plus aucun projet ne garde de classe générée en 0.5.0.
     *
     * @return array{name: string, schema: string|null}
     */
    public function argumentsTableEn050(Entite $entite): array
    {
        $ancien = static fn(string $nom): string => preg_match('/^[a-z_][a-z0-9_]*$/', $nom) === 1 ? $nom : '`' . $nom . '`';

        return [
            'name' => $ancien($entite->table->nom),
            'schema' => $this->avecSchema ? $ancien($entite->table->schema) : null,
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
            ...$this->entete(self::MENTION_UTILISATEUR, $this->espaceDeNoms, [$this->classeBaseQualifiee($entite), 'Doctrine\ORM\Mapping as ORM']),
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
        );

        $imports = ['Doctrine\ORM\Mapping as ORM', ...$rendu['imports']];
        foreach ($entite->traits as $trait) {
            $imports[] = $this->espaceDeNoms . '\\Trait\\' . $trait;
        }
        if ($parent !== null) {
            $imports[] = $this->classes?->qualifiee($parent->nom) ?? $this->espaceDeNoms . '\\' . $parent->nom;
        }

        $lignes = $this->entete(
            sprintf('Généré par Ormeau et réécrit à chaque génération : le code propre à %s va dans %s.php.', $entite->nom, $entite->nom),
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
            $lignes[] = Emetteur::attribut(
                $index->unique ? 'ORM\UniqueConstraint' : 'ORM\Index',
                ['name' => $index->nom, 'columns' => $index->colonnes],
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

        $lignes = $this->entete(self::MENTION_OUTIL, $this->espaceDeNoms . '\\Trait', ['Doctrine\ORM\Mapping as ORM', ...$rendu['imports']]);
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
        $lignes = $this->entete(self::MENTION_OUTIL, $this->espaceDeNoms . '\\Enum', []);
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
