<?php

// Créé par Ormeau, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\DynamiqueBase;
use Doctrine\ORM\Mapping as ORM;

/** séquence choisie à l'exécution : aucun nom à lire. */
#[ORM\Entity]
#[ORM\Table(
    name: 'dynamique',
    schema: 'public',
    options: ['comment' => 'séquence choisie à l\'exécution : aucun nom à lire'],
)]
class Dynamique extends DynamiqueBase {}
