<?php

declare(strict_types=1);

namespace Gescom\Domaine\Entity;

use Gescom\Domaine\Entity\Base\ClientBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: '`T_CLIENTS`')]
class Client extends ClientBase {}
