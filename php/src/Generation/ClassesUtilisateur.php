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
     */
    private function __construct(
        private readonly string $espaceDeNoms,
        private readonly array $trouvees,
    ) {}

    /**
     * Parcourt le répertoire des entités et retient, pour chaque nom donné, les
     * classes de ce nom qui héritent de sa classe de base.
     *
     * Un fichier que PHP ne sait pas lire est ignoré ici : s'il est à la place
     * attendue, le contrôle de la classe de l'utilisateur le signale.
     *
     * @param string       $racine       répertoire des entités, existant
     * @param string       $espaceDeNoms espace de noms des entités du calque
     * @param list<string> $noms         noms des entités du calque
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
        foreach ($chemins as $chemin) {
            $source = file_get_contents($chemin);
            try {
                $ast = $source === false ? null : $parseur->parse($source);
            } catch (Error) {
                continue;
            }
            $ast = (new NodeTraverser(new NameResolver()))->traverse($ast ?? []);

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

        return new self($espaceDeNoms, $trouvees);
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
