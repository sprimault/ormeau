<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\Fixtures;

use Ormeau\Doctrine\OrmeauDoctrineBundle;
use Symfony\Bundle\FrameworkBundle\FrameworkBundle;
use Symfony\Bundle\FrameworkBundle\Kernel\MicroKernelTrait;
use Symfony\Component\Config\Loader\LoaderInterface;
use Symfony\Component\DependencyInjection\ContainerBuilder;
use Symfony\Component\HttpKernel\Bundle\BundleInterface;
use Symfony\Component\HttpKernel\Kernel;

/**
 * Le plus petit noyau qui charge le bundle.
 *
 * FrameworkBundle est indispensable : c'est lui qui apporte l'enregistrement
 * des commandes de console. Sans lui, le test passerait sans rien prouver.
 *
 * Le cache va dans le répertoire temporaire du système et non dans var/ : la
 * suite doit pouvoir tourner sur un dépôt en lecture seule, et ne rien laisser
 * derrière elle.
 */
final class KernelDeTest extends Kernel
{
    use MicroKernelTrait;

    /**
     * @return iterable<BundleInterface>
     */
    public function registerBundles(): iterable
    {
        yield new FrameworkBundle();
        yield new OrmeauDoctrineBundle();
    }

    /**
     * Un cache par variante d'environnement, sinon deux exécutions successives
     * relisent un conteneur compilé pour l'autre.
     */
    public function getCacheDir(): string
    {
        return sys_get_temp_dir() . '/ormeau-doctrine-test/' . $this->environment . '/cache';
    }

    /**
     * Les journaux vont au même endroit que le cache, et disparaissent avec lui.
     */
    public function getLogDir(): string
    {
        return sys_get_temp_dir() . '/ormeau-doctrine-test/' . $this->environment . '/log';
    }

    /**
     * Configure le strict minimum pour que FrameworkBundle démarre.
     */
    protected function configureContainer(ContainerBuilder $conteneur, LoaderInterface $chargeur): void
    {
        // php_errors entièrement éteint : sous Symfony 7, le laisser actif
        // installe un gestionnaire d'exceptions global que l'arrêt du noyau ne
        // retire pas, et PHPUnit signale le test comme à risque — il fuiterait
        // sur les suivants. Rien de ce qu'on vérifie ici n'en dépend.
        $conteneur->loadFromExtension('framework', [
            'test' => true,
            'secret' => 'ormeau-test',
            'http_method_override' => false,
            'php_errors' => ['log' => false, 'throw' => false],
        ]);
    }
}
