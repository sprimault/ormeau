<?php

// Créé par Ormeau, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\GroupBase;
use Doctrine\ORM\Mapping as ORM;

/** un seul schéma, donc aucun schéma écrit : la table réservée se cite. */
#[ORM\Entity]
#[ORM\Table(
    name: '`group`',
    options: ['comment' => 'un seul schéma, donc aucun schéma écrit : la table réservée se cite'],
)]
class Group extends GroupBase {}
