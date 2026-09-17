<?php

// Créé par Ormeau depuis la base index-filtre-sql-server, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\CommercialBase;
use Doctrine\ORM\Mapping as ORM;

/** Index filtrés sous SQL Server : DBAL n'écrit pas le filtre, rendu sans lui et signalé au rapport. */
#[ORM\Entity]
#[ORM\Table(
    name: 'commercial',
    schema: 'ventes',
    options: ['comment' => 'Index filtrés sous SQL Server : DBAL n\'écrit pas le filtre, rendu sans lui et signalé au rapport'],
)]
class Commercial extends CommercialBase {}
