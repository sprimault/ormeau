<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * Un lien entre deux entités, vu depuis celle qui le porte.
 *
 * $proprietaire décide du côté qui porte la colonne de jointure : s'y tromper
 * produit un mapping que Doctrine accepte et qui n'écrit rien en base. Une
 * association bidirectionnelle apparaît donc deux fois dans le calque, une par
 * entité, et les deux se désignent par $inverseePar et $mappeePar.
 *
 * L'origine est requise, contrairement à celle d'une propriété : une clé
 * étrangère est un constat, l'association qu'on en tire — son genre, ses deux
 * côtés — est déjà un jugement.
 */
final class Association
{
    /**
     * @param string                $nom                nom de la propriété PHP qui porte le lien
     * @param GenreAssociation      $genre              cardinalité, vue depuis cette entité
     * @param string                $cible              nom de l'entité visée dans le même calque
     * @param bool                  $proprietaire       vrai du côté qui porte la colonne de jointure
     * @param Origine               $origine            d'où vient l'association
     * @param string|null           $inverseePar        propriété inverse, présente côté propriétaire
     *                                                  d'un lien bidirectionnel
     * @param string|null           $mappeePar          propriété propriétaire, présente côté inverse
     * @param list<ColonneJointure> $jointure           colonnes portantes, côté propriétaire d'un lien
     *                                                  sans table de jointure
     * @param TableJointure|null    $tableJointure      table d'association d'un plusieurs-vers-plusieurs
     * @param bool                  $orphelinsSupprimes orphanRemoval : un élément retiré de la
     *                                                  collection est supprimé en base
     */
    public function __construct(
        public readonly string $nom,
        public readonly GenreAssociation $genre,
        public readonly string $cible,
        public readonly bool $proprietaire,
        public readonly Origine $origine,
        public readonly ?string $inverseePar = null,
        public readonly ?string $mappeePar = null,
        public readonly array $jointure = [],
        public readonly ?TableJointure $tableJointure = null,
        public readonly bool $orphelinsSupprimes = false,
    ) {}

    /**
     * Construit l'association depuis le JSON décodé.
     *
     * @param array<mixed> $donnees objet JSON décodé de l'association
     * @param string       $chemin  chemin de l'objet dans le calque, pour les messages
     *
     * @throws CalqueInvalide champ requis absent, genre ou origine hors vocabulaire
     */
    public static function depuisTableau(array $donnees, string $chemin): self
    {
        $nom = Lecture::chaine($donnees, 'nom', $chemin);
        $genre = Lecture::valeur($donnees, 'genre', $chemin, GenreAssociation::class);
        $cible = Lecture::chaine($donnees, 'cible', $chemin);
        $proprietaire = Lecture::booleen($donnees, 'proprietaire', $chemin);
        $origine = Lecture::valeur($donnees, 'origine', $chemin, Origine::class);
        $inverseePar = Lecture::chaineOptionnelle($donnees, 'inversee_par', $chemin);
        $mappeePar = Lecture::chaineOptionnelle($donnees, 'mappee_par', $chemin);
        $jointure = Lecture::objets($donnees, 'jointure', $chemin, ColonneJointure::depuisTableau(...));
        $tableJointure = Lecture::objet($donnees, 'table_jointure', $chemin, TableJointure::depuisTableau(...));
        $orphelinsSupprimes = Lecture::booleen($donnees, 'orphelins_supprimes', $chemin, defaut: false);

        return new self(
            $nom,
            $genre,
            $cible,
            $proprietaire,
            $origine,
            $inverseePar,
            $mappeePar,
            $jointure,
            $tableJointure,
            $orphelinsSupprimes,
        );
    }
}
