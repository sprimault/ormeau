<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Calque;

/**
 * Une colonne devenue propriété.
 *
 * Le type Doctrine apparaît ici et pas dans le calque physique : il suppose la
 * destination. L'origine porte celle du type, pas celle du nom — la colonne
 * existe, seule sa traduction en type est un jugement. Le type PHP n'y est
 * pas : il dépend de la version de DBAL du projet, et TypesPhp le déduit.
 *
 * Longueur, précision et échelle restent nulles quand le calque ne les porte
 * pas : decimal(10,0) n'est pas un entier, et zéro n'est pas une absence. Même
 * règle pour le défaut, où la chaîne vide est un défaut à part entière.
 */
final class Propriete
{
    /**
     * @param string       $nom          nom de la propriété PHP, en camelCase
     * @param string       $colonne      colonne d'origine, telle que le catalogue l'écrit
     * @param string       $typeDoctrine type DBAL, y compris un type personnalisé forcé par décision
     * @param bool         $nullable     reprise du physique ; une propriété facultative se type ?T
     * @param int|null     $longueur     longueur déclarée, absente pour un text
     * @param int|null     $precision    nombre total de chiffres d'un décimal
     * @param int|null     $echelle      chiffres après la virgule d'un décimal
     * @param string|null  $enumeration  nom de l'énumération PHP du calque qui type la propriété ;
     *                                   le type Doctrine reste celui de la valeur stockée
     * @param string|null  $defaut       défaut littéral, dans sa forme textuelle ; un défaut calculé
     *                                   n'y figure jamais, son sens est dans $defautExpression
     * @param bool         $insertable   faux pour une colonne générée : l'écrire depuis PHP échouerait
     * @param bool         $modifiable   faux pour une colonne générée, pour la même raison
     * @param bool         $unique       vrai quand une contrainte d'unicité porte sur cette seule colonne
     * @param string|null  $commentaire  commentaire de la colonne en base
     * @param Origine|null $origine      d'où vient le type retenu
     * @param ExpressionDefaut|null $defautExpression sens d'un défaut calculé reconnu, exclusif de $defaut ;
     *                                                après les autres pour ne rien décaler chez un appelant
     * @param bool                  $longueurFixe     chaîne ou binaire de longueur fixe, que la base complète :
     *                                                recréée en longueur variable, la colonne se compare
     *                                                autrement ; après les autres, pour la même raison
     * @param Generee|null          $generee          calcul d'une colonne générée, à relire après chaque
     *                                                écriture ; après les autres, pour la même raison
     * @param string|null           $collation        collation explicite, sous son nom de catalogue ; absente
     *                                                pour celle de la base, et pour une collation hors du
     *                                                schéma système, que l'inférence ne reporte pas ; dernier
     *                                                paramètre, pour la même raison
     */
    public function __construct(
        public readonly string $nom,
        public readonly string $colonne,
        public readonly string $typeDoctrine,
        public readonly bool $nullable,
        public readonly ?int $longueur = null,
        public readonly ?int $precision = null,
        public readonly ?int $echelle = null,
        public readonly ?string $enumeration = null,
        public readonly ?string $defaut = null,
        public readonly bool $insertable = true,
        public readonly bool $modifiable = true,
        public readonly bool $unique = false,
        public readonly ?string $commentaire = null,
        public readonly ?Origine $origine = null,
        public readonly ?ExpressionDefaut $defautExpression = null,
        public readonly bool $longueurFixe = false,
        public readonly ?Generee $generee = null,
        public readonly ?string $collation = null,
    ) {}

    /**
     * Construit la propriété depuis le JSON décodé.
     *
     * insertable et modifiable valent vrai quand ils sont absents : le binaire
     * ne les écrit que pour les colonnes qui s'écartent du défaut, sans quoi
     * chaque régénération du calque grossirait le diff.
     *
     * @param array<mixed> $donnees objet JSON décodé de la propriété
     * @param string       $chemin  chemin de l'objet dans le calque, pour les messages
     *
     * @throws CalqueInvalide champ requis absent, champ du mauvais type, origine hors vocabulaire
     */
    public static function depuisTableau(array $donnees, string $chemin): self
    {
        $nom = Lecture::chaine($donnees, 'nom', $chemin);
        $colonne = Lecture::chaine($donnees, 'colonne', $chemin);
        $typeDoctrine = Lecture::chaine($donnees, 'type_doctrine', $chemin);
        $nullable = Lecture::booleen($donnees, 'nullable', $chemin);
        $longueur = Lecture::entierOptionnel($donnees, 'longueur', $chemin);
        $precision = Lecture::entierOptionnel($donnees, 'precision', $chemin);
        $echelle = Lecture::entierOptionnel($donnees, 'echelle', $chemin);
        $enumeration = Lecture::chaineOptionnelle($donnees, 'enumeration', $chemin);
        $defaut = Lecture::chaineOptionnelle($donnees, 'defaut', $chemin);
        $insertable = Lecture::booleen($donnees, 'insertable', $chemin, defaut: true);
        $modifiable = Lecture::booleen($donnees, 'modifiable', $chemin, defaut: true);
        $unique = Lecture::booleen($donnees, 'unique', $chemin, defaut: false);
        $commentaire = Lecture::chaineOptionnelle($donnees, 'commentaire', $chemin);
        $origine = Lecture::valeurOptionnelle($donnees, 'origine', $chemin, Origine::class);
        $defautExpression = Lecture::valeurOptionnelle($donnees, 'defaut_expression', $chemin, ExpressionDefaut::class);
        $longueurFixe = Lecture::booleen($donnees, 'longueur_fixe', $chemin, defaut: false);
        $generee = Lecture::objet($donnees, 'generee', $chemin, Generee::depuisTableau(...));
        $collation = Lecture::chaineOptionnelle($donnees, 'collation', $chemin);

        return new self(
            $nom,
            $colonne,
            $typeDoctrine,
            $nullable,
            $longueur,
            $precision,
            $echelle,
            $enumeration,
            $defaut,
            $insertable,
            $modifiable,
            $unique,
            $commentaire,
            $origine,
            $defautExpression,
            $longueurFixe,
            $generee,
            $collation,
        );
    }
}
