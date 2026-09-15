<?php

// Créé par Ormeau depuis la base defauts, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\CommandeBase;
use Doctrine\ORM\Mapping as ORM;

/** Horodatage avec défaut, comme livraison : un trait. */
#[ORM\Entity]
#[ORM\Table(name: 'commande', options: ['comment' => 'Horodatage avec défaut, comme livraison : un trait'])]
class Commande extends CommandeBase {}
