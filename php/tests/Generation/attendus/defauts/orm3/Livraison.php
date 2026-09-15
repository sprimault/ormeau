<?php

// Créé par Ormeau depuis la base defauts, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\LivraisonBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 'livraison')]
class Livraison extends LivraisonBase {}
