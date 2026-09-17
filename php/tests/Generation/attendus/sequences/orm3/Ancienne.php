<?php

// Créé par Ormeau depuis la base sequences, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\AncienneBase;
use Doctrine\ORM\Mapping as ORM;

/** forme restaurée d'une base antérieure à PostgreSQL 8.1. */
#[ORM\Entity]
#[ORM\Table(name: 'ancienne', options: ['comment' => 'forme restaurée d\'une base antérieure à PostgreSQL 8.1'])]
class Ancienne extends AncienneBase {}
