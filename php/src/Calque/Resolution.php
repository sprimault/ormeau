<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * Ce que l'inférence a fait de l'incertitude qu'un avertissement signale.
 */
enum Resolution: string
{
    /** L'élément visé a été laissé de côté. */
    case Ignoree = 'ignoree';

    /** Un trou laissé ouvert, à trancher. */
    case Aucune = 'aucune';

    /** Un choix appliqué faute de mieux. */
    case ParDefaut = 'par_defaut';

    /** Le fichier de décisions a tranché. */
    case ForceeParDecision = 'forcee_par_decision';
}
