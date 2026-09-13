<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

use LogicException;
use Ormeau\Doctrine\Calque\Association;
use Ormeau\Doctrine\Calque\CalqueLogique;
use Ormeau\Doctrine\Calque\Entite;
use Ormeau\Doctrine\Calque\Enumeration;
use RuntimeException;

/**
 * Traduit un calque logique en entités Doctrine. Ne décide rien : toute
 * heuristique appartient à l'inférence, côté Go.
 *
 * Deux modes de régénération, parce que c'est là que meurent tous les
 * générateurs de ce type :
 *
 *  - ClasseDeBase : ClientBase est régénérée, Client appartient à l'utilisateur.
 *    Robuste, impose une hiérarchie. Mode par défaut, le seul écrit.
 *  - ReecritureAst : nikic/php-parser relit l'entité existante et ne réécrit que
 *    les attributs qui ont bougé, en conservant méthodes métier et formatage.
 *    Phase « Régénération par AST » de la feuille de route.
 *
 * Une entité se génère entière ou pas du tout. Ce que Doctrine ne sait pas
 * représenter écarte l'entité avec sa raison, et l'écart se propage à celles
 * qui la visent — une entité privée d'une association perdrait sa colonne de
 * jointure, et Doctrine la chargerait sans rien dire — comme à toute sa
 * hiérarchie d'héritage.
 */
final class GenerateurEntite
{
    /**
     * Noms que PHP refuse pour une classe, en minuscules.
     *
     * L'inférence peut en produire un — une table list ou match —, et le
     * générateur ne renomme pas : c'est une décision, qui va dans renommages.
     */
    private const NOMS_RESERVES = [
        'abstract', 'and', 'array', 'as', 'bool', 'break', 'callable', 'case', 'catch', 'class', 'clone',
        'const', 'continue', 'declare', 'default', 'do', 'echo', 'else', 'elseif', 'empty', 'enddeclare',
        'endfor', 'endforeach', 'endif', 'endswitch', 'endwhile', 'enum', 'eval', 'exit', 'extends', 'false',
        'final', 'finally', 'float', 'fn', 'for', 'foreach', 'function', 'global', 'goto', 'if', 'implements',
        'include', 'include_once', 'instanceof', 'insteadof', 'int', 'interface', 'isset', 'iterable', 'list',
        'match', 'mixed', 'namespace', 'never', 'new', 'null', 'object', 'or', 'parent', 'print', 'private',
        'protected', 'public', 'readonly', 'require', 'require_once', 'return', 'self', 'static', 'string',
        'switch', 'throw', 'trait', 'true', 'try', 'unset', 'use', 'var', 'void', 'while', 'xor', 'yield',
    ];

    /**
     * Retient le mode de régénération, la classe de base séparée à défaut.
     *
     * @param ModeRegeneration $mode stratégie de préservation du code écrit à la main entre deux
     *                               régénérations
     */
    public function __construct(
        private readonly ModeRegeneration $mode = ModeRegeneration::ClasseDeBase,
    ) {}

    /**
     * Rend le mode retenu à la construction.
     */
    public function mode(): ModeRegeneration
    {
        return $this->mode;
    }

