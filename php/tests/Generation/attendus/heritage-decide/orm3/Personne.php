<?php

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\PersonneBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 'personne')]
#[ORM\InheritanceType('JOINED')]
#[ORM\DiscriminatorColumn(name: 'nature', type: 'string', length: 1)]
#[ORM\DiscriminatorMap(['P' => Personne::class, 'S' => Salarie::class])]
class Personne extends PersonneBase {}
