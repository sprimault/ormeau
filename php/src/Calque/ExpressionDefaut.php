<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * Le sens d'un défaut calculé que l'inférence a reconnu, jamais son texte :
 * now() et CURRENT_TIMESTAMP disent la même chose, et c'est le sens qui se
 * traduit dans la plateforme de la cible.
 *
 * L'instant est celui de la transaction. Un défaut calculé dont le sens n'est
 * pas certain n'arrive pas jusqu'ici : il est signalé dans les avertissements.
 */
enum ExpressionDefaut: string
{
    /** L'instant courant, sur un horodatage. */
    case HorodatageCourant = 'horodatage_courant';

    /** La date du jour, sur une date. */
    case DateCourante = 'date_courante';

    /** L'heure courante, sur une heure. */
    case HeureCourante = 'heure_courante';
}
