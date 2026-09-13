<?php

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\CatégoriesDesNomsBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: '`catégories des noms`')]
class CatégoriesDesNoms extends CatégoriesDesNomsBase {}
