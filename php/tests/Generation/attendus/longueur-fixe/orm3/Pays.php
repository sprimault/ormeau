<?php

// Créé par Ormeau depuis la base longueur-fixe, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\PaysBase;
use Doctrine\ORM\Mapping as ORM;

/** Les mêmes cas sous les formes de PostgreSQL. */
#[ORM\Entity]
#[ORM\Table(name: 'pays', schema: 'ventes', options: ['comment' => 'Les mêmes cas sous les formes de PostgreSQL'])]
class Pays extends PaysBase {}
