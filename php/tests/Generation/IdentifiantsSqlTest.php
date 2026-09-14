<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\Generation;

use Ormeau\Doctrine\Generation\IdentifiantsSql;
use PHPUnit\Framework\Attributes\CoversClass;
use PHPUnit\Framework\TestCase;

/**
 * Les formes que l'essai A2b a validées contre PostgreSQL 17, sous ORM 2.14 et
 * 3.6 : ce qui se cite, et surtout ce qui reste tel que la 0.5.0 l'écrivait.
 */
#[CoversClass(IdentifiantsSql::class)]
final class IdentifiantsSqlTest extends TestCase
{
    /**
     * Une colonne se cite dès que sa forme ou son sens l'exige : elle apparaît
     * sans préfixe dans un INSERT, où un mot réservé est refusé.
     */
    public function testUneColonneSeCiteParSaFormeOuSonSens(): void
    {
        self::assertSame('client_id', IdentifiantsSql::colonne('client_id'));
        self::assertSame('`T_CLIENTS`', IdentifiantsSql::colonne('T_CLIENTS'));
        self::assertSame('`N° Commande`', IdentifiantsSql::colonne('N° Commande'));
        self::assertSame('`1er`', IdentifiantsSql::colonne('1er'));
        self::assertSame('`order`', IdentifiantsSql::colonne('order'));
        self::assertSame('`left`', IdentifiantsSql::colonne('left'));
        self::assertSame('between', IdentifiantsSql::colonne('between'), 'catégorie C, acceptée comme nom');
        self::assertSame('name', IdentifiantsSql::colonne('name'), 'non réservé');
    }

    /**
     * Le schéma reste nu et le nom porte les backticks : Doctrine cite alors
     * les deux. Un nom qui n'exige rien derrière un schéma ordinaire reste tel
     * que la 0.5.0 l'écrivait, mot réservé compris.
     */
    public function testUneTableCiteSonNomQuandLeSchemaOuLeNomLExige(): void
    {
        self::assertSame(['name' => 'facture', 'schema' => 'public'], IdentifiantsSql::table('facture', 'public'));
        self::assertSame(['name' => 'facture', 'schema' => null], IdentifiantsSql::table('facture', null));
        self::assertSame(['name' => '`Bon_Livraison`', 'schema' => 'Compta'], IdentifiantsSql::table('Bon_Livraison', 'Compta'));
        self::assertSame(['name' => '`facture`', 'schema' => 'Compta'], IdentifiantsSql::table('facture', 'Compta'));
        self::assertSame(['name' => '`Majuscule`', 'schema' => 'gescom'], IdentifiantsSql::table('Majuscule', 'gescom'));
        self::assertSame(['name' => '`pays`', 'schema' => 'référentiel'], IdentifiantsSql::table('pays', 'référentiel'));
        self::assertSame(['name' => '`user`', 'schema' => 'order'], IdentifiantsSql::table('user', 'order'));
        self::assertSame(['name' => 'left', 'schema' => 'public'], IdentifiantsSql::table('left', 'public'));
        self::assertSame(['name' => '`user`', 'schema' => null], IdentifiantsSql::table('user', null));
    }

    /**
     * Un schéma de table de jointure s'écrit sans citation possible : ceux qui
     * en exigent une sont signalés, par leur forme comme par leur sens.
     */
    public function testUnSchemaACiterSeReconnait(): void
    {
        self::assertFalse(IdentifiantsSql::schemaACiter('public'));
        self::assertTrue(IdentifiantsSql::schemaACiter('Compta'));
        self::assertTrue(IdentifiantsSql::schemaACiter('order'));
    }
}
