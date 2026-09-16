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
     * comme avant une montée de version. --cible-dbal ne sert qu'à s'écarter de
     * ce que la commande déduit : la version installée, ou la plus récente
     * qu'accepte la majeure d'ORM forcée.
     */
    protected function configure(): void
    {
        $this
            ->addArgument('calque', InputArgument::REQUIRED, 'Fichier .logique.json')
            ->addOption('repertoire', null, InputOption::VALUE_REQUIRED, 'Répertoire de sortie', 'src/Entity')
            ->addOption('cible-orm', null, InputOption::VALUE_REQUIRED, 'Majeure de Doctrine ORM visée (2 ou 3), à défaut de celle installée')
            ->addOption('cible-dbal', null, InputOption::VALUE_REQUIRED, 'Version de DBAL visée (majeure.mineure), à défaut de celle installée ou déduite')
            ->addOption('remplacer', null, InputOption::VALUE_REQUIRED | InputOption::VALUE_IS_ARRAY, 'Base dont les fichiers peuvent être réécrits dans ce répertoire');
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
     * ils n'empêchent pas le reste d'être écrit. Un refus d'écrasement rend 1 :
     * un fichier d'une autre base occupe la place, et la génération demandée
     * n'a pas eu lieu, ce qu'un pipeline doit voir échouer.
     *
     * La base se lit dans le nom du fichier, gescom pour gescom.logique.json,
     * comme ormeau inferer la tire du calque physique ; un nom hors convention
     * perd seulement son extension. --remplacer nomme la base qu'on accepte
     * d'écraser : un fichier renommé, ou un répertoire repris par une autre
     * base, sans supprimer Base/ à la main.
     *
     * Tout texte venu du calque passe par OutputFormatter::escape : une balise
     * dans un nom de table, `<error></info>` par exemple, ferait lever le
     * formateur au milieu du compte rendu, fichiers déjà écrits.
     *
     * @return int 0 si la génération est allée au bout, 1 si un fichier d'une autre base l'a bloquée
     *
     * @throws InvalidArgumentException chemin absent, --cible-orm hors de 2 et 3, --cible-dbal illisible ou
     *                                  incompatible, ORM ou DBAL introuvables, nom de base vide ou porteur
     *                                  d'un caractère de contrôle
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

        $cible = $this->cible($entree->getOption('cible-orm'), $entree->getOption('cible-dbal'));
        $sortie->writeln($cible->annonce());
        $base = self::base($chemin);
        $sortie->writeln('Base : ' . OutputFormatter::escape($base));

        /** @var list<string> $remplacables */
        $remplacables = array_values(array_filter((array) $entree->getOption('remplacer'), 'is_string'));
        $calque = $this->lecteur->lire($chemin);
        $rapport = $this->generateur->generer($calque, $repertoire, $cible, $base, $remplacables);

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
        foreach ($rapport->sequences as $sequence) {
            $sortie->writeln('<comment>séquence ' . OutputFormatter::escape($sequence->message()) . '</comment>');
        }
        foreach ($rapport->generees as $generee) {
            $sortie->writeln('<comment>héritage ' . OutputFormatter::escape($generee->message()) . '</comment>');
        }
        foreach ($rapport->index as $index) {
            $sortie->writeln('<comment>index    ' . OutputFormatter::escape($index->message()) . '</comment>');
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

        // En dernier, pour qu'on les lise sous tout le reste.
        foreach ($rapport->refus as $refus) {
            $sortie->writeln(sprintf(
                '<error>refusé   %s (--remplacer=%s)</error>',
                OutputFormatter::escape($refus->message()),
                OutputFormatter::escape($refus->baseExistante),
            ));
        }

        return $rapport->refus === [] ? Command::SUCCESS : Command::FAILURE;
    }

    /**
     * Rend le nom de la base que désigne un calque logique.
     *
     * gescom.logique.json donne gescom, comme ormeau inferer le fait du calque
     * physique ; un nom qui ne suit pas la convention perd seulement son
     * extension, plutôt que d'être refusé.
     */
    private static function base(string $chemin): string
    {
        $nom = basename($chemin);
        if (str_ends_with($nom, '.logique.json')) {
            return substr($nom, 0, -strlen('.logique.json'));
        }
        $point = strrpos($nom, '.');

        return $point === false || $point === 0 ? $nom : substr($nom, 0, $point);
    }

    /**
     * Rend la cible forcée par --cible-orm, ou celle de l'application ; la
     * version de DBAL vient de --cible-dbal quand elle est donnée.
     *
     * @param mixed $orm  valeur de --cible-orm
     * @param mixed $dbal valeur de --cible-dbal
     *
     * @throws InvalidArgumentException majeure d'ORM inconnue, version de DBAL illisible ou incompatible
     */
    private function cible(mixed $orm, mixed $dbal): Cible
    {
        $dbal = is_string($dbal) ? $dbal : null;
        if ($orm === null) {
            return Cible::detecter($dbal);
        }
        if (!is_string($orm) || preg_match('/^\d+$/', $orm) !== 1) {
            throw new InvalidArgumentException(sprintf(
                '--cible-orm attend une majeure de Doctrine ORM (%s).',
                implode(', ', Cible::MAJEURES_ORM),
            ));
        }

        return Cible::forcer((int) $orm, $dbal);
    }
}
