<?php

// Créé par Ormeau depuis la base enumerations-decidees, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\CommandeBase;
use Doctrine\ORM\Mapping as ORM;

/** Un CHECK portant sur deux colonnes ne décrit pas une énumération. */
#[ORM\Entity]
#[ORM\Table(
    name: 'commande',
    options: ['comment' => 'Un CHECK portant sur deux colonnes ne décrit pas une énumération'],
)]
class Commande extends CommandeBase {}
