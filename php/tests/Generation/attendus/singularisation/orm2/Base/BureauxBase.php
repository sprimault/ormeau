<?php

// Généré par Ormeau et réécrit à chaque génération : le code propre à Bureaux va dans Bureaux.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class BureauxBase
{
    #[ORM\Id]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected int $id;

    public function getId(): int
    {
        return $this->id;
    }

    public function setId(int $id): static
    {
        $this->id = $id;

        return $this;
    }
}
