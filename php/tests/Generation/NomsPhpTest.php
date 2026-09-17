<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\Generation;

use Ormeau\Doctrine\Generation\NomsPhp;
use PHPUnit\Framework\Attributes\CoversClass;
use PHPUnit\Framework\TestCase;
use RuntimeException;

/**
 * Les listes de mots réservés disent ce que PHP refuse, et pas ce que la
 * documentation en dit : chaque mot est repassé à php -l, sous la version de
 * PHP qui exécute les tests. Chaque job de compatibilité vérifie donc la liste
 * contre sa propre version.
 */
#[CoversClass(NomsPhp::class)]
final class NomsPhpTest extends TestCase
{
    /**
     * PHP refuse chaque mot de la liste pour une classe, et accepte les mots
     * que la documentation dit réservés sans qu'il les refuse. __property__ ne
     * l'est qu'à partir de la 8.4 : sous une version antérieure, il passe.
     */
    public function testChaqueMotReserveDeClasseEstRefuseParPhp(): void
    {
        foreach (NomsPhp::RESERVES_DE_CLASSE as $mot) {
            if ($mot === '__property__' && PHP_VERSION_ID < 80400) {
                continue;
            }
            self::assertFalse(self::compile('class ' . ucfirst($mot) . ' {}'), $mot);
        }
        foreach (['Enum', 'Resource', 'Numeric', 'Client'] as $nom) {
            self::assertTrue(self::compile('class ' . $nom . ' {}'), $nom);
            self::assertNull(NomsPhp::raisonClasse($nom), $nom);
        }
    }

    /**
     * Un cas d'énumération et un segment d'espace de noms acceptent les mots
     * réservés, sauf les deux de leur liste.
     */
    public function testUnCasEtUnEspaceDeNomsAcceptentPresqueTout(): void
    {
        foreach (NomsPhp::RESERVES_DE_CAS as $mot) {
            self::assertFalse(self::compile("enum E: string { case {$mot} = 'a'; }"), $mot);
        }
        self::assertTrue(self::compile("enum E: string { case Match = 'a'; case Default = 'b'; }"));
        self::assertNull(NomsPhp::raisonCas('Match'));

        foreach (NomsPhp::RESERVES_D_ESPACE as $mot) {
            self::assertFalse(self::compile("namespace {$mot};"), $mot);
        }
        self::assertTrue(self::compile('namespace App\\Function;'));
        self::assertNull(NomsPhp::raisonEspaceDeNoms('App\\Function'));
    }

    /**
     * La forme d'un identifiant : un nom accentué passe, rien de ce qui
     * sortirait d'une déclaration ou d'un répertoire.
     */
    public function testLaFormeDUnIdentifiant(): void
    {
        foreach (['CatégoriesDesNoms', '_interne', 'T2'] as $nom) {
            self::assertNull(NomsPhp::raisonMembre($nom), $nom);
        }
        foreach (['', '2024Ventes', '../../x', 'a b', "a\nb", 'a;b', 'App\\Client'] as $nom) {
            self::assertNotNull(NomsPhp::raisonMembre($nom), $nom);
        }
        foreach (['\\App\\Entity', 'App\\Entity\\', 'App\\\\Entity', 'App\\Namespace'] as $espace) {
            self::assertNotNull(NomsPhp::raisonEspaceDeNoms($espace), $espace);
        }
    }

    /**
     * Dit si PHP compile un fragment, par php -l dans un processus à part : un
     * nom réservé est une erreur de compilation fatale, qu'aucun try ne
     * rattrape.
     */
    private static function compile(string $code): bool
    {
        $fichier = tempnam(sys_get_temp_dir(), 'ormeau-noms-');
        if ($fichier === false || file_put_contents($fichier, "<?php\n" . $code . "\n") === false) {
            throw new RuntimeException('Fichier temporaire impossible à écrire.');
        }
        try {
            exec(escapeshellarg(PHP_BINARY) . ' -l ' . escapeshellarg($fichier) . ' 2>&1', $sortie, $retour);

            return $retour === 0;
        } finally {
            unlink($fichier);
        }
    }
}
