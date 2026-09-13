<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

/**
 * Un fragment de PHP que l'émetteur écrit tel quel, sans le mettre entre
 * apostrophes : StatutClient::class, StatutClient::Actif.
 *
 * Réservé à ce que le rendu compose lui-même à partir de noms déjà validés.
 * Une valeur venue du calque — un défaut, un commentaire — passe toujours par
 * un littéral : l'écrire comme code ouvrirait le fichier produit à n'importe
 * quelle expression.
 *
 * @internal
 */
final class Code
{
    /**
     * @param string $php expression PHP, écrite telle quelle dans le fichier produit
     */
    public function __construct(
        public readonly string $php,
    ) {}
}
