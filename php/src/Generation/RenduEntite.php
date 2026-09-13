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
     */
    public function __construct(
        Cible $cible,
        private readonly string $espaceDeNoms,
        private readonly bool $avecSchema,
        array $enumerations = [],
    ) {
        $this->membres = new RenduMembres($cible, $espaceDeNoms, $enumerations, $avecSchema);
    }

    /**
     * Rend le nom de la classe de base d'une entité.
     */
    public static function nomBase(Entite $entite): string
    {
        return $entite->nom . 'Base';
    }

    /**
     * Rend un nom de table ou de colonne tel que Doctrine doit l'écrire.
     *
     * Doctrine ne cite un identifiant que s'il est entouré de backticks. Un nom
     * qui n'est pas un identifiant SQL simple en minuscules — « N° Commande »,
     * T_CLIENTS — doit l'être : sans guillemets, PostgreSQL ramène T_CLIENTS à
     * t_clients, et la requête vise une table qui n'existe pas. Les autres
     * restent nus, comme on les écrirait à la main.
     *
     * Les mots réservés d'un SGBD ne sont pas traités : la liste dépend du
     * SGBD, que le calque logique ne porte pas.
     */
    public static function identifiantSql(string $nom): string
    {
        return preg_match('/^[a-z_][a-z0-9_]*$/', $nom) === 1 ? $nom : '`' . $nom . '`';
    }

    /**
     * Rend les arguments attendus de #[ORM\Table] pour une entité. Le contrôle
     * de la classe de l'utilisateur compare à ces valeurs-là, et à rien
     * d'autre.
     *
     * @return array{name: string, schema: string|null}
     */
    public function argumentsTable(Entite $entite): array
    {
        return [
            'name' => self::identifiantSql($entite->table->nom),
            'schema' => $this->avecSchema ? self::identifiantSql($entite->table->schema) : null,
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
     * rien d'autre. Elle n'est écrite qu'une fois, et ce qu'on y ajoute ensuite
     * appartient à l'utilisateur.
     */
    public function classeUtilisateur(Entite $entite): string
    {
        $lignes = [
            '<?php',
            '',
            'declare(strict_types=1);',
            '',
            'namespace ' . $this->espaceDeNoms . ';',
            '',
            'use ' . $this->classeBaseQualifiee($entite) . ';',
            'use Doctrine\ORM\Mapping as ORM;',
            '',
            Emetteur::attribut('ORM\Entity', [], ''),
            Emetteur::attribut('ORM\Table', $this->argumentsTable($entite), ''),
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
     */
    public function classeBase(Entite $entite): string
    {
        $rendu = $this->membres->rendre($entite->proprietes, $entite->identifiant, $entite->associations);

        $imports = ['Doctrine\ORM\Mapping as ORM', ...$rendu['imports']];
        foreach ($entite->traits as $trait) {
            $imports[] = $this->espaceDeNoms . '\\Trait\\' . $trait;
        }

        $lignes = $this->entete($this->espaceDeNoms . '\\Base', $imports);
        $lignes[] = Emetteur::attribut('ORM\MappedSuperclass', [], '');
        foreach ($entite->index as $index) {
            $lignes[] = Emetteur::attribut(
                $index->unique ? 'ORM\UniqueConstraint' : 'ORM\Index',
                ['name' => $index->nom, 'columns' => $index->colonnes],
                '',
            );
        }
        $lignes[] = 'abstract class ' . self::nomBase($entite);
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

        $lignes = $this->entete($this->espaceDeNoms . '\\Trait', ['Doctrine\ORM\Mapping as ORM', ...$rendu['imports']]);
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
        $lignes = $this->entete($this->espaceDeNoms . '\\Enum', []);
        $lignes[] = 'enum ' . $enumeration->nom . ': ' . $enumeration->typeSupport->value;
        $lignes[] = '{';
        foreach ($enumeration->cas as $cas) {
            $lignes[] = Emetteur::INDENTATION . 'case ' . $cas->nom . ' = ' . Emetteur::litteral($cas->valeur) . ';';
        }
        $lignes[] = '}';

        return implode("\n", $lignes) . "\n";
    }

    /**
     * Rend le début d'un fichier : balise, déclaration stricte, espace de noms
     * et imports triés, suivis d'une ligne vide.
     *
     * @param string       $espaceDeNoms espace de noms du fichier
     * @param list<string> $imports      classes à importer, doublons compris
     *
     * @return list<string>
     */
    private function entete(string $espaceDeNoms, array $imports): array
    {
        $lignes = ['<?php', '', 'declare(strict_types=1);', '', 'namespace ' . $espaceDeNoms . ';', ''];
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
