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

        self::assertStringStartsWith("Cible forcée : Doctrine ORM 2, DBAL 3.10 (déduite)\n", $sortie);
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
     * --cible-dbal s'écarte de la version déduite, avec --cible-orm comme sans,
     * et l'annonce dit qu'elle est forcée.
     */
    public function testUneVersionDeDbalForcee(): void
    {
        $avecOrm = $this->executer(['--cible-orm' => '3', '--cible-dbal' => '3.10']);
        self::assertStringStartsWith("Cible forcée : Doctrine ORM 3, DBAL 3.10 (forcée)\n", $avecOrm);

        Repertoires::supprimer($this->sortie);
        $this->sortie = Repertoires::creer();
        $seule = $this->executer(['--cible-dbal' => Cible::detecter()->dbalMajeure . '.0']);
        self::assertSame(Cible::detecter(Cible::detecter()->dbalMajeure . '.0')->annonce(), strtok($seule, "\n"));
        self::assertStringEndsWith('.0 (forcée)', (string) strtok($seule, "\n"));
    }

    /**
     * Une version de DBAL illisible, ou que la majeure d'ORM n'accepte pas, est
     * refusée avant toute écriture.
     */
    public function testRefuseUneVersionDeDbalIllisibleOuIncompatible(): void
    {
        foreach ([
            [['--cible-orm' => '3', '--cible-dbal' => '4'], 'Version de DBAL illisible, « majeure.mineure » attendu : « 4 ».'],
            [['--cible-orm' => '2', '--cible-dbal' => '4.4'], 'Doctrine ORM 2 ne s\'installe pas avec DBAL 4 (DBAL 2, 3).'],
        ] as [$options, $message]) {
            try {
                $this->executer($options);
                self::fail('version de DBAL acceptée : ' . json_encode($options));
            } catch (InvalidArgumentException $e) {
                self::assertSame($message, $e->getMessage());
            }
        }

        self::assertSame([], Repertoires::lire($this->sortie));
    }

    /**
     * Un nom venu du calque s'affiche tel quel : une balise dans un nom de
     * table ou d'entité ne fait pas lever le formateur de la console au milieu
     * du compte rendu.
     */
    public function testUnNomDuCalqueNEstPasPrisPourUneBaliseDeLaConsole(): void
    {
        $entree = Repertoires::creer();
        try {
            $calque = $entree . '/balises.logique.json';
            file_put_contents($calque, json_encode([
                'version_ri' => 2,
                'empreinte_physique' => 'sha256:' . str_repeat('b', 64),
                'espace_de_noms' => 'App\\Entity',
                'entites' => [[
                    'nom' => '</info>',
                    'table' => ['nom' => 't', 'schema' => 'public'],
                    'proprietes' => [],
                ]],
                'avertissements' => [[
                    'code' => 'table_sans_cle_primaire',
                    'cible' => 'public.<error></info>',
                    'message' => 'm',
                    'resolution' => 'aucune',
                    'confiance' => 1,
                ]],
            ], JSON_THROW_ON_ERROR));

            $sortie = $this->executer([], $calque);

            self::assertStringContainsString('écartée  </info> : ', $sortie);
            self::assertStringContainsString('  table_sans_cle_primaire public.<error></info> — m', $sortie);
        } finally {
            Repertoires::supprimer($entree);
        }
    }

    /**
     * La base se lit dans le nom du fichier et s'annonce ; un nom hors
     * convention perd seulement son extension.
     */
    public function testLaBaseSeLitDansLeNomDuFichier(): void
    {
        $entree = Repertoires::creer();
        try {
            copy(Repertoires::REFERENCES . '/minimal/logique.json', $entree . '/gescom.v2.logique.json');
            copy(Repertoires::REFERENCES . '/minimal/logique.json', $entree . '/export.json');

            self::assertStringContainsString("\nBase : gescom.v2\n", $this->executer([], $entree . '/gescom.v2.logique.json'));
            self::assertStringContainsString('depuis la base gescom.v2 ', (string) file_get_contents($this->sortie . '/Base/ClientBase.php'));
            self::assertStringContainsString("\nBase : export\n", $this->executer(['--remplacer' => ['gescom.v2']], $entree . '/export.json'));
        } finally {
            Repertoires::supprimer($entree);
        }
    }

    /**
     * Une autre base dans le même répertoire fait échouer la commande, le
     * refus en dernier avec l'option qui le lève ; --remplacer la laisse
     * passer.
     */
    public function testUnRefusDEcrasementFaitEchouerLaCommande(): void
    {
        $entree = Repertoires::creer();
        try {
            copy(Repertoires::REFERENCES . '/minimal/logique.json', $entree . '/gescom.logique.json');
            copy(Repertoires::REFERENCES . '/minimal/logique.json', $entree . '/paie.logique.json');
            $this->executer([], $entree . '/gescom.logique.json');

            $sortie = $this->executer([], $entree . '/paie.logique.json', Command::FAILURE);

            self::assertStringEndsWith(
                sprintf("refusé   %s/Base/ClientBase.php vient de la base gescom, la génération depuis paie ne l'écrit pas : générer chaque base dans son répertoire et son espace de noms, ou remplacer explicitement la base gescom (--remplacer=gescom)\n", $this->sortie),
                $sortie,
            );
            $this->executer(['--remplacer' => ['gescom']], $entree . '/paie.logique.json');
        } finally {
            Repertoires::supprimer($entree);
        }
    }

    /**
     * Un nom de base porteur d'un saut de ligne ferait du reste une ligne de
     * code dans l'en-tête : il est refusé avant toute écriture.
     */
    public function testUnNomDeBaseAvecSautDeLigneEstRefuse(): void
    {
        if (PHP_OS_FAMILY === 'Windows') {
            self::markTestSkipped('un nom de fichier ne porte pas de saut de ligne sous Windows');
        }

        $entree = Repertoires::creer();
        try {
            $calque = $entree . "/gescom\nsystem('id');.logique.json";
            copy(Repertoires::REFERENCES . '/minimal/logique.json', $calque);

            try {
                $this->executer([], $calque);
                self::fail('nom de base accepté');
            } catch (InvalidArgumentException $e) {
                self::assertStringStartsWith('Nom de base refusé, rien n\'est écrit', $e->getMessage());
            }
            self::assertSame([], Repertoires::lire($this->sortie));
        } finally {
            Repertoires::supprimer($entree);
        }
    }

    /**
     * Exécute la commande sur un calque, le cas de référence cas-limites à
     * défaut, vérifie son code de retour et rend sa sortie.
     *
     * @param array<string, mixed> $options
     * @param string               $calque  calque logique à générer
     * @param int                  $attendu code de retour attendu
     */
    private function executer(array $options, string $calque = Repertoires::REFERENCES . '/cas-limites/logique.json', int $attendu = Command::SUCCESS): string
    {
        $testeur = new CommandTester(new GenererCommande(new LecteurCalque(), new GenerateurEntite()));
        $code = $testeur->execute(array_merge([
            'calque' => $calque,
            '--repertoire' => $this->sortie,
        ], $options));

        self::assertSame($attendu, $code, $testeur->getDisplay());

        return $testeur->getDisplay();
    }
}
