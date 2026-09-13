<?php

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\FournisseurBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 'tbl_fournisseur')]
class Fournisseur extends FournisseurBase {}
