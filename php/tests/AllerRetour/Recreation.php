<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\AllerRetour;

use Composer\InstalledVersions;
use Doctrine\DBAL\Connection;
use Doctrine\DBAL\DriverManager;
use Doctrine\DBAL\Exception as DBALException;
use Doctrine\ORM\Configuration;
use Doctrine\ORM\EntityManager;
use Doctrine\ORM\ORMSetup;
use Doctrine\ORM\Tools\SchemaTool;
use InvalidArgumentException;
use JsonException;
use Ormeau\Doctrine\Calque\CalqueInvalide;
use Ormeau\Doctrine\Calque\LecteurCalque;
use Ormeau\Doctrine\Generation\Cible;
use Ormeau\Doctrine\Generation\GenerateurEntite;
use RuntimeException;

/**
 * La moitié PHP de l'aller-retour : générer les entités d'un calque logique,
 * puis laisser Doctrine créer leur schéma dans une base vierge.
 *
 * Le test Go qui orchestre la chaîne extrait la base d'origine, infère, appelle
 * ce script, puis extrait la base recréée et compare. Ce qui se passe ici est
 * exactement ce qu'un utilisateur obtient : le générateur du paquet, puis
 * SchemaTool, sans projet Symfony autour — ormeau:generer n'ajoute que la
 * lecture des options et l'affichage.
 *
 * Seule cette classe écrit dans une base, et seulement dans la base de test
 * qu'on lui désigne : l'introspection, elle, reste en lecture seule.
 */
final class Recreation
{
    /**
     * Génère, recrée le schéma, et rend ce que le test Go doit savoir pour
     * choisir et conditionner ses tolérances.
     *
     * @param string $parametres chemin d'un fichier JSON écrit par le test Go : sgbd, logique, entites,
     *                           base, schema, hote, port, utilisateur, mot_de_passe
     *
     * @return array{orm: string, dbal: string, ecartees: list<array{entite: string, raison: string}>}
     *
     * @throws RuntimeException         paramètres illisibles, SGBD inconnu, ou dépendance absente
     * @throws DBALException            connexion ou préparation de la base refusée
     * @throws JsonException            paramètres qui ne sont pas du JSON
     * @throws CalqueInvalide           calque logique refusé par le lecteur
     * @throws InvalidArgumentException ORM absent ou d'une majeure inconnue
     */
    public static function executer(string $parametres): array
    {
        $contenu = file_get_contents($parametres);
        if ($contenu === false) {
            throw new RuntimeException('Paramètres illisibles : ' . $parametres);
        }
        /** @var array{sgbd: string, logique: string, entites: string, base: string, schema: string, hote: string, port: int, utilisateur: string, mot_de_passe: string} $p */
        $p = json_decode($contenu, true, 8, JSON_THROW_ON_ERROR);

        $calque = (new LecteurCalque())->lire($p['logique']);
        $cible = Cible::detecter();
        $rapport = (new GenerateurEntite())->generer($calque, $p['entites'], $cible, 'gescom');

        $espace = $calque->espaceDeNoms . '\\';
        $entites = $p['entites'];
        spl_autoload_register(static function (string $classe) use ($espace, $entites): void {
            if (str_starts_with($classe, $espace)) {
                require $entites . '/' . str_replace('\\', '/', substr($classe, strlen($espace))) . '.php';
            }
        });

        $configuration = ORMSetup::createAttributeMetadataConfiguration([$entites], true);
        // Les objets paresseux natifs n'existent qu'à partir d'ORM 3.4 sous PHP 8.4.
        // @phpstan-ignore function.alreadyNarrowedType
        if (PHP_VERSION_ID >= 80400 && method_exists($configuration, 'enableNativeLazyObjects')) {
            $configuration->enableNativeLazyObjects(true);
        }
        // Le calque ne couvre qu'un schéma : le générateur n'écrit alors pas
        // schema:, et Doctrine crée tout dans le schéma courant. Chaque
        // préparation rend ce schéma vide et courant, pour que la base recréée
        // se compare à l'originale sans traduction.
        $connexion = match ($p['sgbd']) {
            'postgres' => self::preparerPostgres($p, $configuration),
            'sqlserver' => self::preparerSqlServer($p, $configuration),
            default => throw new RuntimeException('SGBD de l\'aller-retour inconnu : ' . $p['sgbd']),
        };

        $gestionnaire = new EntityManager($connexion, $configuration);
        $metadonnees = $gestionnaire->getMetadataFactory()->getAllMetadata();
        (new SchemaTool($gestionnaire))->createSchema($metadonnees);

        $ecartees = [];
        foreach ($rapport->ecartees as $ecartee) {
            $ecartees[] = ['entite' => $ecartee->nom, 'raison' => $ecartee->raison];
        }

        return [
            'orm' => self::version('doctrine/orm'),
            'dbal' => self::version('doctrine/dbal'),
            'ecartees' => $ecartees,
        ];
    }

