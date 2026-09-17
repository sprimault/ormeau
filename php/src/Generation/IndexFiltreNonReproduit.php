<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

/**
 * Index filtré de SQL Server, rendu sans son filtre.
 *
 * La plateforme SQL Server de DBAL n'écrit pas options: where. Écrit dans
 * l'entité, le filtre fait proposer à schema:update et migrations:diff de
 * supprimer l'index de la base d'origine puis de le recréer sans filtre, et de
 * nouveau à chaque passage ; sans lui, rien n'est proposé. Dans les deux cas,
 * une base créée par Doctrine a l'index sans filtre, et une unicité filtrée y
 * porte sur toutes les lignes non nulles. Constaté sous ORM 2.14 / DBAL 3.10 et
 * ORM 3.7 / DBAL 4.4, contre SQL Server 2022.
 *
 * L'index reste dans l'entité : l'omettre ferait proposer sa suppression.
 */
final class IndexFiltreNonReproduit
{
    /**
     * @param string      $entite   entité dont la table porte l'index
     * @param string|null $nom      nom de l'index, tel que le catalogue le rend ; null quand il n'en a pas
     * @param list<string> $colonnes colonnes de l'index, pour le désigner sans nom
     * @param string      $predicat filtre de l'index, verbatim du catalogue
     * @param bool        $unique   l'index porte une unicité, qu'une base recréée étend à toutes les lignes
     */
    public function __construct(
        public readonly string $entite,
        public readonly ?string $nom,
        public readonly array $colonnes,
        public readonly string $predicat,
        public readonly bool $unique,
    ) {}

    /**
     * Rend l'avertissement en une ligne : ce que Doctrine ne sait pas écrire,
     * ce qu'une base recréée aura à la place, et comment le rétablir.
     */
    public function message(): string
    {
        return sprintf(
            '%s : %s %s porte le filtre %s, que DBAL n\'écrit pas sous SQL Server. Une base créée par Doctrine l\'aura sans filtre%s : le recréer à la main dans la migration.',
            $this->entite,
            $this->unique ? 'l\'unicité' : 'l\'index',
            $this->nom ?? 'sur ' . implode(', ', $this->colonnes),
            $this->predicat,
            $this->unique ? ', et refusera des lignes que la base d\'origine accepte' : '',
        );
    }
}
