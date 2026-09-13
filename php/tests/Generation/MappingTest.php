<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\Generation;

use Doctrine\DBAL\DriverManager;
use Doctrine\ORM\EntityManager;
use Doctrine\ORM\Mapping\ClassMetadata;
use Doctrine\ORM\ORMSetup;
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

            $ecartees = array_map(static fn($e): string => $e->nom, $rapport->ecartees);
            foreach ($calque->entites as $entite) {
                if (in_array($entite->nom, $ecartees, true)) {
                    continue;
                }
                $classe = $calque->espaceDeNoms . '\\' . $entite->nom;
                $meta = $gestionnaire->getClassMetadata($classe);

                self::assertSame($entite->table->nom, $meta->getTableName(), $classe);
                self::assertSame(
                    array_map(static fn($p): string => $p->colonne, $entite->proprietes),
                    array_map(static fn(string $champ): string => $meta->getColumnName($champ), array_map(static fn($p): string => $p->nom, $entite->proprietes)),
                    $classe,
                );
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
            }
            self::assertCount(
                count(array_filter($metadonnees, static fn(ClassMetadata $m): bool => !$m->isMappedSuperclass)),
                array_filter($rapport->fichiers, static fn($f): bool => !str_contains($f->chemin, '/Base/')),
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
