<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

use Ormeau\Doctrine\Calque\Entite;
use Ormeau\Doctrine\Calque\Propriete;
use Ormeau\Doctrine\Calque\StrategieHeritage;

/**
 * Les hiérarchies d'héritage d'un calque : qui descend de qui, ce que la
 * racine déclare, et ce qui empêche Doctrine d'en charger une.
 *
 * Un héritage n'existe dans le calque que déclaré par décision, avec sa
 * colonne discriminante et une valeur par classe. Constaté sous ORM 2.14, 2.20
 * et 3.7, la traduction en héritage joint impose trois choses que le calque
 * ne dit pas lui-même :
 *
 *  - la colonne discriminante n'est pas un champ : Doctrine refuse de la
 *    mapper deux fois, même en lecture seule ;
 *  - une classe fille ne redéclare pas l'identifiant, qu'elle hérite de la
 *    racine : ORM 2 l'accepte, ORM 3 le refuse ;
 *  - Doctrine joint chaque table à la racine par les noms de colonnes de la
 *    clé de la racine. Une table fille dont la clé s'appelle personne_id
 *    quand celle de la racine s'appelle id ne se reprend pas : Doctrine
 *    chercherait une colonne id qui n'existe pas.
 *
 * Une hiérarchie se génère entière ou pas du tout : la carte de la racine cite
 * chaque classe, et une ligne dont la classe manquerait ferait échouer tout
 * chargement de la racine.
 *
 * @internal
 */
final class Hierarchies
{
    /** @var array<string, Entite> entités par nom, la première l'emportant en cas de collision */
    private array $parNom = [];

    /** @var array<string, true> noms des entités qu'au moins une autre désigne comme parent */
    private array $parents = [];

    /**
     * @param list<Entite> $entites entités du calque, dans son ordre
     */
    public function __construct(
        private readonly array $entites,
    ) {
        foreach ($entites as $entite) {
            $this->parNom[$entite->nom] ??= $entite;
            if ($entite->heritage !== null) {
                $this->parents[$entite->heritage->parent] = true;
            }
        }
    }

    /**
     * Rend l'entité parente, ou null hors héritage et quand le parent manque
     * au calque.
     */
    public function parent(Entite $entite): ?Entite
    {
        return $entite->heritage === null ? null : $this->parNom[$entite->heritage->parent] ?? null;
    }

    /**
     * Rend la racine de la hiérarchie à laquelle l'entité appartient, elle-même
     * comprise, ou null quand elle n'appartient à aucune.
     *
     * Une chaîne de parents qui boucle, ce qu'aucune inférence ne produit mais
     * qu'un calque écrit à la main peut contenir, s'arrête au premier retour.
     */
    public function racine(Entite $entite): ?Entite
    {
        if ($entite->heritage === null && !isset($this->parents[$entite->nom])) {
            return null;
        }

        $vues = [];
        while (($parent = $this->parent($entite)) !== null && !isset($vues[$parent->nom])) {
            $vues[$entite->nom] = true;
            $entite = $parent;
        }

        return $entite;
    }

    /**
     * Rend les entités de la hiérarchie d'une racine, racine comprise, dans
     * l'ordre du calque.
     *
     * @return list<Entite>
     */
    public function membres(Entite $racine): array
    {
        return array_values(array_filter($this->entites, fn(Entite $e): bool => $this->racine($e) === $racine));
    }

    /**
     * Dit ce que la racine d'une hiérarchie déclare, ou null pour une entité
     * qui n'est pas une racine.
     *
     * @param Entite                  $entite       entité examinée
     * @param string                  $espaceDeNoms espace de noms des entités du calque
     * @param ClassesUtilisateur|null $classes      où vivent les classes de l'utilisateur ; la carte
     *                                              cite chacune sous son nom qualifié réel
     */
    public function racineHeritage(Entite $entite, string $espaceDeNoms, ?ClassesUtilisateur $classes = null): ?RacineHeritage
    {
        if ($this->racine($entite) !== $entite) {
            return null;
        }

        $qualifiee = static fn(string $nom): string => $classes?->qualifiee($nom) ?? $espaceDeNoms . '\\' . $nom;
        $propriete = $this->proprieteDiscriminante($entite);
        $carte = [];
        foreach ($this->membres($entite) as $membre) {
            $carte[(string) $membre->valeurDiscriminante] = $qualifiee($membre->nom);
        }

        return new RacineHeritage([
            'name' => RenduEntite::identifiantSql((string) $this->colonneDiscriminante($entite)),
            'type' => $propriete?->typeDoctrine,
            'length' => $propriete?->longueur,
        ], $carte, $qualifiee($entite->nom));
    }

