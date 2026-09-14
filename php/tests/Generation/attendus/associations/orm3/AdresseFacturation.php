<?php

// Créé par Ormeau depuis la base associations, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\AdresseFacturationBase;
use Doctrine\ORM\Mapping as ORM;

/** Un vers un : la FK porte une contrainte d'unicité. */
#[ORM\Entity]
#[ORM\Table(name: 'adresse_facturation', options: ['comment' => 'Un vers un : la FK porte une contrainte d\'unicité'])]
class AdresseFacturation extends AdresseFacturationBase {}
