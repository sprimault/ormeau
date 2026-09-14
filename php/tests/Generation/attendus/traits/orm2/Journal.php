<?php

// Créé par Ormeau depuis la base traits, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\JournalBase;
use Doctrine\ORM\Mapping as ORM;

/** Une seule colonne d'horodatage : rien à factoriser. */
#[ORM\Entity]
#[ORM\Table(name: 'journal', options: ['comment' => 'Une seule colonne d\'horodatage : rien à factoriser'])]
class Journal extends JournalBase {}
