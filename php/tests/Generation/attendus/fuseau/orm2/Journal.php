<?php

// Créé par Ormeau depuis la base fuseau, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\JournalBase;
use Doctrine\ORM\Mapping as ORM;

/** Les mêmes cas sous les formes de PostgreSQL. */
#[ORM\Entity]
#[ORM\Table(name: 'journal', options: ['comment' => 'Les mêmes cas sous les formes de PostgreSQL'])]
class Journal extends JournalBase {}
