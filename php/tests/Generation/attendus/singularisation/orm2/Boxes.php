<?php

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\BoxesBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 'boxes')]
class Boxes extends BoxesBase {}
