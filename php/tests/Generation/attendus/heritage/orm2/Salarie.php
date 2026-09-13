<?php

// Créé par Ormeau, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\SalarieBase;
use Doctrine\ORM\Mapping as ORM;

/** La clé primaire est aussi une clé étrangère : héritage possible, un-vers-un sans décision. */
#[ORM\Entity]
#[ORM\Table(
    name: 'salarie',
    options: ['comment' => 'La clé primaire est aussi une clé étrangère : héritage possible, un-vers-un sans décision'],
)]
class Salarie extends SalarieBase {}