    /**
     * Écrit les énumérations, les traits et les entités du calque dans le
     * répertoire donné.
     *
     * Base/, Enum/ et Trait/ appartiennent à l'outil : leurs fichiers sont
     * réécrits quand leur contenu change, laissés intacts sinon — un fichier
     * identique garde sa date, et un outil qui surveille le répertoire ne voit
     * rien bouger. Les classes de l'utilisateur vont à la racine du répertoire,
     * sont créées si elles manquent, et ne sont jamais réécrites : elles sont
     * seulement relues pour signaler ce qui a divergé.
     *
     * @param CalqueLogique $calque     calque déjà lu et contrôlé
     * @param string        $repertoire racine des entités, src/Entity dans une application Symfony
     * @param Cible         $cible      version d'ORM visée, qui fixe le type PHP de certaines colonnes
     *
     * @throws LogicException   mode de régénération qui n'est pas encore écrit
     * @throws RuntimeException répertoire ou fichier qui ne s'écrit pas
     */
    public function generer(CalqueLogique $calque, string $repertoire, Cible $cible): Rapport
    {
        if ($this->mode !== ModeRegeneration::ClasseDeBase) {
            throw new LogicException(sprintf(
                'Génération en mode %s : à implémenter, phase « Régénération par AST » de la feuille de route.',
                $this->mode->value,
            ));
        }

        $enumerations = [];
        foreach ($calque->enumerations as $enumeration) {
            $enumerations[$enumeration->nom] = $enumeration;
        }
        $traits = [];
        foreach ($calque->traits as $trait) {
            $traits[$trait->nom] = $trait;
        }

        $schemas = array_unique(array_map(static fn(Entite $e): string => $e->table->schema, $calque->entites));
        $rendu = new RenduEntite($cible, $calque->espaceDeNoms, count($schemas) > 1, $enumerations);
        $controle = new ControleClasseUtilisateur();
        $repertoire = rtrim($repertoire, '/\\');

        $fichiers = [];
        $ecartees = [];
        $divergences = [];

        // Énumérations et traits d'abord : les classes de base les importent.
        // Un nom que PHP refuse n'est pas écrit, et les entités qui s'en
        // servent sont écartées plus bas avec cette raison.
        $refus = [];
        foreach ($enumerations as $nom => $enumeration) {
            $refus['enum:' . $nom] = $this->refusEnumeration($enumeration);
            if ($refus['enum:' . $nom] === null) {
                $chemin = $repertoire . '/Enum/' . $nom . '.php';
                $fichiers[] = new Fichier($chemin, $this->ecrire($chemin, $rendu->enumeration($enumeration)));
            }
        }
        foreach ($traits as $nom => $trait) {
            $refus['trait:' . $nom] = self::estReserve($nom) ? sprintf('%s est un mot réservé de PHP', $nom) : null;
            if ($refus['trait:' . $nom] === null) {
                $chemin = $repertoire . '/Trait/' . $nom . '.php';
                $fichiers[] = new Fichier($chemin, $this->ecrire($chemin, $rendu->traitPartage($trait)));
            }
        }

        // Deux entités de même nom viennent d'une collision que l'inférence a
        // signalée sans la trancher. Aucune des deux ne s'écrit : choisir celle
        // qui garde le nom serait une décision, et PHP comme les systèmes de
        // fichiers de Windows et macOS ignorent la casse.
        $occurrences = array_count_values(array_map(static fn(Entite $e): string => strtolower($e->nom), $calque->entites));

        $hierarchies = new Hierarchies($calque->entites);
        $raisons = [];
        foreach ($calque->entites as $rang => $entite) {
            $raisons[$rang] = $occurrences[strtolower($entite->nom)] > 1
                ? sprintf('le nom %s est porté par plusieurs entités, à départager dans renommages', $entite->nom)
                : $this->raisonDEcarter($entite, $refus) ?? $hierarchies->raison($entite);
        }
        $raisons = $this->ecarterLesIdentitesEnChaine($calque->entites, $raisons);
        $raisons = $this->propagerLesEcarts($calque->entites, $raisons, $hierarchies);

        $omises = [];
        $generees = [];
        foreach ($calque->entites as $rang => $entite) {
            if ($raisons[$rang] !== null) {
                $ecartees[] = new EntiteEcartee($entite->nom, $raisons[$rang]);
                continue;
            }
            $generees[] = $this->sansCotesInversesOrphelins($entite, $calque->entites, $raisons, $omises);
        }

        // Les hiérarchies se relisent sur les entités retenues : une classe
        // fille appelle le constructeur de son parent selon les collections
        // qui lui restent, côtés inverses omis déduits.
        $hierarchies = new Hierarchies($generees);
        foreach ($generees as $entite) {
            $base = $repertoire . '/Base/' . RenduEntite::nomBase($entite) . '.php';
            $fichiers[] = new Fichier($base, $this->ecrire($base, $rendu->classeBase($entite, $hierarchies)));

            $racine = $hierarchies->racineHeritage($entite, $calque->espaceDeNoms);
            $utilisateur = $repertoire . '/' . $entite->nom . '.php';
            if (!is_file($utilisateur)) {
                $this->ecrire($utilisateur, $rendu->classeUtilisateur($entite, $racine));
                $fichiers[] = new Fichier($utilisateur, EtatFichier::Cree);
                continue;
            }

            $fichiers[] = new Fichier($utilisateur, EtatFichier::Conserve);
            $source = file_get_contents($utilisateur);
            if ($source === false) {
                throw new RuntimeException(sprintf('Classe illisible : %s', $utilisateur));
            }
            array_push($divergences, ...$controle->comparer(
                $utilisateur,
                $source,
                $entite->nom,
                $rendu->classeBaseQualifiee($entite),
                $rendu->argumentsTable($entite),
                $racine,
            ));
        }

        return new Rapport($fichiers, $ecartees, $divergences, $omises);
    }

