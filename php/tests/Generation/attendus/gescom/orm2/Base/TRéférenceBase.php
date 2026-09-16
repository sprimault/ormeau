<?php

// Généré par Ormeau depuis la base gescom et réécrit à chaque génération : le code propre à TRéférence va dans TRéférence.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
#[ORM\Index(name: 'ix_reference_libelle', columns: ['`Libellé`'])]
#[ORM\Index(name: 'ix_reference_order', columns: ['`order`'])]
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

    #[ORM\Column(name: '`Libellé`', type: 'string', length: 30, nullable: true)]
    protected ?string $libellé = null;

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

    public function getLibellé(): ?string
    {
        return $this->libellé;
    }

    public function setLibellé(?string $libellé): static
    {
        $this->libellé = $libellé;

        return $this;
    }
}
