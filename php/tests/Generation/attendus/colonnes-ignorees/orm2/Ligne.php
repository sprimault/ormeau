<?php

// Créé par Ormeau depuis la base colonnes-ignorees, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\LigneBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 'ligne')]
class Ligne extends LigneBase {}
