<?php

// Créé par Ormeau depuis la base gescom, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\TAvoirBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 't_avoir', schema: 'gescom')]
class TAvoir extends TAvoirBase {}
