<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * Une énumération PHP à générer, que les propriétés désignent par son nom.
 *
 * Son origine — CHECK, type énuméré natif, échantillon ou décision — est ce
 * qui permet d'en discuter : une énumération tirée d'un échantillon peut
 * manquer un cas que la base n'avait pas encore rencontré.
 */
final class Enumeration
{
    /**
     * @param string               $nom         nom de l'énumération PHP, sans espace de noms
     * @param TypeSupport          $typeSupport type sur lequel elle s'adosse
     * @param list<CasEnumeration> $cas         cas dans l'ordre du calque, que la génération
     *                                          reprend tel quel
     * @param Origine              $origine     d'où vient l'énumération
     */
    public function __construct(
        public readonly string $nom,
        public readonly TypeSupport $typeSupport,
        public readonly array $cas,
        public readonly Origine $origine,
    ) {}

    /**
     * Construit l'énumération depuis le JSON décodé.
     *
     * Une énumération sans cas est refusée : PHP la compilerait, mais aucune
     * valeur lue en base ne pourrait s'y hydrater.
     *
     * @param array<mixed> $donnees objet JSON décodé de l'énumération
     * @param string       $chemin  chemin de l'objet dans le calque, pour les messages
     *
     * @throws CalqueInvalide champ requis absent, aucun cas, type ou origine hors vocabulaire
     */
    public static function depuisTableau(array $donnees, string $chemin): self
    {
        $cas = Lecture::objets($donnees, 'cas', $chemin, CasEnumeration::depuisTableau(...), requise: true);
        Lecture::nonVide($cas, 'cas', $chemin);

        $nom = Lecture::chaine($donnees, 'nom', $chemin);
        $typeSupport = Lecture::valeur($donnees, 'type_support', $chemin, TypeSupport::class);
        $origine = Lecture::valeur($donnees, 'origine', $chemin, Origine::class);

        return new self($nom, $typeSupport, $cas, $origine);
    }
}