    /**
     * Rend les propriétés qu'une entité de hiérarchie ne déclare pas : la
     * colonne discriminante pour la racine, l'identifiant hérité pour une
     * classe fille.
     *
     * @return list<Propriete> les propriétés de l'entité, sans celles-là
     */
    public function proprietesDeclarees(Entite $entite): array
    {
        $racine = $this->racine($entite);
        if ($racine === null) {
            return $entite->proprietes;
        }

        $retirees = $racine === $entite
            ? [(string) $this->colonneDiscriminante($racine)]
            : self::colonnesDeCle($entite);

        return array_values(array_filter(
            $entite->proprietes,
            static fn(Propriete $p): bool => !in_array($p->colonne, $retirees, true),
        ));
    }

    /**
     * Dit pourquoi la hiérarchie de cette entité ne peut pas être générée, ou
     * null quand elle le peut ou que l'entité n'hérite de rien.
     *
     * La raison est donnée par l'entité en cause ; les autres membres de la
     * hiérarchie sont écartés ensuite, avec la leur.
     */
    public function raison(Entite $entite): ?string
    {
        $heritage = $entite->heritage;
        if ($heritage !== null && $heritage->strategie === StrategieHeritage::TableUnique) {
            return 'héritage en table unique, que ni l\'inférence ni les décisions ne produisent : ce paquet ne le génère pas';
        }
        if ($heritage !== null && $this->parent($entite) === null) {
            return sprintf('le parent %s est absent du calque', $heritage->parent);
        }

        $racine = $this->racine($entite);
        if ($racine === null) {
            return null;
        }
        if ($this->colonneDiscriminante($racine) === null || $entite->valeurDiscriminante === null) {
            return sprintf(
                'la hiérarchie de %s est incomplète : colonne discriminante ou valeur de %s absente du calque',
                $racine->nom,
                $entite->nom,
            );
        }

        $cle = self::colonnesDeCle($entite);
        $cleRacine = self::colonnesDeCle($racine);
        if ($cle !== $cleRacine) {
            return sprintf(
                'Doctrine joint %s à %s par les colonnes de clé de la racine (%s), que %s nomme %s. '
                . 'Retirer %s des valeurs de heritages',
                self::table($entite),
                self::table($racine),
                implode(', ', $cleRacine),
                self::table($entite),
                implode(', ', $cle),
                self::table($entite),
            );
        }

        return null;
    }

    /**
     * Dit si un ancêtre de l'entité a un constructeur, que celui de l'entité
     * doit alors appeler : il initialise des collections que la classe fille
     * hérite.
     *
     * @param callable(Entite): bool $aUnConstructeur dit si la classe de base d'une entité en déclare un
     */
    public function ancetreAvecConstructeur(Entite $entite, callable $aUnConstructeur): bool
    {
        $vues = [$entite->nom => true];
        while (($parent = $this->parent($entite)) !== null && !isset($vues[$parent->nom])) {
            if ($aUnConstructeur($parent)) {
                return true;
            }
            $vues[$parent->nom] = true;
            $entite = $parent;
        }

        return false;
    }

    /**
     * Rend la colonne discriminante d'une hiérarchie, portée par l'héritage
     * de ses classes filles ; null quand aucune ne la donne.
     */
    public function colonneDiscriminante(Entite $racine): ?string
    {
        foreach ($this->membres($racine) as $membre) {
            if ($membre->heritage?->colonneDiscriminante !== null) {
                return $membre->heritage->colonneDiscriminante;
            }
        }

        return null;
    }

    /**
     * Rend la propriété de la racine posée sur la colonne discriminante, ou
     * null quand la colonne n'en porte pas — écartée par colonnes_ignorees,
     * par exemple.
     */
    private function proprieteDiscriminante(Entite $racine): ?Propriete
    {
        $colonne = $this->colonneDiscriminante($racine);
        foreach ($racine->proprietes as $propriete) {
            if ($propriete->colonne === $colonne) {
                return $propriete;
            }
        }

        return null;
    }

    /**
     * Rend les colonnes de la clé primaire d'une entité, dans l'ordre de son
     * identifiant.
     *
     * @return list<string>
     */
    private static function colonnesDeCle(Entite $entite): array
    {
        $colonnes = [];
        foreach ($entite->identifiant->proprietes ?? [] as $nom) {
            foreach ($entite->proprietes as $propriete) {
                if ($propriete->nom === $nom) {
                    $colonnes[] = $propriete->colonne;
                }
            }
        }

        return $colonnes;
    }

    /**
     * Rend le nom qualifié de la table d'une entité, celui que l'utilisateur
     * retrouve dans le fichier de décisions.
     */
    private static function table(Entite $entite): string
    {
        return $entite->table->schema . '.' . $entite->table->nom;
    }
}
