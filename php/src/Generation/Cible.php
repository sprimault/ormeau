<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

use Composer\InstalledVersions;
use InvalidArgumentException;

/**
 * L'environnement pour lequel le code est produit : les versions de Doctrine
 * ORM et de DBAL.
 *
 * La sortie du générateur dépend du calque logique et de cette cible, et chaque
 * règle se raisonne par le paquet qui en décide. ORM décide de la stratégie
 * d'une séquence ou des paramètres qu'un attribut accepte ; DBAL décide de ce
 * qu'une colonne hydrate — un bigint est une chaîne sous DBAL 2 et 3, un entier
 * sous DBAL 4. Les deux ne vont pas de pair : ORM 3 accepte DBAL 3.8. Une cible
 * réduite à la majeure d'ORM rendait bigint en entier sous ORM 3 et DBAL 3, et
 * l'unité de travail voyait la colonne modifiée à chaque flush (essai du
 * 2026-09-15, ORM 3.7.1 et DBAL 3.10.6).
 *
 * La commande la détecte dans l'application, ou la déduit d'une majeure d'ORM
 * forcée ; un appelant de la bibliothèque la fournit. Elle est toujours
 * annoncée en tête d'exécution, avec l'origine de la version de DBAL.
 */
final class Cible
{
    /** Majeures d'ORM que le générateur sait viser. */
    public const MAJEURES_ORM = [2, 3];

    /** Version de DBAL lue dans l'application. */
    public const DBAL_DETECTEE = 'détectée';

    /** Version de DBAL déduite d'une majeure d'ORM forcée. */
    public const DBAL_DEDUITE = 'déduite';

    /** Version de DBAL donnée par --cible-dbal. */
    public const DBAL_FORCEE = 'forcée';

    /**
     * Majeures de DBAL que chaque majeure d'ORM accepte, lues dans le
     * composer.json de doctrine/orm 2.14, 2.20 et 3.7.
     */
    private const DBAL_ACCEPTEES = [2 => [2, 3], 3 => [3, 4]];

    /**
     * DBAL supposée quand seule la majeure d'ORM est forcée : la plus récente
     * qu'elle accepte parmi celles vues, et sous ORM 3 celle que résout le lock.
     */
    private const DBAL_PAR_DEFAUT = [2 => [3, 10], 3 => [4, 4]];

    /**
     * @param int      $ormMajeure  majeure de Doctrine ORM visée, 2 ou 3
     * @param int|null $ormMineure  mineure d'ORM quand elle est connue, c'est-à-dire lue dans
     *                              l'application ; null quand seule la majeure est forcée, auquel cas
     *                              aucune règle propre à une mineure ne s'applique
     * @param string   $php         version de PHP de l'application, pour l'annonce seulement : le code
     *                              produit tient sous PHP 8.1, quelle qu'elle soit
     * @param int      $dbalMajeure majeure de DBAL visée, qui doit être acceptée par la majeure d'ORM
     * @param int      $dbalMineure mineure de DBAL, pour l'annonce et les règles propres à une mineure
     * @param bool     $forcee      vrai quand la majeure d'ORM vient de --cible-orm et non de l'application
     * @param string   $orm         version d'ORM installée, majeure et mineure, pour l'annonce ; vide quand
     *                              la cible est forcée
     * @param string   $origineDbal d'où vient la version de DBAL : DBAL_DETECTEE, DBAL_DEDUITE ou DBAL_FORCEE
     *
     * @throws InvalidArgumentException majeure d'ORM inconnue, ou DBAL que cette majeure n'accepte pas
     */
    public function __construct(
        public readonly int $ormMajeure,
        public readonly ?int $ormMineure,
        public readonly string $php,
        public readonly int $dbalMajeure,
        public readonly int $dbalMineure,
        public readonly bool $forcee = false,
        public readonly string $orm = '',
        public readonly string $origineDbal = self::DBAL_DETECTEE,
    ) {
        if (!in_array($ormMajeure, self::MAJEURES_ORM, true)) {
            throw new InvalidArgumentException(sprintf(
                'Doctrine ORM %d n\'est pas une cible connue (%s).',
                $ormMajeure,
                implode(', ', self::MAJEURES_ORM),
            ));
        }
        if (!in_array($dbalMajeure, self::DBAL_ACCEPTEES[$ormMajeure], true)) {
            throw new InvalidArgumentException(sprintf(
                'Doctrine ORM %d ne s\'installe pas avec DBAL %d (DBAL %s).',
                $ormMajeure,
                $dbalMajeure,
                implode(', ', self::DBAL_ACCEPTEES[$ormMajeure]),
            ));
        }
    }

