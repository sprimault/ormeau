<?php

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\BureauxBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 'bureaux')]
class Bureaux extends BureauxBase {}
