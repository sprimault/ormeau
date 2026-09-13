<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\Generation;

use LogicException;
use Ormeau\Doctrine\Calque\CalqueLogique;
use Ormeau\Doctrine\Generation\Cible;
use Ormeau\Doctrine\Generation\GenerateurEntite;
use Ormeau\Doctrine\Generation\ModeRegeneration;
use PHPUnit\Framework\Attributes\CoversClass;
use PHPUnit\Framework\TestCase;

/**
 * Le mode de régénération, et ce qui fait écarter une entité plutôt que la
 * générer à moitié. Le rendu lui-même est figé par AttendusTest, et validé par
 * Doctrine dans MappingTest.
 */
#[CoversClass(GenerateurEntite::class)]
final class GenerateurEntiteTest extends TestCase
{
    /**
     * La classe de base est le mode par défaut : le changer serait une rupture
     * pour tout projet déjà généré, et rien d'autre ne le signalerait.
     */
    public function testLeModeParDefautEstLaClasseDeBase(): void
    {
        self::assertSame(ModeRegeneration::ClasseDeBase, (new GenerateurEntite())->mode());
    }

    /**
     * Les valeurs de l'énumération sont un contrat : elles viennent d'un drapeau
     * de commande et d'une configuration Symfony, les renommer casse les deux.
     */
    public function testLesValeursDuModeSontStables(): void
    {
        self::assertSame('classe_de_base', ModeRegeneration::ClasseDeBase->value);
        self::assertSame('reecriture_ast', ModeRegeneration::ReecritureAst->value);
    }

    /**
     * La réécriture par AST n'est pas écrite : elle échoue franchement, avant
     * d'écrire quoi que ce soit, plutôt que de retomber sur l'autre mode.
     */
    public function testLaReecritureParAstEchoueSansEcrire(): void
    {
        $sortie = Repertoires::creer();
        try {
            (new GenerateurEntite(ModeRegeneration::ReecritureAst))->generer(self::calque([]), $sortie, Cible::forcer(3));
            self::fail('la réécriture par AST a généré');
        } catch (LogicException $e) {
            self::assertStringContainsString('reecriture_ast', $e->getMessage());
            self::assertSame([], Repertoires::lire($sortie));
        } finally {
            Repertoires::supprimer($sortie);
        }
    }

    /**
     * Une entité que Doctrine refuserait, ou que ce paquet ne sait pas encore
     * générer entière, est écartée avec sa raison, et les autres s'écrivent.
     */
    public function testEcarteCeQuiNeSeGenerePasEntierEtEcritLeReste(): void
    {
        $sortie = Repertoires::creer();
        try {
            $rapport = (new GenerateurEntite())->generer(self::calque([
                self::entite('Client'),
                self::entite('Journal', ['identifiant' => null]),
                self::entite('List'),
                self::entite('Commande', ['associations' => [[
                    'nom' => 'client', 'genre' => 'plusieurs_vers_un', 'cible' => 'Client', 'proprietaire' => true, 'origine' => 'contrainte',
                ]]]),
                self::entite('Etiquette', ['table' => ['nom' => 'etiquette', 'schema' => 'public']]),
                self::entite('etiquette', ['table' => ['nom' => 'etiquette', 'schema' => 'archive']]),
            ]), $sortie, Cible::forcer(3));

            self::assertSame([
                'Journal' => 'la table n\'a pas de clé primaire, et Doctrine exige un identifiant',
                'List' => 'List est un mot réservé de PHP, à renommer dans renommages',
                'Commande' => 'pas encore générées par ce paquet : associations',
                'Etiquette' => 'le nom Etiquette est porté par plusieurs entités, à départager dans renommages',
                'etiquette' => 'le nom etiquette est porté par plusieurs entités, à départager dans renommages',
            ], array_column(array_map(static fn($e): array => [$e->nom, $e->raison], $rapport->ecartees), 1, 0));
            self::assertSame(['Base/ClientBase.php', 'Client.php'], array_keys(Repertoires::lire($sortie)));
        } finally {
            Repertoires::supprimer($sortie);
        }
    }

    /**
     * Un calque réduit aux entités données.
     *
     * @param list<array<string, mixed>> $entites
     */
    private static function calque(array $entites): CalqueLogique
    {
        return CalqueLogique::depuisTableau([
            'version_ri' => 1,
            'empreinte_physique' => 'sha256:' . str_repeat('b', 64),
            'espace_de_noms' => 'App\\Entity',
            'entites' => $entites,
        ]);
    }

    /**
     * Une entité à clé IDENTITY, modifiable par les champs donnés ; un champ
     * à null est retiré.
     *
     * @param array<string, mixed> $modifications
     *
     * @return array<string, mixed>
     */
    private static function entite(string $nom, array $modifications = []): array
    {
        $entite = array_merge([
            'nom' => $nom,
            'table' => ['nom' => strtolower($nom), 'schema' => 'public'],
            'identifiant' => ['proprietes' => ['id'], 'strategie' => 'identite'],
            'proprietes' => [['nom' => 'id', 'colonne' => 'id', 'type_php' => 'int', 'type_doctrine' => 'integer', 'nullable' => false]],
        ], $modifications);

        return array_filter($entite, static fn($valeur): bool => $valeur !== null);
    }
}
