<?php

// Créé par Ormeau, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\CommandeBase;
use Doctrine\ORM\Mapping as ORM;

/** Plusieurs vers un, plus une FK auto-référencée facultative. */
#[ORM\Entity]
#[ORM\Table(name: 'commande', options: ['comment' => 'Plusieurs vers un, plus une FK auto-référencée facultative'])]
class Commande extends CommandeBase {}
