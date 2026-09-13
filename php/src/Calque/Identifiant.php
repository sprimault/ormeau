<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * Les propriétés qui identifient une entité.
 *
 * Plusieurs propriétés signifient une clé composite, cas courant sur une base
 * reprise et que Doctrine accepte à condition de marquer chacune d'un #[Id].
 * Une table sans clé primaire porte la stratégie Aucune : la génération doit le
 * signaler, pas choisir une colonne à sa place.
 */
final class Identifiant
{
    /**
     * @param list<string>         $proprietes noms de propriétés de l'entité — pas de colonnes —,
     *                                         dans l'ordre de la clé primaire
     * @param StrategieIdentifiant $strategie  qui produit la valeur de la clé
     * @param string|null          $sequence   nom de la séquence, présent avec la stratégie Sequence
     */
    public function __construct(
        public readonly array $proprietes,
        public readonly StrategieIdentifiant $strategie,
        public readonly ?string $sequence = null,
    ) {}

    /**
     * Construit l'identifiant depuis le JSON décodé.
     *
     * Une liste de propriétés vide est refusée : elle ne désignerait rien, et
     * une entité sans identifiant se dit par l'absence de l'objet, pas par une
     * liste vide.
     *
     * @param array<mixed> $donnees objet JSON décodé de l'identifiant
     * @param string       $chemin  chemin de l'objet dans le calque, pour les messages
     *
     * @throws CalqueInvalide propriétés absentes ou vides, stratégie hors vocabulaire
     */
    public static function depuisTableau(array $donnees, string $chemin): self
    {
        $proprietes = Lecture::chaines($donnees, 'proprietes', $chemin, requise: true);
        Lecture::nonVide($proprietes, 'proprietes', $chemin);

        $strategie = Lecture::valeur($donnees, 'strategie', $chemin, StrategieIdentifiant::class);
        $sequence = Lecture::chaineOptionnelle($donnees, 'sequence', $chemin);

        return new self($proprietes, $strategie, $sequence);
    }
}
