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

    /**
     * Énumération adossée à des entiers. Aucune version ne la produit encore, la
     * détection ne reconnaissant que des littéraux chaîne : lue et rendue en
     * attendant l'échantillonnage.
     */
    case Entier = 'int';
}
