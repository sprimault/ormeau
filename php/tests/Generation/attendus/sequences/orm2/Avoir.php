<?php

// Créé par Ormeau, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\AvoirBase;
use Doctrine\ORM\Mapping as ORM;

/** serial hors du search_path de l'extraction : nom qualifié. */
#[ORM\Entity]
#[ORM\Table(
    name: 'avoir',
    schema: 'gescom',
    options: ['comment' => 'serial hors du search_path de l\'extraction : nom qualifié'],
)]
class Avoir extends AvoirBase {}
