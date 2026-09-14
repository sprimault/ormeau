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

    /**
     * Produite par aucune version d'Ormeau : une table sans clé primaire n'a
     * pas d'identifiant du tout. Déclarée en version 1 du format, retirée à la
     * suivante ; la génération écarte d'ici là l'entité qui la porte.
     */
    case Aucune = 'aucune';

    /** Valeur fournie par l'application. */
    case Assignee = 'assignee';
}
