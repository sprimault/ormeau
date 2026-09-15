<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

use Ormeau\Doctrine\Calque\Propriete;

/**
 * Le type PHP qu'hydrate chaque type Doctrine, selon la version de DBAL visée.
 *
 * C'est ici, et non dans le calque, que le type PHP se décide : il dépend de
 * DBAL, que le calque ignore. Chaque entrée suit ce que convertToPHPValue rend
 * réellement dans DBAL 3.10 et 4.4, lu dans leur code et non dans leur
 * documentation. Un type PHP qui ne correspond pas à l'hydratation produit une
 * entité que Doctrine charge, puis croit modifiée à chaque flush.
 *
 * Deux types changent avec DBAL 4 : bigint, rendu en chaîne par DBAL 2 et 3 et
 * en entier par DBAL 4, et binary, rendu en ressource par DBAL 2 et 3 et en
 * chaîne par DBAL 4. La table s'indexe par DBAL et non par ORM : ORM 3 accepte
 * DBAL 3.8, sous lequel binary en chaîne empêchait de charger l'entité
 * (TypeError) et bigint en entier la faisait croire modifiée (essai du
 * 2026-09-15, ORM 3.7.1 avec DBAL 3.10.6, et ORM 2.14.3 avec DBAL 2.13.9, qui
 * hydrate comme DBAL 3).
 *
 * Le champ type_php du calque n'est lu que pour un type absent de cette table,
 * typiquement un type personnalisé forcé par décision : seul son auteur sait ce
 * qu'il hydrate, et le calque porte au moins ce que la colonne suggérait.
 */
final class TypesPhp
{
    /**
     * Type PHP par type Doctrine, sous DBAL 2 ou 3 (clé 3) et sous DBAL 4
     * (clé 4).
     *
     * mixed pour blob, et pour binary avant DBAL 4 : une ressource n'a pas de
     * type déclarable. int pour bigint sous DBAL 4, qui rend une chaîne au-delà
     * de PHP_INT_MAX — un BIGINT UNSIGNED de MySQL, rare en pratique et qu'un
     * type forcé par décision traite. array pour json, comme l'écrit
     * make:entity : DBAL rend ce que json_decode rend, et une colonne JSON de
     * base reprise porte des objets.
     */
    private const TABLE = [
        'integer' => [3 => 'int', 4 => 'int'],
        'smallint' => [3 => 'int', 4 => 'int'],
        'bigint' => [3 => 'string', 4 => 'int'],
        'float' => [3 => 'float', 4 => 'float'],
        'smallfloat' => [3 => 'float', 4 => 'float'],
        'boolean' => [3 => 'bool', 4 => 'bool'],
        'decimal' => [3 => 'string', 4 => 'string'],
        'string' => [3 => 'string', 4 => 'string'],
        'text' => [3 => 'string', 4 => 'string'],
        'guid' => [3 => 'string', 4 => 'string'],
        'blob' => [3 => 'mixed', 4 => 'mixed'],
        'binary' => [3 => 'mixed', 4 => 'string'],
        'json' => [3 => 'array', 4 => 'array'],
        'jsonb' => [3 => 'array', 4 => 'array'],
        'simple_array' => [3 => 'array', 4 => 'array'],
        'date_immutable' => [3 => '\DateTimeImmutable', 4 => '\DateTimeImmutable'],
        'time_immutable' => [3 => '\DateTimeImmutable', 4 => '\DateTimeImmutable'],
        'datetime_immutable' => [3 => '\DateTimeImmutable', 4 => '\DateTimeImmutable'],
        'datetimetz_immutable' => [3 => '\DateTimeImmutable', 4 => '\DateTimeImmutable'],
        'dateinterval' => [3 => '\DateInterval', 4 => '\DateInterval'],
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
        return self::TABLE[$propriete->typeDoctrine][$cible->hydrateCommeDbal4() ? 4 : 3] ?? ltrim($propriete->typePhp, '?');
    }
}
