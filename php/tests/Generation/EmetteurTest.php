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
     * Un commentaire de base qui ferme le docblock n'injecte pas de code : la
     * barre oblique qui suit l'étoile est échappée, et le docblock produit se
     * relit comme un seul commentaire.
     */
    public function testUnCommentaireDeBaseNeFermePasLeDocblock(): void
    {
        $docblock = implode("\n", Emetteur::docblock(Emetteur::commentaire("Fin */ system('id'); /* reprise"), [], ''));

        $jetons = array_values(array_filter(
            token_get_all("<?php\n" . $docblock . "\nclass A {}\n"),
            static fn($j): bool => is_array($j) && !in_array($j[0], [T_OPEN_TAG, T_WHITESPACE], true),
        ));
        self::assertSame(T_DOC_COMMENT, $jetons[0][0], $docblock);
        self::assertSame('class', $jetons[1][1], $docblock);
    }

    /**
     * Le premier paragraphe d'un commentaire finit par une ponctuation, que
     * PHP-CS-Fixer ajouterait sinon à chaque régénération ; la suite garde ses
     * lignes, sans lignes vides en tête ni en fin.
     */
    public function testUnCommentaireFinitSonResumeParUnePonctuation(): void
    {
        self::assertSame(['Dénomination légale.'], Emetteur::commentaire('Dénomination légale'));
        self::assertSame(['Déjà ponctué ?'], Emetteur::commentaire('Déjà ponctué ?'));
        self::assertSame(
            ['Fiche client.', '', 'Tenue par la comptabilité'],
            Emetteur::commentaire("\nFiche client\r\n\nTenue par la comptabilité  \n"),
        );
        self::assertSame([], Emetteur::commentaire("\n \n"));
    }

    /**
     * Un docblock d'une ligne tient sur une ligne ; texte et étiquettes sont
     * séparés par une ligne vide, sans espace en fin de ligne.
     */
    public function testUnDocblockSeRendSurUneLigneOuEnBloc(): void
    {
        self::assertSame([], Emetteur::docblock([], [], '    '));
        self::assertSame(['    /** @var array<mixed> */'], Emetteur::docblock([], ['@var array<mixed>'], '    '));
        self::assertSame(
            ['    /**', '     * Réglages.', '     *', '     * @var array<mixed>', '     */'],
            Emetteur::docblock(['Réglages.'], ['@var array<mixed>'], '    '),
        );
    }
}
