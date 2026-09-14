<?php

// Créé par Ormeau depuis la base sequences, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\ReglementBase;
use Doctrine\ORM\Mapping as ORM;

/** serial extrait sous un search_path vide : public qualifié, retiré par l'inférence. */
#[ORM\Entity]
#[ORM\Table(
    name: 'reglement',
    schema: 'public',
    options: ['comment' => 'serial extrait sous un search_path vide : public qualifié, retiré par l\'inférence'],
)]
class Reglement extends ReglementBase {}
