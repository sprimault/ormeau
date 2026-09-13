<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\Calque;

use Ormeau\Doctrine\Calque\CalqueLogique;
use Ormeau\Doctrine\Calque\LecteurCalque;
use PHPUnit\Framework\Attributes\CoversClass;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;
use RuntimeException;

/**
 * Chaque calque logique attendu de l'inférence se lit en PHP.
 *
 * Ce sont les fichiers que le binaire produit réellement, figés par les tests
 * Go : un champ que le binaire écrit et que le lecteur refuse se voit ici, pas
 * chez l'intégrateur qui lance la génération sur sa base.
 */
#[CoversClass(LecteurCalque::class)]
#[CoversClass(CalqueLogique::class)]
final class ReferenceTest extends TestCase
{
    /** Répertoire des cas d'inférence, à la racine du dépôt. */
    private const REFERENCES = __DIR__ . '/../../../tests/reference/inference';

    /**
     * Le calque se lit, et rien n'y est perdu en route : autant d'entités, de
     * propriétés et d'associations que le JSON en porte.
     */
    #[DataProvider('cas')]
    public function testLitLeCalqueDeReference(string $chemin): void
    {
        $calque = (new LecteurCalque())->lire($chemin);
        $brut = json_decode((string) file_get_contents($chemin), true, 64, JSON_THROW_ON_ERROR);
        self::assertIsArray($brut);
        self::assertIsArray($brut['entites']);

        self::assertCount(count($brut['entites']), $calque->entites);
        foreach ($calque->entites as $rang => $entite) {
            $attendue = $brut['entites'][$rang];
            self::assertIsArray($attendue);
            self::assertSame($attendue['nom'], $entite->nom);
            self::assertCount(count(self::liste($attendue, 'proprietes')), $entite->proprietes);
            self::assertCount(count(self::liste($attendue, 'associations')), $entite->associations);
        }
        self::assertCount(count(self::liste($brut, 'enumerations')), $calque->enumerations);
        self::assertCount(count(self::liste($brut, 'traits')), $calque->traits);
        self::assertCount(count(self::liste($brut, 'avertissements')), $calque->avertissements);
    }

    /**
     * Un cas par répertoire de référence.
     *
     * Un répertoire introuvable échoue plutôt que de rendre une liste vide :
     * la suite passerait en vert sans avoir rien lu.
     *
     * @return iterable<string, array{string}>
     */
    public static function cas(): iterable
    {
        $fichiers = glob(self::REFERENCES . '/*/logique.json');
        if ($fichiers === false || $fichiers === []) {
            throw new RuntimeException('Aucun calque logique de référence sous ' . self::REFERENCES);
        }

        foreach ($fichiers as $fichier) {
            yield basename(dirname($fichier)) => [$fichier];
        }
    }

    /**
     * Rend la liste d'un champ du JSON brut, vide quand il est absent.
     *
     * @param array<mixed> $donnees objet du JSON brut qui porte la liste
     *
     * @return array<mixed>
     */
    private static function liste(array $donnees, string $cle): array
    {
        $valeur = $donnees[$cle] ?? [];
        self::assertIsArray($valeur);

        return $valeur;
    }
}
