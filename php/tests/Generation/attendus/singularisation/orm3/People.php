<?php

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\PeopleBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 'people')]
class People extends PeopleBase {}
