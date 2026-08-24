<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

$finder = PhpCsFixer\Finder::create()
    ->in([__DIR__.'/src', __DIR__.'/tests']);

// PER-CS plutôt que @Symfony : le paquet est un bundle, mais le code généré ne
// l'est pas, et aligner les deux sur le standard PHP-FIG évite d'avoir à
// expliquer pourquoi une entité produite ne ressemble pas au reste du dépôt.
//
// header_comment n'est volontairement pas activé : il réécrirait l'en-tête de
// copyright en bloc /* */, alors que la convention du dépôt est en //.
return (new PhpCsFixer\Config())
    ->setRiskyAllowed(true)
    ->setRules([
        '@PER-CS2.0' => true,
        '@PHP83Migration' => true,
        'declare_strict_types' => true,
        'strict_comparison' => true,
        'strict_param' => true,
        'no_unused_imports' => true,
        'ordered_imports' => ['sort_algorithm' => 'alpha'],
        'global_namespace_import' => [
            'import_classes' => true,
            'import_constants' => false,
            'import_functions' => false,
        ],
    ])
    ->setFinder($finder);
