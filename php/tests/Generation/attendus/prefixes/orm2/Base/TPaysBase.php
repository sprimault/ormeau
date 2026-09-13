<?php

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class TPaysBase
{
    #[ORM\Id]
    #[ORM\Column(name: '`CODE`', type: 'string', length: 2)]
    protected string $code;

    public function getCode(): string
    {
        return $this->code;
    }

    public function setCode(string $code): static
    {
        $this->code = $code;

        return $this;
    }
}
