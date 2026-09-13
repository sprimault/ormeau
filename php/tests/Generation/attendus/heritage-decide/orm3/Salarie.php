<?php

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\SalarieBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 'salarie')]
class Salarie extends SalarieBase {}
