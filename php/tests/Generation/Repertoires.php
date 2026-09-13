<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\Generation;

use FilesystemIterator;
use RecursiveDirectoryIterator;
use RecursiveIteratorIterator;
use RuntimeException;
use SplFileInfo;

/**
 * Répertoires jetables et lecture d'arborescences, pour les tests qui
 * écrivent des entités.
 *
 * Tout va sous le répertoire temporaire du système, un sous-répertoire par
 * appel : les tests tournent dans un ordre aléatoire, et deux tests ne doivent
 * jamais écrire au même endroit.
 */
final class Repertoires
{
    /** Cas d'inférence dont chaque entité relève de la génération actuelle, ou en est écartée. */
    public const CAS = [
        'cas-limites',
        'cible-ignoree',
        'decisions',
        'enumerations',
        'enumerations-decidees',
        'identifiants-penibles',
        'minimal',
        'nommage-court',
        'prefixes',
        'singularisation',
        'traits',
        'types',
    ];

    /** Calques logiques de référence, à la racine du dépôt. */
    public const REFERENCES = __DIR__ . '/../../../tests/reference/inference';

    /**
     * Crée un répertoire vide et rend son chemin.
     */
    public static function creer(): string
    {
        $chemin = sys_get_temp_dir() . '/ormeau-generation-' . bin2hex(random_bytes(6));
        if (!mkdir($chemin, 0o777, true)) {
            throw new RuntimeException('Répertoire temporaire impossible à créer : ' . $chemin);
        }

        return $chemin;
    }

    /**
     * Supprime un répertoire et tout ce qu'il contient.
     */
    public static function supprimer(string $chemin): void
    {
        if (!is_dir($chemin)) {
            return;
        }
        $elements = new RecursiveIteratorIterator(
            new RecursiveDirectoryIterator($chemin, FilesystemIterator::SKIP_DOTS),
            RecursiveIteratorIterator::CHILD_FIRST,
        );
        foreach ($elements as $element) {
            /** @var SplFileInfo $element */
            $element->isDir() ? rmdir($element->getPathname()) : unlink($element->getPathname());
        }
        rmdir($chemin);
    }

    /**
     * Rend les fichiers d'une arborescence, chemin relatif vers contenu, triés.
     *
     * @return array<string, string>
     */
    public static function lire(string $racine): array
    {
        $fichiers = [];
        if (!is_dir($racine)) {
            return $fichiers;
        }
        $elements = new RecursiveIteratorIterator(new RecursiveDirectoryIterator($racine, FilesystemIterator::SKIP_DOTS));
        foreach ($elements as $element) {
            /** @var SplFileInfo $element */
            $relatif = str_replace('\\', '/', substr($element->getPathname(), strlen($racine) + 1));
            $fichiers[$relatif] = (string) file_get_contents($element->getPathname());
        }
        ksort($fichiers);

        return $fichiers;
    }
}
