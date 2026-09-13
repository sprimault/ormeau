<?php

// Créé par Ormeau, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\MajusculeBase;
use Doctrine\ORM\Mapping as ORM;

/** bigserial au nom en majuscules : PostgreSQL cite le nom de la séquence. */
#[ORM\Entity]
#[ORM\Table(
    name: '`Majuscule`',
    schema: 'public',
    options: ['comment' => 'bigserial au nom en majuscules : PostgreSQL cite le nom de la séquence'],
)]
class Majuscule extends MajusculeBase {}
