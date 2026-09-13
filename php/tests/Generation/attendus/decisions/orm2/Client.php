<?php

// Créé par Ormeau, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace Gescom\Domaine\Entity;

use Doctrine\ORM\Mapping as ORM;
use Gescom\Domaine\Entity\Base\ClientBase;

#[ORM\Entity]
#[ORM\Table(name: '`T_CLIENTS`')]
class Client extends ClientBase {}
