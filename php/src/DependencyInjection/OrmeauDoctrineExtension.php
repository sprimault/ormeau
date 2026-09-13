<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\DependencyInjection;

use Symfony\Component\Config\FileLocator;
use Symfony\Component\DependencyInjection\ContainerBuilder;
use Symfony\Component\DependencyInjection\Extension\Extension;
use Symfony\Component\DependencyInjection\Loader\PhpFileLoader;

/**
 * Charge les services du paquet. Il n'y a rien à configurer — ni connexion, ni
 * chemin, ni seuil —, d'où l'absence de classe Configuration.
 *
 * L'Extension du composant DependencyInjection, pas celle de HttpKernel :
 * marquée interne en 7.1 et dépréciée en 8.1, elle romprait le haut de la
 * plage de versions. Celle-ci existe telle quelle depuis la 5.4.
 */
final class OrmeauDoctrineExtension extends Extension
{
    /**
     * @param array<array<string, mixed>> $configs
     */
    public function load(array $configs, ContainerBuilder $container): void
    {
        $chargeur = new PhpFileLoader($container, new FileLocator(__DIR__ . '/../../config'));
        $chargeur->load('services.php');
    }
}
