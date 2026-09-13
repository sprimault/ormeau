<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

use Ormeau\Doctrine\Calque\Entite;
use Ormeau\Doctrine\Calque\Propriete;
use Ormeau\Doctrine\Calque\StrategieIdentifiant;

/**
 * Écrit le source des deux classes d'une entité, en mode classe de base.
 *
 * Ce qui va où n'est pas un choix de présentation, c'est ce que Doctrine
 * accepte, constaté sous ORM 2.14 à 3.7 : #[ORM\Table] posé sur un
 * MappedSuperclass est ignoré, et l'héritage y est refusé. Entity et Table vont
 * donc dans la classe de l'utilisateur, écrite une fois ; propriétés,
 * identifiant, accesseurs et index vont dans la classe de base, réécrite à
 * chaque passage.
 *
 * Le rendu ne lit rien d'autre que l'entité, la cible et l'espace de noms : deux
 * appels identiques rendent deux sources identiques octet pour octet.
 *
 * @internal
 */
final class RenduEntite
{
    /**
     * @param Cible  $cible        détermine le type PHP de chaque propriété
     * @param string $espaceDeNoms espace de noms des classes de l'utilisateur ; celui des classes
     *                             de base y ajoute \Base
     * @param bool   $avecSchema   écrire le schéma dans #[ORM\Table] ; vrai quand les entités du
     *                             calque viennent de plusieurs schémas, faute de quoi deux tables
     *                             de même nom se confondraient
     */
    public function __construct(
        private readonly Cible $cible,
        private readonly string $espaceDeNoms,
        private readonly bool $avecSchema,
    ) {}

    /**
     * Rend le nom de la classe de base d'une entité.
     */
    public static function nomBase(Entite $entite): string
    {
        return $entite->nom . 'Base';
    }

    /**
     * Rend les arguments attendus de #[ORM\Table] pour une entité. Le contrôle
     * de la classe de l'utilisateur compare à ces valeurs-là, et à rien
     * d'autre.
     *
     * @return array{name: string, schema: string|null}
     */
    public function argumentsTable(Entite $entite): array
    {
        return [
            'name' => self::identifiantSql($entite->table->nom),
            'schema' => $this->avecSchema ? self::identifiantSql($entite->table->schema) : null,
        ];
    }

    /**
     * Rend un nom de table ou de colonne tel que Doctrine doit l'écrire.
     *
     * Doctrine ne cite un identifiant que s'il est entouré de backticks. Un nom
     * qui n'est pas un identifiant SQL simple en minuscules — « N° Commande »,
     * T_CLIENTS — doit l'être : sans guillemets, PostgreSQL ramène T_CLIENTS à
     * t_clients, et la requête vise une table qui n'existe pas. Les autres
     * restent nus, comme on les écrirait à la main.
     *
     * Les mots réservés d'un SGBD ne sont pas traités : la liste dépend du
     * SGBD, que le calque logique ne porte pas.
     */
    public static function identifiantSql(string $nom): string
    {
        return preg_match('/^[a-z_][a-z0-9_]*$/', $nom) === 1 ? $nom : '`' . $nom . '`';
    }

    /**
     * Rend l'espace de noms complet de la classe de base d'une entité.
     */
    public function classeBaseQualifiee(Entite $entite): string
    {
        return $this->espaceDeNoms . '\\Base\\' . self::nomBase($entite);
    }

    /**
     * Rend le source de la classe de l'utilisateur : l'entité, sa table, et
     * rien d'autre. Elle n'est écrite qu'une fois, et ce qu'on y ajoute ensuite
     * appartient à l'utilisateur.
     */
    public function classeUtilisateur(Entite $entite): string
    {
        $lignes = [
            '<?php',
            '',
            'declare(strict_types=1);',
            '',
            'namespace ' . $this->espaceDeNoms . ';',
            '',
            'use ' . $this->classeBaseQualifiee($entite) . ';',
            'use Doctrine\ORM\Mapping as ORM;',
            '',
            Emetteur::attribut('ORM\Entity', [], ''),
            Emetteur::attribut('ORM\Table', $this->argumentsTable($entite), ''),
            'class ' . $entite->nom . ' extends ' . self::nomBase($entite) . ' {}',
        ];

        return implode("\n", $lignes) . "\n";
    }

