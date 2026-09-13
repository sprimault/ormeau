<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

/**
 * Un écart entre une classe de l'utilisateur et ce que le calque attend.
 *
 * La classe de l'utilisateur n'est jamais modifiée, sans exception : c'est la
 * garantie qui rend la régénération sûre, et une liste d'exceptions ne la
 * tiendrait pas. Le prix est une correction à la main, et c'est pourquoi la
 * divergence nomme le fichier, la ligne, ce qui s'y trouve et ce qui est
 * attendu — trente secondes de travail, pas une recherche.
 */
final class Divergence
{
    /**
     * @param string $fichier chemin de la classe de l'utilisateur
     * @param int    $ligne   ligne de l'élément en cause, ou de la classe quand il manque
     * @param string $actuel  ce qui s'y trouve, tel qu'écrit dans le fichier ; vide quand il manque
     * @param string $attendu ce que le calque attend, dit en une phrase
     */
    public function __construct(
        public readonly string $fichier,
        public readonly int $ligne,
        public readonly string $actuel,
        public readonly string $attendu,
    ) {}

    /**
     * La divergence en une ligne : « Client.php ligne 14 :
     * #[ORM\Table(name: 't_clients')], la table s'appelle maintenant t_client ».
     */
    public function message(): string
    {
        return $this->actuel === ''
            ? sprintf('%s ligne %d : %s', $this->fichier, $this->ligne, $this->attendu)
            : sprintf('%s ligne %d : %s, %s', $this->fichier, $this->ligne, $this->actuel, $this->attendu);
    }
}
