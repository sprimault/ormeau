<?php

// Créé par Ormeau depuis la base fuseau, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\EvenementBase;
use Doctrine\ORM\Mapping as ORM;

/** Fuseau sous les formes de SQL Server : le fait vient du champ fuseau, pas du nom du type. */
#[ORM\Entity]
#[ORM\Table(
    name: 'evenement',
    options: ['comment' => 'Fuseau sous les formes de SQL Server : le fait vient du champ fuseau, pas du nom du type'],
)]
class Evenement extends EvenementBase {}
