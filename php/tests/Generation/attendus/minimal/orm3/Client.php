<?php

// Créé par Ormeau, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\ClientBase;
use Doctrine\ORM\Mapping as ORM;

/** Fiche client. */
#[ORM\Entity]
#[ORM\Table(name: 'client', options: ['comment' => 'Fiche client'])]
class Client extends ClientBase {}
