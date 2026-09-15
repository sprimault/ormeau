<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

use Ormeau\Doctrine\Tests\AllerRetour\Recreation;

require __DIR__ . '/../../vendor/autoload.php';

if (!isset($argv[1])) {
    fwrite(STDERR, "usage : php recreer.php <paramètres.json>\n");
    exit(2);
}

// Sortie lue par le test Go de l'aller-retour. Ce n'est pas un calque : JSON
// direct, sans exigence de déterminisme.
echo json_encode(Recreation::executer($argv[1]), JSON_THROW_ON_ERROR | JSON_UNESCAPED_UNICODE), "\n";
