<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * Un trait du calque : des propriétés récurrentes, déclarées une fois et
 * partagées par les entités qui le nomment.
 *
 * C'est le cas des horodatages techniques — created_at, updated_at — qu'une
 * base reprise répète dans chaque table. L'inférence ne regroupe que des
 * propriétés identiques en nom, type et nullabilité : deux tables dont l'une a
 * updated_at facultative et l'autre obligatoire ne partagent pas le même trait.
 *
 * TraitPartage et non Trait, mot réservé de PHP.
 */
final class TraitPartage
{
    /**
     * @param string          $nom        nom du trait PHP, sans espace de noms
     * @param list<Propriete> $proprietes propriétés que le trait déclare, retirées des entités
     *                                    qui l'utilisent
     */
    public function __construct(
        public readonly string $nom,
        public readonly array $proprietes,
    ) {}

    /**
     * Construit le trait depuis le JSON décodé.
     *
     * @param array<mixed> $donnees objet JSON décodé du trait
     * @param string       $chemin  chemin de l'objet dans le calque, pour les messages
     *
     * @throws CalqueInvalide nom ou propriétés absents, propriété invalide
     */
    public static function depuisTableau(array $donnees, string $chemin): self
    {
        $nom = Lecture::chaine($donnees, 'nom', $chemin);
        $proprietes = Lecture::objets($donnees, 'proprietes', $chemin, Propriete::depuisTableau(...), requise: true);

        return new self($nom, $proprietes);
    }
}
