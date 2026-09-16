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
 * Une table sans clé primaire n'en a pas : l'entité n'a pas d'identifiant, et
 * la génération l'écarte plutôt que de choisir une colonne à sa place.
 */
final class Identifiant
{
    /**
     * @param list<string>         $proprietes        noms de propriétés de l'entité — pas de colonnes —,
     *                                                dans l'ordre de la clé primaire
     * @param StrategieIdentifiant $strategie         qui produit la valeur de la clé
     * @param string|null          $sequence          nom de la séquence, présent avec la stratégie Sequence
     * @param int|null             $sequenceIncrement incrément déclaré de la séquence, présent quand l'inférence
     *                                                l'a retrouvée dans le calque physique ; un fait, pas une
     *                                                taille d'allocation : au-delà de 1, il réserve des blocs
     *                                                ou sépare plusieurs nœuds, et la base ne dit pas lequel
     * @param int|null             $sequenceMinimum   valeur minimale déclarée de la séquence, présente dans le
     *                                                même cas
     * @param int|null             $sequenceDepart    valeur de départ déclarée de la séquence, présente dans le
     *                                                même cas quand le calque physique la porte ; SQL Server
     *                                                part de 1 une séquence dont le minimum est celui du type
     */
    public function __construct(
        public readonly array $proprietes,
        public readonly StrategieIdentifiant $strategie,
        public readonly ?string $sequence = null,
        public readonly ?int $sequenceIncrement = null,
        public readonly ?int $sequenceMinimum = null,
        public readonly ?int $sequenceDepart = null,
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
        $increment = Lecture::entierOptionnel($donnees, 'sequence_increment', $chemin);
        $minimum = Lecture::entierOptionnel($donnees, 'sequence_minimum', $chemin);
        $depart = Lecture::entierOptionnel($donnees, 'sequence_depart', $chemin);

        return new self($proprietes, $strategie, $sequence, $increment, $minimum, $depart);
    }
}
