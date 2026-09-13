<?php

// Créé par Ormeau, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\AffectationBase;
use Doctrine\ORM\Mapping as ORM;

/** Identifiée par salarie : possible une fois salarie déclaré enfant, sa clé devenant celle de personne. */
#[ORM\Entity]
#[ORM\Table(
    name: 'affectation',
    options: ['comment' => 'Identifiée par salarie : possible une fois salarie déclaré enfant, sa clé devenant celle de personne'],
)]
class Affectation extends AffectationBase {}
