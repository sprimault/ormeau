<?php

// Créé par Ormeau depuis la base sequence-sql-server, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\NoteBase;
use Doctrine\ORM\Mapping as ORM;

/** Séquence du schéma par défaut, qui part de 1000 : le départ n'est pas le minimum. */
#[ORM\Entity]
#[ORM\Table(
    name: 'note',
    options: ['comment' => 'Séquence du schéma par défaut, qui part de 1000 : le départ n\'est pas le minimum'],
)]
class Note extends NoteBase {}
