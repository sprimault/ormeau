<?php

// Créé par Ormeau depuis la base sequence-sql-server, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\AncienBase;
use Doctrine\ORM\Mapping as ORM;

/** Calque extrait avant le champ sequence : rien ne désigne la séquence, rien n'est inventé. */
#[ORM\Entity]
#[ORM\Table(
    name: 'ancien',
    schema: 'ventes',
    options: ['comment' => 'Calque extrait avant le champ sequence : rien ne désigne la séquence, rien n\'est inventé'],
)]
class Ancien extends AncienBase {}
