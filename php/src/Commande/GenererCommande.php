<?php

declare(strict_types=1);

namespace Ormeau\Doctrine\Commande;

use Symfony\Component\Console\Attribute\AsCommand;
use Symfony\Component\Console\Command\Command;
use Symfony\Component\Console\Input\InputArgument;
use Symfony\Component\Console\Input\InputInterface;
use Symfony\Component\Console\Output\OutputInterface;

#[AsCommand(
    name: 'ormeau:generer',
    description: 'Génère les entités Doctrine à partir d\'un calque logique',
)]
final class GenererCommande extends Command
{
    protected function configure(): void
    {
        $this
            ->addArgument('calque', InputArgument::REQUIRED, 'Fichier .logique.json')
            ->addOption('repertoire', null, InputArgument::OPTIONAL, 'Répertoire de sortie', 'src/Entity');
    }

    protected function execute(InputInterface $entree, OutputInterface $sortie): int
    {
        throw new \LogicException('À implémenter : phase 4 de la feuille de route.');
    }
}
