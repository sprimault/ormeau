<?php

// Créé par Ormeau depuis la base sequences, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\JournalBase;
use Doctrine\ORM\Mapping as ORM;

/** schéma public écrit entre guillemets, forme que PostgreSQL ne rend pas : rien n'est retiré. */
#[ORM\Entity]
#[ORM\Table(
    name: '`Journal`',
    options: ['comment' => 'schéma public écrit entre guillemets, forme que PostgreSQL ne rend pas : rien n\'est retiré'],
)]
class Journal extends JournalBase {}