    /**
     * Dit pourquoi une entité ne peut pas être générée, ou null quand elle le
     * peut.
     *
     * Les raisons nomment ce que l'utilisateur peut faire : une table sans clé
     * se résout par une décision, un nom réservé par un renommage. Ce qui
     * tient à une hiérarchie d'héritage est dit par Hierarchies::raison().
     *
     * @param Entite                     $entite entité à examiner
     * @param array<string, string|null> $refus  raison du refus de chaque énumération (enum:Nom) et de
     *                                           chaque trait (trait:Nom), null quand il est écrit
     */
    private function raisonDEcarter(Entite $entite, array $refus): ?string
    {
        if (self::estReserve($entite->nom)) {
            return sprintf('%s est un mot réservé de PHP, à renommer dans renommages', $entite->nom);
        }
        if ($entite->identifiant === null) {
            return 'la table n\'a pas de clé primaire, et Doctrine exige un identifiant';
        }

        foreach ($entite->traits as $trait) {
            if (!array_key_exists('trait:' . $trait, $refus)) {
                return sprintf('le trait %s est absent du calque', $trait);
            }
            if ($refus['trait:' . $trait] !== null) {
                return sprintf('le trait %s n\'est pas généré : %s', $trait, $refus['trait:' . $trait]);
            }
        }
        foreach ($entite->proprietes as $propriete) {
            $enumeration = $propriete->enumeration;
            if ($enumeration === null) {
                continue;
            }
            if (!array_key_exists('enum:' . $enumeration, $refus)) {
                return sprintf('l\'énumération %s de la propriété %s est absente du calque', $enumeration, $propriete->nom);
            }
            if ($refus['enum:' . $enumeration] !== null) {
                return sprintf('l\'énumération %s n\'est pas générée : %s', $enumeration, $refus['enum:' . $enumeration]);
            }
        }

        return null;
    }

    /**
     * Écarte une entité identifiée par une association vers une entité elle-
     * même identifiée par une association.
     *
     * Doctrine refuse cette identité dérivée en chaîne : il n'identifie par
     * association que vers un identifiant scalaire. Constaté sous ORM 2.14 et
     * 3.6, et aucune forme ne la contourne sans défaut. Garder une propriété
     * scalaire modifiable à côté de l'association rend deux écrivains à la
     * colonne, et l'association écrase la propriété en silence ; la garder en
     * lecture seule rend l'entité impossible à persister, faute d'identifiant
     * renseigné. L'entité est donc écartée, et la raison nomme la sortie : une
     * table déclarée dans un héritage prend l'identifiant scalaire de sa
     * racine, et la chaîne disparaît.
     *
     * @param list<Entite>             $entites entités du calque, dans son ordre
     * @param array<int, string|null> $raisons raison d'écarter chaque entité, par rang
     *
     * @return array<int, string|null> les mêmes raisons, identités en chaîne comprises
     */
    private function ecarterLesIdentitesEnChaine(array $entites, array $raisons): array
    {
        $parNom = [];
        foreach ($entites as $entite) {
            $parNom[$entite->nom] ??= $entite;
        }

        foreach ($entites as $rang => $entite) {
            if ($raisons[$rang] !== null) {
                continue;
            }
            foreach (RenduMembres::associationsDeCle($entite->proprietes, $entite->identifiant, $entite->associations) as $nom) {
                $association = self::association($entite, $nom);
                $cible = $parNom[$association->cible] ?? null;
                if ($cible === null || RenduMembres::associationsDeCle($cible->proprietes, $cible->identifiant, $cible->associations) === []) {
                    continue;
                }
                $raisons[$rang] = sprintf(
                    'Doctrine ne sait pas identifier %s (%s) par %s : %s (%s) est elle-même identifiée par une association. '
                    . 'Déclarer %s dans un héritage (heritages), ou écarter l\'une des deux tables (tables_ignorees)',
                    $entite->nom,
                    self::table($entite),
                    $nom,
                    $cible->nom,
                    self::table($cible),
                    self::table($cible),
                );
                break;
            }
        }

        return $raisons;
    }

