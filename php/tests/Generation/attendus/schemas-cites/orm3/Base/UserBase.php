<?php

// Généré par Ormeau et réécrit à chaque génération : le code propre à User va dans User.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class UserBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    #[ORM\Column(name: '`select`', type: 'string', length: 20, nullable: true)]
    protected ?string $select = null;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getSelect(): ?string
    {
        return $this->select;
    }

    public function setSelect(?string $select): static
    {
        $this->select = $select;

        return $this;
    }
}
