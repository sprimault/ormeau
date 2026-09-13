<?php

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\JournauxBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 'journaux')]
class Journaux extends JournauxBase {}
