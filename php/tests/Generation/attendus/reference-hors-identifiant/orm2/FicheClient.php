<?php

// Créé par Ormeau, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\FicheClientBase;
use Doctrine\ORM\Mapping as ORM;

/** clé primaire étrangère vers client.code : ni héritage ni un-vers-un possibles. */
#[ORM\Entity]
#[ORM\Table(
    name: 'fiche_client',
    options: ['comment' => 'clé primaire étrangère vers client.code : ni héritage ni un-vers-un possibles'],
)]
class FicheClient extends FicheClientBase {}
