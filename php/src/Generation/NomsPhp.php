<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

/**
 * Dit si un nom lu dans le calque logique peut être écrit là où la génération
 * l'écrit : déclaration de classe, cas d'énumération, propriété, espace de
 * noms. Le calque ne porte plus de type PHP : la génération le déduit.
 *
 * Le calque logique n'est pas une source sûre. Il se modifie à la main, vient
 * d'une autre version de l'outil, se recopie d'un autre projet ; et les noms
 * qu'il porte viennent en partie d'un fichier de décisions reçu par une pull
 * request. Un nom recopié tel quel deviendrait une ligne de code et un chemin
 * de fichier. L'inférence refuse déjà ces noms côté Go ; ce contrôle ne s'y fie
 * pas.
 *
 * Les mots réservés ont été relevés en passant chaque mot à php -l sous PHP
 * 8.1 et 8.4, et NomsPhpTest les repasse à la version de PHP qui l'exécute :
 * enum, resource et numeric, réservés selon la documentation, nomment une
 * classe sans erreur. La liste Go de l'inférence est la même.
 *
 * @internal
 */
final class NomsPhp
{
    /**
     * Noms refusés pour une classe, une énumération ou un trait, en minuscules :
     * PHP les compare sans la casse. __property__ n'est refusé qu'à partir de
     * la 8.4, et retenu, le code produit devant se charger sur toute la plage.
     */
    public const RESERVES_DE_CLASSE = [
        '__class__', '__dir__', '__file__', '__function__', '__halt_compiler', '__line__', '__method__',
        '__namespace__', '__property__', '__trait__', 'abstract', 'and', 'array', 'as', 'bool', 'break',
        'callable', 'case', 'catch', 'class', 'clone', 'const', 'continue', 'declare', 'default', 'die', 'do',
        'echo', 'else', 'elseif', 'empty', 'enddeclare', 'endfor', 'endforeach', 'endif', 'endswitch',
        'endwhile', 'eval', 'exit', 'extends', 'false', 'final', 'finally', 'float', 'fn', 'for', 'foreach',
        'function', 'global', 'goto', 'if', 'implements', 'include', 'include_once', 'instanceof', 'insteadof',
        'int', 'interface', 'isset', 'iterable', 'list', 'match', 'mixed', 'namespace', 'never', 'new', 'null',
        'object', 'or', 'parent', 'print', 'private', 'protected', 'public', 'readonly', 'require',
        'require_once', 'return', 'self', 'static', 'string', 'switch', 'throw', 'trait', 'true', 'try',
        'unset', 'use', 'var', 'void', 'while', 'xor', 'yield',
    ];

    /** Seuls noms qu'un cas d'énumération ne peut pas porter. */
    public const RESERVES_DE_CAS = ['class', '__halt_compiler'];

    /** Seuls noms qu'un segment d'espace de noms refuse. */
    public const RESERVES_D_ESPACE = ['namespace', '__halt_compiler'];

    /**
     * Forme d'un identifiant, octet par octet comme le lexer de PHP : un nom
     * accentué passe, un point, une barre oblique, un espace ou un
     * point-virgule non.
     */
    private const IDENTIFIANT = '/^[a-zA-Z_\x80-\xff][a-zA-Z0-9_\x80-\xff]*$/D';

    /**
     * Dit pourquoi un nom ne peut pas nommer une classe, une énumération ou un
     * trait, ou null quand il le peut.
     */
    public static function raisonClasse(string $nom): ?string
    {
        return self::raison($nom, self::RESERVES_DE_CLASSE);
    }

    /**
     * Dit pourquoi un nom ne peut pas nommer un cas d'énumération, ou null.
     */
    public static function raisonCas(string $nom): ?string
    {
        return self::raison($nom, self::RESERVES_DE_CAS);
    }

    /**
     * Dit pourquoi un nom ne peut pas nommer une propriété ou une association,
     * ou null. PHP accepte tout mot réservé à cette place : seule la forme
     * compte.
     */
    public static function raisonMembre(string $nom): ?string
    {
        return self::raison($nom, []);
    }

    /**
     * Dit pourquoi un espace de noms est refusé, ou null.
     */
    public static function raisonEspaceDeNoms(string $espace): ?string
    {
        foreach (explode('\\', $espace) as $segment) {
            $raison = self::raison($segment, self::RESERVES_D_ESPACE);
            if ($raison !== null) {
                return sprintf('« %s » n\'est pas un espace de noms PHP : %s', $espace, $raison);
            }
        }

        return null;
    }

    /**
     * Dit pourquoi un nom est refusé : forme, puis mot réservé à cette place.
     *
     * @param list<string> $reserves mots refusés à cette place, en minuscules
     */
    private static function raison(string $nom, array $reserves): ?string
    {
        if (preg_match(self::IDENTIFIANT, $nom) !== 1) {
            return sprintf('« %s » n\'est pas un identifiant PHP', $nom);
        }
        if (in_array(strtolower($nom), $reserves, true)) {
            return sprintf('%s est un mot réservé de PHP', $nom);
        }

        return null;
    }
}
