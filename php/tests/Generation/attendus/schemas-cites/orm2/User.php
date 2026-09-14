<?php

// Créé par Ormeau depuis la base schemas-cites, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\UserBase;
use Doctrine\ORM\Mapping as ORM;

/** schéma et table en mots réservés, sans majuscule ni accent. */
#[ORM\Entity]
#[ORM\Table(
    name: '`user`',
    schema: 'order',
    options: ['comment' => 'schéma et table en mots réservés, sans majuscule ni accent'],
)]
class User extends UserBase {}
