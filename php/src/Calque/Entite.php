<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * Une classe à générer.
 *
 * Toutes les tables n'en produisent pas : une table de jointure pure devient
 * une association, et une table écartée par décision n'apparaît pas.
 *
 * L'origine porte celle du nom de classe, et rien d'autre. L'existence de
 * l'entité est un constat — la table est là —, et l'héritage comme les
 * associations portent chacun la leur.
 */
final class Entite
{
    /**
     * @param string              $nom          nom de classe PHP, sans espace de noms
     * @param ReferenceTable      $table        table dont l'entité est issue
     * @param list<Propriete>     $proprietes   propriétés propres à l'entité, sans celles de ses traits
     * @param Heritage|null       $heritage     place dans une hiérarchie, quand le schéma en révèle une
     * @param list<string>        $traits       noms des traits du calque que l'entité utilise
     * @param Identifiant|null    $identifiant  clé de l'entité ; absent, la table n'en a pas
     * @param list<Association>   $associations liens vers les autres entités
     * @param list<IndexEntite>   $index        index de la table, reportés pour le schéma
     * @param Origine|null        $origine      d'où vient le nom de classe
     */
    public function __construct(
        public readonly string $nom,
        public readonly ReferenceTable $table,
        public readonly array $proprietes,
        public readonly ?Heritage $heritage = null,
        public readonly array $traits = [],
        public readonly ?Identifiant $identifiant = null,
        public readonly array $associations = [],
        public readonly array $index = [],
        public readonly ?Origine $origine = null,
    ) {}

    /**
     * Construit l'entité depuis le JSON décodé.
     *
     * Les listes facultatives absentes valent la liste vide : le binaire omet
     * les champs vides, et la génération itère dessus sans avoir à s'en
     * méfier.
     *
     * @param array<mixed> $donnees
     * @param string       $chemin  chemin de l'objet dans le calque, pour les messages
     *
     * @throws CalqueInvalide champ requis absent, élément imbriqué invalide
     */
    public static function depuisTableau(array $donnees, string $chemin): self
    {
        $nom = Lecture::chaine($donnees, 'nom', $chemin);
        $table = Lecture::objet($donnees, 'table', $chemin, ReferenceTable::depuisTableau(...), requis: true);
        $proprietes = Lecture::objets($donnees, 'proprietes', $chemin, Propriete::depuisTableau(...), requise: true);
        $heritage = Lecture::objet($donnees, 'heritage', $chemin, Heritage::depuisTableau(...));
        $traits = Lecture::chaines($donnees, 'traits', $chemin);
        $identifiant = Lecture::objet($donnees, 'identifiant', $chemin, Identifiant::depuisTableau(...));
        $associations = Lecture::objets($donnees, 'associations', $chemin, Association::depuisTableau(...));
        $index = Lecture::objets($donnees, 'index', $chemin, IndexEntite::depuisTableau(...));
        $origine = Lecture::valeurOptionnelle($donnees, 'origine', $chemin, Origine::class);

        return new self($nom, $table, $proprietes, $heritage, $traits, $identifiant, $associations, $index, $origine);
    }
}
