<?php

// Créé par Ormeau depuis la base decisions-invalides, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\ClientBase;
use Doctrine\ORM\Mapping as ORM;

/** Visée par des décisions dont les noms ne sont pas du PHP. */
#[ORM\Entity]
#[ORM\Table(name: 'client', options: ['comment' => 'Visée par des décisions dont les noms ne sont pas du PHP'])]
class Client extends ClientBase {}
