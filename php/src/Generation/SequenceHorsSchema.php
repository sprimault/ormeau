<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

/**
 * Séquence tirée par un générateur produit, que les outils de schéma de
 * Doctrine ne voient pas.
 *
 * Sous ORM 3 et SQL Server, la clé passe par un générateur CUSTOM (voir
 * RenduEntite::tireParGenerateur) : l'identifiant est juste, mais la séquence
 * n'est plus déclarée dans le mapping. schema:update et migrations:diff
 * proposent donc de la supprimer, ce que SQL Server refuse tant que le défaut
 * de la colonne la référence, et schema:create ne la crée pas. Constaté sous
 * ORM 3.7, contre SQL Server 2022 (essai du 2026-09-16).
 */
final class SequenceHorsSchema
{
    /**
     * @param string $entite   entité dont la clé lit la séquence
     * @param string $sequence nom de la séquence, tel que le générateur l'écrit
     */
    public function __construct(
        public readonly string $entite,
        public readonly string $sequence,
    ) {}

    /**
     * Rend l'avertissement en une ligne : ce qui va apparaître, qu'il ne faut
     * pas l'appliquer, et ce qu'une base recréée par Doctrine n'aura pas.
     */
    public function message(): string
    {
        return sprintf(
            '%s : la séquence %s est tirée par un générateur que les outils de schéma de Doctrine ne voient pas. schema:update et migrations:diff proposeront « DROP SEQUENCE %s » : ne pas l\'appliquer, SQL Server le refuse tant que la colonne la référence. Une base créée par schema:create n\'aura pas la séquence.',
            $this->entite,
            $this->sequence,
            $this->sequence,
        );
    }
}
