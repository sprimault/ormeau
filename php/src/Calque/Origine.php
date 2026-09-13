<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * D'où vient une inférence, de la plus sûre à la plus discutable. Sans elle,
 * l'outil n'est pas auditable.
 */
enum Origine: string
{
    /** Une contrainte déclarée en base : clé, unicité, type. */
    case Contrainte = 'contrainte';

    /** Une expression CHECK. */
    case Verification = 'verification';

    /** Un échantillon de valeurs distinctes. */
    case Cardinalite = 'cardinalite';

    /** Le seul nom d'une table ou d'une colonne. */
    case Nommage = 'nommage';

    /** Le fichier de décisions, qui l'emporte sur tout le reste. */
    case Decision = 'decision';
}
