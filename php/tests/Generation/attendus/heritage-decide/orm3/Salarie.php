<?php

// Créé par Ormeau depuis la base heritage-decide, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\SalarieBase;
use Doctrine\ORM\Mapping as ORM;

/** Déclaré enfant de personne par le fichier de décisions. */
#[ORM\Entity]
#[ORM\Table(name: 'salarie', options: ['comment' => 'Déclaré enfant de personne par le fichier de décisions'])]
class Salarie extends SalarieBase {}
