<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

/**
 * Racine d'une hiérarchie jointe portant une colonne que la base calcule,
 * visée par une version d'ORM qui l'écrit quand même.
 *
 * Jusqu'à ORM 2.15 comprise, JoinedSubclassPersister::getInsertColumnList()
 * garde la colonne dans l'INSERT sans lui fournir de valeur : PostgreSQL
 * répond « cannot insert a non-DEFAULT value into column », et aucune entité
 * de la hiérarchie ne s'enregistre, racine comprise. Corrigé en 2.16.0 par
 * doctrine/orm#10598, jamais rétroporté : la seule sortie est de monter de
 * version.
 *
 * Le rendu, lui, est juste — c'est l'ORM installé qui est en défaut. La
 * hiérarchie est donc générée entière, et le rapport prévient, plutôt que
 * d'écarter des classes qu'une mise à jour d'ORM rendrait utilisables telles
 * quelles. Le défaut ne se voit qu'à la première écriture, jamais à la
 * génération ni à validateMapping.
 */
final class ColonneGenereeNonExclue
{
    /**
     * @param string $racine    entité racine de la hiérarchie jointe
     * @param string $propriete propriété que la base calcule, sous son nom de classe
     * @param string $colonne   colonne correspondante, sous son nom de base
     * @param string $orm       version d'ORM visée, majeure et mineure
     */
    public function __construct(
        public readonly string $racine,
        public readonly string $propriete,
        public readonly string $colonne,
        public readonly string $orm,
    ) {}

    /**
     * Rend l'avertissement en une ligne : ce qui ne marchera pas, quand cela
     * se verra, et la seule sortie.
     */
    public function message(): string
    {
        return sprintf(
            '%s : la racine de cette hiérarchie jointe porte une colonne générée (%s), que Doctrine ORM %s n\'exclut pas de l\'INSERT. Aucune entité de la hiérarchie ne s\'enregistrera, racine comprise, et rien ne le signale avant la première écriture. Corrigé en ORM 2.16 ; prendre 2.16.1, la 2.16.0 perdant des mises à jour sur les classes à héritage. Ni 2.14 ni 2.15 n\'ont de version corrective.',
            $this->racine,
            $this->colonne,
            $this->orm,
        );
    }
}
