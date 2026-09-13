<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * Qui produit la valeur de la clé.
 */
enum StrategieIdentifiant: string
{
    /** Colonne IDENTITY ou auto-incrémentée. */
    case Identite = 'identite';

    /** Séquence nommée, lue avant l'insertion. */
    case Sequence = 'sequence';

    /** Table sans clé primaire : signalée, jamais inventée. */
    case Aucune = 'aucune';

    /** Valeur fournie par l'application. */
    case Assignee = 'assignee';
}