    /**
     * Rend l'association d'une entité qui porte ce nom.
     */
    private static function association(Entite $entite, string $nom): Association
    {
        foreach ($entite->associations as $association) {
            if ($association->nom === $nom) {
                return $association;
            }
        }

        throw new LogicException(sprintf('Association %s introuvable sur %s.', $nom, $entite->nom));
    }

    /**
     * Rend le nom qualifié de la table d'une entité : c'est lui que
     * l'utilisateur retrouve dans le fichier de décisions.
     */
    private static function table(Entite $entite): string
    {
        return $entite->table->schema . '.' . $entite->table->nom;
    }

    /**
     * Rend l'entité sans ses côtés inverses dont l'entité propriétaire est
     * écartée, et note chaque omission.
     *
     * Un côté inverse ne porte aucune colonne : Doctrine accepte l'association
     * unidirectionnelle, et l'entité reste entière. L'écarter pour autant
     * remonterait l'écart à toute la hiérarchie — Salarie, puis Personne, puis
     * Prestataire, pour une seule Affectation écartée.
     *
     * @param Entite                  $entite  entité générée
     * @param list<Entite>            $entites entités du calque, dans son ordre
     * @param array<int, string|null> $raisons raison d'écarter chaque entité, par rang
     * @param list<AssociationOmise>  $omises  omissions notées jusqu'ici, complétées
     */
    private function sansCotesInversesOrphelins(Entite $entite, array $entites, array $raisons, array &$omises): Entite
    {
        $ecartees = [];
        foreach ($entites as $rang => $autre) {
            if ($raisons[$rang] !== null) {
                $ecartees[$autre->nom] = $autre;
            }
        }

        $gardees = [];
        foreach ($entite->associations as $association) {
            $cible = $ecartees[$association->cible] ?? null;
            if ($association->proprietaire || $cible === null) {
                $gardees[] = $association;
                continue;
            }
            $omises[] = new AssociationOmise($entite->nom, $association->nom, sprintf(
                'côté inverse de %s (%s), écartée',
                $cible->nom,
                self::table($cible),
            ));
        }
        if (count($gardees) === count($entite->associations)) {
            return $entite;
        }

        return new Entite(
            $entite->nom,
            $entite->table,
            $entite->proprietes,
            $entite->heritage,
            $entite->traits,
            $entite->identifiant,
            $gardees,
            $entite->index,
            $entite->origine,
            $entite->valeurDiscriminante,
            $entite->commentaire,
        );
    }

