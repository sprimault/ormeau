<?php

// Généré par Ormeau depuis la base mots-reserves et réécrit à chaque génération : le code propre à User va dans User.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Group;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class UserBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    #[ORM\Column(name: '`order`', type: 'integer')]
    protected int $order;

    #[ORM\Column(name: '`left`', type: 'text', nullable: true)]
    protected ?string $left = null;

    #[ORM\Column(name: 'between', type: 'integer', nullable: true)]
    protected ?int $between = null;

    /** @var Collection<int, Group> */
    #[ORM\OneToMany(targetEntity: Group::class, mappedBy: 'user')]
    protected Collection $group;

    public function __construct()
    {
        $this->group = new ArrayCollection();
    }

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getOrder(): int
    {
        return $this->order;
    }

    public function setOrder(int $order): static
    {
        $this->order = $order;

        return $this;
    }

    public function getLeft(): ?string
    {
        return $this->left;
    }

    public function setLeft(?string $left): static
    {
        $this->left = $left;

        return $this;
    }

    public function getBetween(): ?int
    {
        return $this->between;
    }

    public function setBetween(?int $between): static
    {
        $this->between = $between;

        return $this;
    }

    /** @return Collection<int, Group> */
    public function getGroup(): Collection
    {
        return $this->group;
    }
}
