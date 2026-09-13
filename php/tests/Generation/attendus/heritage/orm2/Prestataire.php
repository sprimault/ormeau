<?php

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\PrestataireBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 'prestataire')]
class Prestataire extends PrestataireBase {}
