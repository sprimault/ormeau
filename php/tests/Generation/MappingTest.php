<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\Generation;

use Doctrine\DBAL\DriverManager;
use Doctrine\ORM\EntityManager;
use Doctrine\ORM\Mapping\ClassMetadata;
use Doctrine\ORM\ORMSetup;
use Doctrine\ORM\Tools\SchemaTool;
use Doctrine\ORM\Tools\SchemaValidator;
use Ormeau\Doctrine\Calque\LecteurCalque;
use Ormeau\Doctrine\Generation\Cible;
use Ormeau\Doctrine\Generation\GenerateurEntite;
use PHPUnit\Framework\Attributes\CoversClass;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\Attributes\RunInSeparateProcess;
use PHPUnit\Framework\TestCase;

/**
 * Doctrine accepte ce que le générateur produit, sous la version d'ORM
 * installée.
 *
 * Un attendu figé prouve que le rendu ne bouge pas, pas qu'il est juste. Ici,
 * les entités sont chargées par l'ORM réel et passées à son validateur de
 * mapping, sans base de données. Chaque job de compatibilité installe une
 * majeure différente : ORM 2 valide le rendu ORM 2, ORM 3 le rendu ORM 3.
 *
 * Un processus par cas : tous les calques écrivent dans App\Entity, et une
 * classe ne se déclare qu'une fois par processus.
 */
#[CoversClass(GenerateurEntite::class)]
final class MappingTest extends TestCase
{
    /**
     * Le mapping se charge, le validateur n'y trouve rien, et ce que le
     * calque dit de chaque table se retrouve dans les métadonnées : un
     * attribut que Doctrine ignorerait en silence — un #[Table] sur une classe
     * de base — se verrait ici.
     */
    #[DataProvider('cas')]
    #[RunInSeparateProcess]
    public function testDoctrineAccepteLeMapping(string $cas): void
    {
        $cible = Cible::detecter();
        $calque = (new LecteurCalque())->lire(Repertoires::REFERENCES . '/' . $cas . '/logique.json');
        $sortie = Repertoires::creer();

        try {
            $rapport = (new GenerateurEntite())->generer($calque, $sortie, $cible);
            spl_autoload_register(static function (string $classe) use ($calque, $sortie): void {
                $prefixe = $calque->espaceDeNoms . '\\';
                if (str_starts_with($classe, $prefixe)) {
                    require $sortie . '/' . str_replace('\\', '/', substr($classe, strlen($prefixe))) . '.php';
                }
            });

            $gestionnaire = self::gestionnaire($sortie);
            $metadonnees = $gestionnaire->getMetadataFactory()->getAllMetadata();

            self::assertSame([], (new SchemaValidator($gestionnaire))->validateMapping());

            // Le validateur ne voit pas tout : Doctrine doit aussi savoir écrire
            // le DDL de l'ensemble, ce qui échoue sur une colonne de jointure
            // qui vise une colonne inexistante ou une table de jointure mal
            // formée.
            self::assertNotEmpty((new SchemaTool($gestionnaire))->getCreateSchemaSql($metadonnees));

            $ecartees = array_map(static fn($e): string => $e->nom, $rapport->ecartees);
            foreach ($calque->entites as $entite) {
                if (in_array($entite->nom, $ecartees, true)) {
                    continue;
                }
                $classe = $calque->espaceDeNoms . '\\' . $entite->nom;
                $meta = $gestionnaire->getClassMetadata($classe);
                $jointures = self::colonnesDeJointure($meta);

                self::assertSame($entite->table->nom, $meta->getTableName(), $classe);

                // Chaque propriété du calque est un champ sur sa colonne, sauf
                // celle dont la colonne porte une association de la clé
                // primaire : l'identité passe alors par l'association.
                foreach ($entite->proprietes as $propriete) {
                    if ($meta->hasField($propriete->nom)) {
                        self::assertSame($propriete->colonne, $meta->getColumnName($propriete->nom), $classe . '::' . $propriete->nom);
                        continue;
                    }
                    self::assertArrayHasKey($propriete->colonne, $jointures, $classe . '::' . $propriete->nom . ' absente sans être une jointure');
                    self::assertContains($jointures[$propriete->colonne], $meta->getIdentifierFieldNames(), $classe . '::' . $propriete->nom);
                }

                // Une colonne, un seul écrivain : un champ posé sur une colonne
                // de jointure ne l'écrit pas, sans quoi l'association l'écrase
                // en silence à l'insertion — Doctrine l'accepte, et le
                // validateur ne dit rien.
                foreach (array_keys($meta->fieldMappings) as $champ) {
                    $colonne = $meta->getColumnName($champ);
                    if (isset($jointures[$colonne])) {
                        self::assertSame([true, true], self::lectureSeule($meta, $champ), $classe . '::' . $champ . ' écrit la colonne de ' . $jointures[$colonne]);
                    }
                }

                // Chaque association du calque est mappée, sauf un côté inverse
                // que le rapport déclare omis — et celui-là ne l'est pas.
                $omises = array_map(
                    static fn($o): string => $o->association,
                    array_filter($rapport->omises, static fn($o): bool => $o->entite === $entite->nom),
                );
                foreach ($entite->associations as $association) {
                    if (in_array($association->nom, $omises, true)) {
                        self::assertFalse($meta->hasAssociation($association->nom), $classe . '::' . $association->nom . ' omise mais mappée');
                        continue;
                    }
                    self::assertTrue($meta->hasAssociation($association->nom), $classe . '::' . $association->nom);
                    self::assertSame($calque->espaceDeNoms . '\\' . $association->cible, $meta->getAssociationTargetClass($association->nom));
                }
                self::assertCount(
                    count(array_filter($entite->index, static fn($i): bool => $i->unique)),
                    $meta->table['uniqueConstraints'] ?? [],
                    $classe . ' : contraintes d\'unicité',
                );
                self::assertCount(
                    count(array_filter($entite->index, static fn($i): bool => !$i->unique)),
                    $meta->table['indexes'] ?? [],
                    $classe . ' : index',
                );

                foreach ($entite->proprietes as $propriete) {
                    if ($propriete->enumeration !== null) {
                        self::assertSame(
                            $calque->espaceDeNoms . '\\Enum\\' . $propriete->enumeration,
                            self::enumType($meta, $propriete->nom),
                            $classe . '::' . $propriete->nom,
                        );
                    }
                }

                // Les colonnes d'un trait sont mappées sur chaque entité qui
                // l'utilise, comme si elles y étaient déclarées.
                foreach ($calque->traits as $trait) {
                    if (!in_array($trait->nom, $entite->traits, true)) {
                        continue;
                    }
                    foreach ($trait->proprietes as $propriete) {
                        self::assertSame($propriete->colonne, $meta->getColumnName($propriete->nom), $classe . ' : trait ' . $trait->nom);
                    }
                }
            }
            self::assertCount(
                count(array_filter($metadonnees, static fn(ClassMetadata $m): bool => !$m->isMappedSuperclass)),
                array_filter($rapport->fichiers, static fn($f): bool => dirname($f->chemin) === $sortie),
            );
        } finally {
            Repertoires::supprimer($sortie);
        }
    }

