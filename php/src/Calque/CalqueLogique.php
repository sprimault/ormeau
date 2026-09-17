<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * Représentation en mémoire d'un calque logique : ce que la génération
 * traduit, et rien d'autre.
 *
 * Contrairement au calque physique, il n'est pas neutre : il parle le
 * vocabulaire de la famille Hibernate — entités, associations avec côté
 * propriétaire, héritage. Ce biais est assumé, et c'est ce qui permet au
 * générateur de traduire sans décider.
 *
 * Chaque objet imbriqué est typé et contrôlé à la lecture : un calque qui passe
 * ici a ses champs requis, dans le bon type, avec des valeurs connues de chaque
 * vocabulaire fermé. La génération n'a pas à revérifier.
 */
final class CalqueLogique
{
    /**
     * @param int                 $versionRi         version du format, déjà contrôlée par le lecteur
     * @param string              $empreintePhysique empreinte du calque physique dont ce jugement
     *                                               découle ; elle dit si la base a bougé sans
     *                                               relire les entités
     * @param string              $espaceDeNoms      espace de noms des entités générées
     * @param string              $sgbd              SGBD du calque physique : le rendu dépend de la plateforme
     *                                               DBAL, pas seulement des versions — schéma écrit d'une table,
     *                                               clé par séquence, défaut calculé
     * @param list<Entite>        $entites           classes à générer, dans l'ordre du calque
     * @param list<Enumeration>   $enumerations      énumérations PHP que des propriétés désignent
     * @param list<TraitPartage>  $traits            traits que des entités nomment
     * @param list<Avertissement> $avertissements    ce que l'inférence n'a pas résolu
     */
    public function __construct(
        public readonly int $versionRi,
        public readonly string $empreintePhysique,
        public readonly string $espaceDeNoms,
        public readonly string $sgbd,
        public readonly array $entites,
        public readonly array $enumerations = [],
        public readonly array $traits = [],
        public readonly array $avertissements = [],
    ) {}

    /**
     * Construit un calque depuis le JSON décodé.
     *
     * Les champs optionnels du format — énumérations, traits, avertissements —
     * valent la liste vide quand ils sont absents : leur absence est
     * légitime. Le premier champ invalide arrête la lecture, avec son chemin
     * dans le message.
     *
     * @param array<mixed> $donnees document JSON décodé, objet racine du calque
     *
     * @throws CalqueInvalide champ requis absent, du mauvais type, ou hors vocabulaire
     */
    public static function depuisTableau(array $donnees): self
    {
        $versionRi = Lecture::entier($donnees, 'version_ri', '');
        $empreintePhysique = Lecture::chaine($donnees, 'empreinte_physique', '');
        $espaceDeNoms = Lecture::chaine($donnees, 'espace_de_noms', '');
        $sgbd = Lecture::chaine($donnees, 'sgbd', '');
        $entites = Lecture::objets($donnees, 'entites', '', Entite::depuisTableau(...), requise: true);
        $enumerations = Lecture::objets($donnees, 'enumerations', '', Enumeration::depuisTableau(...));
        $traits = Lecture::objets($donnees, 'traits', '', TraitPartage::depuisTableau(...));
        $avertissements = Lecture::objets($donnees, 'avertissements', '', Avertissement::depuisTableau(...));

        return new self($versionRi, $empreintePhysique, $espaceDeNoms, $sgbd, $entites, $enumerations, $traits, $avertissements);
    }

    /**
     * Rend les avertissements dont la confiance atteint le seuil.
     *
     * Sert à la CI : elle échoue sur les avertissements sûrs et laisse passer
     * les incertains. Le résultat est une liste réindexée, que json_encode rend
     * en tableau et non en objet.
     *
     * @return list<Avertissement>
     */
    public function avertissementsAuDessusDe(float $confiance): array
    {
        return array_values(array_filter(
            $this->avertissements,
            static fn(Avertissement $a): bool => $a->confiance >= $confiance,
        ));
    }
}
