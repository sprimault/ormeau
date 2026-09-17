<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

/**
 * Ce qu'une génération a fait, et ce qu'elle laisse à l'utilisateur.
 *
 * Neuf listes, parce que ce sont neuf questions différentes : qu'est-ce qui a
 * été écrit, qu'est-ce qui n'a pas pu l'être, qu'est-ce qui a été écrit sans
 * une partie, qu'est-ce que l'utilisateur doit reprendre à la main dans ses
 * propres classes, qu'est-ce qui appartient à une autre base, qu'est-ce que
 * son outil de migration lui proposera à tort sur une séquence, qu'est-ce que
 * la version d'ORM installée empêchera d'enregistrer, quel index de la base ne
 * se retrouve pas dans les entités, et quel défaut la génération ne reproduit
 * pas fidèlement. Aucune ne bloque la génération du reste ; seul un
 * refus d'écrasement dit que la génération demandée n'a pas eu lieu, et fait
 * échouer la commande.
 */
final class Rapport
{
    /**
     * @param list<Fichier>            $fichiers    chaque fichier touché ou examiné, dans l'ordre des
     *                                              entités du calque
     * @param list<EntiteEcartee>      $ecartees    entités qui n'ont produit aucun fichier, avec la raison
     * @param list<Divergence>         $divergences écarts entre une classe de l'utilisateur et le calque,
     *                                              signalés sans jamais y écrire
     * @param list<AssociationOmise>   $omises      côtés inverses non écrits, parce que l'entité qui
     *                                              porte le côté propriétaire est écartée
     * @param list<EcrasementRefuse>   $refus       fichiers d'une autre base laissés intacts ; seul cas où la
     *                                              génération demandée n'a pas eu lieu
     * @param list<SequenceNonAlignee|SequenceHorsSchema> $sequences séquences dont migrations:diff proposera
     *                                                               de changer l'incrément ou de les supprimer,
     *                                                               ce qu'il ne faut pas appliquer
     * @param list<ColonneGenereeNonExclue> $generees  hiérarchies jointes qu'une version d'ORM antérieure
     *                                                 à 2.16 n'enregistrera pas, à cause d'une colonne
     *                                                 calculée par la base sur la racine
     * @param list<IndexOmis>               $index     index laissés de côté, DBAL refusant leur nom
     * @param list<DefautRepropose|DefautNonReproduit> $defauts défauts que chaque diff de schéma proposera
     *                                                          de recréer sans effet, ou que la génération
     *                                                          n'écrit pas et que l'application doit fournir
     */
    public function __construct(
        public readonly array $fichiers,
        public readonly array $ecartees,
        public readonly array $divergences,
        public readonly array $omises = [],
        public readonly array $refus = [],
        public readonly array $sequences = [],
        public readonly array $generees = [],
        public readonly array $index = [],
        public readonly array $defauts = [],
    ) {}
}
