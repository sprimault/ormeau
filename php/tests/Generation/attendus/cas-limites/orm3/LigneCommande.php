<?php

// Créé par Ormeau depuis la base cas-limites, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\LigneCommandeBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 'ligne_commande', schema: 'public')]
class LigneCommande extends LigneCommandeBase {}
