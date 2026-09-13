<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\Commande;

use InvalidArgumentException;
use Ormeau\Doctrine\Calque\LecteurCalque;
use Ormeau\Doctrine\Commande\GenererCommande;
use Ormeau\Doctrine\Generation\Cible;
use Ormeau\Doctrine\Generation\GenerateurEntite;
use Ormeau\Doctrine\Tests\Generation\Repertoires;
use PHPUnit\Framework\Attributes\CoversClass;
use PHPUnit\Framework\TestCase;
use Symfony\Component\Console\Command\Command;
use Symfony\Component\Console\Tester\CommandTester;

/**
 * Ce que la commande dit à l'utilisateur : la cible d'abord, puis ce qui a été
 * écrit, écarté, et ce qu'il doit reprendre.
 */
#[CoversClass(GenererCommande::class)]
final class GenererCommandeTest extends TestCase
{
    /** Répertoire de sortie du test en cours. */
    private string $sortie;

    /**
     * Prépare un répertoire de sortie vide.
     */
    protected function setUp(): void
    {
        $this->sortie = Repertoires::creer();
    }

    /**
     * Supprime ce que le test a écrit.
     */
    protected function tearDown(): void
    {
        Repertoires::supprimer($this->sortie);
    }

    /**
     * La première ligne annonce la cible détectée : sans elle, un diff dû à
     * une montée de version d'ORM ressemblerait à un défaut.
     */
    public function testAnnonceLaCibleDetecteeEnPremiereLigne(): void
    {
        $sortie = $this->executer([]);

        self::assertSame(Cible::detecter()->annonce(), strtok($sortie, "\n"));
    }

    /**
     * --cible-orm force la majeure et le dit ; les fichiers écrits, l'entité
     * écartée et les avertissements du calque suivent, sans échec.
     */
    public function testUneCibleForceeEtLeCompteRendu(): void
    {
        $sortie = $this->executer(['--cible-orm' => '2']);

        self::assertStringStartsWith("Cible forcée : Doctrine ORM 2\n", $sortie);
        self::assertStringContainsString('créé     ' . $this->sortie . '/Base/LigneCommandeBase.php', $sortie);
        self::assertStringContainsString('écartée  Journal : la table n\'a pas de clé primaire', $sortie);
        self::assertStringContainsString('avertissement(s) dans le calque :', $sortie);
        self::assertStringContainsString('  table_sans_cle_primaire public.journal — ', $sortie);
    }

    /**
     * Une majeure hors de celles que le rendu connaît est refusée avant toute
     * écriture.
     */
    public function testRefuseUneCibleInconnueAvantDEcrire(): void
    {
        try {
            $this->executer(['--cible-orm' => 'orm3']);
            self::fail('cible acceptée');
        } catch (InvalidArgumentException $e) {
            self::assertSame('--cible-orm attend une majeure de Doctrine ORM (2, 3).', $e->getMessage());
        }

        self::assertSame([], Repertoires::lire($this->sortie));
    }

    /**
     * Exécute la commande sur le cas de référence cas-limites et rend sa
     * sortie.
     *
     * @param array<string, string> $options
     */
    private function executer(array $options): string
    {
        $testeur = new CommandTester(new GenererCommande(new LecteurCalque(), new GenerateurEntite()));
        $code = $testeur->execute(array_merge([
            'calque' => Repertoires::REFERENCES . '/cas-limites/logique.json',
            '--repertoire' => $this->sortie,
        ], $options));

        self::assertSame(Command::SUCCESS, $code);

        return $testeur->getDisplay();
    }
}
