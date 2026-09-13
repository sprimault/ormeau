<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Commande;

use InvalidArgumentException;
use JsonException;
use Ormeau\Doctrine\Calque\CalqueInvalide;
use Ormeau\Doctrine\Calque\LecteurCalque;
use Ormeau\Doctrine\Generation\Cible;
use Ormeau\Doctrine\Generation\GenerateurEntite;
use RuntimeException;
use Symfony\Component\Console\Attribute\AsCommand;
use Symfony\Component\Console\Command\Command;
use Symfony\Component\Console\Formatter\OutputFormatter;
use Symfony\Component\Console\Helper\Helper;
use Symfony\Component\Console\Input\InputArgument;
use Symfony\Component\Console\Input\InputInterface;
use Symfony\Component\Console\Input\InputOption;
use Symfony\Component\Console\Output\OutputInterface;

/**
 * Produit les entités d'un calque logique. Ne lit jamais le calque physique et
 * ne se connecte jamais à la base : le générateur traduit, il ne décide pas.
 *
 * La cible — la version de Doctrine ORM de l'application — est détectée et
 * annoncée en première ligne. Elle change le type PHP de certaines colonnes :
 * sans l'annonce, un passage d'ORM 2 à ORM 3 produirait un diff large sur des
 * entités que personne n'a touchées, qui ressemblerait à un défaut.
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
     *
     * --cible-orm fait partie du contrat, pas seulement des tests : c'est ce
     * qui permet de générer pour une version qu'on n'a pas installée, en CI
     * comme avant une montée de version.
     */
    protected function configure(): void
    {
        $this
            ->addArgument('calque', InputArgument::REQUIRED, 'Fichier .logique.json')
            ->addOption('repertoire', null, InputOption::VALUE_REQUIRED, 'Répertoire de sortie', 'src/Entity')
            ->addOption('cible-orm', null, InputOption::VALUE_REQUIRED, 'Majeure de Doctrine ORM visée (2 ou 3), à défaut de celle installée');
    }

    /**
     * Annonce la cible, lit le calque, écrit les entités, puis rend compte.
     *
     * Le calque est lu entièrement avant la première écriture : un calque
     * invalide arrête la commande sans avoir touché au répertoire. Les
     * exceptions du lecteur remontent telles quelles, et la console de Symfony
     * affiche leur message, qui nomme le champ fautif.
     *
     * Les avertissements du calque, les entités écartées et les divergences
     * s'affichent sans changer le code de retour : ils disent quoi reprendre,
     * ils n'empêchent pas le reste d'être écrit.
     *
     * Tout texte venu du calque passe par OutputFormatter::escape : une balise
     * dans un nom de table, `<error></info>` par exemple, ferait lever le
     * formateur au milieu du compte rendu, fichiers déjà écrits.
     *
     * @return int 0 si la génération est allée au bout
     *
     * @throws InvalidArgumentException chemin absent, --cible-orm hors de 2 et 3, ORM introuvable
     * @throws CalqueInvalide           calque illisible, d'une version inconnue, ou invalide
     * @throws JsonException            fichier qui n'est pas du JSON
     * @throws RuntimeException         fichier ou répertoire qui ne s'écrit pas
     */
    protected function execute(InputInterface $entree, OutputInterface $sortie): int
    {
        $chemin = $entree->getArgument('calque');
        $repertoire = $entree->getOption('repertoire');
        if (!is_string($chemin) || !is_string($repertoire)) {
            throw new InvalidArgumentException('Le calque et le répertoire sont des chemins de fichier.');
        }

        $cible = $this->cible($entree->getOption('cible-orm'));
        $sortie->writeln($cible->annonce());

        $calque = $this->lecteur->lire($chemin);
        $rapport = $this->generateur->generer($calque, $repertoire, $cible);

        // Aligné sur la largeur affichée et non en octets : « créé » en compte
        // six, et un %-9s décalerait les chemins d'une ligne à l'autre.
        foreach ($rapport->fichiers as $fichier) {
            $etat = $fichier->etat->value;
            $sortie->writeln($etat . str_repeat(' ', 9 - Helper::width($etat)) . $fichier->chemin);
        }
        foreach ($rapport->ecartees as $ecartee) {
            $sortie->writeln(sprintf(
                '<comment>écartée  %s : %s</comment>',
                OutputFormatter::escape($ecartee->nom),
                OutputFormatter::escape($ecartee->raison),
            ));
        }
        foreach ($rapport->omises as $omise) {
            $sortie->writeln(sprintf(
                '<comment>omise    %s::%s : %s</comment>',
                OutputFormatter::escape($omise->entite),
                OutputFormatter::escape($omise->association),
                OutputFormatter::escape($omise->raison),
            ));
        }
        foreach ($rapport->divergences as $divergence) {
            $sortie->writeln('<comment>à reprendre ' . OutputFormatter::escape($divergence->message()) . '</comment>');
        }

        if ($calque->avertissements !== []) {
            $sortie->writeln(sprintf('%d avertissement(s) dans le calque :', count($calque->avertissements)));
            foreach ($calque->avertissements as $avertissement) {
                $sortie->writeln(sprintf(
                    '  %s %s — %s',
                    OutputFormatter::escape($avertissement->code),
                    OutputFormatter::escape($avertissement->cible),
                    OutputFormatter::escape($avertissement->message),
                ));
            }
        }

        return Command::SUCCESS;
    }

    /**
     * Rend la cible forcée par --cible-orm, ou celle de l'application.
     *
     * @throws InvalidArgumentException valeur qui n'est pas une majeure connue
     */
    private function cible(mixed $option): Cible
    {
        if ($option === null) {
            return Cible::detecter();
        }
        if (!is_string($option) || preg_match('/^\d+$/', $option) !== 1) {
            throw new InvalidArgumentException(sprintf(
                '--cible-orm attend une majeure de Doctrine ORM (%s).',
                implode(', ', Cible::MAJEURES_ORM),
            ));
        }

        return Cible::forcer((int) $option);
    }
}
