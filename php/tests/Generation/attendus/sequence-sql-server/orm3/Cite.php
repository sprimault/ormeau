<?php

// Créé par Ormeau depuis la base sequence-sql-server, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\CiteBase;
use Doctrine\ORM\Mapping as ORM;

/** Séquence au nom à citer : non écrite plutôt que citée au hasard. */
#[ORM\Entity]
#[ORM\Table(
    name: 'cite',
    schema: 'ventes',
    options: ['comment' => 'Séquence au nom à citer : non écrite plutôt que citée au hasard'],
)]
class Cite extends CiteBase {}