    /**
     * Lit la cible dans l'application qui exécute la commande : les versions
     * d'ORM et de DBAL installées par Composer, et le PHP en cours.
     *
     * doctrine/orm et doctrine/dbal sont des dépendances du paquet : s'il en
     * manque un, l'installation est cassée, et le dire vaut mieux que de
     * supposer une version.
     *
     * @param string|null $dbal version de DBAL à viser au lieu de celle installée, « majeure.mineure »
     *
     * @throws InvalidArgumentException ORM ou DBAL absents, versions inconnues ou incompatibles
     */
    public static function detecter(?string $dbal = null): self
    {
        [$ormMajeure, $ormMineure] = self::installee('doctrine/orm');
        [$dbalMajeure, $dbalMineure] = $dbal === null ? self::installee('doctrine/dbal') : self::lireDbal($dbal);

        return new self(
            $ormMajeure,
            $ormMineure,
            self::phpCourant(),
            $dbalMajeure,
            $dbalMineure,
            orm: $ormMajeure . '.' . $ormMineure,
            origineDbal: $dbal === null ? self::DBAL_DETECTEE : self::DBAL_FORCEE,
        );
    }

    /**
     * Fixe la majeure d'ORM, pour générer à destination d'une version qui
     * n'est pas installée. Sans version de DBAL, la plus récente que cette
     * majeure accepte est supposée, et l'annonce le dit.
     *
     * La mineure d'ORM reste inconnue : rien du rendu n'en dépend, et supposer
     * la dernière publiée reviendrait à taire un défaut propre aux premières
     * mineures pour quelqu'un qui vise justement l'une d'elles.
     *
     * @param int         $ormMajeure majeure d'ORM visée
     * @param string|null $dbal       version de DBAL visée, « majeure.mineure »
     *
     * @throws InvalidArgumentException majeure inconnue, version de DBAL illisible ou incompatible
     */
    public static function forcer(int $ormMajeure, ?string $dbal = null): self
    {
        $parDefaut = self::DBAL_PAR_DEFAUT[$ormMajeure] ?? [0, 0];
        [$dbalMajeure, $dbalMineure] = $dbal === null ? $parDefaut : self::lireDbal($dbal);

        return new self(
            $ormMajeure,
            null,
            self::phpCourant(),
            $dbalMajeure,
            $dbalMineure,
            forcee: true,
            origineDbal: $dbal === null ? self::DBAL_DEDUITE : self::DBAL_FORCEE,
        );
    }

    /**
     * Dit si DBAL hydrate bigint en entier et binary en chaîne, comme DBAL 4.
     * DBAL 2 et 3 rendent une chaîne et une ressource.
     */
    public function hydrateCommeDbal4(): bool
    {
        return $this->dbalMajeure >= 4;
    }

    /**
     * Version d'ORM visée, « majeure.mineure », ou la majeure seule quand la
     * mineure est inconnue.
     *
     * Distincte de $orm, qui sert l'annonce et reste vide sur une cible
     * forcée : un message qui nomme la version ne peut pas dépendre d'un champ
     * d'affichage.
     */
    public function versionOrm(): string
    {
        return $this->ormMineure === null ? (string) $this->ormMajeure : $this->ormMajeure . '.' . $this->ormMineure;
    }

