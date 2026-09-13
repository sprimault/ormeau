<?php

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\LigneCommandeBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 'ligne_commande')]
class LigneCommande extends LigneCommandeBase {}
