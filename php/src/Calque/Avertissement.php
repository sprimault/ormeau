<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * Ce que l'inférence n'a pas résolu, dit plutôt qu'inventé.
 *
 * C'est une sortie de premier ordre, pas un journal : la commande de génération
 * les affiche, et une CI filtre sur leur code et leur confiance.
 *
 * Le code reste une chaîne et non un vocabulaire fermé. Un code nouveau ne
 * change pas le format, et la génération doit pouvoir afficher un
 * avertissement qu'elle ne connaît pas plutôt que refuser le calque.
 */
final class Avertissement
{
    /**
     * @param string     $code       identifiant stable entre versions, en snake_case
     * @param string     $cible      élément visé, qualifié : schéma, schema.table ou
     *                               schema.table.colonne
     * @param string     $message    explication en français, produite par le binaire
     * @param Resolution $resolution ce que l'inférence a fait de l'incertitude
     * @param float      $confiance  dans [0, 1] ; 1 pour un constat, moins pour une supposition
     */
    public function __construct(
        public readonly string $code,
        public readonly string $cible,
        public readonly string $message,
        public readonly Resolution $resolution,
        public readonly float $confiance,
    ) {}

    /**
     * Construit l'avertissement depuis le JSON décodé.
     *
     * @param array<mixed> $donnees objet JSON décodé de l'avertissement
     * @param string       $chemin  chemin de l'objet dans le calque, pour les messages
     *
     * @throws CalqueInvalide champ requis absent, résolution hors vocabulaire
     */
    public static function depuisTableau(array $donnees, string $chemin): self
    {
        $code = Lecture::chaine($donnees, 'code', $chemin);
        $cible = Lecture::chaine($donnees, 'cible', $chemin);
        $message = Lecture::chaine($donnees, 'message', $chemin);
        $resolution = Lecture::valeur($donnees, 'resolution', $chemin, Resolution::class);
        $confiance = Lecture::nombre($donnees, 'confiance', $chemin);

        return new self($code, $cible, $message, $resolution, $confiance);
    }
}
