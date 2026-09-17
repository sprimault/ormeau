<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\Generation;

use Ormeau\Doctrine\Generation\SchemaParDefaut;
use PHPUnit\Framework\Attributes\CoversClass;
use PHPUnit\Framework\TestCase;

/**
 * Quel schéma s'écrit dans #[ORM\Table], par SGBD.
 */
#[CoversClass(SchemaParDefaut::class)]
final class SchemaParDefautTest extends TestCase
{
    /**
     * Le schéma par défaut de la plateforme ne s'écrit pas, tout autre si ;
     * la comparaison est exacte, Public n'est pas public.
     */
    public function testSeulLeSchemaParDefautSeTait(): void
    {
        self::assertNull(SchemaParDefaut::ecrit('postgres', 'public'));
        self::assertSame('ventes', SchemaParDefaut::ecrit('postgres', 'ventes'));
        self::assertSame('Public', SchemaParDefaut::ecrit('postgres', 'Public'));
        self::assertNull(SchemaParDefaut::ecrit('sqlserver', 'dbo'));
        self::assertSame('ventes', SchemaParDefaut::ecrit('sqlserver', 'ventes'));
        self::assertSame('public', SchemaParDefaut::ecrit('sqlserver', 'public'), 'public n\'est pas le défaut de SQL Server');
    }

    /**
     * Sous MySQL et MariaDB, le schéma est la base : il ne s'écrit jamais.
     */
    public function testLeSchemaDUneBaseMySqlNeSEcritPas(): void
    {
        self::assertNull(SchemaParDefaut::ecrit('mariadb', 'achats'));
        self::assertNull(SchemaParDefaut::ecrit('mysql', 'achats'));
    }
}
