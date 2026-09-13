<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

/**
 * Écrit les fragments de PHP que le rendu assemble : littéraux, attributs et
 * docblocks.
 *
 * Un émetteur direct plutôt que l'imprimante de nikic/php-parser : celle-ci
 * n'écrit aucune ligne vide entre les membres, rend declare (strict_types=1)
 * avec une espace, et laisse un attribut de 130 colonnes sur une ligne. Le code
 * produit doit se lire comme du code écrit à la main, et son format ne doit
 * dépendre d'aucune version de bibliothèque : une mise à jour de php-parser ne
 * doit pas faire bouger le diff des entités d'un client.
 *
 * @internal
 */
final class Emetteur
{
    /** Au-delà, un attribut passe ses arguments à la ligne, un par ligne. */
    public const LARGEUR = 120;

    /** Une indentation, selon PER-CS. */
    public const INDENTATION = '    ';

    /**
     * Écrit un docblock : sur une ligne quand il n'en a qu'une, en bloc sinon,
     * le texte séparé des étiquettes par une ligne vide.
     *
     * Rien ne s'écrit sans texte ni étiquette : un docblock vide n'apprend
     * rien.
     *
     * @param list<string> $texte       lignes de texte ; ce qui vient de la base passe d'abord par
     *                                  commentaire()
     * @param list<string> $etiquettes  étiquettes composées par le rendu, @var ou @return
     * @param string       $indentation indentation des lignes du docblock
     *
     * @return list<string> les lignes du docblock, aucune quand il n'y a rien à dire
     */
    public static function docblock(array $texte, array $etiquettes, string $indentation): array
    {
        $lignes = $texte !== [] && $etiquettes !== [] ? [...$texte, '', ...$etiquettes] : [...$texte, ...$etiquettes];
        if (count($lignes) <= 1) {
            return array_map(static fn(string $l): string => $indentation . '/** ' . $l . ' */', $lignes);
        }

        return [
            $indentation . '/**',
            ...array_map(static fn(string $l): string => rtrim($indentation . ' * ' . $l), $lignes),
            $indentation . ' */',
        ];
    }

    /**
     * Prépare un commentaire venu de la base pour un docblock.
     *
     * Le commentaire est une donnée, écrite par n'importe qui ayant accès au
     * schéma : une étoile suivie d'une barre oblique y fermerait le docblock,
     * et la suite s'exécuterait comme du code dans le fichier produit — une
     * injection. La barre oblique est donc échappée. La valeur exacte reste
     * dans options: ['comment' => …], qui passe par un littéral échappé.
     *
     * Le premier paragraphe finit par une ponctuation : sinon PHP-CS-Fixer, en
     * règles @Symfony, lui ajoute un point dans le projet, et chaque
     * régénération le retire — un diff à chaque passage sur des fichiers que
     * personne n'a touchés.
     *
     * @return list<string> les lignes du commentaire, sans lignes vides en tête ni en fin
     */
    public static function commentaire(string $brut): array
    {
        $lignes = array_map('rtrim', preg_split('/\R/u', str_replace('*/', '*\/', $brut)) ?: []);
        while ($lignes !== [] && $lignes[0] === '') {
            array_shift($lignes);
        }
        while ($lignes !== [] && end($lignes) === '') {
            array_pop($lignes);
        }
        if ($lignes === []) {
            return [];
        }

        $fin = array_search('', $lignes, true);
        $fin = $fin === false ? count($lignes) - 1 : $fin - 1;
        if (preg_match('/[.!?]$/u', $lignes[$fin]) !== 1) {
            $lignes[$fin] .= '.';
        }

        return array_values($lignes);
    }

    /**
     * Écrit une valeur PHP sous forme de littéral.
     *
     * Les chaînes sont entre apostrophes, seules l'apostrophe et la barre
     * oblique inverse s'échappent. Un flottant garde son point décimal, sans
     * quoi 1.0 se relirait en entier.
     *
     * Un Code s'écrit tel quel : c'est le seul chemin pour une expression.
     *
     * @param array<array-key, mixed>|bool|Code|float|int|string|null $valeur
     */
    public static function litteral(array|bool|Code|float|int|string|null $valeur): string
    {
        return match (true) {
            $valeur === null => 'null',
            $valeur instanceof Code => $valeur->php,
            is_bool($valeur) => $valeur ? 'true' : 'false',
            is_int($valeur) => (string) $valeur,
            is_float($valeur) => self::flottant($valeur),
            is_string($valeur) => "'" . strtr($valeur, ['\\' => '\\\\', "'" => "\\'"]) . "'",
            default => self::tableau($valeur),
        };
    }

    /**
     * Écrit un attribut, sur une ligne quand il tient dans la largeur, avec un
     * argument par ligne sinon.
     *
     * Les arguments sont nommés, dans l'ordre donné : c'est ce qui rend le code
     * stable d'une régénération à l'autre, et lisible sans connaître l'ordre
     * des paramètres du constructeur de l'attribut. Seul un attribut à valeur
     * unique, comme #[ORM\InheritanceType('JOINED')], prend un argument
     * positionnel, écrit comme on l'écrirait à la main.
     *
     * @param string                   $nom         nom de l'attribut, tel qu'il s'écrit après #[
     * @param array<array-key, mixed>  $arguments   arguments nommés, ou positionnels sous une clé
     *                                              entière ; un argument null est omis
     * @param string                   $indentation indentation de la ligne de l'attribut
     */
    public static function attribut(string $nom, array $arguments, string $indentation): string
    {
        $parties = [];
        foreach ($arguments as $cle => $valeur) {
            if ($valeur === null) {
                continue;
            }
            /** @var array<array-key, mixed>|bool|Code|float|int|string $valeur */
            $parties[] = (is_int($cle) ? '' : $cle . ': ') . self::litteral($valeur);
        }

        if ($parties === []) {
            return $indentation . '#[' . $nom . ']';
        }

        $ligne = $indentation . '#[' . $nom . '(' . implode(', ', $parties) . ')]';
        if (strlen($ligne) <= self::LARGEUR) {
            return $ligne;
        }

        $interieur = $indentation . self::INDENTATION;

        return $indentation . '#[' . $nom . "(\n"
            . $interieur . implode(",\n" . $interieur, $parties) . ",\n"
            . $indentation . ')]';
    }

    /**
     * Écrit un tableau court : liste sans clés, ou tableau associatif.
     *
     * @param array<array-key, mixed> $tableau
     */
    private static function tableau(array $tableau): string
    {
        $elements = [];
        foreach ($tableau as $cle => $valeur) {
            /** @var array<array-key, mixed>|bool|Code|float|int|string|null $valeur */
            $elements[] = array_is_list($tableau)
                ? self::litteral($valeur)
                : self::litteral($cle) . ' => ' . self::litteral($valeur);
        }

        return '[' . implode(', ', $elements) . ']';
    }

    /**
     * Écrit un flottant sans perte, avec son point décimal.
     */
    private static function flottant(float $valeur): string
    {
        $texte = var_export($valeur, true);

        return str_contains($texte, '.') || str_contains($texte, 'E') ? $texte : $texte . '.0';
    }
}
