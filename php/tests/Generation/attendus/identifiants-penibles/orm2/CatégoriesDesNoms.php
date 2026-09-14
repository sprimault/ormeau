<?php

// Créé par Ormeau depuis la base identifiants-penibles, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\CatégoriesDesNomsBase;
use Doctrine\ORM\Mapping as ORM;

/** Accent et espaces : l'accent suffit à désigner le français. */
#[ORM\Entity]
#[ORM\Table(
    name: '`catégories des noms`',
    options: ['comment' => 'Accent et espaces : l\'accent suffit à désigner le français'],
)]
class CatégoriesDesNoms extends CatégoriesDesNomsBase {}
