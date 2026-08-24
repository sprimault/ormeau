<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

/**
 * Stratégie de préservation du travail humain à la régénération.
 *
 * Adossée à une chaîne : le mode vient d'un drapeau de commande, et les valeurs
 * doivent rester lisibles dans une ligne de commande comme dans une
 * configuration Symfony.
 */
enum ModeRegeneration: string
{
    /** ClientBase est régénérée, Client appartient à l'utilisateur. */
    case ClasseDeBase = 'classe_de_base';

    /** L'entité existante est relue et seuls les attributs qui ont bougé sont réécrits. */
    case ReecritureAst = 'reecriture_ast';
}
