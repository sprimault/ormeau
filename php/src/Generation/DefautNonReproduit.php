<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

/**
 * UUID tiré par la base à l'insertion, que la génération ne reproduit pas.
 *
 * DBAL n'a aucune expression pour gen_random_uuid() ou newid() : une chaîne
 * serait écrite en littéral. La propriété reste modifiable, comme le défaut
 * l'est en base — une application qui pose son propre UUID doit pouvoir le
 * faire —, mais Doctrine écrit la colonne à chaque insertion : sans valeur,
 * l'insertion échoue. La rendre générée par la base (insertable: false) aurait
 * ignoré en silence la valeur posée par l'application.
 *
 * schema:update et migrations:diff proposent de supprimer le défaut, qui
 * n'existe pas dans le mapping : l'appliquer le retirerait aux autres
 * applications qui insèrent sans valeur.
 */
final class DefautNonReproduit
{
    /**
     * @param string $classe    entité ou trait qui porte la propriété
     * @param string $propriete propriété dont la colonne porte le défaut
     * @param string $colonne   colonne, telle que la base la nomme
     */
    public function __construct(
        public readonly string $classe,
        public readonly string $propriete,
        public readonly string $colonne,
    ) {}

    /**
     * Rend l'avertissement en une ligne : ce qui manque, ce qu'il faut faire,
     * et ce qu'il ne faut pas appliquer.
     */
    public function message(): string
    {
        return sprintf(
            '%s::%s : la base tire un UUID pour la colonne %s, et Doctrine ne sait pas écrire ce défaut. Fournir la valeur avant la première insertion, par exemple dans le constructeur de la classe. schema:update et migrations:diff proposeront de supprimer le défaut : ne pas l\'appliquer.',
            $this->classe,
            $this->propriete,
            $this->colonne,
        );
    }
}
