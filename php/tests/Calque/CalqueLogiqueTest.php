<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\Calque;

use Ormeau\Doctrine\Calque\CalqueInvalide;
use Ormeau\Doctrine\Calque\CalqueLogique;
use PHPUnit\Framework\Attributes\CoversClass;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;

/**
 * Ce qui se joue ici est la distinction entre champ requis et champ optionnel
 * du format : la confondre ferait échouer la génération sur un calque
 * parfaitement valide, ou l'inverse.
 */
#[CoversClass(CalqueLogique::class)]
final class CalqueLogiqueTest extends TestCase
{
    /**
     * Les champs optionnels absents valent le tableau vide, pas null : la
     * génération itère dessus sans avoir à s'en méfier.
     */
    public function testLesChampsOptionnelsValentLeTableauVide(): void
    {
        $calque = CalqueLogique::depuisTableau([
            'version_ri' => 1,
            'empreinte_physique' => 'sha256:' . str_repeat('c', 64),
            'espace_de_noms' => 'App\\Entity',
            'entites' => [],
        ]);

        self::assertSame([], $calque->enumerations);
        self::assertSame([], $calque->avertissements);
    }

    /**
     * Un champ requis manquant est nommé dans le message : sur un calque de
     * plusieurs milliers de lignes, « champ absent » sans plus ne sert à rien.
     */
    #[DataProvider('champsRequis')]
    public function testRefuseUnChampRequisAbsent(string $champ): void
    {
        $donnees = [
            'version_ri' => 1,
            'empreinte_physique' => 'sha256:' . str_repeat('d', 64),
            'espace_de_noms' => 'App\\Entity',
            'entites' => [],
        ];
        unset($donnees[$champ]);

        $this->expectException(CalqueInvalide::class);
        $this->expectExceptionMessageMatches('/' . preg_quote($champ, '/') . '/');

        CalqueLogique::depuisTableau($donnees);
    }

    /**
     * @return iterable<string, array{string}>
     */
    public static function champsRequis(): iterable
    {
        yield 'empreinte' => ['empreinte_physique'];
        yield 'espace de noms' => ['espace_de_noms'];
        yield 'entites' => ['entites'];
    }

    /**
     * Un champ optionnel ajouté au format n'incrémente pas version_ri, donc un
     * calque qui le porte doit se lire tel quel. C'est ce que ce test vérifie
     * sur origine, le dernier ajouté : un lecteur qui filtrerait les clés
     * connues le perdrait en silence.
     */
    public function testUnChampOptionnelRecentEstConserve(): void
    {
        $calque = CalqueLogique::depuisTableau([
            'version_ri' => 1,
            'empreinte_physique' => 'sha256:' . str_repeat('1', 64),
            'espace_de_noms' => 'App\\Entity',
            'entites' => [[
                'nom' => 'Client',
                'table' => ['nom' => 't_clients', 'schema' => 'public'],
                'proprietes' => [],
                'origine' => 'nommage',
            ]],
        ]);

        self::assertSame('nommage', $calque->entites[0]['origine']);
    }

    /**
     * Le filtre sur la confiance sert à la CI : elle échoue sur les
     * avertissements sûrs et laisse passer les incertains.
     */
    public function testFiltreLesAvertissementsParConfiance(): void
    {
        $calque = CalqueLogique::depuisTableau([
            'version_ri' => 1,
            'empreinte_physique' => 'sha256:' . str_repeat('e', 64),
            'espace_de_noms' => 'App\\Entity',
            'entites' => [],
            'avertissements' => [
                ['code' => 'table_sans_cle_primaire', 'confiance' => 1.0],
                ['code' => 'fk_implicite_probable', 'confiance' => 0.6],
                ['code' => 'type_non_reconnu', 'confiance' => 0.3],
            ],
        ]);

        $retenus = $calque->avertissementsAuDessusDe(0.6);

        self::assertCount(2, $retenus);
        self::assertSame('table_sans_cle_primaire', $retenus[0]['code']);
        self::assertSame('fk_implicite_probable', $retenus[1]['code']);
    }

    /**
     * Le résultat est une liste réindexée : les trous laissés par array_filter
     * casseraient un json_encode, qui rendrait un objet au lieu d'un tableau.
     */
    public function testLeFiltreRendUneListeReindexee(): void
    {
        $calque = CalqueLogique::depuisTableau([
            'version_ri' => 1,
            'empreinte_physique' => 'sha256:' . str_repeat('f', 64),
            'espace_de_noms' => 'App\\Entity',
            'entites' => [],
            'avertissements' => [
                ['code' => 'ecarte', 'confiance' => 0.1],
                ['code' => 'retenu', 'confiance' => 0.9],
            ],
        ]);

        self::assertSame([0], array_keys($calque->avertissementsAuDessusDe(0.5)));
    }

    /**
     * Un avertissement sans confiance vaut zéro : il n'est jamais retenu par un
     * seuil, plutôt que de faire échouer une CI sur un champ manquant.
     */
    public function testUnAvertissementSansConfianceNestPasRetenu(): void
    {
        $calque = CalqueLogique::depuisTableau([
            'version_ri' => 1,
            'empreinte_physique' => 'sha256:' . str_repeat('0', 64),
            'espace_de_noms' => 'App\\Entity',
            'entites' => [],
            'avertissements' => [['code' => 'sans_confiance']],
        ]);

        self::assertSame([], $calque->avertissementsAuDessusDe(0.1));
    }
}
