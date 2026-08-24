<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

use RuntimeException;

/**
 * Un calque illisible, mal formé, ou dans une version que ce paquet ne connaît
 * pas.
 *
 * C'est la seule exception que le paquet lève à la lecture : la génération
 * s'arrête, elle ne poursuit pas sur un calque partiel. Un calque est produit
 * par un outil, pas saisi à la main — s'il est invalide, quelque chose est
 * cassé en amont et le signaler vaut mieux que de générer à moitié.
 */
final class CalqueInvalide extends RuntimeException {}
