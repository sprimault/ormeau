<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * La place d'une entité dans une hiérarchie déduite du schéma.
 *
 * Sur une base reprise, l'héritage se lit presque toujours dans une clé
 * primaire qui est aussi une clé étrangère vers la table parente : c'est la
 * stratégie jointe. L'origine porte celle de cette déduction, distincte de
 * celle du nom de classe.
 *
 * Doctrine exige une colonne discriminante pour les deux stratégies, et en
 * ajoute une quand le mapping n'en déclare pas. Une base legacy n'en a
 * généralement pas : $colonneDiscriminante absente est le cas courant, pas une
 * anomalie du calque.
 */
final class Heritage
{
    /**
     * @param StrategieHeritage $strategie            projection de la hiérarchie sur les tables
     * @param string            $parent               nom de l'entité parente dans le même calque,
     *                                                pas un nom de table
     * @param string|null       $colonneDiscriminante colonne qui départage les classes, quand la
     *                                                base en a une
     * @param Origine|null      $origine              d'où vient la déduction de l'héritage
     */
    public function __construct(
        public readonly StrategieHeritage $strategie,
        public readonly string $parent,
        public readonly ?string $colonneDiscriminante = null,
        public readonly ?Origine $origine = null,
    ) {}

    /**
     * Construit l'héritage depuis le JSON décodé.
     *
     * @param array<mixed> $donnees objet JSON décodé de l'héritage
     * @param string       $chemin  chemin de l'objet dans le calque, pour les messages
     *
     * @throws CalqueInvalide stratégie absente ou hors vocabulaire, parent absent
     */
    public static function depuisTableau(array $donnees, string $chemin): self
    {
        $strategie = Lecture::valeur($donnees, 'strategie', $chemin, StrategieHeritage::class);
        $parent = Lecture::chaine($donnees, 'parent', $chemin);
        $colonneDiscriminante = Lecture::chaineOptionnelle($donnees, 'colonne_discriminante', $chemin);
        $origine = Lecture::valeurOptionnelle($donnees, 'origine', $chemin, Origine::class);

        return new self($strategie, $parent, $colonneDiscriminante, $origine);
    }
}
