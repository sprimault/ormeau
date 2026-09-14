<?php

// Créé par Ormeau depuis la base schemas-cites, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\AvoirBase;
use Doctrine\ORM\Mapping as ORM;

/** seul le schéma est à citer. */
#[ORM\Entity]
#[ORM\Table(name: '`avoir`', schema: 'Compta', options: ['comment' => 'seul le schéma est à citer'])]
class Avoir extends AvoirBase {}
