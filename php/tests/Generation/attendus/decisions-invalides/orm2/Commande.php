<?php

// Créé par Ormeau depuis la base decisions-invalides, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\CommandeBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 'commande')]
class Commande extends CommandeBase {}
