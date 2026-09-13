<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

/**
 * Ce qu'une génération a fait d'un fichier.
 *
 * Inchangé et conservé ne se confondent pas : le premier est une classe de base
 * dont le contenu n'a pas bougé, le second une classe de l'utilisateur, que la
 * génération ne réécrit jamais, qu'elle ait bougé ou non.
 */
enum EtatFichier: string
{
    /** Le fichier n'existait pas. */
    case Cree = 'créé';

    /** Une classe de base dont le contenu a changé. */
    case Reecrit = 'réécrit';

    /** Une classe de base identique à ce qui était sur disque, laissée intacte. */
    case Inchange = 'inchangé';

    /** Une classe de l'utilisateur déjà présente, jamais touchée. */
    case Conserve = 'conservé';
}
