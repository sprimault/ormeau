<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

/**
 * Ce qu'une génération a fait, et ce qu'elle laisse à l'utilisateur.
 *
 * Trois listes, parce que ce sont trois questions différentes : qu'est-ce qui a
 * été écrit, qu'est-ce qui n'a pas pu l'être, et qu'est-ce que l'utilisateur
 * doit reprendre à la main dans ses propres classes. Aucune ne bloque la
 * génération : une entité écartée ou une divergence n'empêche pas les autres
 * d'être écrites.
 */
final class Rapport
{
    /**
     * @param list<Fichier>        $fichiers    chaque fichier touché ou examiné, dans l'ordre des
     *                                          entités du calque
     * @param list<EntiteEcartee>  $ecartees    entités qui n'ont produit aucun fichier, avec la raison
     * @param list<Divergence>     $divergences écarts entre une classe de l'utilisateur et le calque,
     *                                          signalés sans jamais y écrire
     */
    public function __construct(
        public readonly array $fichiers,
        public readonly array $ecartees,
        public readonly array $divergences,
    ) {}
}
