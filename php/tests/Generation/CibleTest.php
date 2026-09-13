<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\Generation;

use Composer\InstalledVersions;
use InvalidArgumentException;
use Ormeau\Doctrine\Generation\Cible;
use PHPUnit\Framework\Attributes\CoversClass;
use PHPUnit\Framework\TestCase;

/**
 * La cible dit pour quelle version d'ORM le code est produit, et l'annonce.
 */
#[CoversClass(Cible::class)]
final class CibleTest extends TestCase
{
    /**
     * La détection lit la version installée : c'est celle qui hydratera les
     * entités, pas une version supposée.
     */
    public function testDetecteLaVersionInstallee(): void
    {
        $version = (string) InstalledVersions::getVersion('doctrine/orm');
        [$majeure, $mineure] = explode('.', $version);

        $cible = Cible::detecter();

        self::assertSame((int) $majeure, $cible->ormMajeure);
        self::assertFalse($cible->forcee);
        self::assertSame(
            sprintf('Cible détectée : PHP %d.%d, Doctrine ORM %s.%s', PHP_MAJOR_VERSION, PHP_MINOR_VERSION, $majeure, $mineure),
            $cible->annonce(),
        );
    }

    /**
     * Une cible forcée se dit forcée : un diff produit pour une version qui
     * n'est pas installée ne doit pas passer pour la version courante.
     */
    public function testUneCibleForceeSeDitForcee(): void
    {
        self::assertSame('Cible forcée : Doctrine ORM 2', Cible::forcer(2)->annonce());
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
}
