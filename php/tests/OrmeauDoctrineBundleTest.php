<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests;

use Ormeau\Doctrine\Calque\LecteurCalque;
use Ormeau\Doctrine\Generation\GenerateurEntite;
use Ormeau\Doctrine\OrmeauDoctrineBundle;
use Ormeau\Doctrine\Tests\Fixtures\KernelDeTest;
use PHPUnit\Framework\Attributes\CoversClass;
use Psr\Container\ContainerInterface;
use Symfony\Bundle\FrameworkBundle\Console\Application;
use Symfony\Bundle\FrameworkBundle\Test\KernelTestCase;

/**
 * Le seul test qui répond à « le paquet s'installe-t-il ».
 *
 * Un bundle peut être parfaitement écrit et n'enregistrer aucune commande : le
 * conteneur se compile, l'application démarre, et bin/console ne montre rien.
 * Rien d'autre que le démarrage réel d'un noyau ne l'attrape.
 *
 * KernelTestCase plutôt qu'un noyau instancié à la main : Symfony 7 installe au
 * démarrage un gestionnaire d'exceptions qu'il ne retire pas, et PHPUnit
 * signale à juste titre le test comme à risque — il fuiterait sur les suivants.
 * C'est cette classe qui sait démonter ce qu'un noyau laisse derrière lui, et
 * s'appuyer dessus évite d'avoir à suivre ce que chaque version installe.
 */
#[CoversClass(OrmeauDoctrineBundle::class)]
final class OrmeauDoctrineBundleTest extends KernelTestCase
{
    /** Gestionnaire d'exceptions en place avant le démarrage du noyau. */
    private mixed $gestionnaireInitial = null;

    /**
     * Désigne le noyau minimal du répertoire de fixtures.
     *
     * @return class-string<KernelDeTest>
     */
    protected static function getKernelClass(): string
    {
        return KernelDeTest::class;
    }

    /**
     * Retient le gestionnaire d'exceptions en place avant le test.
     *
     * Le couple set/restore lit la valeur courante sans la modifier : il n'y a
     * pas d'accesseur pour ça en PHP.
     */
    protected function setUp(): void
    {
        parent::setUp();

        $this->gestionnaireInitial = set_exception_handler(null);
        restore_exception_handler();
    }

    /**
     * Remet la pile des gestionnaires dans l'état où le test l'a trouvée.
     *
     * Symfony 7 installe un gestionnaire d'exceptions au démarrage du noyau et
     * ne le retire pas à l'arrêt, ni de lui-même ni via KernelTestCase. PHPUnit
     * le signale comme un test à risque, à juste titre : laissé en place, il
     * intercepterait les exceptions des tests suivants.
     *
     * Le dépilage s'arrête sur le gestionnaire d'origine plutôt que de compter
     * les appels : rien ne garantit combien de couches une version donnée
     * ajoute, et en retirer une de trop emporterait celui de PHPUnit.
     */
    protected function tearDown(): void
    {
        parent::tearDown();

        for ($garde = 0; $garde < 16; ++$garde) {
            $courant = set_exception_handler(null);
            restore_exception_handler();

            if ($courant === $this->gestionnaireInitial) {
                return;
            }
            restore_exception_handler();
        }
    }

    /**
     * Les deux commandes apparaissent dans la console de l'application hôte.
     */
    public function testLesCommandesSontEnregistrees(): void
    {
        $application = new Application(self::bootKernel());

        self::assertTrue($application->has('ormeau:generer'));
        self::assertTrue($application->has('ormeau:synchroniser'));
    }

    /**
     * Les descriptions viennent de l'attribut AsCommand : vides, elles
     * signaleraient que l'autoconfiguration ne s'applique pas.
     */
    public function testLesCommandesPortentLeurDescription(): void
    {
        $application = new Application(self::bootKernel());

        self::assertStringContainsString(
            'entités Doctrine',
            $application->find('ormeau:generer')->getDescription(),
        );
        self::assertStringContainsString(
            'calque physique',
            $application->find('ormeau:synchroniser')->getDescription(),
        );
    }

    /**
     * Le lecteur et le générateur sont injectables.
     *
     * Le générateur a un paramètre d'énumération avec valeur par défaut, que
     * l'autowiring doit laisser tel quel plutôt qu'échouer à le résoudre.
     */
    public function testLesServicesSontConstructibles(): void
    {
        self::bootKernel();
        $conteneur = self::getContainer();
        self::assertInstanceOf(ContainerInterface::class, $conteneur);

        self::assertInstanceOf(LecteurCalque::class, $conteneur->get(LecteurCalque::class));
        self::assertInstanceOf(GenerateurEntite::class, $conteneur->get(GenerateurEntite::class));
    }
}
