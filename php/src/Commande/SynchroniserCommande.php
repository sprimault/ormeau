<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Commande;

use LogicException;
use Symfony\Component\Console\Attribute\AsCommand;
use Symfony\Component\Console\Command\Command;
use Symfony\Component\Console\Input\InputArgument;
use Symfony\Component\Console\Input\InputInterface;
use Symfony\Component\Console\Output\OutputInterface;

/**
 * Répond à « qu'est-ce qui a changé en base depuis mes entités ». C'est
 * l'inverse de doctrine:schema:update, et le mode qui sert au quotidien sur du
 * legacy où le schéma bouge sans passer par les migrations.
 *
 * N'écrit rien. Code de retour non nul en cas de divergence, pour échouer un
 * pipeline.
 */
#[AsCommand(
    name: 'ormeau:synchroniser',
    description: 'Compare un calque physique aux entités Doctrine existantes',
)]
final class SynchroniserCommande extends Command
{
    /**
     * Le calque attendu est physique, seul cas où ce paquet en lit un : c'est
     * l'état réel de la base qu'on compare aux entités.
     */
    protected function configure(): void
    {
        $this->addArgument('calque', InputArgument::REQUIRED, 'Fichier .calque.json');
    }

    /**
     * @return int 0 si les entités correspondent à la base, non nul sinon
     */
    protected function execute(InputInterface $entree, OutputInterface $sortie): int
    {
        throw new LogicException('A implementer : jalon diff.');
    }
}
