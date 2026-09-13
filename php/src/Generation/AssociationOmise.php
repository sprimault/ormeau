<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

/**
 * Le côté inverse d'une association, omis parce que l'entité qui porte le côté
 * propriétaire est écartée.
 *
 * L'entité reste générée entière : un côté inverse ne porte aucune colonne,
 * Doctrine accepte l'association unidirectionnelle, et c'est le côté
 * propriétaire qui fait foi en base. L'omettre plutôt qu'écarter l'entité
 * évite qu'un écart ne remonte toute la hiérarchie — mais l'omission se dit,
 * comme un écart : une propriété absente sans raison ferait douter du reste.
 */
final class AssociationOmise
{
    /**
     * @param string $entite      nom de l'entité générée sans ce côté inverse
     * @param string $association nom de l'association omise
     * @param string $raison      pourquoi, en nommant la cible et sa table
     */
    public function __construct(
        public readonly string $entite,
        public readonly string $association,
        public readonly string $raison,
    ) {}
}
