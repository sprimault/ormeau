<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Commande;

use InvalidArgumentException;
use Ormeau\Doctrine\Calque\LecteurCalque;
use Ormeau\Doctrine\Generation\GenerateurEntite;
use Symfony\Component\Console\Attribute\AsCommand;
use Symfony\Component\Console\Command\Command;
use Symfony\Component\Console\Input\InputArgument;
use Symfony\Component\Console\Input\InputInterface;
use Symfony\Component\Console\Input\InputOption;
use Symfony\Component\Console\Output\OutputInterface;

/**
 * Produit les entités d'un calque logique. Ne lit jamais le calque physique et
 * ne se connecte jamais à la base : le générateur traduit, il ne décide pas.
 */
#[AsCommand(
    name: 'ormeau:generer',
    description: 'Génère les entités Doctrine à partir d\'un calque logique',
)]
final class GenererCommande extends Command
{
    public function __construct(
        private readonly LecteurCalque $lecteur,
        private readonly GenerateurEntite $generateur,
    ) {
        parent::__construct();
    }

    /**
     * Le répertoire de sortie vaut src/Entity, disposition Symfony par défaut.
     */
    protected function configure(): void
    {
        $this
            ->addArgument('calque', InputArgument::REQUIRED, 'Fichier .logique.json')
            ->addOption('repertoire', null, InputOption::VALUE_REQUIRED, 'Répertoire de sortie', 'src/Entity');
    }

    /**
     * @return int 0 si les entités ont été écrites
     */
    protected function execute(InputInterface $entree, OutputInterface $sortie): int
    {
        $chemin = $entree->getArgument('calque');
        $repertoire = $entree->getOption('repertoire');
        if (!is_string($chemin) || !is_string($repertoire)) {
            throw new InvalidArgumentException('Le calque et le repertoire sont des chemins de fichier.');
        }

        $calque = $this->lecteur->lire($chemin);

        foreach ($this->generateur->generer($calque, $repertoire) as $fichier) {
            $sortie->writeln($fichier);
        }

        return Command::SUCCESS;
    }
}
