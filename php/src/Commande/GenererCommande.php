<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Commande;

use InvalidArgumentException;
use JsonException;
use Ormeau\Doctrine\Calque\CalqueInvalide;
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
    /**
     * Reçoit le lecteur et le générateur par le conteneur de Symfony.
     *
     * @param LecteurCalque    $lecteur    lit et contrôle le calque logique avant toute écriture
     * @param GenerateurEntite $generateur écrit les entités, dans le mode de régénération retenu
     */
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
     * Lit le calque, puis écrit les entités et affiche un chemin par fichier
     * écrit.
     *
     * Le calque est lu entièrement avant la première écriture : un calque
     * invalide arrête la commande sans avoir touché au répertoire. Les
     * exceptions du lecteur remontent telles quelles, et la console de Symfony
     * affiche leur message, qui nomme le champ fautif.
     *
     * @return int 0 si les entités ont été écrites
     *
     * @throws CalqueInvalide calque illisible, d'une version inconnue, ou invalide
     * @throws JsonException  fichier qui n'est pas du JSON
     */
    protected function execute(InputInterface $entree, OutputInterface $sortie): int
    {
        $chemin = $entree->getArgument('calque');
        $repertoire = $entree->getOption('repertoire');
        if (!is_string($chemin) || !is_string($repertoire)) {
            throw new InvalidArgumentException('Le calque et le répertoire sont des chemins de fichier.');
        }

        $calque = $this->lecteur->lire($chemin);

        foreach ($this->generateur->generer($calque, $repertoire) as $fichier) {
            $sortie->writeln($fichier);
        }

        return Command::SUCCESS;
    }
}
