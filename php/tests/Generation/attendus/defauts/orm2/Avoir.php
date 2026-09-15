<?php

// Créé par Ormeau depuis la base defauts, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\AvoirBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 'avoir')]
class Avoir extends AvoirBase {}
