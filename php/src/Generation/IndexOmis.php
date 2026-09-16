<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

/**
 * Index de la table, omis parce que DBAL refuse son nom.
 *
 * DBAL n'accepte qu'un nom en [a-zA-Z0-9_] et lève IndexNameInvalid en
 * construisant la table : gardé, il empêcherait l'entité de produire son
 * schéma, index compris. Les deux autres sorties sont pires — citer le nom ne
 * le sauve pas, le contrôle ayant lieu avant la décitation, et l'omettre
 * laisserait Doctrine en inventer un à la place de celui que porte la base.
 *
 * L'entité reste générée avec ses autres index : ce qui manque est nommé
 * plutôt que deviné, et se répare en base d'un renommage.
 */
final class IndexOmis
{
    /**
     * @param string $entite entité dont la table porte l'index
     * @param string $nom    nom de l'index, tel que le catalogue le rend
     * @param bool   $unique l'index porte une unicité, que le schéma régénéré perd avec lui
     */
    public function __construct(
        public readonly string $entite,
        public readonly string $nom,
        public readonly bool $unique,
    ) {}

    /**
     * Rend l'avertissement en une ligne : ce qui manque, ce que ça coûte, et
     * comment le retrouver.
     */
    public function message(): string
    {
        return sprintf(
            '%s : %s laissé de côté, Doctrine n\'acceptant qu\'un nom en [a-zA-Z0-9_].%s Le renommer en base le rétablit ; le garder empêcherait l\'entité de produire son schéma.',
            $this->entite,
            ($this->unique ? 'unicité ' : 'index ') . $this->nom,
            $this->unique ? ' Le schéma régénéré perd cette unicité, que la base, elle, garde.' : '',
        );
    }
}
