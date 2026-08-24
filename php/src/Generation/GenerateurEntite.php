<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

use LogicException;
use Ormeau\Doctrine\Calque\CalqueLogique;

/**
 * Traduit un calque logique en entités Doctrine. Ne décide rien : toute
 * heuristique appartient à l'inférence, côté Go.
 *
 * Deux modes de régénération, parce que c'est là que meurent tous les
 * générateurs de ce type :
 *
 *  - ClasseDeBase : ClientBase est régénérée, Client appartient à l'utilisateur.
 *    Robuste, impose une hiérarchie. Mode par défaut.
 *  - ReecritureAst : nikic/php-parser relit l'entité existante et ne réécrit que
 *    les attributs qui ont bougé, en conservant méthodes métier et formatage.
 */
final class GenerateurEntite
{
    public function __construct(
        private readonly ModeRegeneration $mode = ModeRegeneration::ClasseDeBase,
    ) {}

    /**
     * Rend le mode retenu à la construction.
     */
    public function mode(): ModeRegeneration
    {
        return $this->mode;
    }

    /**
     * Écrit les entités du calque dans le répertoire donné.
     *
     * @return list<string> chemins des fichiers écrits
     */
    public function generer(CalqueLogique $calque, string $repertoire): array
    {
        throw new LogicException(sprintf(
            'Generation en mode %s : a implementer.',
            $this->mode->value,
        ));
    }
}
