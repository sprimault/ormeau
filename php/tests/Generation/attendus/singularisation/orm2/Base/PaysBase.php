<?php

// Généré par Ormeau depuis la base singularisation et réécrit à chaque génération : le code propre à Pays va dans Pays.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class PaysBase
{
    #[ORM\Id]
    #[ORM\Column(name: 'code', type: 'string', length: 2)]
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
