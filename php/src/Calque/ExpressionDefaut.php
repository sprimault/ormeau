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
 * L'instant est celui que la base donne à CURRENT_TIMESTAMP : le début de la
 * transaction sous PostgreSQL, l'instruction sous SQL Server. Un défaut
 * calculé dont le sens n'est pas certain n'arrive pas jusqu'ici : il est
 * signalé dans les avertissements.
 */
enum ExpressionDefaut: string
{
    /** L'instant courant, sur un horodatage. */
    case HorodatageCourant = 'horodatage_courant';

    /** La date du jour, sur une date. */
    case DateCourante = 'date_courante';

    /** L'heure courante, sur une heure. */
    case HeureCourante = 'heure_courante';

    /**
     * Un UUID aléatoire tiré par la base à l'insertion. DBAL n'a aucune
     * expression pour l'écrire : la génération ne le reproduit pas, et le
     * rapport dit de fournir la valeur (DefautNonReproduit).
     */
    case UuidGenere = 'uuid_genere';
}
