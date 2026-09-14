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
    /** Calques logiques de référence, à la racine du dépôt. */
    public const REFERENCES = __DIR__ . '/../../../tests/reference/inference';

    /**
     * Rend les cas de référence, triés : chaque répertoire qui porte un
     * logique.json.
     *
     * Lus sur le disque et non tenus en liste : la règle du projet veut qu'une
     * heuristique arrive avec son cas, et un cas oublié dans une liste passait
     * sans attendu ni validation par Doctrine. Un cas nouveau fait échouer
     * AttendusTest tant que ses attendus ne sont pas écrits.
     *
     * @return list<string>
     *
     * @throws RuntimeException répertoire des références introuvable ou vide
     */
    public static function cas(): array
    {
        $fichiers = glob(self::REFERENCES . '/*/logique.json');
        if ($fichiers === false || $fichiers === []) {
            throw new RuntimeException('Aucun cas de référence sous ' . self::REFERENCES);
        }
        $cas = array_map(static fn(string $fichier): string => basename(dirname($fichier)), $fichiers);
        sort($cas);

        return $cas;
    }

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
     * Range une classe de l'utilisateur dans un sous-répertoire, comme le
     * ferait un développeur : le fichier change de place et son espace de noms
     * suit. Rend le nouveau chemin.
     *
     * @param string $racine     répertoire des entités
     * @param string $nom        nom de la classe, à la racine
     * @param string $sousEspace sous-répertoire, qui est aussi le segment d'espace de noms ajouté
     * @param bool   $copier     laisser l'original en place, pour obtenir deux classes
     */
    public static function ranger(string $racine, string $nom, string $sousEspace, bool $copier = false): string
    {
        $source = (string) file_get_contents($racine . '/' . $nom . '.php');
        $cible = $racine . '/' . $sousEspace . '/' . $nom . '.php';
        if (!is_dir(dirname($cible))) {
            mkdir(dirname($cible), 0o777, true);
        }
        file_put_contents($cible, str_replace("namespace App\\Entity;\n", "namespace App\\Entity\\" . $sousEspace . ";\n", $source));
        if (!$copier) {
            unlink($racine . '/' . $nom . '.php');
        }

        return $cible;
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
