<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\Calque;

use BackedEnum;
use Ormeau\Doctrine\Calque\ActionSuppression;
use Ormeau\Doctrine\Calque\Association;
use Ormeau\Doctrine\Calque\Avertissement;
use Ormeau\Doctrine\Calque\CalqueLogique;
use Ormeau\Doctrine\Calque\CasEnumeration;
use Ormeau\Doctrine\Calque\ColonneJointure;
use Ormeau\Doctrine\Calque\Entite;
use Ormeau\Doctrine\Calque\Enumeration;
use Ormeau\Doctrine\Calque\ExpressionDefaut;
use Ormeau\Doctrine\Calque\Generee;
use Ormeau\Doctrine\Calque\GenreAssociation;
use Ormeau\Doctrine\Calque\Heritage;
use Ormeau\Doctrine\Calque\Identifiant;
use Ormeau\Doctrine\Calque\IndexEntite;
use Ormeau\Doctrine\Calque\Origine;
use Ormeau\Doctrine\Calque\Propriete;
use Ormeau\Doctrine\Calque\ReferenceTable;
use Ormeau\Doctrine\Calque\Resolution;
use Ormeau\Doctrine\Calque\StrategieHeritage;
use Ormeau\Doctrine\Calque\StrategieIdentifiant;
use Ormeau\Doctrine\Calque\TableJointure;
use Ormeau\Doctrine\Calque\TraitPartage;
use Ormeau\Doctrine\Calque\TypeSupport;
use PHPUnit\Framework\Attributes\CoversNothing;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;
use ReflectionClass;
use ReflectionProperty;

/**
 * Le modèle PHP et le JSON Schema publié décrivent le même format.
 *
 * Un champ ajouté au calque appelle trois modifications — le schéma, les
 * structures Go, le lecteur PHP. Les tests Go tiennent les deux premiers ;
 * celui-ci tient le troisième, dans les deux sens : aucun champ du schéma sans
 * propriété, aucune propriété sans champ, aucun vocabulaire qui diverge.
 */
#[CoversNothing]
final class SchemaTest extends TestCase
{
    /** Le contrat publié, à la racine du dépôt. */
    private const SCHEMA = __DIR__ . '/../../../schemas/calque-logique.v2.json';

    /**
     * Chaque champ d'un objet du schéma a sa propriété, et réciproquement.
     *
     * @param list<string> $cheminSchema clés à parcourir jusqu'à la définition
     * @param class-string $classe       classe du modèle qui la représente
     */
    #[DataProvider('objets')]
    public function testLesChampsCorrespondent(array $cheminSchema, string $classe): void
    {
        $definition = self::definition($cheminSchema);
        self::assertIsArray($definition['properties'] ?? null);

        $champs = array_map(self::camel(...), array_keys($definition['properties']));
        $proprietes = array_map(
            static fn(ReflectionProperty $p): string => $p->getName(),
            (new ReflectionClass($classe))->getProperties(ReflectionProperty::IS_PUBLIC),
        );
        sort($champs);
        sort($proprietes);

        self::assertSame($champs, $proprietes);
    }

    /**
     * Les objets du schéma et leur classe.
     *
     * @return iterable<string, array{list<string>, class-string}>
     */
    public static function objets(): iterable
    {
        yield 'racine' => [[], CalqueLogique::class];
        yield 'entite' => [['$defs', 'entite'], Entite::class];
        yield 'referenceTable' => [['$defs', 'referenceTable'], ReferenceTable::class];
        yield 'heritage' => [['$defs', 'heritage'], Heritage::class];
        yield 'identifiant' => [['$defs', 'identifiant'], Identifiant::class];
        yield 'propriete' => [['$defs', 'propriete'], Propriete::class];
        yield 'generee' => [['$defs', 'propriete', 'properties', 'generee'], Generee::class];
        yield 'association' => [['$defs', 'association'], Association::class];
        yield 'colonneJointure' => [['$defs', 'colonneJointure'], ColonneJointure::class];
        yield 'tableJointure' => [['$defs', 'tableJointure'], TableJointure::class];
        yield 'indexEntite' => [['$defs', 'indexEntite'], IndexEntite::class];
        yield 'enumeration' => [['$defs', 'enumeration'], Enumeration::class];
        yield 'cas' => [['$defs', 'enumeration', 'properties', 'cas', 'items'], CasEnumeration::class];
        yield 'trait' => [['$defs', 'trait'], TraitPartage::class];
        yield 'avertissement' => [['$defs', 'avertissement'], Avertissement::class];
    }

    /**
     * Chaque vocabulaire fermé du schéma a exactement les valeurs de son
     * énumération PHP. Une valeur ajoutée d'un seul côté ferait refuser un
     * calque valide, ou accepter un calque que le binaire n'écrit pas.
     *
     * @param list<string>             $cheminSchema clés à parcourir jusqu'à la propriété qui porte l'enum
     * @param class-string<BackedEnum> $vocabulaire  énumération PHP qui doit en avoir les valeurs
     */
    #[DataProvider('vocabulaires')]
    public function testLesVocabulairesCorrespondent(array $cheminSchema, string $vocabulaire): void
    {
        $definition = self::definition($cheminSchema);
        self::assertIsArray($definition['enum'] ?? null);

        $valeurs = array_map(static fn(BackedEnum $cas): string|int => $cas->value, $vocabulaire::cases());

        self::assertSame($definition['enum'], $valeurs);
    }

    /**
     * Les vocabulaires du schéma et leur énumération.
     *
     * @return iterable<string, array{list<string>, class-string<BackedEnum>}>
     */
    public static function vocabulaires(): iterable
    {
        $champ = static fn(string $objet, string $cle): array => ['$defs', $objet, 'properties', $cle];

        yield 'origine' => [['$defs', 'origine'], Origine::class];
        yield 'heritage.strategie' => [$champ('heritage', 'strategie'), StrategieHeritage::class];
        yield 'identifiant.strategie' => [$champ('identifiant', 'strategie'), StrategieIdentifiant::class];
        yield 'association.genre' => [$champ('association', 'genre'), GenreAssociation::class];
        yield 'colonneJointure.a_la_suppression' => [$champ('colonneJointure', 'a_la_suppression'), ActionSuppression::class];
        yield 'enumeration.type_support' => [$champ('enumeration', 'type_support'), TypeSupport::class];
        yield 'propriete.defaut_expression' => [$champ('propriete', 'defaut_expression'), ExpressionDefaut::class];
        yield 'avertissement.resolution' => [$champ('avertissement', 'resolution'), Resolution::class];
    }

    /**
     * Rend la définition désignée par un chemin de clés dans le schéma.
     *
     * @param list<string> $chemin clés à parcourir depuis la racine du schéma
     *
     * @return array<mixed>
     */
    private static function definition(array $chemin): array
    {
        $noeud = json_decode((string) file_get_contents(self::SCHEMA), true, 64, JSON_THROW_ON_ERROR);
        foreach ($chemin as $cle) {
            self::assertIsArray($noeud);
            $noeud = $noeud[$cle] ?? null;
        }
        self::assertIsArray($noeud, 'définition introuvable : ' . implode('.', $chemin));

        return $noeud;
    }

    /**
     * Passe un nom de champ du format en nom de propriété : a_la_suppression
     * devient aLaSuppression.
     */
    private static function camel(string $champ): string
    {
        return lcfirst(str_replace('_', '', ucwords($champ, '_')));
    }
}
