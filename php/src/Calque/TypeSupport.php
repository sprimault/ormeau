<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * Type sur lequel s'adosse une énumération PHP. Les valeurs sont celles que
 * PHP attend après « enum Nom: ».
 */
enum TypeSupport: string
{
    /** Énumération adossée à des chaînes. */
    case Chaine = 'string';

    /** Énumération adossée à des entiers. */
    case Entier = 'int';
}