    /**
     * Ouvre la base recréée PostgreSQL, créée par le DDL de test, et y rend le
     * schéma vide et courant par search_path.
     *
     * @param array{base: string, schema: string, hote: string, port: int, utilisateur: string, mot_de_passe: string} $p paramètres du test Go
     * @param Configuration                                                                                             $configuration configuration de l'ORM, partagée avec l'EntityManager
     *
     * @throws DBALException connexion ou instruction refusée
     */
    private static function preparerPostgres(array $p, Configuration $configuration): Connection
    {
        $connexion = DriverManager::getConnection([
            'driver' => 'pdo_pgsql',
            'host' => $p['hote'],
            'port' => $p['port'],
            'user' => $p['utilisateur'],
            'password' => $p['mot_de_passe'],
            'dbname' => $p['base'],
        ], $configuration);

        $schema = $connexion->quoteIdentifier($p['schema']);
        $connexion->executeStatement('DROP SCHEMA IF EXISTS ' . $schema . ' CASCADE');
        $connexion->executeStatement('CREATE SCHEMA ' . $schema);
        $connexion->executeStatement('SET search_path TO ' . $schema);

        return $connexion;
    }

    /**
     * Recrée la base SQL Server vide.
     *
     * Les tables y atterrissent dans dbo, et non dans le schéma du calque :
     * SQL Server n'a pas de search_path, et la plateforme SQL Server de DBAL
     * tient dbo pour le schéma d'une table non qualifiée — elle y pose les
     * commentaires de table, même quand le schéma par défaut de la session est
     * un autre (essai du 2026-09-16). Le test Go renomme le schéma de la base
     * recréée avant de comparer.
     *
     * La base est supprimée et recréée plutôt que vidée : SQL Server n'a pas de
     * DROP SCHEMA … CASCADE, et vider un schéma demanderait de défaire les
     * clés étrangères dans l'ordre.
     *
     * @param array{base: string, schema: string, hote: string, port: int, utilisateur: string, mot_de_passe: string} $p paramètres du test Go
     * @param Configuration                                                                                             $configuration configuration de l'ORM, partagée avec l'EntityManager
     *
     * @throws DBALException connexion ou instruction refusée
     */
    private static function preparerSqlServer(array $p, Configuration $configuration): Connection
    {
        // Le conteneur de test sert un certificat auto-signé, que le pilote
        // ODBC 18 refuse par défaut.
        $parametres = [
            'driver' => 'pdo_sqlsrv',
            'host' => $p['hote'],
            'port' => $p['port'],
            'user' => $p['utilisateur'],
            'password' => $p['mot_de_passe'],
            'driverOptions' => ['TrustServerCertificate' => 'yes'],
        ];

        $serveur = DriverManager::getConnection($parametres + ['dbname' => 'master']);
        $base = $serveur->quoteIdentifier($p['base']);
        $serveur->executeStatement(
            'IF DB_ID(' . $serveur->quote($p['base']) . ') IS NOT NULL BEGIN '
            . 'ALTER DATABASE ' . $base . ' SET SINGLE_USER WITH ROLLBACK IMMEDIATE; '
            . 'DROP DATABASE ' . $base . '; END',
        );
        $serveur->executeStatement('CREATE DATABASE ' . $base);
        $serveur->close();

        return DriverManager::getConnection($parametres + ['dbname' => $p['base']], $configuration);
    }

    /**
     * Rend la version installée d'un paquet.
     *
     * @param string $paquet nom Composer du paquet
     *
     * @throws RuntimeException paquet absent de l'installation
     */
    private static function version(string $paquet): string
    {
        $version = InstalledVersions::isInstalled($paquet) ? InstalledVersions::getVersion($paquet) : null;
        if ($version === null) {
            throw new RuntimeException($paquet . ' est introuvable dans l\'installation.');
        }

        return $version;
    }
}
