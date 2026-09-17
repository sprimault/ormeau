<?php

// Créé par Ormeau depuis la base schemas-cites, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\LeftBase;
use Doctrine\ORM\Mapping as ORM;

/** mot réservé derrière un schéma ordinaire : accepté tel quel, comme en 0.5.0. */
#[ORM\Entity]
#[ORM\Table(
    name: '`left`',
    options: ['comment' => 'mot réservé derrière un schéma ordinaire : accepté tel quel, comme en 0.5.0'],
)]
class Left extends LeftBase {}
