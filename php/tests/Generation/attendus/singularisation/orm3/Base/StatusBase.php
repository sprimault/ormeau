<?php

// Généré par Ormeau et réécrit à chaque génération : le code propre à Status va dans Status.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class StatusBase
{
    #[ORM\Id]
    #[ORM\Column(name: 'code', type: 'string', length: 16)]
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
