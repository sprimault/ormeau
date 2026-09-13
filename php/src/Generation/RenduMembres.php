<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

use Ormeau\Doctrine\Calque\Enumeration;
use Ormeau\Doctrine\Calque\Identifiant;
use Ormeau\Doctrine\Calque\Propriete;
use Ormeau\Doctrine\Calque\StrategieIdentifiant;
use Ormeau\Doctrine\Calque\TypeSupport;

/**
 * Écrit les propriétés mappées et leurs accesseurs, pour une classe de base
 * comme pour un trait.
 *
 * Les deux reçoivent exactement le même rendu : une colonne created_at doit
 * s'écrire de la même façon qu'elle vive dans ClientBase ou dans le trait
 * Horodatage, sans quoi sortir des colonnes dans un trait changerait leur
 * mapping.
 *
 * @internal
 */
final class RenduMembres
{
    /**
     * @param Cible                      $cible        détermine le type PHP de chaque propriété
     * @param string                     $espaceDeNoms espace de noms des entités ; les énumérations
     *                                                 vivent dans \Enum sous lui
     * @param array<string, Enumeration> $enumerations énumérations du calque par nom, pour typer une
     *                                                 propriété et retrouver le cas de son défaut
     */
    public function __construct(
        private readonly Cible $cible,
        private readonly string $espaceDeNoms,
        private readonly array $enumerations,
    ) {}

    /**
     * Rend les propriétés et leurs accesseurs, avec les imports qu'ils
     * demandent.
     *
     * Les propriétés d'abord, les accesseurs ensuite, dans l'ordre du calque :
     * c'est l'ordre des colonnes de la table, que l'on retrouve en lisant la
     * classe.
     *
     * @param list<Propriete> $proprietes  propriétés à rendre, dans l'ordre du calque
     * @param Identifiant|null $identifiant clé de l'entité ; null pour un trait, qui n'en porte pas
     *
     * @return array{imports: list<string>, proprietes: list<string>, accesseurs: list<string>}
     */
    public function rendre(array $proprietes, ?Identifiant $identifiant): array
    {
        $imports = [];
        $membres = [];
        $accesseurs = [];

        foreach ($proprietes as $propriete) {
            [$type, $import] = $this->type($propriete, $identifiant);
            if ($import !== null) {
                $imports[] = $import;
            }
            $membres[] = $this->propriete($propriete, $identifiant, $type);
            $accesseurs[] = $this->accesseurs($propriete, $identifiant, $type);
        }

        return ['imports' => $imports, 'proprietes' => $membres, 'accesseurs' => $accesseurs];
    }

    /**
     * Rend le nom qualifié de la classe PHP d'une énumération du calque.
     */
    public function enumerationQualifiee(string $nom): string
    {
        return $this->espaceDeNoms . '\\Enum\\' . $nom;
    }

    /**
     * Dit si la valeur d'une propriété est produite par la base à l'insertion :
     * clé IDENTITY ou séquence, sur une clé à une seule colonne.
     */
    private function estGeneree(Propriete $propriete, ?Identifiant $identifiant): bool
    {
        return $identifiant !== null
            && $identifiant->proprietes === [$propriete->nom]
            && in_array($identifiant->strategie, [StrategieIdentifiant::Identite, StrategieIdentifiant::Sequence], true);
    }

    /**
     * Rend le type déclaré d'une propriété, et la classe à importer pour
     * l'écrire par son nom court.
     *
     * Une propriété énumérée prend le type de l'énumération. Une clé produite
     * par la base est nullable en PHP même quand la colonne ne l'est pas : elle
     * vaut null tant que l'entité n'est pas persistée. Une classe qualifiée
     * s'importe, comme on l'écrirait à la main : DateTimeImmutable, pas
     * \DateTimeImmutable à chaque déclaration.
     *
     * @return array{string, string|null}
     */
    private function type(Propriete $propriete, ?Identifiant $identifiant): array
    {
        $nu = $propriete->enumeration !== null
            ? '\\' . $this->enumerationQualifiee($propriete->enumeration)
            : TypesPhp::nu($propriete, $this->cible);

        $nullable = $nu !== 'mixed' && ($propriete->nullable || $this->estGeneree($propriete, $identifiant));

        $import = null;
        if (str_starts_with($nu, '\\')) {
            $import = ltrim($nu, '\\');
            $segments = explode('\\', $import);
            $nu = end($segments);
        }

        return [($nullable ? '?' : '') . $nu, $import];
    }

