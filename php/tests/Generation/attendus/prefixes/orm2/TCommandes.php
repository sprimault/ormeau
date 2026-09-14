<?php

// Créé par Ormeau depuis la base prefixes, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\TCommandesBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: '`T_COMMANDES`')]
class TCommandes extends TCommandesBase {}
