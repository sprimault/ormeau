<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

use PhpParser\Error;
use PhpParser\Node;
use PhpParser\Node\Attribute;
use PhpParser\Node\Scalar\String_;
use PhpParser\Node\Stmt\Class_;
use PhpParser\NodeFinder;
use PhpParser\NodeTraverser;
use PhpParser\NodeVisitor\NameResolver;
use PhpParser\ParserFactory;

/**
 * Relit une classe de l'utilisateur et dit en quoi elle ne correspond plus au
 * calque, sans jamais l'écrire.
 *
 * Trois choses sont contrôlées, les seules que la génération y a mises : la
 * classe hérite de sa classe de base, elle porte #[ORM\Entity], et
 * #[ORM\Table] nomme la bonne table. Le reste — méthodes, docblocks, attributs
 * ajoutés, comme un repositoryClass — appartient à l'utilisateur et n'est pas
 * regardé.
 *
 * La lecture passe par nikic/php-parser et non par une expression régulière :
 * un attribut peut s'écrire sur plusieurs lignes, avec ou sans alias, et un
 * commentaire peut contenir #[ORM\Table].
 *
 * @internal
 */
final class ControleClasseUtilisateur
{
    /** Nom qualifié de l'attribut d'entité, une fois les alias résolus. */
    private const ENTITY = 'Doctrine\ORM\Mapping\Entity';

    /** Nom qualifié de l'attribut de table, une fois les alias résolus. */
    private const TABLE = 'Doctrine\ORM\Mapping\Table';

    /**
     * Compare une classe de l'utilisateur à ce que le calque attend.
     *
     * @param string                                   $chemin     fichier de la classe, tel qu'on l'affiche
     * @param string                                   $source     contenu du fichier
     * @param string                                   $classe     nom court de la classe attendue
     * @param string                                   $base       nom qualifié de sa classe de base
     * @param array{name: string, schema: string|null} $table      arguments attendus de #[ORM\Table]
     *
     * @return list<Divergence> vide quand la classe correspond
     */
    public function comparer(string $chemin, string $source, string $classe, string $base, array $table): array
    {
        try {
            $ast = (new ParserFactory())->createForNewestSupportedVersion()->parse($source) ?? [];
        } catch (Error $e) {
            return [new Divergence($chemin, max(1, $e->getStartLine()), '', 'fichier illisible : ' . $e->getRawMessage())];
        }
        $ast = (new NodeTraverser(new NameResolver()))->traverse($ast);

        $noeud = (new NodeFinder())->findFirst(
            $ast,
            static fn(Node $n): bool => $n instanceof Class_ && $n->name?->toString() === $classe,
        );
        if (!$noeud instanceof Class_) {
            return [new Divergence($chemin, 1, '', sprintf('classe %s introuvable, la génération ne peut pas s\'y appuyer', $classe))];
        }

        $divergences = [];
        $ligne = $noeud->getStartLine();

        if ($noeud->extends?->toString() !== $base) {
            $divergences[] = new Divergence($chemin, $ligne, '', sprintf(
                '%s n\'hérite pas de %s : la classe de base régénérée ne s\'applique pas',
                $classe,
                $base,
            ));
        }

        if ($this->attribut($noeud, self::ENTITY) === null) {
            $divergences[] = new Divergence($chemin, $ligne, '', 'attribut #[ORM\Entity] absent');
        }

        $attribut = $this->attribut($noeud, self::TABLE);
        if ($attribut === null) {
            $divergences[] = new Divergence($chemin, $ligne, '', sprintf(
                'attribut #[ORM\Table] absent, la table s\'appelle %s',
                $table['name'],
            ));

            return $divergences;
        }

        $actuel = '#[' . substr($source, $attribut->getStartFilePos(), $attribut->getEndFilePos() - $attribut->getStartFilePos() + 1) . ']';
        if ($this->argument($attribut, 'name', 0) !== $table['name']) {
            $divergences[] = new Divergence($chemin, $attribut->getStartLine(), $actuel, sprintf(
                'la table s\'appelle maintenant %s',
                $table['name'],
            ));
        }
        if ($table['schema'] !== null && $this->argument($attribut, 'schema', 1) !== $table['schema']) {
            $divergences[] = new Divergence($chemin, $attribut->getStartLine(), $actuel, sprintf(
                'la table est dans le schéma %s',
                $table['schema'],
            ));
        }

        return $divergences;
    }

    /**
     * Rend l'attribut d'une classe qui porte ce nom qualifié.
     */
    private function attribut(Class_ $classe, string $nom): ?Attribute
    {
        foreach ($classe->attrGroups as $groupe) {
            foreach ($groupe->attrs as $attribut) {
                if ($attribut->name->toString() === $nom) {
                    return $attribut;
                }
            }
        }

        return null;
    }

    /**
     * Rend la valeur littérale d'un argument, nommé ou à sa position.
     *
     * Une valeur qui n'est pas une chaîne littérale — une constante, une
     * concaténation — rend null : elle ne se compare pas sans exécuter le code,
     * et la signaler vaut mieux que la supposer juste.
     */
    private function argument(Attribute $attribut, string $nom, int $position): ?string
    {
        foreach ($attribut->args as $rang => $argument) {
            $vise = $argument->name === null ? $rang === $position : $argument->name->toString() === $nom;
            if ($vise) {
                return $argument->value instanceof String_ ? $argument->value->value : null;
            }
        }

        return null;
    }
}