    /**
     * Dit si le persister d'un héritage joint laisse hors de l'INSERT une
     * colonne que la base calcule.
     *
     * ORM l'y incluait jusqu'à la 2.15 comprise, sans lui donner de valeur : la
     * base refuse alors d'écrire une ligne de la hiérarchie, racine comprise
     * (essai du 2026-09-16, ORM 2.14.3 et 2.15.5 contre PostgreSQL 17). Corrigé
     * en 2.16.0 par doctrine/orm#10598, sans rétroportage sur 2.14 ni 2.15.
     * Hors héritage joint, la colonne était déjà exclue.
     *
     * Vrai quand la mineure est inconnue : une cible forcée ne désigne aucune
     * version précise, et avertir sur une supposition vaudrait moins que se
     * taire.
     */
    public function ometLesColonnesGenereesEnHeritageJoint(): bool
    {
        return $this->ormMineure === null || $this->ormMajeure >= 3 || $this->ormMineure >= 16;
    }

    /**
     * Dit si DBAL décrit un défaut calculé par un objet de
     * Doctrine\DBAL\Schema\DefaultExpression, introduit en 4.4 : la chaîne
     * 'CURRENT_TIMESTAMP' y est dépréciée, et reste la seule forme avant.
     */
    public function defautParExpression(): bool
    {
        return $this->dbalMajeure === 4 && $this->dbalMineure >= 4;
    }

    /**
     * Dit si DBAL connaît le type smallfloat, introduit en 4.1 : avant, une
     * simple précision n'a que float, recréé en DOUBLE PRECISION.
     */
    public function connaitSmallfloat(): bool
    {
        return $this->dbalMajeure === 4 && $this->dbalMineure >= 1;
    }

    /**
     * Dit si DBAL connaît le type jsonb, introduit en 4.3, qui y déprécie
     * l'option de colonne jsonb. Avant, seule l'option le décrit, et le type
     * est inconnu (essai du 2026-09-15, DBAL 4.2.5).
     */
    public function connaitJsonb(): bool
    {
        return $this->dbalMajeure === 4 && $this->dbalMineure >= 3;
    }

    /**
     * La ligne annoncée en tête d'exécution : « Cible détectée : PHP 8.2,
     * Doctrine ORM 2.14, DBAL 3.10 » ou « Cible forcée : Doctrine ORM 3,
     * DBAL 4.4 (déduite) ».
     */
    public function annonce(): string
    {
        $dbal = sprintf('DBAL %d.%d', $this->dbalMajeure, $this->dbalMineure);
        if ($this->origineDbal !== self::DBAL_DETECTEE) {
            $dbal .= ' (' . $this->origineDbal . ')';
        }

        return $this->forcee
            ? sprintf('Cible forcée : Doctrine ORM %d, %s', $this->ormMajeure, $dbal)
            : sprintf('Cible détectée : PHP %s, Doctrine ORM %s, %s', $this->php, $this->orm, $dbal);
    }

    /**
     * Lit une version de DBAL écrite « majeure.mineure ».
     *
     * @return array{int, int}
     *
     * @throws InvalidArgumentException forme illisible
     */
    private static function lireDbal(string $version): array
    {
        if (preg_match('/^(\d+)\.(\d+)$/', $version, $parties) !== 1) {
            throw new InvalidArgumentException(sprintf('Version de DBAL illisible, « majeure.mineure » attendu : « %s ».', $version));
        }

        return [(int) $parties[1], (int) $parties[2]];
    }

    /**
     * Rend la majeure et la mineure installées d'un paquet.
     *
     * @return array{int, int}
     *
     * @throws InvalidArgumentException paquet absent de l'application
     */
    private static function installee(string $paquet): array
    {
        $version = InstalledVersions::isInstalled($paquet) ? InstalledVersions::getVersion($paquet) : null;
        if ($version === null) {
            throw new InvalidArgumentException(sprintf('%s est introuvable dans l\'application.', $paquet));
        }
        $parties = explode('.', $version);

        return [(int) $parties[0], (int) ($parties[1] ?? 0)];
    }

    /**
     * Rend la version majeure et mineure du PHP en cours.
     */
    private static function phpCourant(): string
    {
        return PHP_MAJOR_VERSION . '.' . PHP_MINOR_VERSION;
    }
}
