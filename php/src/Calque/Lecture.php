<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

use BackedEnum;

/**
 * Accès typés à un tableau issu de json_decode, qui nomment le champ fautif.
 *
 * Chaque objet du calque se construit par ces méthodes plutôt que par des
 * accès directs : sur un calque de plusieurs milliers de lignes, « chaîne
 * attendue » ne sert à rien sans « entites[12].proprietes[3].type_doctrine ».
 *
 * Les objets lisent leurs champs dans des variables, puis appellent
 * new self(...) avec elles, jamais new self(Lecture::…) directement. Sous
 * PHP 8.1, une exception levée pendant l'évaluation des arguments d'un new —
 * l'objet est déjà alloué — corrompt la mémoire : constaté sous 8.1.34 avec
 * PHPUnit 10.5, en erreur de segmentation au test suivant, et pas au-delà de
 * 8.1. Le job de compatibilité PHP 8.1 le rattrape s'il revient.
 *
 * @internal
 */
final class Lecture
{
    /**
     * Lit une chaîne requise.
     *
     * @param array<mixed> $donnees objet JSON décodé qui porte le champ
     * @param string       $cle     nom du champ, tel que le format l'écrit
     * @param string       $chemin  chemin de l'objet dans le calque, vide à la racine
     *
     * @throws CalqueInvalide
     */
    public static function chaine(array $donnees, string $cle, string $chemin): string
    {
        $valeur = self::requis($donnees, $cle, $chemin);
        if (!is_string($valeur)) {
            throw self::attendu($chemin, $cle, 'chaîne attendue');
        }

        return $valeur;
    }

    /**
     * Lit une chaîne facultative.
     *
     * @param array<mixed> $donnees objet JSON décodé qui porte le champ
     * @param string       $cle     nom du champ, tel que le format l'écrit
     * @param string       $chemin  chemin de l'objet dans le calque, vide à la racine
     *
     * @throws CalqueInvalide
     */
    public static function chaineOptionnelle(array $donnees, string $cle, string $chemin): ?string
    {
        if (!array_key_exists($cle, $donnees)) {
            return null;
        }

        return self::chaine($donnees, $cle, $chemin);
    }

    /**
     * Lit un entier requis. Un nombre à virgule est refusé, même s'il tombe
     * juste : 1.0 dans un calque ne vient pas du binaire.
     *
     * @param array<mixed> $donnees objet JSON décodé qui porte le champ
     * @param string       $cle     nom du champ, tel que le format l'écrit
     * @param string       $chemin  chemin de l'objet dans le calque, vide à la racine
     *
     * @throws CalqueInvalide
     */
    public static function entier(array $donnees, string $cle, string $chemin): int
    {
        $valeur = self::requis($donnees, $cle, $chemin);
        if (!is_int($valeur)) {
            throw self::attendu($chemin, $cle, 'entier attendu');
        }

        return $valeur;
    }

    /**
     * Lit un entier facultatif. Absent et zéro restent distinguables.
     *
     * @param array<mixed> $donnees objet JSON décodé qui porte le champ
     * @param string       $cle     nom du champ, tel que le format l'écrit
     * @param string       $chemin  chemin de l'objet dans le calque, vide à la racine
     *
     * @throws CalqueInvalide
     */
    public static function entierOptionnel(array $donnees, string $cle, string $chemin): ?int
    {
        if (!array_key_exists($cle, $donnees)) {
            return null;
        }

        return self::entier($donnees, $cle, $chemin);
    }

    /**
     * Lit un nombre requis. json_decode rend 1 en entier et 0.5 en flottant,
     * les deux sont acceptés.
     *
     * @param array<mixed> $donnees objet JSON décodé qui porte le champ
     * @param string       $cle     nom du champ, tel que le format l'écrit
     * @param string       $chemin  chemin de l'objet dans le calque, vide à la racine
     *
     * @throws CalqueInvalide
     */
    public static function nombre(array $donnees, string $cle, string $chemin): float
    {
        $valeur = self::requis($donnees, $cle, $chemin);
        if (!is_int($valeur) && !is_float($valeur)) {
            throw self::attendu($chemin, $cle, 'nombre attendu');
        }

        return (float) $valeur;
    }

