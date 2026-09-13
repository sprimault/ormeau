<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

/**
 * Un fichier que la génération a créé, réécrit, ou laissé tel quel.
 */
final class Fichier
{
    /**
     * @param string      $chemin chemin composé à partir du répertoire de sortie, tel qu'on
     *                            l'affiche
     * @param EtatFichier $etat   ce qui lui est arrivé
     */
    public function __construct(
        public readonly string $chemin,
        public readonly EtatFichier $etat,
    ) {}
}
