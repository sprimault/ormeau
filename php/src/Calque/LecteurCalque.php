<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * Lit un calque logique produit par le binaire ormeau.
 *
 * Ce paquet ne lit jamais le calque physique et ne se connecte jamais à la base
 * pour générer : si une information manque ici, c'est un champ manquant du
 * calque ou une heuristique manquante côté inférence, jamais quelque chose à
 * deviner en PHP.
 */
final class LecteurCalque
{
    public const VERSION_CONNUE = 1;

    public function lire(string $chemin): CalqueLogique
    {
        $contenu = @file_get_contents($chemin);
        if ($contenu === false) {
            throw new CalqueInvalide(sprintf('Calque illisible : %s', $chemin));
        }

        $donnees = json_decode($contenu, true, 64, JSON_THROW_ON_ERROR);

        $version = $donnees['version_ri'] ?? null;
        if (!is_int($version)) {
            throw new CalqueInvalide('version_ri absente ou invalide');
        }

        if ($version > self::VERSION_CONNUE) {
            throw new CalqueInvalide(sprintf(
                'Calque en version %d, ce paquet ne connaît que la version %d',
                $version,
                self::VERSION_CONNUE,
            ));
        }

        return CalqueLogique::depuisTableau($donnees);
    }
}
