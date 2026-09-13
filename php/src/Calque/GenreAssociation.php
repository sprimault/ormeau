<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * Cardinalité d'une association, vue depuis l'entité qui la porte.
 */
enum GenreAssociation: string
{
    /** OneToOne : la colonne portante est unique. */
    case UnVersUn = 'un_vers_un';

    /** ManyToOne. */
    case PlusieursVersUn = 'plusieurs_vers_un';

    /** OneToMany, toujours côté inverse. */
    case UnVersPlusieurs = 'un_vers_plusieurs';

    /** ManyToMany, par une table de jointure pure. */
    case PlusieursVersPlusieurs = 'plusieurs_vers_plusieurs';
}