    /**
     * Rend le source de la classe de base : propriétés mappées, index, puis
     * accesseurs.
     */
    public function classeBase(Entite $entite): string
    {
        $imports = ['Doctrine\ORM\Mapping as ORM'];
        $membres = [];
        $accesseurs = [];

        foreach ($entite->proprietes as $propriete) {
            $type = $this->type($entite, $propriete);

            // Une classe qualifiée s'importe et s'écrit par son nom court,
            // comme on l'écrirait à la main : DateTimeImmutable, pas
            // \DateTimeImmutable à chaque déclaration.
            $nu = ltrim($type, '?');
            if (str_starts_with($nu, '\\')) {
                $qualifiee = ltrim($nu, '\\');
                $imports[] = $qualifiee;
                $segments = explode('\\', $qualifiee);
                $type = ($nu === $type ? '' : '?') . end($segments);
            }

            $membres[] = $this->propriete($entite, $propriete, $type);
            $accesseurs[] = $this->accesseurs($entite, $propriete, $type);
        }

        $imports = array_unique($imports);
        sort($imports);

        $lignes = ['<?php', '', 'declare(strict_types=1);', '', 'namespace ' . $this->espaceDeNoms . '\\Base;', ''];
        foreach ($imports as $import) {
            $lignes[] = 'use ' . $import . ';';
        }
        $lignes[] = '';
        $lignes[] = Emetteur::attribut('ORM\MappedSuperclass', [], '');
        foreach ($entite->index as $index) {
            $lignes[] = Emetteur::attribut(
                $index->unique ? 'ORM\UniqueConstraint' : 'ORM\Index',
                ['name' => $index->nom, 'columns' => $index->colonnes],
                '',
            );
        }
        $lignes[] = 'abstract class ' . self::nomBase($entite);
        $lignes[] = '{';
        $lignes[] = implode("\n\n", array_merge($membres, $accesseurs));
        $lignes[] = '}';

        return implode("\n", $lignes) . "\n";
    }

    /**
     * Dit si la valeur d'une propriété est produite par la base à l'insertion :
     * clé IDENTITY ou séquence, sur une clé à une seule colonne.
     */
    private function estGeneree(Entite $entite, Propriete $propriete): bool
    {
        $identifiant = $entite->identifiant;

        return $identifiant !== null
            && $identifiant->proprietes === [$propriete->nom]
            && in_array($identifiant->strategie, [StrategieIdentifiant::Identite, StrategieIdentifiant::Sequence], true);
    }

    /**
     * Rend le type déclaré d'une propriété.
     *
     * Une clé produite par la base est nullable en PHP même quand la colonne ne
     * l'est pas : elle vaut null tant que l'entité n'est pas persistée.
     */
    private function type(Entite $entite, Propriete $propriete): string
    {
        $type = TypesPhp::declaration($propriete, $this->cible);
        if ($this->estGeneree($entite, $propriete) && $type !== 'mixed' && !str_starts_with($type, '?')) {
            return '?' . $type;
        }

        return $type;
    }

    /**
     * Rend une propriété mappée, attributs compris.
     */
    private function propriete(Entite $entite, Propriete $propriete, string $type): string
    {
        $indentation = Emetteur::INDENTATION;
        $lignes = [];

        $identifiant = $entite->identifiant;
        if ($identifiant !== null && in_array($propriete->nom, $identifiant->proprietes, true)) {
            $lignes[] = Emetteur::attribut('ORM\Id', [], $indentation);
            if ($this->estGeneree($entite, $propriete)) {
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

        $lignes[] = Emetteur::attribut('ORM\Column', [
            'name' => self::identifiantSql($propriete->colonne),
            'type' => $propriete->typeDoctrine,
            'length' => $propriete->longueur,
            'precision' => $propriete->precision,
            'scale' => $propriete->echelle,
            'unique' => $propriete->unique ? true : null,
            'nullable' => $propriete->nullable ? true : null,
            'insertable' => $propriete->insertable ? null : false,
            'updatable' => $propriete->modifiable ? null : false,
            'options' => $options === [] ? null : $options,
        ], $indentation);

        $initialisation = match (true) {
            $this->estGeneree($entite, $propriete) => ' = null',
            $defaut !== null && $type !== 'mixed' => ' = ' . Emetteur::litteral($defaut),
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
    private function accesseurs(Entite $entite, Propriete $propriete, string $type): string
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
        if ($ecrivable && !$this->estGeneree($entite, $propriete)) {
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
     * Convertit le défaut littéral du calque dans le type PHP de la propriété.
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
}
