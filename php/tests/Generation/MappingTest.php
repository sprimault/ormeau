<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\Generation;

use Doctrine\DBAL\DriverManager;
use Doctrine\DBAL\Types\StringType;
use Doctrine\DBAL\Types\Type;
use Doctrine\Deprecations\Deprecation;
use Doctrine\ORM\EntityManager;
use Doctrine\ORM\Id\SequenceGenerator;
use Doctrine\ORM\Mapping\ClassMetadata;
use Doctrine\ORM\ORMSetup;
use Doctrine\ORM\Tools\SchemaTool;
use Doctrine\ORM\Tools\SchemaValidator;
use Ormeau\Doctrine\Calque\LecteurCalque;
use Ormeau\Doctrine\Calque\StrategieIdentifiant;
use Ormeau\Doctrine\Generation\Cible;
use Ormeau\Doctrine\Generation\GenerateurEntite;
use Ormeau\Doctrine\Generation\IndexOmis;
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
            $rapport = (new GenerateurEntite())->generer($calque, $sortie, $cible, $cas);
            spl_autoload_register(static function (string $classe) use ($calque, $sortie): void {
                $prefixe = $calque->espaceDeNoms . '\\';
                if (str_starts_with($classe, $prefixe)) {
                    require $sortie . '/' . str_replace('\\', '/', substr($classe, strlen($prefixe))) . '.php';
                }
            });

            // Un type forcé par décision que DBAL ne connaît pas est un type du
            // projet, que le projet enregistre : le test en fait autant, sur le
            // type le plus neutre, puisque seul le mapping est éprouvé ici.
            foreach ([...$calque->entites, ...$calque->traits] as $porteur) {
                foreach ($porteur->proprietes as $propriete) {
                    if (!Type::hasType($propriete->typeDoctrine)) {
                        Type::addType($propriete->typeDoctrine, StringType::class);
                    }
                }
            }

            $gestionnaire = self::gestionnaire($sortie);
            [$metadonnees, $depreciations] = self::chargerEnEcoutant($gestionnaire);

            // Ne sont retenues que les dépréciations qui nomment une classe
            // produite : celles de l'environnement — fabrique de gestionnaire
            // de schéma, getEventManager, IDENTITY sous PostgreSQL — ne citent
            // jamais l'espace de noms du calque, ce qui évite une liste
            // d'exclusions à tenir à jour. C'est un filet, pas une preuve : une
            // dépréciation qui ne nommerait aucune classe passerait au travers.
            self::assertSame([], array_values(array_filter(
                $depreciations,
                static fn(string $m): bool => str_contains($m, $calque->espaceDeNoms . '\\'),
            )), 'le rendu déclenche une dépréciation de Doctrine');

            // Un cas dont toutes les tables sont écartées ne produit aucune
            // entité. La condition se lit sur le calque et le rapport, jamais
            // sur les métadonnées obtenues : un cas qui devait produire des
            // entités échoue toujours plus bas sur un DDL vide.
            if (count($calque->entites) === count($rapport->ecartees)) {
                self::assertSame([], $metadonnees, 'aucune entité générée, aucune métadonnée attendue');

                return;
            }

            self::assertSame([], (new SchemaValidator($gestionnaire))->validateMapping());

            // Le validateur ne voit pas tout : Doctrine doit aussi savoir écrire
            // le DDL de l'ensemble, ce qui échoue sur une colonne de jointure
            // qui vise une colonne inexistante ou une table de jointure mal
            // formée. SQLite n'a pas de séquence : une clé par séquence, rendue
            // telle sous ORM 2, est contrôlée plus bas sur ses métadonnées, et
            // son DDL relève de l'aller-retour contre PostgreSQL.
            self::assertNotEmpty((new SchemaTool($gestionnaire))->getCreateSchemaSql(array_values(array_filter(
                $metadonnees,
                static fn(ClassMetadata $m): bool => !$m->isIdGeneratorSequence(),
            ))));

            $ecartees = array_map(static fn($e): string => $e->nom, $rapport->ecartees);
            foreach ($calque->entites as $entite) {
                if (in_array($entite->nom, $ecartees, true)) {
                    continue;
                }
                $classe = $calque->espaceDeNoms . '\\' . $entite->nom;
                $meta = $gestionnaire->getClassMetadata($classe);
                $jointures = self::colonnesDeJointure($meta);

                self::assertSame($entite->table->nom, $meta->getTableName(), $classe);
                // Une clé par séquence : ORM 2 lit le nom de la séquence, ORM 3
                // passe par IDENTITY, le générateur d'une classe de base
                // mappée n'y étant pas repris — sauf sous SQL Server, où un
                // générateur produit tire la séquence nommée.
                if ($entite->identifiant?->strategie === StrategieIdentifiant::Sequence) {
                    if ($cible->ormMajeure === 2) {
                        self::assertTrue($meta->isIdGeneratorSequence(), $classe . ' : génération par séquence');
                        self::assertSame($entite->identifiant->sequence, $meta->sequenceGeneratorDefinition['sequenceName'] ?? null, $classe . ' : nom de séquence');
                    } elseif ($calque->sgbd === 'sqlserver') {
                        self::assertSame(ClassMetadata::GENERATOR_TYPE_CUSTOM, $meta->generatorType, $classe . ' : génération par générateur produit');
                        self::assertInstanceOf(SequenceGenerator::class, $meta->idGenerator, $classe . ' : générateur de séquence');
                    } else {
                        self::assertTrue($meta->isIdGeneratorIdentity(), $classe . ' : génération par IDENTITY');
                    }
                }
                self::assertSame($entite->commentaire, $meta->table['options']['comment'] ?? null, $classe . ' : commentaire de table');

                // Une classe d'une hiérarchie décidée est jointe, hérite de son
                // parent et porte sa valeur : Doctrine ajouterait sinon sa
                // propre colonne dtype, absente de la base.
                $discriminante = self::colonneDiscriminante($meta);
                if ($entite->valeurDiscriminante !== null) {
                    self::assertSame(ClassMetadata::INHERITANCE_TYPE_JOINED, $meta->inheritanceType, $classe);
                    self::assertSame($entite->valeurDiscriminante, (string) $meta->discriminatorValue, $classe);
                    self::assertNotNull($discriminante, $classe);
                }
                if ($entite->heritage !== null) {
                    self::assertContains($calque->espaceDeNoms . '\\' . $entite->heritage->parent, $meta->parentClasses, $classe);
                    self::assertSame($entite->heritage->colonneDiscriminante, $discriminante, $classe);
                }

                // Chaque propriété du calque est un champ sur sa colonne, sauf
                // celle dont la colonne porte une association de la clé
                // primaire : l'identité passe alors par l'association. La
                // colonne discriminante n'est pas un champ non plus, Doctrine
                // refusant de la mapper deux fois.
                foreach ($entite->proprietes as $propriete) {
                    if ($propriete->colonne === $discriminante) {
                        self::assertFalse($meta->hasField($propriete->nom), $classe . '::' . $propriete->nom . ' mappe la colonne discriminante');
                        continue;
                    }
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
                // Un index dont DBAL refuse le nom n'est pas mappé, et le
                // rapport le déclare : le compte se fait sur ce qui reste. Un
                // index filtré sous SQL Server reste mappé, sans son filtre.
                $refuses = array_map(
                    static fn(IndexOmis $i): string => $i->nom,
                    array_filter($rapport->index, static fn($i): bool => $i instanceof IndexOmis && $i->entite === $entite->nom),
                );
                $retenus = array_filter($entite->index, static fn($i): bool => !in_array($i->nom, $refuses, true));
                self::assertCount(
                    count(array_filter($retenus, static fn($i): bool => $i->unique)),
                    $meta->table['uniqueConstraints'] ?? [],
                    $classe . ' : contraintes d\'unicité',
                );
                self::assertCount(
                    count(array_filter($retenus, static fn($i): bool => !$i->unique)),
                    $meta->table['indexes'] ?? [],
                    $classe . ' : index',
                );
                // Le prédicat arrive à Doctrine tel que le calque le porte :
                // un échappement faux le changerait sans rien refuser. Sous
                // SQL Server, il n'arrive pas, DBAL ne sachant pas l'écrire.
                foreach ($retenus as $index) {
                    if ($index->predicat === null || $index->nom === null) {
                        continue;
                    }
                    $mappe = $meta->table[$index->unique ? 'uniqueConstraints' : 'indexes'][$index->nom] ?? [];
                    $attendu = $calque->sgbd === 'sqlserver' ? null : $index->predicat;
                    self::assertSame($attendu, $mappe['options']['where'] ?? null, $classe . ' : prédicat de ' . $index->nom);
                }

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
     * Des classes de l'utilisateur rangées dans des sous-répertoires, puis
     * régénérées, se chargent encore : Doctrine résout chaque cible
     * d'association, le parent d'une classe fille et la carte de la racine vers
     * les classes rangées, et rien n'est recréé à la racine.
     */
    #[RunInSeparateProcess]
    public function testDoctrineAccepteDesClassesRangees(): void
    {
        $cible = Cible::detecter();
        $calque = (new LecteurCalque())->lire(Repertoires::REFERENCES . '/heritage-decide/logique.json');
        $sortie = Repertoires::creer();

        try {
            (new GenerateurEntite())->generer($calque, $sortie, $cible, 'heritage-decide');
            Repertoires::ranger($sortie, 'Personne', 'Rh');
            Repertoires::ranger($sortie, 'Salarie', 'Rh');
            Repertoires::ranger($sortie, 'Adresse', 'Contact');
            $rapport = (new GenerateurEntite())->generer($calque, $sortie, $cible, 'heritage-decide');

            self::assertSame([], $rapport->ecartees);
            self::assertSame([], $rapport->divergences);
            foreach (['Personne', 'Salarie', 'Adresse'] as $nom) {
                self::assertFileDoesNotExist($sortie . '/' . $nom . '.php');
            }

            spl_autoload_register(static function (string $classe) use ($sortie): void {
                if (str_starts_with($classe, 'App\\Entity\\')) {
                    require $sortie . '/' . str_replace('\\', '/', substr($classe, strlen('App\\Entity\\'))) . '.php';
                }
            });
            $gestionnaire = self::gestionnaire($sortie);
            $metadonnees = $gestionnaire->getMetadataFactory()->getAllMetadata();

            self::assertSame([], (new SchemaValidator($gestionnaire))->validateMapping());
            self::assertNotEmpty((new SchemaTool($gestionnaire))->getCreateSchemaSql($metadonnees));
            self::assertSame('App\\Entity\\Rh\\Salarie', $gestionnaire->getClassMetadata('App\\Entity\\Affectation')->getAssociationTargetClass('salarie'));
            self::assertSame('App\\Entity\\Contact\\Adresse', $gestionnaire->getClassMetadata('App\\Entity\\Rh\\Personne')->getAssociationTargetClass('adresse'));
            self::assertContains('App\\Entity\\Rh\\Personne', $gestionnaire->getClassMetadata('App\\Entity\\Rh\\Salarie')->parentClasses);
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
        foreach (Repertoires::cas() as $cas) {
            yield $cas => [$cas];
        }
    }

    /**
     * Rend le nom de la colonne discriminante d'une entité, ou null hors
     * héritage.
     *
     * ORM 2 la décrit par un tableau, ORM 3 par un objet aux propriétés
     * publiques. Convertie en tableau, elle se lit de la même façon sous les
     * deux, sans l'accès par clé qu'ORM 3 déprécie.
     *
     * @param ClassMetadata<object> $meta métadonnées de l'entité
     */
    private static function colonneDiscriminante(ClassMetadata $meta): ?string
    {
        $colonne = $meta->discriminatorColumn;

        return $colonne === null ? null : (string) ((array) $colonne)['name'];
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
     *
     * La plateforme PostgreSQL ne s'en passe pas : SchemaTool y lit
     * current_schema() pour écrire le DDL, donc une vraie base.
     */
    /**
     * Charge toutes les métadonnées en écoutant les dépréciations émises.
     *
     * doctrine/deprecations n'émet rien tant qu'on ne l'active pas, et ne
     * signale chaque dépréciation qu'une fois : sans le processus par cas de ce
     * test, la première entité concernée masquerait toutes les suivantes.
     *
     * @return array{list<ClassMetadata<object>>, list<string>}
     */
    private static function chargerEnEcoutant(EntityManager $gestionnaire): array
    {
        Deprecation::enableWithTriggerError();

        /** @var list<string> $depreciations */
        $depreciations = [];
        set_error_handler(static function (int $niveau, string $message) use (&$depreciations): bool {
            $depreciations[] = $message;

            return true;
        }, E_USER_DEPRECATED);

        try {
            $metadonnees = $gestionnaire->getMetadataFactory()->getAllMetadata();
        } finally {
            restore_error_handler();
        }

        return [$metadonnees, $depreciations];
    }

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
