<?php

// Créé par Ormeau, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\ClientBase;
use Doctrine\ORM\Mapping as ORM;

/** code est unique mais n'est pas la clé primaire. */
#[ORM\Entity]
#[ORM\Table(name: 'client', options: ['comment' => 'code est unique mais n\'est pas la clé primaire'])]
class Client extends ClientBase {}
