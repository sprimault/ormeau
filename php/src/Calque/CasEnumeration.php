<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * Un cas d'énumération PHP et la valeur que la base stocke.
 *
 * Le nom est déjà un identifiant PHP valide : c'est l'inférence, ou une
 * décision, qui l'a tiré de la valeur. Une valeur que rien ne permet de nommer
 * — « 1 », « ? » — produit un avertissement cas_enumeration_opaque, pas un nom
 * inventé ici.
 */
final class CasEnumeration
{
    /**
     * @param string     $nom    nom du cas, tel qu'il s'écrira après « case »
     * @param int|string $valeur valeur stockée, du type support de l'énumération
     */
    public function __construct(
        public readonly string $nom,
        public readonly int|string $valeur,
    ) {}

    /**
     * Construit le cas depuis le JSON décodé.
     *
     * La valeur est lue à part : c'est le seul champ du format qui admet deux
     * types, et Lecture n'en connaît qu'un par méthode.
     *
     * @param array<mixed> $donnees objet JSON décodé du cas
     * @param string       $chemin  chemin de l'objet dans le calque, pour les messages
     *
     * @throws CalqueInvalide nom absent, valeur absente ou ni chaîne ni entier
     */
    public static function depuisTableau(array $donnees, string $chemin): self
    {
        $valeur = $donnees['valeur'] ?? null;
        if (!is_int($valeur) && !is_string($valeur)) {
            throw new CalqueInvalide(sprintf('Champ %s.valeur : chaîne ou entier attendu', $chemin));
        }

        $nom = Lecture::chaine($donnees, 'nom', $chemin);

        return new self($nom, $valeur);
    }
}
