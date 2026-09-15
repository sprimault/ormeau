<?php

// Généré par Ormeau depuis la base gescom et réécrit à chaque génération : le code propre à TRéférence va dans TRéférence.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class TRéférenceBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    #[ORM\Column(name: '`order`', type: 'integer', nullable: true)]
    protected ?int $order = null;

    #[ORM\Column(name: '`select`', type: 'string', length: 10, nullable: true)]
    protected ?string $select = null;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getOrder(): ?int
    {
        return $this->order;
    }

    public function setOrder(?int $order): static
    {
        $this->order = $order;

        return $this;
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
