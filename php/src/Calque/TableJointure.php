<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * La table d'association d'un plusieurs-vers-plusieurs.
 *
 * Elle ne devient pas une entité : l'inférence ne la reconnaît comme jointure
 * que si elle ne porte rien d'autre que ses deux clés étrangères. Une colonne
 * de plus — une date, une quantité — en ferait une entité à part entière.
 */
final class TableJointure
{
    /**
     * @param string                $nom             nom de la table, tel que le catalogue l'écrit
     * @param string                $schema          schéma qui la contient
     * @param list<ColonneJointure> $jointure        colonnes qui mènent à l'entité propriétaire —
     *                                               #[JoinColumn]
     * @param list<ColonneJointure> $jointureInverse colonnes qui mènent à l'entité cible —
     *                                               #[InverseJoinColumn]
     */
    public function __construct(
        public readonly string $nom,
        public readonly string $schema,
        public readonly array $jointure,
        public readonly array $jointureInverse,
    ) {}

    /**
     * Construit la table de jointure depuis le JSON décodé. Les deux listes
     * sont requises, même vides : le format les écrit toujours.
     *
     * @param array<mixed> $donnees objet JSON décodé de la table de jointure
     * @param string       $chemin  chemin de l'objet dans le calque, pour les messages
     *
     * @throws CalqueInvalide champ requis absent, appariement invalide
     */
    public static function depuisTableau(array $donnees, string $chemin): self
    {
        $nom = Lecture::chaine($donnees, 'nom', $chemin);
        $schema = Lecture::chaine($donnees, 'schema', $chemin);
        $jointure = Lecture::objets($donnees, 'jointure', $chemin, ColonneJointure::depuisTableau(...), requise: true);
        $jointureInverse = Lecture::objets($donnees, 'jointure_inverse', $chemin, ColonneJointure::depuisTableau(...), requise: true);

        return new self($nom, $schema, $jointure, $jointureInverse);
    }
}
