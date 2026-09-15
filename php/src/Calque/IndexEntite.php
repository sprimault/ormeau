<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * Un index reporté du physique, pour que la régénération du schéma soit
 * fidèle.
 *
 * La méthode et la classe d'opérateurs n'y survivent pas : Doctrine ne sait
 * pas les exprimer, et c'est le calque physique qui les garde. Le prédicat
 * d'un index partiel, lui, suit : recréé complet, l'index mentirait, et une
 * unicité partielle deviendrait plus stricte que la base.
 */
final class IndexEntite
{
    /**
     * @param list<string> $colonnes noms de colonnes — pas de propriétés —, dans l'ordre de
     *                               l'index, qui compte pour l'optimiseur
     * @param bool         $unique   vrai pour un index d'unicité, y compris composite
     * @param string|null  $nom      nom de l'index en base, absent quand le SGBD n'en expose pas
     * @param string|null  $predicat condition d'un index partiel, verbatim du catalogue et
     *                               jamais interprétée ; absente pour un index complet
     */
    public function __construct(
        public readonly array $colonnes,
        public readonly bool $unique,
        public readonly ?string $nom = null,
        public readonly ?string $predicat = null,
    ) {}

    /**
     * Construit l'index depuis le JSON décodé. Un index sans colonne est
     * refusé.
     *
     * @param array<mixed> $donnees objet JSON décodé de l'index
     * @param string       $chemin  chemin de l'objet dans le calque, pour les messages
     *
     * @throws CalqueInvalide colonnes absentes ou vides, unicité absente, prédicat qui n'est pas une chaîne
     */
    public static function depuisTableau(array $donnees, string $chemin): self
    {
        $colonnes = Lecture::chaines($donnees, 'colonnes', $chemin, requise: true);
        Lecture::nonVide($colonnes, 'colonnes', $chemin);

        $unique = Lecture::booleen($donnees, 'unique', $chemin);
        $nom = Lecture::chaineOptionnelle($donnees, 'nom', $chemin);
        $predicat = Lecture::chaineOptionnelle($donnees, 'predicat', $chemin);

        return new self($colonnes, $unique, $nom, $predicat);
    }
}
