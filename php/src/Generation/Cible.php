<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

use Composer\InstalledVersions;
use InvalidArgumentException;

/**
 * L'environnement pour lequel le code est produit.
 *
 * La sortie du générateur dépend du calque logique et de cette cible : un
 * bigint est une chaîne sous Doctrine ORM 2, qui s'appuie sur DBAL 3, et un
 * entier sous ORM 3 et DBAL 4. C'est ce qui rend un passage d'ORM 2 à ORM 3
 * visible dans le diff des entités, et c'est pourquoi la cible est toujours
 * annoncée en tête d'exécution.
 *
 * La commande la détecte dans l'application ; un appelant de la bibliothèque la
 * fournit. Aucune détection n'a lieu dans le générateur lui-même : il reste une
 * traduction du calque et de ce qu'on lui passe.
 */
final class Cible
{
    /** Majeures d'ORM que le générateur sait viser. */
    public const MAJEURES_ORM = [2, 3];

    /**
     * @param int    $ormMajeure majeure de Doctrine ORM visée, 2 ou 3
     * @param string $php        version de PHP de l'application, pour l'annonce seulement : le code
     *                           produit tient sous PHP 8.1, quelle qu'elle soit
     * @param bool   $forcee     vrai quand la majeure vient de --cible-orm et non de l'application
     * @param string $orm        version installée, majeure et mineure, pour l'annonce ; vide quand
     *                           la cible est forcée
     *
     * @throws InvalidArgumentException majeure qu'aucun rendu ne connaît
     */
    public function __construct(
        public readonly int $ormMajeure,
        public readonly string $php,
        public readonly bool $forcee = false,
        public readonly string $orm = '',
    ) {
        if (!in_array($ormMajeure, self::MAJEURES_ORM, true)) {
            throw new InvalidArgumentException(sprintf(
                'Doctrine ORM %d n\'est pas une cible connue (%s).',
                $ormMajeure,
                implode(', ', self::MAJEURES_ORM),
            ));
        }
    }

    /**
     * Lit la cible dans l'application qui exécute la commande : la version
     * d'ORM installée par Composer, et le PHP en cours.
     *
     * doctrine/orm est une dépendance du paquet : s'il manque, l'installation
     * est cassée, et le dire vaut mieux que de supposer une version.
     *
     * @throws InvalidArgumentException ORM absent, ou d'une majeure inconnue
     */
    public static function detecter(): self
    {
        $version = InstalledVersions::isInstalled('doctrine/orm')
            ? InstalledVersions::getVersion('doctrine/orm')
            : null;
        if ($version === null) {
            throw new InvalidArgumentException('Doctrine ORM est introuvable dans l\'application.');
        }

        $parties = explode('.', $version);

        return new self((int) $parties[0], self::phpCourant(), orm: $parties[0] . '.' . ($parties[1] ?? '0'));
    }

    /**
     * Fixe la majeure d'ORM, pour générer à destination d'une version qui
     * n'est pas installée.
     *
     * @throws InvalidArgumentException majeure inconnue
     */
    public static function forcer(int $ormMajeure): self
    {
        return new self($ormMajeure, self::phpCourant(), forcee: true);
    }

    /**
     * La ligne annoncée en tête d'exécution : « Cible détectée : PHP 8.2,
     * Doctrine ORM 2.14 » ou « Cible forcée : Doctrine ORM 3 ».
     */
    public function annonce(): string
    {
        return $this->forcee
            ? sprintf('Cible forcée : Doctrine ORM %d', $this->ormMajeure)
            : sprintf('Cible détectée : PHP %s, Doctrine ORM %s', $this->php, $this->orm);
    }

    /**
     * Rend la version majeure et mineure du PHP en cours.
     */
    private static function phpCourant(): string
    {
        return PHP_MAJOR_VERSION . '.' . PHP_MINOR_VERSION;
    }
}
