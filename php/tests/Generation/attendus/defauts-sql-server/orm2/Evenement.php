<?php

// Créé par Ormeau depuis la base defauts-sql-server, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\EvenementBase;
use Doctrine\ORM\Mapping as ORM;

/** Défauts calculés, sous les formes que rend sys.default_constraints : l'instant est celui de l'instruction, à l'heure locale du serveur. */
#[ORM\Entity]
#[ORM\Table(
    name: 'evenement',
    options: ['comment' => 'Défauts calculés, sous les formes que rend sys.default_constraints : l\'instant est celui de l\'instruction, à l\'heure locale du serveur'],
)]
class Evenement extends EvenementBase {}
