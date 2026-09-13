<?php

// Créé par Ormeau, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\LigneCommandeBase;
use Doctrine\ORM\Mapping as ORM;

/** Deux FK et une PK composite, mais une colonne en plus : ce n'est pas une jointure pure. */
#[ORM\Entity]
#[ORM\Table(
    name: 'ligne_commande',
    options: ['comment' => 'Deux FK et une PK composite, mais une colonne en plus : ce n\'est pas une jointure pure'],
)]
class LigneCommande extends LigneCommandeBase {}
