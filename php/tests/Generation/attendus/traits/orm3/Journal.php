<?php

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\JournalBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 'journal')]
class Journal extends JournalBase {}
