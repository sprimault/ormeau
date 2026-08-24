<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine;

use Symfony\Component\DependencyInjection\ContainerBuilder;
use Symfony\Component\DependencyInjection\Loader\Configurator\ContainerConfigurator;
use Symfony\Component\HttpKernel\Bundle\AbstractBundle;

/**
 * Enregistre les deux commandes dans l'application hôte.
 *
 * AbstractBundle plutôt qu'un Bundle avec son répertoire DependencyInjection :
 * il n'y a rien à configurer ici — pas de connexion, pas de chemin, pas de
 * seuil —, et une Extension vide aurait été trois fichiers pour rien.
 *
 * Aucune recette Flex à publier : le type symfony-bundle du composer.json suffit
 * à Flex pour en générer une et inscrire la classe dans config/bundles.php.
 * C'est ce type, et lui seul, qui rend le paquet installable d'un seul
 * composer require.
 */
final class OrmeauDoctrineBundle extends AbstractBundle
{
    /**
     * @param array<string, mixed> $config
     */
    public function loadExtension(array $config, ContainerConfigurator $container, ContainerBuilder $builder): void
    {
        $container->import(__DIR__ . '/../config/services.php');
    }
}
