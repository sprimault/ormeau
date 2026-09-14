<?php

// Créé par Ormeau, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\CommandeBase;
use Doctrine\ORM\Mapping as ORM;

/** clé étrangère déclarée vers client.code, et deux relations forcées vers autre chose que l'identifiant. */
#[ORM\Entity]
#[ORM\Table(
    name: 'commande',
    options: ['comment' => 'clé étrangère déclarée vers client.code, et deux relations forcées vers autre chose que l\'identifiant'],
)]
class Commande extends CommandeBase {}
