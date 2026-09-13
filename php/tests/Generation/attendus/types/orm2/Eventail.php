<?php

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\EventailBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 'eventail')]
class Eventail extends EventailBase {}
