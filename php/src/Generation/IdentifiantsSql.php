<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

/**
 * Écrit un nom de schéma, de table ou de colonne comme Doctrine doit le
 * recevoir pour que la base le retrouve.
 *
 * Doctrine ne cite un identifiant que s'il commence par un backtick, et ne le
 * cite que là : un nom qui n'est pas un identifiant SQL simple en minuscules
 * (T_CLIENTS, « N° Commande ») serait replié ou refusé par la base, un mot
 * réservé (order, user) refusé. Les autres restent nus, comme on les écrirait à
 * la main.
 *
 * Le calque logique ne porte pas le SGBD : la liste des mots réservés est
 * celle des SGBD dont Ormeau extrait les calques, PostgreSQL seul aujourd'hui.
 * Citer un nom exact qui n'en avait pas besoin ne casse rien — essai A2b, sous
 * ORM 2.14 et 3.6 contre PostgreSQL 17 —, l'union des listes suffira donc
 * quand un autre SGBD arrivera.
 *
 * La liste vit ici et non dans DBAL : le code produit ne doit pas changer
 * parce qu'une dépendance s'est mise à jour.
 *
 * @internal
 */
final class IdentifiantsSql
{
    /**
     * Mots que PostgreSQL refuse comme nom de schéma, de table ou de colonne
     * sans guillemets : catégories R (réservés) et T (noms de types et de
     * fonctions) de pg_get_keywords(). La catégorie C s'accepte comme nom, et U
     * n'est pas réservée.
     *
     * Relevé le 2026-09-14 sous PostgreSQL 17.11 :
     *
     *   SELECT catcode, string_agg(word, ' ' ORDER BY word)
     *   FROM pg_get_keywords() WHERE catcode IN ('R', 'T') GROUP BY catcode;
     */
    public const RESERVES_POSTGRESQL = [
        'all', 'analyse', 'analyze', 'and', 'any', 'array', 'as', 'asc', 'asymmetric', 'authorization', 'binary',
        'both', 'case', 'cast', 'check', 'collate', 'collation', 'column', 'concurrently', 'constraint', 'create',
        'cross', 'current_catalog', 'current_date', 'current_role', 'current_schema', 'current_time',
        'current_timestamp', 'current_user', 'default', 'deferrable', 'desc', 'distinct', 'do', 'else', 'end',
        'except', 'false', 'fetch', 'for', 'foreign', 'freeze', 'from', 'full', 'grant', 'group', 'having', 'ilike',
        'in', 'initially', 'inner', 'intersect', 'into', 'is', 'isnull', 'join', 'lateral', 'leading', 'left',
        'like', 'limit', 'localtime', 'localtimestamp', 'natural', 'not', 'notnull', 'null', 'offset', 'on', 'only',
        'or', 'order', 'outer', 'overlaps', 'placing', 'primary', 'references', 'returning', 'right', 'select',
        'session_user', 'similar', 'some', 'symmetric', 'system_user', 'table', 'tablesample', 'then', 'to',
        'trailing', 'true', 'union', 'unique', 'user', 'using', 'variadic', 'verbose', 'when', 'where', 'window',
        'with',
    ];

    /**
     * Rend un nom de colonne, cité s'il le faut.
     *
     * Une colonne apparaît sans préfixe dans la liste d'un INSERT, et un mot
     * réservé y est refusé : elle se cite dès que sa forme ou son sens
     * l'exige.
     */
    public static function colonne(string $nom): string
    {
        return self::aCiter($nom) ? self::citer($nom) : $nom;
    }

    /**
     * Rend les arguments name et schema de #[ORM\Table] ou #[ORM\JoinTable].
     *
     * Le schéma reste nu, et c'est le nom qui porte les backticks : Doctrine ne
     * pose son drapeau de citation que sur le nom, puis cite le schéma avec lui
     * (DefaultQuoteStrategy::getTableName). Des backticks sur le schéma
     * seraient cités une seconde fois, défaut de la 0.5.0.
     *
     * Le nom se cite quand sa forme l'exige, quand le schéma écrit l'exige, ou
     * quand il est réservé sans schéma devant lui. Réservé derrière un schéma
     * ordinaire, il reste nu : PostgreSQL accepte tout mot après un point
     * (public.left), la 0.5.0 l'écrivait ainsi, et le citer ferait signaler
     * une divergence sur une classe de l'utilisateur qui fonctionnait.
     *
     * @param string      $nom    nom de la table, tel que le catalogue le rend
     * @param string|null $schema schéma écrit dans l'attribut, null quand il ne l'est pas
     *
     * @return array{name: string, schema: string|null}
     */
    public static function table(string $nom, ?string $schema): array
    {
        $citer = self::forme($nom) || ($schema === null ? self::reserve($nom) : self::aCiter($schema));

        return ['name' => $citer ? self::citer($nom) : $nom, 'schema' => $schema];
    }

    /**
     * Dit si un schéma exige d'être cité, forme ou mot réservé.
     *
     * Sert à la table de jointure : Doctrine y colle le schéma sans jamais le
     * citer (DefaultQuoteStrategy::getJoinTableName), et un tel schéma n'a
     * aucune forme portable.
     */
    public static function schemaACiter(string $schema): bool
    {
        return self::aCiter($schema);
    }

    /**
     * Dit si un nom exige d'être cité, où qu'il soit écrit.
     */
    private static function aCiter(string $nom): bool
    {
        return self::forme($nom) || self::reserve($nom);
    }

    /**
     * Dit si la forme du nom sort d'un identifiant SQL simple en minuscules.
     */
    private static function forme(string $nom): bool
    {
        return preg_match('/^[a-z_][a-z0-9_]*$/', $nom) !== 1;
    }

    /**
     * Dit si le nom est un mot réservé. La comparaison est exacte : ORDER
     * sort déjà par la forme.
     */
    private static function reserve(string $nom): bool
    {
        return in_array($nom, self::RESERVES_POSTGRESQL, true);
    }

    /**
     * Entoure un nom de backticks, la marque de citation de Doctrine.
     */
    private static function citer(string $nom): string
    {
        return '`' . $nom . '`';
    }
}
