<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * La place d'une entité dans une hiérarchie déclarée par décision.
 *
 * Une clé primaire qui est aussi une clé étrangère autorise un héritage sans
 * l'imposer : « un salarié est une personne » et « un salarié a une personne »
 * sont deux modèles que le schéma permet également. Sans décision, l'inférence
 * relie la table à son parent par un un-vers-un, et cet objet est absent.
 *
 * Doctrine exige une colonne discriminante pour les deux stratégies, et en
 * ajoute une quand le mapping n'en déclare pas — ce qui casse toute requête
 * sur une base qui ne la porte pas. C'est pourquoi un héritage n'existe ici
 * que déclaré, avec sa colonne, et que la valeur de chaque classe est portée
 * par Entite::$valeurDiscriminante, racine comprise.
 */
final class Heritage
{
    /**
     * @param StrategieHeritage $strategie            projection de la hiérarchie sur les tables
     * @param string            $parent               nom de l'entité parente dans le même calque,
     *                                                pas un nom de table
     * @param string|null       $colonneDiscriminante colonne de la table racine qui départage les
     *                                                classes ; la décision la fournit
     * @param Origine|null      $origine              d'où vient l'héritage : une décision
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
