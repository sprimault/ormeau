<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

/**
 * Ce que la classe de l'utilisateur d'une racine d'héritage porte en plus de
 * #[ORM\Entity] et #[ORM\Table] : la stratégie, la colonne discriminante et
 * la carte des valeurs.
 *
 * Ces trois attributs vont dans la classe de l'utilisateur et non dans la
 * classe de base, Doctrine les refusant sur un MappedSuperclass. Ils sont donc
 * écrits une fois, et relus ensuite pour signaler ce qui a divergé : une
 * sous-classe déclarée après coup change la carte, que l'utilisateur recopie
 * depuis le message. Le rendu et le contrôle partagent cet objet, pour que le
 * message cite exactement ce que la création aurait écrit.
 *
 * @internal
 */
final class RacineHeritage
{
    /** Seule stratégie que l'inférence produit : une table par classe. */
    public const STRATEGIE = 'JOINED';

    /**
     * @param array{name: string, type: string|null, length: int|null, options?: array<string, bool|string>|null} $colonne arguments de
     *                                                                           #[ORM\DiscriminatorColumn] ; type
     *                                                                           et longueur absents quand la
     *                                                                           colonne n'est pas une propriété,
     *                                                                           options sous ORM 3 seulement
     * @param array<string, string>                                     $carte   valeur discriminante vers le nom
     *                                                                           qualifié de la classe, dans
     *                                                                           l'ordre du calque
     * @param string                                                    $classe  nom qualifié de la classe de la
     *                                                                           racine, qui porte ces attributs
     */
    public function __construct(
        public readonly array $colonne,
        public readonly array $carte,
        public readonly string $classe,
    ) {}

    /**
     * Rend les trois attributs tels qu'ils s'écrivent dans la classe, indexés
     * par leur nom qualifié.
     *
     * La carte est composée ici plutôt que confiée à l'émetteur : une carte
     * aux valeurs 0 et 1 serait une liste pour PHP, et s'écrirait sans ses
     * clés. Une classe du même espace de noms que la racine s'écrit par son nom
     * court ; une classe rangée ailleurs, qualifiée, puisque l'attribut ne
     * peut pas ajouter d'import à un fichier qui appartient à l'utilisateur.
     *
     * @return array<string, string>
     */
    public function attributs(): array
    {
        $espace = self::espace($this->classe);
        $entrees = [];
        foreach ($this->carte as $valeur => $classe) {
            $nom = self::espace($classe) === $espace ? substr($classe, strlen($espace) + ($espace === '' ? 0 : 1)) : '\\' . $classe;
            $entrees[] = Emetteur::litteral((string) $valeur) . ' => ' . $nom . '::class';
        }

        return [
            'Doctrine\ORM\Mapping\InheritanceType' => Emetteur::attribut('ORM\InheritanceType', [self::STRATEGIE], ''),
            'Doctrine\ORM\Mapping\DiscriminatorColumn' => Emetteur::attribut('ORM\DiscriminatorColumn', $this->colonne, ''),
            'Doctrine\ORM\Mapping\DiscriminatorMap' => Emetteur::attribut('ORM\DiscriminatorMap', [new Code('[' . implode(', ', $entrees) . ']')], ''),
        ];
    }

    /**
     * Rend l'espace de noms d'un nom qualifié, vide pour une classe globale.
     */
    private static function espace(string $classe): string
    {
        $position = strrpos($classe, '\\');

        return $position === false ? '' : substr($classe, 0, $position);
    }
}
