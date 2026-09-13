<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * La table dont une entité est issue, sous son nom de catalogue.
 *
 * Le nom n'est jamais traduit ni normalisé : c'est lui que l'attribut
 * #[ORM\Table] doit écrire, et un T_CLIENTS rendu en t_clients viserait une
 * autre table sur un SGBD sensible à la casse.
 */
final class ReferenceTable
{
    /**
     * @param string $nom    nom de la table, tel que le catalogue l'écrit
     * @param string $schema schéma qui la contient ; deux tables de même nom
     *                       dans deux schémas sont deux entités distinctes
     */
    public function __construct(
        public readonly string $nom,
        public readonly string $schema,
    ) {}

    /**
     * Construit la référence depuis le JSON décodé.
     *
     * @param array<mixed> $donnees
     * @param string       $chemin  chemin de l'objet dans le calque, pour les messages
     *
     * @throws CalqueInvalide nom ou schéma absent
     */
    public static function depuisTableau(array $donnees, string $chemin): self
    {
        $nom = Lecture::chaine($donnees, 'nom', $chemin);
        $schema = Lecture::chaine($donnees, 'schema', $chemin);

        return new self($nom, $schema);
    }
}
