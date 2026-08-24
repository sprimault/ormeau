<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * Représentation en mémoire d'un calque logique.
 *
 * Les entités sont laissées en tableaux à ce stade : les objets typés viendront
 * quand la génération sera écrite et qu'on saura ce dont elle a réellement
 * besoin. Rule of three avant de figer une hiérarchie de valeurs.
 */
final class CalqueLogique
{
    /**
     * Privé : un calque se construit depuis un tableau décodé, jamais champ par
     * champ. C'est ce qui garantit que les contrôles de depuisTableau ont eu
     * lieu.
     *
     * @param list<array<string, mixed>> $entites
     * @param list<array<string, mixed>> $enumerations
     * @param list<array<string, mixed>> $avertissements
     */
    private function __construct(
        public readonly int $versionRi,
        public readonly string $empreintePhysique,
        public readonly string $espaceDeNoms,
        public readonly array $entites,
        public readonly array $enumerations,
        public readonly array $avertissements,
    ) {}

    /**
     * Construit un calque depuis le JSON décodé.
     *
     * Les champs optionnels du format — énumérations, avertissements — valent
     * le tableau vide quand ils sont absents : leur absence est légitime, seuls
     * les champs requis sont contrôlés.
     *
     * @param array<string, mixed> $donnees
     *
     * @throws CalqueInvalide champ requis absent
     */
    public static function depuisTableau(array $donnees): self
    {
        foreach (['empreinte_physique', 'espace_de_noms', 'entites'] as $requis) {
            if (!isset($donnees[$requis])) {
                throw new CalqueInvalide(sprintf('Champ %s absent du calque', $requis));
            }
        }

        return new self(
            $donnees['version_ri'],
            $donnees['empreinte_physique'],
            $donnees['espace_de_noms'],
            $donnees['entites'],
            $donnees['enumerations'] ?? [],
            $donnees['avertissements'] ?? [],
        );
    }

    /**
     * Les avertissements sont une sortie de premier ordre : la commande de
     * génération les affiche, et la CI peut filtrer sur leur code.
     *
     * @return list<array<string, mixed>>
     */
    public function avertissementsAuDessusDe(float $confiance): array
    {
        return array_values(array_filter(
            $this->avertissements,
            static fn(array $a): bool => ($a['confiance'] ?? 0.0) >= $confiance,
        ));
    }
}
