<?php

// Créé par Ormeau depuis la base schemas-cites, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\ClientBase;
use Doctrine\ORM\Mapping as ORM;

/** schéma par défaut : rien ne change. */
#[ORM\Entity]
#[ORM\Table(name: 'client', options: ['comment' => 'schéma par défaut : rien ne change'])]
class Client extends ClientBase {}
