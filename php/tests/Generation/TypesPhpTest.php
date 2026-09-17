<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\Generation;

use Ormeau\Doctrine\Calque\Propriete;
use Ormeau\Doctrine\Generation\Cible;
use Ormeau\Doctrine\Generation\TypesPhp;
use PHPUnit\Framework\Attributes\CoversClass;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;

/**
 * Le type PHP qu'hydrate chaque type Doctrine, et ce qui change avec la cible.
 */
#[CoversClass(TypesPhp::class)]
final class TypesPhpTest extends TestCase
{
    /**
     * bigint et binary sont les deux types dont l'hydratation change entre
     * DBAL 3 et DBAL 4, et c'est DBAL qui décide, pas ORM : sous ORM 3 avec
     * DBAL 3, un int ferait croire l'entité modifiée à chaque flush — DBAL rend
     * "9000000000", la propriété garde 9000000000 — et un string refuserait la
     * ressource que DBAL rend pour binary, au chargement.
     */
    #[DataProvider('typesQuiDependentDeLaCible')]
    public function testLeTypeSuitDbal(string $typeDoctrine, string $avantDbal4, string $dbal4): void
    {
        $propriete = self::propriete($typeDoctrine);

        self::assertSame($avantDbal4, TypesPhp::declaration($propriete, Cible::forcer(2)));
        self::assertSame($avantDbal4, TypesPhp::declaration($propriete, Cible::forcer(2, '2.13')));
        self::assertSame($avantDbal4, TypesPhp::declaration($propriete, Cible::forcer(3, '3.10')));
        self::assertSame($dbal4, TypesPhp::declaration($propriete, Cible::forcer(3)));
    }

    /**
     * Les types qui changent avec DBAL 4.
     *
     * @return iterable<string, array{string, string, string}>
     */
    public static function typesQuiDependentDeLaCible(): iterable
    {
        yield 'bigint' => ['bigint', 'string', 'int'];
        yield 'binary' => ['binary', 'mixed', 'string'];
    }

    /**
     * Une propriété facultative prend « ? », sauf mixed, qui contient déjà
     * null et refuse la syntaxe.
     */
    public function testLaNullabiliteSaufPourMixed(): void
    {
        self::assertSame('?string', TypesPhp::declaration(self::propriete('text', true), Cible::forcer(3)));
        self::assertSame('mixed', TypesPhp::declaration(self::propriete('blob', true), Cible::forcer(3)));
    }

    /**
     * Hors table — un type personnalisé forcé par décision —, mixed, sans
     * « ? » : rien ne dit ce que le type hydrate.
     */
    public function testUnTypeInconnuSeDeclareMixed(): void
    {
        $propriete = self::propriete('geometrie_maison', true);

        self::assertFalse(TypesPhp::connu('geometrie_maison'));
        self::assertSame('mixed', TypesPhp::declaration($propriete, Cible::forcer(2)));
    }

    /**
     * Tout type Doctrine qu'écrit l'inférence dans un calque de référence est
     * connu du générateur. C'est la parité entre la table Go et celle-ci : un
     * type ajouté côté Go sans son entrée ici serait déclaré mixed. Un type
     * forcé par décision en est exclu : un type personnalisé n'a pas à être
     * connu, c'est le cas même de mixed.
     */
    public function testConnaitChaqueTypeDesCalquesDeReference(): void
    {
        $fichiers = glob(Repertoires::REFERENCES . '/*/logique.json');
        self::assertIsArray($fichiers);
        self::assertNotEmpty($fichiers);

        $inconnus = [];
        foreach ($fichiers as $fichier) {
            $calque = json_decode((string) file_get_contents($fichier), true, 64, JSON_THROW_ON_ERROR);
            self::assertIsArray($calque);
            foreach ([...$calque['entites'] ?? [], ...$calque['traits'] ?? []] as $porteur) {
                foreach ($porteur['proprietes'] ?? [] as $propriete) {
                    if (($propriete['origine'] ?? null) !== 'decision' && !TypesPhp::connu($propriete['type_doctrine'])) {
                        $inconnus[$propriete['type_doctrine']] = basename(dirname($fichier));
                    }
                }
            }
        }

        self::assertSame([], $inconnus);
    }

    /**
     * Fabrique une propriété réduite à ce que le typage regarde.
     */
    private static function propriete(string $typeDoctrine, bool $nullable = false): Propriete
    {
        return new Propriete('valeur', 'valeur', $typeDoctrine, $nullable);
    }
}
