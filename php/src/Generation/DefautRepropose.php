<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

/**
 * Date ou heure du jour par défaut, sous SQL Server avant DBAL 4.4 : juste,
 * mais reproposée par chaque diff de schéma.
 *
 * DBAL n'y écrit la date du jour en expression que sous sa propre forme,
 * CONVERT(date, GETDATE()) ; toute autre chaîne part en littéral cité, que
 * chaque insertion refuse. SQL Server la stocke CONVERT([date],getdate()), DBAL
 * la relit ainsi, et la compare en texte : schema:update et migrations:diff
 * proposent de supprimer puis recréer le défaut, face à la base reprise comme
 * face à celle que Doctrine vient de créer. Appliqué, cela ne change rien.
 *
 * Omettre le défaut ferait taire le diff face à une base recréée, mais
 * proposerait face à la base reprise un DROP CONSTRAINT qui, appliqué,
 * retirerait le défaut aux autres applications qui insèrent sans valeur.
 * DBAL 4.4 écrit l'instant courant par un objet, relu à l'identique : la
 * proposition disparaît avec la montée de version.
 */
final class DefautRepropose
{
    /**
     * @param string $classe    entité ou trait qui porte la propriété
     * @param string $propriete propriété dont la colonne porte le défaut
     * @param string $colonne   colonne, telle que la base la nomme
     * @param string $defaut    défaut écrit dans #[ORM\Column], tel que DBAL le reconnaît
     */
    public function __construct(
        public readonly string $classe,
        public readonly string $propriete,
        public readonly string $colonne,
        public readonly string $defaut,
    ) {}

    /**
     * Rend l'avertissement en une ligne : ce qui va apparaître, qu'il est sans
     * effet, et ce qui le fait disparaître.
     */
    public function message(): string
    {
        return sprintf(
            '%s::%s : schema:update et migrations:diff proposeront à chaque passage de supprimer puis recréer le défaut %s de la colonne %s. L\'appliquer ne change rien : DBAL relit ce défaut sous une autre forme que celle qu\'il écrit. Il n\'est plus proposé à partir de DBAL 4.4.',
            $this->classe,
            $this->propriete,
            $this->defaut,
            $this->colonne,
        );
    }
}
