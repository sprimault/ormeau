<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\Generation;

use Ormeau\Doctrine\Generation\GenerateurEntite;
use Ormeau\Doctrine\Generation\ModeRegeneration;
use PHPUnit\Framework\Attributes\CoversClass;
use PHPUnit\Framework\TestCase;

/**
 * La génération n'est pas écrite. Ce qui se teste dès maintenant, c'est le mode
 * par défaut : le changer serait une rupture pour tout projet déjà généré, et
 * rien d'autre ne le signalerait.
 */
#[CoversClass(GenerateurEntite::class)]
final class GenerateurEntiteTest extends TestCase
{
    /**
     * La classe de base est le mode par défaut : c'est le seul robuste sans
     * relire l'entité existante.
     */
    public function testLeModeParDefautEstLaClasseDeBase(): void
    {
        self::assertSame(ModeRegeneration::ClasseDeBase, (new GenerateurEntite())->mode());
    }

    /**
     * Le mode demandé est retenu tel quel.
     */
    public function testLeModeDemandeEstRetenu(): void
    {
        $generateur = new GenerateurEntite(ModeRegeneration::ReecritureAst);

        self::assertSame(ModeRegeneration::ReecritureAst, $generateur->mode());
    }

    /**
     * Les valeurs de l'énumération sont un contrat : elles viennent d'un drapeau
     * de commande et d'une configuration Symfony, les renommer casse les deux.
     */
    public function testLesValeursDuModeSontStables(): void
    {
        self::assertSame('classe_de_base', ModeRegeneration::ClasseDeBase->value);
        self::assertSame('reecriture_ast', ModeRegeneration::ReecritureAst->value);
    }
}
