<?php

// Créé par Ormeau depuis la base reference-hors-identifiant, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\LigneBase;
use Doctrine\ORM\Mapping as ORM;

/** clé étrangère composite vers toute la clé primaire d'article : l'association reste. */
#[ORM\Entity]
#[ORM\Table(
    name: 'ligne',
    options: ['comment' => 'clé étrangère composite vers toute la clé primaire d\'article : l\'association reste'],
)]
class Ligne extends LigneBase {}
