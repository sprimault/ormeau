<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * Un index reporté du physique, pour que la régénération du schéma soit
 * fidèle.
 *
 * Prédicat et méthode d'un index partiel n'y survivent pas : Doctrine ne sait
 * pas les exprimer, et un index partiel rendu en index complet serait un
 * mensonge plus grave que son absence. C'est le calque physique qui les garde.
 */
final class IndexEntite
{
    /**
     * @param list<string> $colonnes noms de colonnes — pas de propriétés —, dans l'ordre de
     *                               l'index, qui compte pour l'optimiseur
     * @param bool         $unique   vrai pour un index d'unicité, y compris composite
     * @param string|null  $nom      nom de l'index en base, absent quand le SGBD n'en expose pas
     */
    public function __construct(
        public readonly array $colonnes,
        public readonly bool $unique,
        public readonly ?string $nom = null,
    ) {}

    /**
     * Construit l'index depuis le JSON décodé. Un index sans colonne est
     * refusé.
     *
     * @param array<mixed> $donnees
     * @param string       $chemin  chemin de l'objet dans le calque, pour les messages
     *
     * @throws CalqueInvalide colonnes absentes ou vides, unicité absente
     */
    public static function depuisTableau(array $donnees, string $chemin): self
    {
        $colonnes = Lecture::chaines($donnees, 'colonnes', $chemin, requise: true);
        Lecture::nonVide($colonnes, 'colonnes', $chemin);

        $unique = Lecture::booleen($donnees, 'unique', $chemin);
        $nom = Lecture::chaineOptionnelle($donnees, 'nom', $chemin);

        return new self($colonnes, $unique, $nom);
    }
}
