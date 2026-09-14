<?php

// Créé par Ormeau depuis la base reference-hors-identifiant, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\ProduitBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 'produit')]
class Produit extends ProduitBase {}
