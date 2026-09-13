<?php

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\AdressesBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 'adresses')]
class Adresses extends AdressesBase {}
