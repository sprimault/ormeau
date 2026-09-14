<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

/**
 * Ce qu'une génération a fait, et ce qu'elle laisse à l'utilisateur.
 *
 * Cinq listes, parce que ce sont cinq questions différentes : qu'est-ce qui a
 * été écrit, qu'est-ce qui n'a pas pu l'être, qu'est-ce qui a été écrit sans
 * une partie, qu'est-ce que l'utilisateur doit reprendre à la main dans ses
 * propres classes, et qu'est-ce qui appartient à une autre base. Aucune ne
 * bloque la génération du reste ; seul un refus d'écrasement dit que la
 * génération demandée n'a pas eu lieu, et fait échouer la commande.
 */
final class Rapport
{
    /**
     * @param list<Fichier>          $fichiers    chaque fichier touché ou examiné, dans l'ordre des
     *                                            entités du calque
     * @param list<EntiteEcartee>    $ecartees    entités qui n'ont produit aucun fichier, avec la raison
     * @param list<Divergence>       $divergences écarts entre une classe de l'utilisateur et le calque,
     *                                            signalés sans jamais y écrire
     * @param list<AssociationOmise> $omises      côtés inverses non écrits, parce que l'entité qui
     *                                            porte le côté propriétaire est écartée
     * @param list<EcrasementRefuse> $refus       fichiers d'une autre base laissés intacts ; seul cas où la
     *                                            génération demandée n'a pas eu lieu
     */
    public function __construct(
        public readonly array $fichiers,
        public readonly array $ecartees,
        public readonly array $divergences,
        public readonly array $omises = [],
        public readonly array $refus = [],
    ) {}
}
