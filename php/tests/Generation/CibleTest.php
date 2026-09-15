<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\Generation;

use Composer\InstalledVersions;
use InvalidArgumentException;
use Ormeau\Doctrine\Generation\Cible;
use PHPUnit\Framework\Attributes\CoversClass;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;

/**
 * La cible dit pour quelles versions d'ORM et de DBAL le code est produit, et
 * l'annonce.
 */
#[CoversClass(Cible::class)]
final class CibleTest extends TestCase
{
    /**
     * La détection lit les versions installées : ce sont elles qui hydrateront
     * les entités, pas des versions supposées.
     */
    public function testDetecteLesVersionsInstallees(): void
    {
        [$orm, $ormMineure] = explode('.', (string) InstalledVersions::getVersion('doctrine/orm'));
        [$dbal, $dbalMineure] = explode('.', (string) InstalledVersions::getVersion('doctrine/dbal'));

        $cible = Cible::detecter();

        self::assertSame((int) $orm, $cible->ormMajeure);
        self::assertSame((int) $dbal, $cible->dbalMajeure);
        self::assertFalse($cible->forcee);
        self::assertSame(
            sprintf('Cible détectée : PHP %d.%d, Doctrine ORM %s.%s, DBAL %s.%s', PHP_MAJOR_VERSION, PHP_MINOR_VERSION, $orm, $ormMineure, $dbal, $dbalMineure),
            $cible->annonce(),
        );
    }

    /**
     * Une majeure d'ORM forcée suppose la DBAL la plus récente qu'elle
     * accepte, et l'annonce la dit déduite : un diff produit pour une version
     * qui n'est pas installée ne doit pas passer pour la version courante.
     */
    public function testUneCibleForceeDeduitDbalEtLeDit(): void
    {
        self::assertSame('Cible forcée : Doctrine ORM 2, DBAL 3.10 (déduite)', Cible::forcer(2)->annonce());
        self::assertSame('Cible forcée : Doctrine ORM 3, DBAL 4.4 (déduite)', Cible::forcer(3)->annonce());
    }

    /**
     * Une version de DBAL forcée l'emporte sur la déduction comme sur la
     * détection, et se dit forcée.
     */
    public function testUneVersionDeDbalForceeSeDitForcee(): void
    {
        self::assertSame('Cible forcée : Doctrine ORM 3, DBAL 3.10 (forcée)', Cible::forcer(3, '3.10')->annonce());

        $detectee = Cible::detecter(Cible::detecter()->dbalMajeure . '.0');
        self::assertStringEndsWith(sprintf('DBAL %d.0 (forcée)', $detectee->dbalMajeure), $detectee->annonce());
    }

    /**
     * L'hydratation de bigint et binary change avec DBAL 4, pas avec ORM 3 :
     * ORM 3 accepte DBAL 3.8, et DBAL 2 hydrate comme DBAL 3.
     */
    #[DataProvider('combinaisons')]
    public function testHydrateCommeDbal4SuitDbal(int $orm, string $dbal, bool $attendu): void
    {
        self::assertSame($attendu, Cible::forcer($orm, $dbal)->hydrateCommeDbal4());
    }

    /**
     * Les combinaisons que Composer résout.
     *
     * @return iterable<string, array{int, string, bool}>
     */
    public static function combinaisons(): iterable
    {
        yield 'ORM 2, DBAL 2' => [2, '2.13', false];
        yield 'ORM 2, DBAL 3' => [2, '3.10', false];
        yield 'ORM 3, DBAL 3' => [3, '3.10', false];
        yield 'ORM 3, DBAL 4' => [3, '4.0', true];
    }

    /**
     * L'objet qui décrit un défaut calculé n'existe qu'à partir de DBAL 4.4 ;
     * ORM 3 accepte DBAL 4.0 à 4.3, où seule la chaîne est lue.
     */
    public function testLeDefautParExpressionAttendDbal44(): void
    {
        self::assertFalse(Cible::forcer(2)->defautParExpression());
        self::assertFalse(Cible::forcer(3, '3.10')->defautParExpression());
        self::assertFalse(Cible::forcer(3, '4.3')->defautParExpression());
        self::assertTrue(Cible::forcer(3, '4.4')->defautParExpression());
        self::assertTrue(Cible::forcer(3)->defautParExpression());
    }

    /**
     * Une majeure qu'aucun rendu ne connaît est refusée, y compris une
     * version future : on ne promet pas ce qu'on n'a pas vu.
     */
    public function testRefuseUneMajeureInconnue(): void
    {
        $this->expectException(InvalidArgumentException::class);
        $this->expectExceptionMessage('Doctrine ORM 4 n\'est pas une cible connue (2, 3).');

        Cible::forcer(4);
    }

    /**
     * Une combinaison que Composer ne résout pas est refusée : générer pour
     * elle promettrait un environnement qui n'existe pas.
     */
    #[DataProvider('combinaisonsImpossibles')]
    public function testRefuseUneCombinaisonImpossible(int $orm, string $dbal, string $message): void
    {
        $this->expectException(InvalidArgumentException::class);
        $this->expectExceptionMessage($message);

        Cible::forcer($orm, $dbal);
    }

    /**
     * Combinaisons refusées, et version illisible.
     *
     * @return iterable<string, array{int, string, string}>
     */
    public static function combinaisonsImpossibles(): iterable
    {
        yield 'ORM 2, DBAL 4' => [2, '4.4', 'Doctrine ORM 2 ne s\'installe pas avec DBAL 4 (DBAL 2, 3).'];
        yield 'ORM 3, DBAL 2' => [3, '2.13', 'Doctrine ORM 3 ne s\'installe pas avec DBAL 2 (DBAL 3, 4).'];
        yield 'DBAL 5' => [3, '5.0', 'Doctrine ORM 3 ne s\'installe pas avec DBAL 5 (DBAL 3, 4).'];
        yield 'majeure seule' => [3, '4', 'Version de DBAL illisible, « majeure.mineure » attendu : « 4 ».'];
    }
}
