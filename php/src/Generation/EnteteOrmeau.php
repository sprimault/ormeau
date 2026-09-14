<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

/**
 * Écrit et relit la ligne d'en-tête des fichiers produits, qui dit à qui le
 * fichier appartient et de quelle base il vient.
 *
 * Le format vit ici seul : l'écriture et la relecture qui protège d'un
 * écrasement ne doivent pas diverger, sans quoi un fichier de la base voisine
 * ne serait plus reconnu et se réécrirait en silence. La ligne ne porte ni
 * version ni date, qui changeraient chaque fichier à chaque mise à jour de
 * l'outil.
 *
 * @internal
 */
final class EnteteOrmeau
{
    /** Début commun à tout en-tête qui nomme sa base. */
    private const DEPUIS = 'par Ormeau depuis la base ';

    /** Suite de l'en-tête d'un fichier réécrit à chaque génération. */
    private const REECRIT = ' et réécrit à chaque génération';

    /** Suite de l'en-tête de la classe de l'utilisateur, écrite une fois. */
    private const JAMAIS_REECRIT = ', jamais réécrit : ce fichier appartient au projet.';

    /**
     * Rend l'en-tête d'une énumération ou d'un trait, réécrits à chaque
     * génération.
     */
    public static function outil(string $base): string
    {
        return 'Généré ' . self::DEPUIS . $base . self::REECRIT . '.';
    }

    /**
     * Rend l'en-tête d'une classe de base, qui dit aussi où va le code propre
     * à l'entité.
     */
    public static function classeBase(string $base, string $entite): string
    {
        return sprintf('Généré %s%s%s : le code propre à %s va dans %s.php.', self::DEPUIS, $base, self::REECRIT, $entite, $entite);
    }

    /**
     * Rend l'en-tête de la classe de l'utilisateur : la garantie du projet,
     * dite là où on la lit.
     */
    public static function utilisateur(string $base): string
    {
        return 'Créé ' . self::DEPUIS . $base . self::JAMAIS_REECRIT;
    }

    /**
     * Rend la base que nomme l'en-tête d'un source produit, ou null.
     *
     * Seule la ligne d'en-tête est lue — la troisième, après <?php et une
     * ligne vide —, jamais une ligne qui contiendrait « Ormeau » ailleurs :
     * un commentaire venu de la base peut porter le mot. Un fichier de la
     * 0.5.0, dont l'en-tête ne nomme pas de base, rend null.
     */
    public static function base(string $source): ?string
    {
        $ligne = explode("\n", $source, 4)[2] ?? '';
        $ligne = rtrim($ligne, "\r");
        foreach (['// Généré ', '// Créé '] as $debut) {
            if (!str_starts_with($ligne, $debut . self::DEPUIS)) {
                continue;
            }
            $reste = substr($ligne, strlen($debut . self::DEPUIS));
            foreach ([self::REECRIT, self::JAMAIS_REECRIT] as $suite) {
                $fin = strrpos($reste, $suite);
                if ($fin !== false && $fin > 0) {
                    return substr($reste, 0, $fin);
                }
            }
        }

        return null;
    }
}
