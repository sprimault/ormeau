<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

// Ce que le générateur réécrit à chaque passage — Base/, Enum/, Trait/ — ne
// doit rien donner à corriger aux règles @Symfony, les plus répandues dans les
// projets visés : un projet qui les applique à src/ corrigerait le fichier, et
// chaque régénération le réécrirait, un diff sur des fichiers que personne n'a
// touchés. Deux défauts de cette nature ont été trouvés avant ce contrôle : un
// point final ajouté au résumé d'un docblock, une classe globale importée.
//
// Les classes de l'utilisateur n'en sont pas : écrites une fois, ce que son
// outil de style y change lui appartient.
//
// @Symfony évolue avec PHP-CS-Fixer, dont la version est celle du lock : un
// échec qui suit une mise à jour de composer.lock vient de là, pas du rendu.
$finder = PhpCsFixer\Finder::create()
    ->in(__DIR__.'/tests/Generation/attendus')
    ->path('#/orm\d+/(Base|Enum|Trait)/#');

return (new PhpCsFixer\Config())
    ->setRules(['@Symfony' => true])
    ->setFinder($finder);
