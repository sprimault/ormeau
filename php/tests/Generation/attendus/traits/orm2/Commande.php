<?php

// Créé par Ormeau depuis la base traits, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\CommandeBase;
use Doctrine\ORM\Mapping as ORM;

/** Mêmes colonnes d'horodatage que client : le trait est partagé. */
#[ORM\Entity]
#[ORM\Table(
    name: 'commande',
    options: ['comment' => 'Mêmes colonnes d\'horodatage que client : le trait est partagé'],
)]
class Commande extends CommandeBase {}