    /**
     * Lit un booléen, requis quand aucun défaut n'est donné.
     *
     * @param array<mixed> $donnees objet JSON décodé qui porte le champ
     * @param string       $cle     nom du champ, tel que le format l'écrit
     * @param string       $chemin  chemin de l'objet dans le calque, vide à la racine
     * @param bool|null    $defaut  valeur du champ absent ; null le rend requis
     *
     * @throws CalqueInvalide
     */
    public static function booleen(array $donnees, string $cle, string $chemin, ?bool $defaut = null): bool
    {
        if ($defaut !== null && !array_key_exists($cle, $donnees)) {
            return $defaut;
        }

        $valeur = self::requis($donnees, $cle, $chemin);
        if (!is_bool($valeur)) {
            throw self::attendu($chemin, $cle, 'booléen attendu');
        }

        return $valeur;
    }

    /**
     * Lit une valeur d'un vocabulaire fermé.
     *
     * Une valeur inconnue est refusée plutôt que lue au mieux : le format
     * incrémente version_ri quand un vocabulaire grandit, donc une valeur
     * inconnue dans une version connue est un calque abîmé.
     *
     * @template T of BackedEnum
     *
     * @param array<mixed>    $donnees     objet JSON décodé qui porte le champ
     * @param string          $cle         nom du champ, tel que le format l'écrit
     * @param string          $chemin      chemin de l'objet dans le calque, vide à la racine
     * @param class-string<T> $vocabulaire énumération PHP dont les valeurs forment le vocabulaire
     *
     * @return T
     *
     * @throws CalqueInvalide
     */
    public static function valeur(array $donnees, string $cle, string $chemin, string $vocabulaire): BackedEnum
    {
        $brute = self::chaine($donnees, $cle, $chemin);
        $valeur = $vocabulaire::tryFrom($brute);
        if ($valeur === null) {
            throw new CalqueInvalide(sprintf(
                'Champ %s : « %s » hors du vocabulaire (%s)',
                self::champ($chemin, $cle),
                $brute,
                implode(', ', array_map(static fn(BackedEnum $cas): string => (string) $cas->value, $vocabulaire::cases())),
            ));
        }

        return $valeur;
    }

    /**
     * Lit une valeur facultative d'un vocabulaire fermé.
     *
     * @template T of BackedEnum
     *
     * @param array<mixed>    $donnees     objet JSON décodé qui porte le champ
     * @param string          $cle         nom du champ, tel que le format l'écrit
     * @param string          $chemin      chemin de l'objet dans le calque, vide à la racine
     * @param class-string<T> $vocabulaire énumération PHP dont les valeurs forment le vocabulaire
     *
     * @return T|null
     *
     * @throws CalqueInvalide
     */
    public static function valeurOptionnelle(array $donnees, string $cle, string $chemin, string $vocabulaire): ?BackedEnum
    {
        if (!array_key_exists($cle, $donnees)) {
            return null;
        }

        return self::valeur($donnees, $cle, $chemin, $vocabulaire);
    }

    /**
     * Lit une liste de chaînes, vide quand elle est facultative et absente.
     *
     * @param array<mixed> $donnees objet JSON décodé qui porte le champ
     * @param string       $cle     nom du champ, tel que le format l'écrit
     * @param string       $chemin  chemin de l'objet dans le calque, vide à la racine
     * @param bool         $requise refuser la liste absente plutôt que la rendre vide
     *
     * @return list<string>
     *
     * @throws CalqueInvalide
     */
    public static function chaines(array $donnees, string $cle, string $chemin, bool $requise = false): array
    {
        $chaines = [];
        foreach (self::liste($donnees, $cle, $chemin, $requise) as $rang => $element) {
            if (!is_string($element)) {
                throw self::attendu($chemin, sprintf('%s[%d]', $cle, $rang), 'chaîne attendue');
            }
            $chaines[] = $element;
        }

        return $chaines;
    }

    /**
     * Construit chaque objet d'une liste, en lui passant son chemin.
     *
     * @template T
     *
     * @param array<mixed>                      $donnees    objet JSON décodé qui porte la liste
     * @param string                            $cle        nom du champ, tel que le format l'écrit
     * @param string                            $chemin     chemin de l'objet dans le calque, vide à la racine
     * @param callable(array<mixed>, string): T $construire construit un élément depuis son objet et son chemin
     * @param bool                              $requise    refuser la liste absente plutôt que la rendre vide
     *
     * @return list<T>
     *
     * @throws CalqueInvalide
     */
    public static function objets(array $donnees, string $cle, string $chemin, callable $construire, bool $requise = false): array
    {
        $objets = [];
        foreach (self::liste($donnees, $cle, $chemin, $requise) as $rang => $element) {
            $sous = sprintf('%s[%d]', self::champ($chemin, $cle), $rang);
            if (!is_array($element) || ($element !== [] && array_is_list($element))) {
                throw new CalqueInvalide(sprintf('Champ %s : objet attendu', $sous));
            }
            $objets[] = $construire($element, $sous);
        }

        return $objets;
    }

