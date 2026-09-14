<?php

// Créé par Ormeau depuis la base heritage-decide, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\AdresseBase;
use Doctrine\ORM\Mapping as ORM;

/** Vise la racine : la collection vit dans la classe de base de personne. */
#[ORM\Entity]
#[ORM\Table(
    name: 'adresse',
    options: ['comment' => 'Vise la racine : la collection vit dans la classe de base de personne'],
)]
class Adresse extends AdresseBase {}
