<?php

// Créé par Ormeau, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\LignesDeCommandeBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: '`Lignes de commande`')]
class LignesDeCommande extends LignesDeCommandeBase {}
