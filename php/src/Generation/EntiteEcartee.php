<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

/**
 * Une entité du calque qui n'a produit aucun fichier.
 *
 * Écarter une entité entière plutôt que la générer à moitié : une entité sans
 * ses associations perdrait ses colonnes de jointure, et Doctrine la chargerait
 * sans rien signaler.
 */
final class EntiteEcartee
{
    /**
     * @param string $nom    nom de l'entité dans le calque
     * @param string $raison ce qui empêche de la générer, et ce que l'utilisateur peut y faire
     */
    public function __construct(
        public readonly string $nom,
        public readonly string $raison,
    ) {}
}
