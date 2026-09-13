<?php

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\AffectationBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 'affectation')]
class Affectation extends AffectationBase {}