    /**
     * Écarte toute entité dont une association propriétaire vise une entité
     * écartée ou absente du calque, ou dont la hiérarchie compte une entité
     * écartée, jusqu'à ce que plus rien ne bouge.
     *
     * Une classe de base qui importe une classe jamais écrite ne se charge
     * pas, et l'écart se propage : une entité qui vise Affectation, écartée
     * pour son identité en chaîne, l'est à son tour. Une hiérarchie se génère
     * entière ou pas du tout, la carte de sa racine citant chaque classe.
     * Chaque entité écartée a sa propre raison : elle nomme l'association ou
     * la hiérarchie, la cible et sa table, et la cause première se lit sur la
     * ligne de la cible. Une entité absente sans explication ferait douter de
     * l'outil entier.
     *
     * @param list<Entite>             $entites     entités du calque, dans son ordre
     * @param array<int, string|null> $raisons     raison d'écarter chaque entité, par rang ; null
     *                                              pour une entité générée
     * @param Hierarchies              $hierarchies hiérarchies d'héritage du calque
     *
     * @return array<int, string|null> les mêmes raisons, écarts propagés compris
     */
    private function propagerLesEcarts(array $entites, array $raisons, Hierarchies $hierarchies): array
    {
        $rangs = [];
        foreach ($entites as $rang => $entite) {
            $rangs[$entite->nom][] = $rang;
        }

        do {
            $stable = true;
            foreach ($entites as $rang => $entite) {
                if ($raisons[$rang] !== null) {
                    continue;
                }
                $racine = $hierarchies->racine($entite);
                foreach ($racine === null ? [] : $hierarchies->membres($racine) as $membre) {
                    $rangMembre = array_search($membre, $entites, true);
                    if ($rangMembre !== false && $raisons[$rangMembre] !== null) {
                        $raisons[$rang] = sprintf('même hiérarchie que %s (%s), écartée', $membre->nom, self::table($membre));
                        $stable = false;
                        continue 2;
                    }
                }
                foreach ($entite->associations as $association) {
                    // Un côté inverse ne porte aucune colonne : il est omis
                    // plus loin, et l'entité reste générée.
                    if (!$association->proprietaire) {
                        continue;
                    }
                    $cibles = $rangs[$association->cible] ?? [];
                    if ($cibles === []) {
                        $raisons[$rang] = sprintf('l\'association %s vise %s, absente du calque', $association->nom, $association->cible);
                    } elseif (array_filter($cibles, static fn(int $r): bool => $raisons[$r] !== null) !== []) {
                        $raisons[$rang] = sprintf(
                            'l\'association %s vise %s (%s), écartée',
                            $association->nom,
                            $association->cible,
                            self::table($entites[$cibles[0]]),
                        );
                    }
                    if ($raisons[$rang] !== null) {
                        $stable = false;
                        break;
                    }
                }
            }
        } while (!$stable);

        return $raisons;
    }

    /**
     * Dit pourquoi une énumération ne peut pas être écrite, ou null.
     *
     * Deux refus de PHP : un nom réservé pour l'énumération, et un cas nommé
     * class, le seul nom de cas qu'il interdit.
     */
    private function refusEnumeration(Enumeration $enumeration): ?string
    {
        if (self::estReserve($enumeration->nom)) {
            return sprintf('%s est un mot réservé de PHP', $enumeration->nom);
        }
        foreach ($enumeration->cas as $cas) {
            if (strtolower($cas->nom) === 'class') {
                return 'un cas ne peut pas s\'appeler class';
            }
        }

        return null;
    }

    /**
     * Dit si PHP refuse un nom de classe, d'énumération ou de trait.
     */
    private static function estReserve(string $nom): bool
    {
        return in_array(strtolower($nom), self::NOMS_RESERVES, true);
    }

    /**
     * Écrit un fichier quand son contenu diffère de ce qui est sur disque.
     *
     * @throws RuntimeException répertoire ou fichier qui ne s'écrit pas
     */
    private function ecrire(string $chemin, string $contenu): EtatFichier
    {
        if (is_file($chemin)) {
            if (file_get_contents($chemin) === $contenu) {
                return EtatFichier::Inchange;
            }
            $etat = EtatFichier::Reecrit;
        } else {
            $etat = EtatFichier::Cree;
        }

        $dossier = dirname($chemin);
        if (!is_dir($dossier) && !@mkdir($dossier, 0o777, true) && !is_dir($dossier)) {
            throw new RuntimeException(sprintf('Répertoire impossible à créer : %s', $dossier));
        }
        if (@file_put_contents($chemin, $contenu) === false) {
            throw new RuntimeException(sprintf('Fichier impossible à écrire : %s', $chemin));
        }

        return $etat;
    }
}
