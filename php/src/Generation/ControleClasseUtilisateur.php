<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

use PhpParser\Node;
use PhpParser\Node\Attribute;
use PhpParser\Node\Expr;
use PhpParser\Node\Expr\Array_;
use PhpParser\Node\Expr\ClassConstFetch;
use PhpParser\Node\Identifier;
use PhpParser\Node\Name;
use PhpParser\Node\Scalar\Int_;
use PhpParser\Node\Scalar\String_;
use PhpParser\Node\Stmt\Class_;
use PhpParser\NodeFinder;

/**
 * Relit une classe de l'utilisateur et dit en quoi elle ne correspond plus au
 * calque, sans jamais l'écrire.
 *
 * Ce qui est contrôlé est ce que la génération y a mis, et rien d'autre : la
 * classe hérite de sa classe de base, elle porte #[ORM\Entity], #[ORM\Table]
 * nomme la bonne table avec son commentaire, et une racine d'héritage porte
 * sa stratégie, sa colonne discriminante et sa carte. Le reste — méthodes,
 * docblocks, attributs ajoutés, comme un repositoryClass — appartient à
 * l'utilisateur et n'est pas regardé.
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

    /** Nom qualifié de l'attribut de stratégie d'héritage. */
    private const INHERITANCE_TYPE = 'Doctrine\ORM\Mapping\InheritanceType';

    /** Nom qualifié de l'attribut de colonne discriminante. */
    private const DISCRIMINATOR_COLUMN = 'Doctrine\ORM\Mapping\DiscriminatorColumn';

    /** Nom qualifié de l'attribut de carte des classes d'une hiérarchie. */
    private const DISCRIMINATOR_MAP = 'Doctrine\ORM\Mapping\DiscriminatorMap';

    /**
     * Compare une classe de l'utilisateur à ce que le calque attend.
     *
     * Le fichier n'est pas relu : la source et l'AST viennent du parcours des
     * classes de l'utilisateur, qui a déjà refusé la génération sur un fichier
     * illisible. Une seule lecture, et aucun écart entre ce qui a été retrouvé
     * et ce qui est contrôlé.
     *
     * @param string                                   $chemin     fichier de la classe, tel qu'on l'affiche
     * @param string                                   $source     contenu du fichier, pour citer un attribut tel
     *                                                             qu'il est écrit
     * @param array<Node>                              $ast        AST du fichier, noms résolus
     * @param string                                   $classe     nom court de la classe attendue
     * @param string                                   $base       nom qualifié de sa classe de base
     * @param array{name: string, schema: string|null, options: array{comment: string}|null} $table arguments attendus de
     *                                                                                             #[ORM\Table]
     * @param RacineHeritage|null                      $racine     ce que la classe déclare comme racine d'un
     *                                                             héritage ; null quand elle n'en est pas une
     * @param array{name: string, schema: string|null}|null $tableEn050 name et schema tels que la 0.5.0 les
     *                                                                   écrivait (RenduEntite::argumentsTableEn050)
     *
     * @return list<Divergence> vide quand la classe correspond
     */
    public function comparer(string $chemin, string $source, array $ast, string $classe, string $base, array $table, ?RacineHeritage $racine = null, ?array $tableEn050 = null): array
    {
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
        } else {
            $actuel = self::texte($source, $attribut);
            $nomActuel = $this->argument($attribut, 'name', 0);
            $schemaActuel = $this->argument($attribut, 'schema', 1);
            $en050 = self::cassee050($nomActuel, $schemaActuel, $table, $tableEn050);
            if ($nomActuel !== $table['name']) {
                $divergences[] = new Divergence($chemin, $attribut->getStartLine(), $actuel, sprintf(
                    'la table s\'appelle maintenant %s%s',
                    $table['name'],
                    $en050,
                ));
            }
            if ($table['schema'] !== null && $schemaActuel !== $table['schema']) {
                $divergences[] = new Divergence($chemin, $attribut->getStartLine(), $actuel, sprintf(
                    'la table est dans le schéma %s%s',
                    $table['schema'],
                    $en050,
                ));
            }
            // Le schéma par défaut ne s'écrit pas : DBAL 3 propose sinon de
            // supprimer la table puis de la recréer. Une classe générée quand
            // la règle était autre le porte encore.
            if ($table['schema'] === null && $this->valeur($attribut, 'schema', 1) !== null) {
                $divergences[] = new Divergence(
                    $chemin,
                    $attribut->getStartLine(),
                    $actuel,
                    'la table est dans le schéma par défaut, schema à retirer : DBAL 3 proposerait de la supprimer puis de la recréer',
                );
            }
            $commentaire = $table['options']['comment'] ?? null;
            if ($this->commentaire($attribut) !== $commentaire) {
                $divergences[] = new Divergence($chemin, $attribut->getStartLine(), $actuel, $commentaire === null
                    ? 'la table n\'a plus de commentaire'
                    : sprintf('le commentaire de la table est maintenant %s', Emetteur::litteral($commentaire)));
            }
        }

        return [...$divergences, ...$this->comparerHeritage($chemin, $source, $noeud, $racine)];
    }

    /**
     * Compare les attributs d'héritage d'une classe à ce que le calque attend.
     *
     * Une racine doit porter la stratégie, la colonne discriminante et la
     * carte des classes : une classe fille déclarée après la création de la
     * racine change la carte, et c'est la divergence la plus probable. Une
     * classe qui n'est plus racine ne doit plus en porter aucun, Doctrine
     * cherchant sinon des classes filles qui ne descendent plus d'elle. Le
     * message cite l'attribut attendu en entier, à recopier tel quel.
     *
     * @return list<Divergence>
     */
    private function comparerHeritage(string $chemin, string $source, Class_ $classe, ?RacineHeritage $racine): array
    {
        $divergences = [];
        $attendus = $racine?->attributs() ?? [];

        foreach ([self::INHERITANCE_TYPE, self::DISCRIMINATOR_COLUMN, self::DISCRIMINATOR_MAP] as $nom) {
            $attribut = $this->attribut($classe, $nom);
            if ($racine === null) {
                if ($attribut !== null) {
                    $divergences[] = new Divergence($chemin, $attribut->getStartLine(), self::texte($source, $attribut), sprintf(
                        '%s n\'est la racine d\'aucun héritage déclaré, attribut à retirer',
                        $classe->name?->toString(),
                    ));
                }
                continue;
            }

            if ($attribut === null) {
                $divergences[] = new Divergence($chemin, $classe->getStartLine(), '', sprintf(
                    'attribut %s absent, %s est la racine d\'un héritage',
                    $attendus[$nom],
                    $classe->name?->toString(),
                ));
                continue;
            }

            $conforme = match ($nom) {
                self::INHERITANCE_TYPE => $this->argument($attribut, 'value', 0) === RacineHeritage::STRATEGIE,
                self::DISCRIMINATOR_COLUMN => $this->argument($attribut, 'name', 0) === $racine->colonne['name']
                    && $this->argument($attribut, 'type', 1) === $racine->colonne['type']
                    && $this->entier($attribut, 'length', 2) === $racine->colonne['length'],
                default => $this->carte($attribut) === $racine->carte,
            };
            if (!$conforme) {
                $divergences[] = new Divergence($chemin, $attribut->getStartLine(), self::texte($source, $attribut), 'attendu ' . $attendus[$nom]);
            }
        }

        return $divergences;
    }

    /**
     * Rend un attribut tel qu'il est écrit dans le fichier, crochets compris.
     */
    private static function texte(string $source, Attribute $attribut): string
    {
        return '#[' . substr($source, $attribut->getStartFilePos(), $attribut->getEndFilePos() - $attribut->getStartFilePos() + 1) . ']';
    }

    /**
     * Rend le commentaire que #[ORM\Table] donne à la table, dans ses options ;
     * null quand il n'y en a pas, ou qu'il n'est pas écrit en clair.
     */
    private function commentaire(Attribute $table): ?string
    {
        $options = $this->valeur($table, 'options', 4);
        if (!$options instanceof Array_) {
            return null;
        }
        foreach ($options->items as $element) {
            if ($element->key instanceof String_ && $element->key->value === 'comment') {
                return $element->value instanceof String_ ? $element->value->value : null;
            }
        }

        return null;
    }

    /**
     * Rend la valeur entière littérale d'un argument, nommé ou à sa position ;
     * null quand il manque ou n'est pas un entier écrit en clair.
     */
    private function entier(Attribute $attribut, string $nom, int $position): ?int
    {
        $valeur = $this->valeur($attribut, $nom, $position);

        return $valeur instanceof Int_ ? $valeur->value : null;
    }

    /**
     * Rend la carte de #[ORM\DiscriminatorMap] : valeur vers nom qualifié de
     * classe, dans l'ordre du fichier, ou null quand elle n'est pas écrite
     * comme un tableau de X::class.
     *
     * @return array<string, string>|null
     */
    private function carte(Attribute $attribut): ?array
    {
        $tableau = $this->valeur($attribut, 'value', 0);
        if (!$tableau instanceof Array_) {
            return null;
        }

        $carte = [];
        foreach ($tableau->items as $element) {
            $cle = $element->key;
            $classe = $element->value;
            if (!($cle instanceof String_ || $cle instanceof Int_)
                || !$classe instanceof ClassConstFetch || !$classe->class instanceof Name
                || !$classe->name instanceof Identifier || $classe->name->toLowerString() !== 'class') {
                return null;
            }
            $carte[(string) $cle->value] = $classe->class->toString();
        }

        return $carte;
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
     * Rend le complément du message quand la classe porte la forme cassée de
     * la 0.5.0, ou une chaîne vide.
     *
     * Rétro-compatibilité datée, isolée ici avec RenduEntite::
     * argumentsTableEn050 : sans elle, l'utilisateur lirait qu'une
     * régénération casse une table qui ne fonctionnait pas. La condition est
     * stricte — l'attribut est exactement ce que la 0.5.0 écrivait, et cette
     * forme diffère de l'attendue — : hors de là, la phrase serait fausse. À
     * retirer avec argumentsTableEn050.
     *
     * @param array{name: string, schema: string|null, options: array{comment: string}|null} $table      attendus
     * @param array{name: string, schema: string|null}|null                                  $tableEn050 forme 0.5.0
     */
    private static function cassee050(?string $nom, ?string $schema, array $table, ?array $tableEn050): string
    {
        if ($tableEn050 === null
            || ($tableEn050['name'] === $table['name'] && $tableEn050['schema'] === $table['schema'])
            || $nom !== $tableEn050['name']
            || $schema !== $tableEn050['schema']) {
            return '';
        }

        return ' ; forme écrite par la 0.5.0, que PostgreSQL refuse : cette table était inutilisable';
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
        $valeur = $this->valeur($attribut, $nom, $position);

        return $valeur instanceof String_ ? $valeur->value : null;
    }

    /**
     * Rend l'expression d'un argument, nommé ou à sa position, ou null quand
     * il manque.
     */
    private function valeur(Attribute $attribut, string $nom, int $position): ?Expr
    {
        foreach ($attribut->args as $rang => $argument) {
            $vise = $argument->name === null ? $rang === $position : $argument->name->toString() === $nom;
            if ($vise) {
                return $argument->value;
            }
        }

        return null;
    }
}
