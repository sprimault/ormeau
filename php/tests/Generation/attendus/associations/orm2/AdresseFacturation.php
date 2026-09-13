<?php

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\AdresseFacturationBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 'adresse_facturation')]
class AdresseFacturation extends AdresseFacturationBase {}
