<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Symfony\Component\DependencyInjection\Loader\Configurator;

use Ormeau\Doctrine\Calque\LecteurCalque;
use Ormeau\Doctrine\Commande\GenererCommande;
use Ormeau\Doctrine\Commande\SynchroniserCommande;
use Ormeau\Doctrine\Generation\GenerateurEntite;

// Les services sont déclarés un par un plutôt que par chargement de répertoire :
// le paquet en expose quatre, et une déclaration explicite dit exactement ce que
// l'intégrateur récupère dans son conteneur.
//
// Aucun n'est public : ils s'injectent, ils ne se récupèrent pas par $container->get().
// Les commandes font exception par autoconfigure, qui les tague console.command.
return static function (ContainerConfigurator $container): void {
    $services = $container->services()
        ->defaults()
        ->autowire()
        ->autoconfigure();

    $services->set(LecteurCalque::class);
    $services->set(GenerateurEntite::class);

    $services->set(GenererCommande::class);
    $services->set(SynchroniserCommande::class);
};
