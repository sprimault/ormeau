<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

use Ormeau\Doctrine\Calque\Propriete;

/**
 * Le type PHP qu'hydrate chaque type Doctrine, selon la cible.
 *
 * C'est ici, et non dans le calque, que le type PHP se décide : il dépend de
 * DBAL, que le calque ignore. Chaque entrée suit ce que convertToPHPValue rend
 * réellement dans DBAL 3.10 et 4.4, lu dans leur code et non dans leur
 * documentation. Un type PHP qui ne correspond pas à l'hydratation produit une
 * entité que Doctrine charge, puis croit modifiée à chaque flush.
 *
 * Deux types changent d'une majeure à l'autre, et justifient à eux seuls la
 * cible : bigint, rendu en chaîne par DBAL 3 et en entier par DBAL 4, et binary,
 * rendu en ressource par DBAL 3 et en chaîne par DBAL 4.
 *
 * Le champ type_php du calque n'est lu que pour un type absent de cette table,
 * typiquement un type personnalisé forcé par décision : seul son auteur sait ce
 * qu'il hydrate, et le calque porte au moins ce que la colonne suggérait.
 */
final class TypesPhp
{
    /**
     * Type PHP par type Doctrine et par majeure d'ORM.
     *
     * mixed pour blob, et pour binary sous DBAL 3 : une ressource n'a pas de
     * type déclarable. int pour bigint sous DBAL 4, qui rend une chaîne au-delà
     * de PHP_INT_MAX — un BIGINT UNSIGNED de MySQL, rare en pratique et qu'un
     * type forcé par décision traite. array pour json, comme l'écrit
     * make:entity : DBAL rend ce que json_decode rend, et une colonne JSON de
     * base reprise porte des objets.
     */
    private const TABLE = [
        'integer' => [2 => 'int', 3 => 'int'],
        'smallint' => [2 => 'int', 3 => 'int'],
        'bigint' => [2 => 'string', 3 => 'int'],
        'float' => [2 => 'float', 3 => 'float'],
        'boolean' => [2 => 'bool', 3 => 'bool'],
        'decimal' => [2 => 'string', 3 => 'string'],
        'string' => [2 => 'string', 3 => 'string'],
        'text' => [2 => 'string', 3 => 'string'],
        'guid' => [2 => 'string', 3 => 'string'],
        'blob' => [2 => 'mixed', 3 => 'mixed'],
        'binary' => [2 => 'mixed', 3 => 'string'],
        'json' => [2 => 'array', 3 => 'array'],
        'simple_array' => [2 => 'array', 3 => 'array'],
        'date_immutable' => [2 => '\DateTimeImmutable', 3 => '\DateTimeImmutable'],
        'time_immutable' => [2 => '\DateTimeImmutable', 3 => '\DateTimeImmutable'],
        'datetime_immutable' => [2 => '\DateTimeImmutable', 3 => '\DateTimeImmutable'],
        'datetimetz_immutable' => [2 => '\DateTimeImmutable', 3 => '\DateTimeImmutable'],
        'dateinterval' => [2 => '\DateInterval', 3 => '\DateInterval'],
    ];

    /**
     * Dit si la table connaît un type Doctrine.
     */
    public static function connu(string $typeDoctrine): bool
    {
        return isset(self::TABLE[$typeDoctrine]);
    }

    /**
     * Rend la déclaration de type d'une propriété, nullabilité comprise.
     *
     * mixed ne prend pas de « ? » : il contient déjà null.
     */
    public static function declaration(Propriete $propriete, Cible $cible): string
    {
        $type = self::nu($propriete, $cible);
        if (!$propriete->nullable || $type === 'mixed') {
            return $type;
        }

        return '?' . $type;
    }

    /**
     * Rend le type PHP sans sa nullabilité.
     *
     * Hors table, le type_php du calque est repris, débarrassé d'un « ? » que
     * la nullabilité de la propriété redira.
     */
    public static function nu(Propriete $propriete, Cible $cible): string
    {
        return self::TABLE[$propriete->typeDoctrine][$cible->ormMajeure] ?? ltrim($propriete->typePhp, '?');
    }
}
