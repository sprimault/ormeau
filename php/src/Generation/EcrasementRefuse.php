<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

/**
 * Fichier qu'une génération n'a pas écrit parce qu'il vient d'une autre base.
 *
 * Deux bases générées dans le même répertoire d'entités se réécriraient l'une
 * l'autre en silence : la seconde efface la classe de base de la première, et
 * rien ne le signale avant que l'application ne charge des entités fausses.
 * Contrairement à une entité écartée ou à une divergence, un refus veut dire
 * que la génération demandée n'a pas eu lieu, et la commande échoue.
 */
final class EcrasementRefuse
{
    /**
     * @param string $fichier       fichier laissé intact
     * @param string $baseExistante base que nomme son en-tête
     * @param string $baseDemandee  base de la génération refusée
     */
    public function __construct(
        public readonly string $fichier,
        public readonly string $baseExistante,
        public readonly string $baseDemandee,
    ) {}

    /**
     * Rend le refus en une ligne, avec les deux bases et la marche à suivre.
     */
    public function message(): string
    {
        return sprintf(
            '%s vient de la base %s, la génération depuis %s ne l\'écrit pas : générer chaque base dans son répertoire et son espace de noms, ou remplacer explicitement la base %s',
            $this->fichier,
            $this->baseExistante,
            $this->baseDemandee,
            $this->baseExistante,
        );
    }
}
