<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * Le calcul d'une colonne générée, repris du calque physique.
 *
 * L'expression est du SQL du dialecte d'origine, jamais interprétée : le
 * générateur ne la réécrit pas en DDL, que Doctrine ne sait pas produire pour
 * une colonne générée. Elle documente la propriété, et sa présence dit que la
 * valeur est à relire après chaque écriture — ce que insertable et modifiable
 * à faux ne disent pas.
 */
final class Generee
{
    /**
     * @param string $expression expression de calcul, telle que le catalogue la rend
     * @param bool   $stockee    vrai pour un calcul matérialisé à l'écriture, faux pour un calcul
     *                           à la lecture
     */
    public function __construct(
        public readonly string $expression,
        public readonly bool $stockee,
    ) {}

    /**
     * Construit le calcul depuis le JSON décodé.
     *
     * @param array<mixed> $donnees objet JSON décodé du calcul
     * @param string       $chemin  chemin de l'objet dans le calque, pour les messages
     *
     * @throws CalqueInvalide expression ou caractère stocké absent
     */
    public static function depuisTableau(array $donnees, string $chemin): self
    {
        $expression = Lecture::chaine($donnees, 'expression', $chemin);
        $stockee = Lecture::booleen($donnees, 'stockee', $chemin);

        return new self($expression, $stockee);
    }
}
