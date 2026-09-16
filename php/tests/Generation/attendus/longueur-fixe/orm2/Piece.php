<?php

// Créé par Ormeau depuis la base longueur-fixe, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\PieceBase;
use Doctrine\ORM\Mapping as ORM;

/** Longueur fixe et binaires sous les formes de SQL Server : le fait vient du champ longueur_fixe, pas du nom du type. */
#[ORM\Entity]
#[ORM\Table(
    name: 'piece',
    options: ['comment' => 'Longueur fixe et binaires sous les formes de SQL Server : le fait vient du champ longueur_fixe, pas du nom du type'],
)]
class Piece extends PieceBase {}
