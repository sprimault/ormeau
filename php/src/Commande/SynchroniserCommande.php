<?php

declare(strict_types=1);

namespace Ormeau\Doctrine\Commande;

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
    protected function configure(): void
    {
        $this->addArgument('calque', InputArgument::REQUIRED, 'Fichier .calque.json');
    }

    protected function execute(InputInterface $entree, OutputInterface $sortie): int
    {
        throw new \LogicException('À implémenter : phase 8 de la feuille de route.');
    }
}
