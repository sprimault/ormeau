<?php

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\CommandesClientsBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: '`Commandes Clients`')]
class CommandesClients extends CommandesClientsBase {}
