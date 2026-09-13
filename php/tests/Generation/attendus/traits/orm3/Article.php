<?php

// Créé par Ormeau, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\ArticleBase;
use Doctrine\ORM\Mapping as ORM;

/** Colonnes françaises, et une suppression logique : autre signature. */
#[ORM\Entity]
#[ORM\Table(
    name: 'article',
    options: ['comment' => 'Colonnes françaises, et une suppression logique : autre signature'],
)]
class Article extends ArticleBase {}