    /**
     * Rend une propriété mappée, attributs compris.
     */
    private function propriete(Propriete $propriete, ?Identifiant $identifiant, string $type): string
    {
        $indentation = Emetteur::INDENTATION;
        $lignes = [];
        $generee = $this->estGeneree($propriete, $identifiant);

        if ($identifiant !== null && in_array($propriete->nom, $identifiant->proprietes, true)) {
            $lignes[] = Emetteur::attribut('ORM\Id', [], $indentation);
            if ($generee) {
                $sequence = $identifiant->strategie === StrategieIdentifiant::Sequence;
                $lignes[] = Emetteur::attribut('ORM\GeneratedValue', ['strategy' => $sequence ? 'SEQUENCE' : 'IDENTITY'], $indentation);
                if ($sequence && $identifiant->sequence !== null) {
                    $lignes[] = Emetteur::attribut('ORM\SequenceGenerator', ['sequenceName' => $identifiant->sequence], $indentation);
                }
            }
        }

        $defaut = $this->defaut($propriete);
        $options = array_filter([
            'default' => $propriete->defaut === null ? null : ($defaut ?? $propriete->defaut),
            'comment' => $propriete->commentaire,
        ], static fn($valeur): bool => $valeur !== null);

        $enumType = null;
        if ($propriete->enumeration !== null) {
            $segments = explode('\\', $this->enumerationQualifiee($propriete->enumeration));
            $enumType = new Code(end($segments) . '::class');
        }

        $lignes[] = Emetteur::attribut('ORM\Column', [
            'name' => RenduEntite::identifiantSql($propriete->colonne),
            'type' => $propriete->typeDoctrine,
            'length' => $propriete->longueur,
            'precision' => $propriete->precision,
            'scale' => $propriete->echelle,
            'unique' => $propriete->unique ? true : null,
            'nullable' => $propriete->nullable ? true : null,
            'insertable' => $propriete->insertable ? null : false,
            'updatable' => $propriete->modifiable ? null : false,
            'enumType' => $enumType,
            'options' => $options === [] ? null : $options,
        ], $indentation);

        $cas = $this->casDuDefaut($propriete);
        $initialisation = match (true) {
            $generee => ' = null',
            $cas !== null => ' = ' . $cas->php,
            $propriete->enumeration === null && $defaut !== null && $type !== 'mixed' => ' = ' . Emetteur::litteral($defaut),
            $propriete->nullable => ' = null',
            default => '',
        };
        $lignes[] = $indentation . 'protected ' . $type . ' $' . $propriete->nom . $initialisation . ';';

        return implode("\n", $lignes);
    }

    /**
     * Rend l'accesseur et, quand la propriété s'écrit depuis PHP, le
     * mutateur.
     *
     * Pas de mutateur pour une clé produite par la base, ni pour une colonne
     * générée : Doctrine ne l'écrirait pas, et une méthode qui ne fait rien
     * est pire qu'une méthode absente.
     */
    private function accesseurs(Propriete $propriete, ?Identifiant $identifiant, string $type): string
    {
        $i = Emetteur::INDENTATION;
        $suffixe = ucfirst($propriete->nom);
        $lecture = ltrim($type, '?') === 'bool' ? 'is' . $suffixe : 'get' . $suffixe;

        $methodes = [implode("\n", [
            $i . 'public function ' . $lecture . '(): ' . $type,
            $i . '{',
            $i . $i . 'return $this->' . $propriete->nom . ';',
            $i . '}',
        ])];

        $ecrivable = $propriete->insertable || $propriete->modifiable;
        if ($ecrivable && !$this->estGeneree($propriete, $identifiant)) {
            $methodes[] = implode("\n", [
                $i . 'public function set' . $suffixe . '(' . $type . ' $' . $propriete->nom . '): static',
                $i . '{',
                $i . $i . '$this->' . $propriete->nom . ' = $' . $propriete->nom . ';',
                '',
                $i . $i . 'return $this;',
                $i . '}',
            ]);
        }

        return implode("\n\n", $methodes);
    }

    /**
     * Convertit le défaut littéral du calque dans le type de la valeur stockée.
     *
     * Rend null quand la conversion n'est pas certaine — une date, un booléen
     * écrit « O » — : le défaut reste alors une option de colonne, dans sa
     * forme textuelle, sans initialiser la propriété.
     */
    private function defaut(Propriete $propriete): bool|float|int|string|null
    {
        $valeur = $propriete->defaut;
        if ($valeur === null) {
            return null;
        }

        return match (TypesPhp::nu($propriete, $this->cible)) {
            'int' => preg_match('/^-?\d+$/', $valeur) === 1 ? (int) $valeur : null,
            'float' => is_numeric($valeur) ? (float) $valeur : null,
            'bool' => match (strtolower($valeur)) {
                'true', 't', '1' => true,
                'false', 'f', '0' => false,
                default => null,
            },
            'string' => $valeur,
            default => null,
        };
    }

    /**
     * Rend le cas d'énumération qui initialise une propriété énumérée avec un
     * défaut, ou null.
     *
     * Le défaut est comparé à la valeur stockée de chaque cas, dans le type
     * support de l'énumération. Un défaut qu'aucun cas ne porte n'initialise
     * rien : inventer un cas serait décider à la place du calque.
     */
    private function casDuDefaut(Propriete $propriete): ?Code
    {
        if ($propriete->enumeration === null || $propriete->defaut === null) {
            return null;
        }
        $enumeration = $this->enumerations[$propriete->enumeration] ?? null;
        if ($enumeration === null) {
            return null;
        }

        foreach ($enumeration->cas as $cas) {
            $egal = $enumeration->typeSupport === TypeSupport::Entier
                ? (string) $cas->valeur === $propriete->defaut
                : $cas->valeur === $propriete->defaut;
            if ($egal) {
                return new Code($enumeration->nom . '::' . $cas->nom);
            }
        }

        return null;
    }
}
