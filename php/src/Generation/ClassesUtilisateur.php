<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

use FilesystemIterator;
use PhpParser\Error;
use PhpParser\Node;
use PhpParser\Node\Stmt\Class_;
use PhpParser\NodeFinder;
use PhpParser\NodeTraverser;
use PhpParser\NodeVisitor\NameResolver;
use PhpParser\ParserFactory;
use RecursiveDirectoryIterator;
use RecursiveIteratorIterator;
use RuntimeException;
use SplFileInfo;

/**
 * Où vivent les classes de l'utilisateur, et sous quel nom qualifié.
 *
 * Une classe générée naît à la racine du répertoire des entités, et le
 * développeur la range ensuite où il veut — src/Entity/Ventes/Client.php,
 * espace de noms App\Entity\Ventes —, comme Symfony et Doctrine le permettent
 * sans configuration. La régénération doit la retrouver : sinon elle en
 * recrée une à la racine, deux entités se disputent la même table, et les
 * autres classes de base visent la mauvaise.
 *
 * C'est un constat sur le projet, pas une décision : la classe Client qui
 * hérite de Base\ClientBase est celle de l'entité Client, où qu'elle soit. Le
 * nom court doit rester celui de l'entité — c'est ce qui garde les imports sans
 * collision —, l'espace de noms est celui que la classe déclare. Base/, Enum/
 * et Trait/ appartiennent à l'outil et ne sont pas parcourus.
 *
 * @internal
 */
final class ClassesUtilisateur
{
    /** Répertoires de l'outil, à la racine des entités. */
    private const REPERTOIRES_DE_L_OUTIL = ['Base', 'Enum', 'Trait'];

    /**
     * @param string                                                   $espaceDeNoms espace de noms des entités du calque
     * @param array<string, list<array{fichier: string, classe: string}>> $trouvees     pour chaque entité, les classes qui
     *                                                                                  la portent, dans l'ordre des chemins
     * @param array<string, array{source: string, ast: array<Node>}>     $lectures     chaque fichier lu, source et AST noms
     *                                                                                  résolus, par chemin
     */
    private function __construct(
        private readonly string $espaceDeNoms,
        private readonly array $trouvees,
        private readonly array $lectures,
    ) {}

    /**
     * Parcourt le répertoire des entités et retient, pour chaque nom donné, les
     * classes de ce nom qui héritent de sa classe de base.
     *
     * Un fichier PHP que php-parser refuse arrête tout, avant la première
     * écriture de la génération. Il peut porter la classe d'une entité que la
     * génération étend, et rien ne permet de le savoir sans le lire : l'ignorer
     * faisait créer une seconde classe à la racine, et écarter l'entité aurait
     * réécrit les classes de base qui la visent sans leur côté inverse. C'est un
     * état passager, qui se corrige avant de relancer, et non une propriété du
     * calque, qu'on écarterait avec sa raison.
     *
     * @param string       $racine       répertoire des entités, existant
     * @param string       $espaceDeNoms espace de noms des entités du calque
     * @param list<string> $noms         noms des entités du calque
     *
     * @throws RuntimeException un fichier au moins est illisible : le message les nomme tous, ligne comprise
     */
    public static function parcourir(string $racine, string $espaceDeNoms, array $noms): self
    {
        $attendues = [];
        foreach ($noms as $nom) {
            $attendues[$nom] = $espaceDeNoms . '\\Base\\' . $nom . 'Base';
        }

        $chemins = [];
        $elements = new RecursiveIteratorIterator(new RecursiveDirectoryIterator($racine, FilesystemIterator::SKIP_DOTS));
        foreach ($elements as $element) {
            /** @var SplFileInfo $element */
            $relatif = str_replace('\\', '/', substr($element->getPathname(), strlen($racine) + 1));
            $premier = explode('/', $relatif)[0];
            if ($element->isFile() && $element->getExtension() === 'php' && !in_array($premier, self::REPERTOIRES_DE_L_OUTIL, true)) {
                $chemins[] = $racine . '/' . $relatif;
            }
        }
        sort($chemins);

        $parseur = (new ParserFactory())->createForNewestSupportedVersion();
        $trouvees = [];
        $lectures = [];
        $illisibles = [];
        foreach ($chemins as $chemin) {
            $source = file_get_contents($chemin);
            if ($source === false) {
                throw new RuntimeException(sprintf('Génération refusée, rien n\'est écrit : %s ne se lit pas', $chemin));
            }
            try {
                $ast = $parseur->parse($source) ?? [];
            } catch (Error $e) {
                $illisibles[] = sprintf('%s ligne %d : %s', substr($chemin, strlen($racine) + 1), max(1, $e->getStartLine()), $e->getRawMessage());
                continue;
            }
            $ast = (new NodeTraverser(new NameResolver()))->traverse($ast);
            $lectures[$chemin] = ['source' => $source, 'ast' => $ast];

            $classes = (new NodeFinder())->find($ast, static fn(Node $n): bool => $n instanceof Class_);
            foreach ($classes as $classe) {
                /** @var Class_ $classe */
                $nom = $classe->name?->toString();
                if ($nom === null || !isset($attendues[$nom]) || $classe->extends?->toString() !== $attendues[$nom]) {
                    continue;
                }
                $qualifiee = $classe->namespacedName?->toString() ?? $nom;
                $trouvees[$nom][] = ['fichier' => $chemin, 'classe' => $qualifiee];
            }
        }

        if ($illisibles !== []) {
            throw new RuntimeException(sprintf(
                'Génération refusée, rien n\'est écrit : %s. Sous le répertoire des entités, un fichier illisible peut porter la classe d\'une entité que la génération étend, et elle ne peut pas le savoir sans le lire : le corriger, puis relancer.',
                implode(' ; ', $illisibles),
            ));
        }

        return new self($espaceDeNoms, $trouvees, $lectures);
    }

    /**
     * Rend la source et l'AST, noms résolus, d'un fichier lu pendant le
     * parcours, ou null pour un fichier que le parcours n'a pas vu : le
     * contrôle de la classe de l'utilisateur compare ce qui a été lu, sans
     * relire le fichier.
     *
     * @return array{source: string, ast: array<Node>}|null
     */
    public function lecture(string $chemin): ?array
    {
        return $this->lectures[$chemin] ?? null;
    }

    /**
     * Rend le nom qualifié de la classe de l'utilisateur d'une entité : celui
     * de la classe trouvée, ou celui qu'elle aura à sa création.
     */
    public function qualifiee(string $nom): string
    {
        $trouvees = $this->trouvees[$nom] ?? [];

        return count($trouvees) === 1 ? $trouvees[0]['classe'] : $this->espaceDeNoms . '\\' . $nom;
    }

    /**
     * Rend le fichier de la classe de l'utilisateur d'une entité : celui de la
     * classe trouvée, ou la place qu'elle prend à sa création.
     */
    public function fichier(string $nom, string $racine): string
    {
        $trouvees = $this->trouvees[$nom] ?? [];

        return count($trouvees) === 1 ? $trouvees[0]['fichier'] : $racine . '/' . $nom . '.php';
    }

    /**
     * Rend les fichiers qui portent chacun une classe de l'utilisateur pour la
     * même entité, quand il y en a plusieurs ; une liste vide sinon.
     *
     * @return list<string>
     */
    public function doublons(string $nom): array
    {
        $trouvees = $this->trouvees[$nom] ?? [];

        return count($trouvees) > 1 ? array_column($trouvees, 'fichier') : [];
    }
}
