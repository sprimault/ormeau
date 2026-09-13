<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine;

use Symfony\Component\HttpKernel\Bundle\Bundle;

/**
 * Enregistre les deux commandes dans l'application hôte.
 *
 * Bundle et non AbstractBundle, qui n'existe qu'à partir de Symfony 6.1 : le
 * paquet vise aussi les applications en 5.4, celles qu'on reprend. Le prix est
 * une Extension à part, trouvée par convention de nommage dans
 * DependencyInjection/.
 *
 * Aucune recette Flex à publier : le type symfony-bundle du composer.json suffit
 * à Flex pour en générer une et inscrire la classe dans config/bundles.php.
 * C'est ce type, et lui seul, qui rend le paquet installable d'un seul
 * composer require.
 */
final class OrmeauDoctrineBundle extends Bundle {}
