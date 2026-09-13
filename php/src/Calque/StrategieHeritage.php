<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * Projection d'une hiérarchie sur les tables. La table par classe concrète
 * manque volontairement : rien dans un schéma ne permet de la déduire.
 */
enum StrategieHeritage: string
{
    /** Une table par classe, liées par une clé primaire étrangère. */
    case Jointe = 'jointe';

    /** Toute la hiérarchie dans une table, départagée par une colonne. */
    case TableUnique = 'table_unique';
}
