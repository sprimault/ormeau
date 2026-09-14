<?php

// Généré par Ormeau et réécrit à chaque génération : le code propre à Group va dans Group.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\User;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class GroupBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    /** Lecture seule : écrite par l'association user. */
    #[ORM\Column(name: 'user_id', type: 'integer', nullable: true, insertable: false, updatable: false)]
    protected ?int $userId = null;

    #[ORM\ManyToOne(targetEntity: User::class, inversedBy: 'group')]
    #[ORM\JoinColumn(name: 'user_id', referencedColumnName: 'id')]
    protected ?User $user = null;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getUserId(): ?int
    {
        return $this->userId;
    }

    public function getUser(): ?User
    {
        return $this->user;
    }

    public function setUser(?User $user): static
    {
        $this->user = $user;

        return $this;
    }
}