    /**
     * Construit l'objet d'un champ, ou rend null quand il est facultatif et
     * absent.
     *
     * @template T
     *
     * @param array<mixed>                      $donnees    objet JSON décodé qui porte le champ
     * @param string                            $cle        nom du champ, tel que le format l'écrit
     * @param string                            $chemin     chemin de l'objet dans le calque, vide à la racine
     * @param callable(array<mixed>, string): T $construire construit l'objet depuis son tableau et son chemin
     * @param bool                              $requis     refuser le champ absent plutôt que rendre null
     *
     * @return ($requis is true ? T : T|null)
     *
     * @throws CalqueInvalide
     */
    public static function objet(array $donnees, string $cle, string $chemin, callable $construire, bool $requis = false): mixed
    {
        if (!$requis && !array_key_exists($cle, $donnees)) {
            return null;
        }

        $valeur = self::requis($donnees, $cle, $chemin);
        if (!is_array($valeur) || ($valeur !== [] && array_is_list($valeur))) {
            throw self::attendu($chemin, $cle, 'objet attendu');
        }

        return $construire($valeur, self::champ($chemin, $cle));
    }

    /**
     * Vérifie qu'une liste a au moins un élément, là où le format l'exige.
     *
     * @param list<mixed> $liste  liste déjà lue
     * @param string      $cle    nom du champ, pour le message
     * @param string      $chemin chemin de l'objet dans le calque, vide à la racine
     *
     * @throws CalqueInvalide
     */
    public static function nonVide(array $liste, string $cle, string $chemin): void
    {
        if ($liste === []) {
            throw new CalqueInvalide(sprintf('Champ %s : au moins un élément attendu', self::champ($chemin, $cle)));
        }
    }

    /**
     * Rend la liste d'un champ.
     *
     * @param array<mixed> $donnees objet JSON décodé qui porte le champ
     * @param string       $cle     nom du champ, tel que le format l'écrit
     * @param string       $chemin  chemin de l'objet dans le calque, vide à la racine
     * @param bool         $requise refuser la liste absente plutôt que la rendre vide
     *
     * @return list<mixed>
     *
     * @throws CalqueInvalide
     */
    private static function liste(array $donnees, string $cle, string $chemin, bool $requise): array
    {
        if (!$requise && !array_key_exists($cle, $donnees)) {
            return [];
        }

        $valeur = self::requis($donnees, $cle, $chemin);
        if (!is_array($valeur) || !array_is_list($valeur)) {
            throw self::attendu($chemin, $cle, 'liste attendue');
        }

        return $valeur;
    }

    /**
     * Rend la valeur d'un champ qui doit être présent.
     *
     * Un null explicite compte comme absent : le binaire omet les champs vides,
     * il n'écrit jamais null.
     *
     * @param array<mixed> $donnees objet JSON décodé qui porte le champ
     * @param string       $cle     nom du champ, tel que le format l'écrit
     * @param string       $chemin  chemin de l'objet dans le calque, vide à la racine
     *
     * @throws CalqueInvalide
     */
    private static function requis(array $donnees, string $cle, string $chemin): mixed
    {
        if (!isset($donnees[$cle])) {
            throw new CalqueInvalide(sprintf('Champ %s absent du calque', self::champ($chemin, $cle)));
        }

        return $donnees[$cle];
    }

    /**
     * Fabrique l'exception d'un champ du mauvais type. L'accord se fait chez
     * l'appelant : « chaîne attendue », « entier attendu ».
     */
    private static function attendu(string $chemin, string $cle, string $attendu): CalqueInvalide
    {
        return new CalqueInvalide(sprintf('Champ %s : %s', self::champ($chemin, $cle), $attendu));
    }

    /**
     * Qualifie une clé par le chemin de son objet.
     */
    private static function champ(string $chemin, string $cle): string
    {
        return $chemin === '' ? $cle : $chemin . '.' . $cle;
    }
}
