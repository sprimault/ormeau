<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

use JsonException;

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
    /**
     * Version de format que ce paquet sait lire.
     *
     * À tenir alignée sur les structures Go et le JSON Schema : les trois
     * annoncent toujours la même, une divergence est un défaut et non un
     * décalage temporaire.
     */
    public const VERSION_CONNUE = 2;

    /**
     * Charge un calque logique depuis un fichier.
     *
     * Un calque de version supérieure est refusé plutôt que lu au mieux : il
     * peut porter des champs dont l'absence de traitement produirait des
     * entités silencieusement fausses. Une version antérieure l'est aussi : elle
     * vient d'une inférence antérieure, sans ses corrections, et ce paquet ne
     * peut pas la recalculer — il n'a ni le binaire ni le droit de lire le
     * calque physique. Le message donne la commande qui le fait, hors ligne.
     *
     * @throws CalqueInvalide fichier illisible, document sans objet racine,
     *                        version absente ou non gérée, champ requis
     *                        absent, du mauvais type ou hors vocabulaire —
     *                        le message nomme le chemin du champ
     * @throws JsonException JSON mal formé, laissé tel quel : le message du
     *                       décodeur situe l'erreur mieux qu'une réécriture
     */
    public function lire(string $chemin): CalqueLogique
    {
        $contenu = @file_get_contents($chemin);
        if ($contenu === false) {
            throw new CalqueInvalide(sprintf('Calque illisible : %s', $chemin));
        }

        $donnees = json_decode($contenu, true, 64, JSON_THROW_ON_ERROR);
        if (!is_array($donnees)) {
            throw new CalqueInvalide(sprintf('Calque sans objet racine : %s', $chemin));
        }

        $version = $donnees['version_ri'] ?? null;
        if (!is_int($version)) {
            throw new CalqueInvalide('version_ri absente ou invalide');
        }

        if ($version < 1) {
            throw new CalqueInvalide(sprintf('version_ri %d invalide : un calque commence en version 1', $version));
        }

        if ($version > self::VERSION_CONNUE) {
            throw new CalqueInvalide(sprintf(
                'Calque en version %d, ce paquet ne connaît que la version %d',
                $version,
                self::VERSION_CONNUE,
            ));
        }

        if ($version < self::VERSION_CONNUE) {
            $base = preg_replace('/\.logique\.json$/', '', basename($chemin));
            throw new CalqueInvalide(sprintf(
                'Calque logique en version %d, ce paquet ne lit que la version %d : le recalculer par « ormeau inferer %s.calque.json --decisions %s.decisions.yaml --sortie %s.logique.json »',
                $version,
                self::VERSION_CONNUE,
                $base,
                $base,
                $base,
            ));
        }

        return CalqueLogique::depuisTableau($donnees);
    }
}
