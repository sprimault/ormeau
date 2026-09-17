<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

/**
 * Dit si le schéma d'une table s'écrit dans #[ORM\Table] ou #[ORM\JoinTable].
 *
 * Il s'écrit, sauf le schéma par défaut de la plateforme. Sans lui, une table
 * rangée hors de ce schéma n'est trouvée que si l'utilisateur de connexion y a
 * son schéma par défaut : ailleurs, l'application échoue et schema:update
 * propose de créer la table dans le schéma par défaut puis de supprimer
 * l'originale. Écrit pour le schéma par défaut, il fait proposer à DBAL 3 la
 * suppression de la table suivie de sa recréation, et schema:create échoue sur
 * CREATE SCHEMA public ; DBAL 4 ne propose rien. Constaté sous ORM 2.14 /
 * DBAL 3.10 et ORM 3.7 / DBAL 4.4, contre PostgreSQL 17 et SQL Server 2022.
 *
 * Le schéma par défaut est celui que DBAL 3 déclarait par plateforme
 * (AbstractPlatform::getDefaultSchemaName, retiré en DBAL 4) : public, dbo.
 * Sous MySQL et MariaDB, le schéma est la base elle-même, et l'écrire
 * attacherait les entités au nom d'une base qui change d'un environnement à
 * l'autre. La comparaison est exacte : Public n'est pas public.
 *
 * @internal
 */
final class SchemaParDefaut
{
    /**
     * Rend le schéma à écrire, ou null quand il ne s'écrit pas.
     *
     * @param string $sgbd   SGBD du calque logique
     * @param string $schema schéma de la table, tel que le catalogue le rend
     */
    public static function ecrit(string $sgbd, string $schema): ?string
    {
        $implicite = match ($sgbd) {
            'postgres' => $schema === 'public',
            'sqlserver' => $schema === 'dbo',
            'mysql', 'mariadb' => true,
            // Aucun autre dialecte n'a de pilote : sans schéma écrit, la table
            // ne serait trouvée que par chance.
            default => false,
        };

        return $implicite ? null : $schema;
    }
}
