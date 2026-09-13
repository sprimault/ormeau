<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * Une colonne portante appariée à la colonne qu'elle référence.
 *
 * Une clé étrangère composite donne plusieurs appariements, dans l'ordre de la
 * contrainte : c'est cet ordre qui fait correspondre les colonnes deux à deux,
 * pas leurs noms.
 */
final class ColonneJointure
{
    /**
     * @param string                 $colonne           colonne de la table qui porte la clé
     * @param string                 $colonneReferencee colonne visée dans la table cible
     * @param bool                   $nullable          reprise du physique : une colonne facultative
     *                               rend l'association facultative
     * @param ActionSuppression|null $aLaSuppression    comportement déclaré de la clé étrangère,
     *                               absent quand elle n'en déclare pas ; il se reporte en onDelete
     */
    public function __construct(
        public readonly string $colonne,
        public readonly string $colonneReferencee,
        public readonly bool $nullable,
        public readonly ?ActionSuppression $aLaSuppression = null,
    ) {}

    /**
     * Construit l'appariement depuis le JSON décodé.
     *
     * @param array<mixed> $donnees
     * @param string       $chemin  chemin de l'objet dans le calque, pour les messages
     *
     * @throws CalqueInvalide colonne absente, action hors vocabulaire
     */
    public static function depuisTableau(array $donnees, string $chemin): self
    {
        $colonne = Lecture::chaine($donnees, 'colonne', $chemin);
        $colonneReferencee = Lecture::chaine($donnees, 'colonne_referencee', $chemin);
        $nullable = Lecture::booleen($donnees, 'nullable', $chemin);
        $aLaSuppression = Lecture::valeurOptionnelle($donnees, 'a_la_suppression', $chemin, ActionSuppression::class);

        return new self($colonne, $colonneReferencee, $nullable, $aLaSuppression);
    }
}
