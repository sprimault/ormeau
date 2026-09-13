<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

use LogicException;
use Ormeau\Doctrine\Calque\CalqueLogique;
use Ormeau\Doctrine\Calque\Entite;
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
 * Une entité se génère entière ou pas du tout. Ce que ce générateur ne sait pas
 * encore traduire — associations, héritage, traits, énumérations — écarte
 * l'entité avec sa raison : une entité sans ses associations perdrait ses
 * colonnes de jointure, et Doctrine la chargerait sans rien dire.
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
     * Écrit les entités du calque dans le répertoire donné.
     *
     * Les classes de base vont dans Base/ et sont réécrites quand leur contenu
     * change, laissées intactes sinon : un fichier identique garde sa date, et
     * un outil qui surveille le répertoire ne voit rien bouger. Les classes de
     * l'utilisateur vont à la racine du répertoire, sont créées si elles
     * manquent, et ne sont jamais réécrites : elles sont seulement relues pour
     * signaler ce qui a divergé.
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

        $schemas = array_unique(array_map(static fn(Entite $e): string => $e->table->schema, $calque->entites));
        $rendu = new RenduEntite($cible, $calque->espaceDeNoms, count($schemas) > 1);
        $controle = new ControleClasseUtilisateur();
        $repertoire = rtrim($repertoire, '/\\');

        // Deux entités de même nom viennent d'une collision que l'inférence a
        // signalée sans la trancher. Aucune des deux ne s'écrit : choisir celle
        // qui garde le nom serait une décision, et PHP comme les systèmes de
        // fichiers de Windows et macOS ignorent la casse.
        $occurrences = array_count_values(array_map(static fn(Entite $e): string => strtolower($e->nom), $calque->entites));

        $fichiers = [];
        $ecartees = [];
        $divergences = [];

        foreach ($calque->entites as $entite) {
            $raison = $occurrences[strtolower($entite->nom)] > 1
                ? sprintf('le nom %s est porté par plusieurs entités, à départager dans renommages', $entite->nom)
                : $this->raisonDEcarter($entite);
            if ($raison !== null) {
                $ecartees[] = new EntiteEcartee($entite->nom, $raison);
                continue;
            }

            $base = $repertoire . '/Base/' . RenduEntite::nomBase($entite) . '.php';
            $fichiers[] = new Fichier($base, $this->ecrire($base, $rendu->classeBase($entite)));

            $utilisateur = $repertoire . '/' . $entite->nom . '.php';
            if (!is_file($utilisateur)) {
                $this->ecrire($utilisateur, $rendu->classeUtilisateur($entite));
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
            ));
        }

        return new Rapport($fichiers, $ecartees, $divergences);
    }

    /**
     * Dit pourquoi une entité ne peut pas être générée, ou null quand elle le
     * peut.
     *
     * Les raisons nomment ce que l'utilisateur peut faire : une table sans clé
     * se résout par une décision, un nom réservé par un renommage ; le reste
     * attend une version du générateur.
     */
    private function raisonDEcarter(Entite $entite): ?string
    {
        if (in_array(strtolower($entite->nom), self::NOMS_RESERVES, true)) {
            return sprintf('%s est un mot réservé de PHP, à renommer dans renommages', $entite->nom);
        }
        if ($entite->identifiant === null) {
            return 'la table n\'a pas de clé primaire, et Doctrine exige un identifiant';
        }

        $manques = [];
        if ($entite->associations !== []) {
            $manques[] = 'associations';
        }
        if ($entite->heritage !== null) {
            $manques[] = 'héritage';
        }
        if ($entite->traits !== []) {
            $manques[] = 'traits';
        }
        foreach ($entite->proprietes as $propriete) {
            if ($propriete->enumeration !== null) {
                $manques[] = 'énumérations';
                break;
            }
        }

        return $manques === [] ? null : 'pas encore générées par ce paquet : ' . implode(', ', $manques);
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