    /**
     * Les cas que la génération actuelle couvre.
     *
     * @return iterable<string, array{string}>
     */
    public static function cas(): iterable
    {
        foreach (Repertoires::CAS as $cas) {
            yield $cas => [$cas];
        }
    }

    /**
     * Rend, pour chaque colonne de jointure portée par l'entité, le nom de
     * l'association qui l'écrit.
     *
     * ORM 2 décrit une association par un tableau, ORM 3 par un objet dont
     * seul le côté propriétaire d'un objet porte des colonnes de jointure.
     *
     * @param ClassMetadata<object> $meta métadonnées de l'entité
     *
     * @return array<string, string>
     */
    private static function colonnesDeJointure(ClassMetadata $meta): array
    {
        $colonnes = [];
        foreach ($meta->associationMappings as $nom => $mapping) {
            $jointures = is_array($mapping) ? ($mapping['joinColumns'] ?? []) : (property_exists($mapping, 'joinColumns') ? $mapping->joinColumns : []);
            foreach ($jointures as $jointure) {
                $colonnes[is_array($jointure) ? $jointure['name'] : $jointure->name] = $nom;
            }
        }

        return $colonnes;
    }

    /**
     * Rend [non insérable, non modifiable] pour un champ mappé, sous ORM 2 ou 3.
     *
     * @param ClassMetadata<object> $meta  métadonnées de l'entité
     * @param string                $champ nom de la propriété mappée
     *
     * @return array{bool, bool}
     */
    private static function lectureSeule(ClassMetadata $meta, string $champ): array
    {
        $mapping = $meta->fieldMappings[$champ];

        return is_array($mapping)
            ? [(bool) ($mapping['notInsertable'] ?? false), (bool) ($mapping['notUpdatable'] ?? false)]
            : [(bool) $mapping->notInsertable, (bool) $mapping->notUpdatable];
    }

    /**
     * Rend l'enumType d'un champ mappé.
     *
     * ORM 2 décrit un champ par un tableau, ORM 3 par un objet FieldMapping :
     * le test tourne sous les deux.
     *
     * @param ClassMetadata<object> $meta  métadonnées de l'entité
     * @param string                $champ nom de la propriété mappée
     */
    private static function enumType(ClassMetadata $meta, string $champ): ?string
    {
        $mapping = $meta->fieldMappings[$champ];

        return is_array($mapping) ? ($mapping['enumType'] ?? null) : $mapping->enumType;
    }

    /**
     * Construit un gestionnaire d'entités sur les attributs du répertoire,
     * sans base : SQLite en mémoire ne sert qu'à fournir une plateforme.
     */
    private static function gestionnaire(string $repertoire): EntityManager
    {
        $configuration = ORMSetup::createAttributeMetadataConfiguration([$repertoire], true);
        // Les objets paresseux natifs n'existent qu'à partir d'ORM 3.4 : PHPStan,
        // qui analyse sous ORM 3, voit la méthode partout ; ORM 2 ne l'a pas.
        // @phpstan-ignore function.alreadyNarrowedType
        if (PHP_VERSION_ID >= 80400 && method_exists($configuration, 'enableNativeLazyObjects')) {
            $configuration->enableNativeLazyObjects(true);
        }
        $connexion = DriverManager::getConnection(['driver' => 'pdo_sqlite', 'memory' => true], $configuration);

        return new EntityManager($connexion, $configuration);
    }
}
