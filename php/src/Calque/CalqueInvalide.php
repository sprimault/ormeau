<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

use RuntimeException;

/**
 * Un calque illisible, incomplet, ou dans une version que ce paquet ne connaît
 * pas.
 *
 * Un JSON mal formé n'en est pas une : il remonte en JsonException, dont le
 * message situe l'erreur mieux que ce qu'on écrirait à sa place. Qui entoure la
 * lecture attrape donc les deux.
 *
 * Dans tous les cas, la génération s'arrête, elle ne poursuit pas sur un calque
 * partiel. Un calque est produit par un outil, pas saisi à la main — s'il est
 * invalide, quelque chose est cassé en amont et le signaler vaut mieux que de
 * générer à moitié.
 */
final class CalqueInvalide extends RuntimeException {}
