<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\AllerRetour;

use Composer\InstalledVersions;
use Doctrine\DBAL\DriverManager;
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
     * @param string $parametres chemin d'un fichier JSON écrit par le test Go : logique, entites, base,
     *                           schema, hote, port, utilisateur, mot_de_passe
     *
     * @return array{orm: string, dbal: string, ecartees: list<array{entite: string, raison: string}>}
     *
     * @throws RuntimeException         paramètres illisibles, ou dépendance absente
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
        /** @var array{logique: string, entites: string, base: string, schema: string, hote: string, port: int, utilisateur: string, mot_de_passe: string} $p */
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
        $connexion = DriverManager::getConnection([
            'driver' => 'pdo_pgsql',
            'host' => $p['hote'],
            'port' => $p['port'],
            'user' => $p['utilisateur'],
            'password' => $p['mot_de_passe'],
            'dbname' => $p['base'],
        ], $configuration);

        // Le calque ne couvre qu'un schéma : le générateur n'écrit alors pas
        // schema:, et Doctrine crée tout dans le schéma courant. On le recrée
        // vide et on le rend courant, pour que la base recréée se compare à
        // l'originale sans traduction.
        $schema = $connexion->quoteIdentifier($p['schema']);
        $connexion->executeStatement('DROP SCHEMA IF EXISTS ' . $schema . ' CASCADE');
        $connexion->executeStatement('CREATE SCHEMA ' . $schema);
        $connexion->executeStatement('SET search_path TO ' . $schema);

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
