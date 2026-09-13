<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

use LogicException;
use Ormeau\Doctrine\Calque\CalqueLogique;

/**
 * Traduit un calque logique en entités Doctrine. Ne décide rien : toute
 * heuristique appartient à l'inférence, côté Go.
 *
 * Deux modes de régénération, parce que c'est là que meurent tous les
 * générateurs de ce type :
 *
 *  - ClasseDeBase : ClientBase est régénérée, Client appartient à l'utilisateur.
 *    Robuste, impose une hiérarchie. Mode par défaut.
 *  - ReecritureAst : nikic/php-parser relit l'entité existante et ne réécrit que
 *    les attributs qui ont bougé, en conservant méthodes métier et formatage.
 */
final class GenerateurEntite
{
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
     * Pas encore écrite : elle échoue en nommant le mode demandé, plutôt que de
     * rendre une liste vide qui passerait pour une génération sans entité.
     *
     * @param CalqueLogique $calque     calque déjà lu et contrôlé
     * @param string        $repertoire racine des entités, src/Entity dans une application Symfony
     *
     * @return list<string> chemins des fichiers écrits
     *
     * @throws LogicException tant que la génération n'est pas écrite
     */
    public function generer(CalqueLogique $calque, string $repertoire): array
    {
        throw new LogicException(sprintf(
            'Génération en mode %s : à implémenter.',
            $this->mode->value,
        ));
    }
}
