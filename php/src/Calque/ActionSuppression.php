<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * Ce que la base fait des lignes qui référencent une ligne supprimée.
 */
enum ActionSuppression: string
{
    /** NO ACTION : la suppression échoue en fin d'instruction. */
    case Aucune = 'aucune';

    /** Les lignes qui référencent sont supprimées aussi. */
    case Cascade = 'cascade';

    /** La référence passe à NULL. */
    case SetNull = 'set_null';

    /** La référence reprend son défaut. */
    case SetDefault = 'set_default';

    /** RESTRICT : la suppression échoue immédiatement. */
    case Restrict = 'restrict';
}
