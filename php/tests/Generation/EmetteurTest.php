<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\Generation;

use Ormeau\Doctrine\Generation\Emetteur;
use Ormeau\Doctrine\Generation\RenduEntite;
use PHPUnit\Framework\Attributes\CoversClass;
use PHPUnit\Framework\TestCase;

/**
 * Ce que les attendus ne montrent pas : les valeurs piégeuses qu'aucun calque
 * de référence ne porte, et qui produiraient un fichier PHP invalide.
 */
#[CoversClass(Emetteur::class)]
#[CoversClass(RenduEntite::class)]
final class EmetteurTest extends TestCase
{
    /**
     * Un commentaire de colonne peut contenir une apostrophe ou une barre
     * oblique inverse : le littéral produit doit se relire à l'identique.
     */
    public function testUnLitteralSeRelitALIdentique(): void
    {
        foreach (["l'adresse", 'C:\\chemin\\', "fin \\'", '', 'é'] as $chaine) {
            $litteral = Emetteur::litteral($chaine);
            self::assertSame($chaine, eval('return ' . $litteral . ';'), $litteral);
        }
    }

    /**
     * Un flottant garde son point : 1.0 relu en PHP ne doit pas devenir un
     * entier, qu'une propriété float refuserait sous strict_types.
     */
    public function testUnFlottantGardeSonPoint(): void
    {
        self::assertSame('1.0', Emetteur::litteral(1.0));
        self::assertSame('0.5', Emetteur::litteral(0.5));
    }

    /**
     * Un attribut qui dépasse la largeur passe ses arguments à la ligne, un par
     * ligne, avec une virgule finale ; un argument null est omis.
     */
    public function testUnAttributLongPasseALaLigne(): void
    {
        $attribut = Emetteur::attribut('ORM\Column', [
            'name' => 'une_colonne_au_nom_particulierement_long',
            'type' => 'decimal',
            'precision' => 14,
            'scale' => null,
            'options' => ['comment' => 'Un commentaire assez long pour dépasser la largeur'],
        ], '    ');

        self::assertSame(<<<'PHP'
                #[ORM\Column(
                    name: 'une_colonne_au_nom_particulierement_long',
                    type: 'decimal',
                    precision: 14,
                    options: ['comment' => 'Un commentaire assez long pour dépasser la largeur'],
                )]
            PHP, $attribut);
    }

    /**
     * Seul un identifiant SQL simple en minuscules reste nu : une majuscule, un
     * espace ou un accent imposent les backticks, sans quoi PostgreSQL
     * chercherait une autre table.
     */
    public function testCiteCeQuiNestPasUnIdentifiantSimple(): void
    {
        self::assertSame('client_id', RenduEntite::identifiantSql('client_id'));
        self::assertSame('`T_CLIENTS`', RenduEntite::identifiantSql('T_CLIENTS'));
        self::assertSame('`N° Commande`', RenduEntite::identifiantSql('N° Commande'));
        self::assertSame('`1er`', RenduEntite::identifiantSql('1er'));
    }
}
