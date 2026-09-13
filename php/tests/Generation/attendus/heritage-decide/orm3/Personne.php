<?php

// Créé par Ormeau, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\PersonneBase;
use Doctrine\ORM\Mapping as ORM;

/** La racine porte la colonne qui dit quelle classe chaque ligne instancie. */
#[ORM\Entity]
#[ORM\Table(
    name: 'personne',
    options: ['comment' => 'La racine porte la colonne qui dit quelle classe chaque ligne instancie'],
)]
#[ORM\InheritanceType('JOINED')]
#[ORM\DiscriminatorColumn(name: 'nature', type: 'string', length: 1)]
#[ORM\DiscriminatorMap(['P' => Personne::class, 'S' => Salarie::class])]
class Personne extends PersonneBase {}
