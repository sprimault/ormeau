<?php

// Généré par Ormeau depuis la base sequences et réécrit à chaque génération : le code propre à Dynamique va dans Dynamique.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class DynamiqueBase
{
    #[ORM\Id]
    #[ORM\Column(name: 'id', type: 'bigint')]
    protected int $id;

    #[ORM\Column(name: 'numero', type: 'string', length: 20)]
    protected string $numero;

    public function getId(): int
    {
        return $this->id;
    }

    public function setId(int $id): static
    {
        $this->id = $id;

        return $this;
    }

    public function getNumero(): string
    {
        return $this->numero;
    }

    public function setNumero(string $numero): static
    {
        $this->numero = $numero;

        return $this;
    }
}
