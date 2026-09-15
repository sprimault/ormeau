<?php

// Généré par Ormeau depuis la base defauts et réécrit à chaque génération.

declare(strict_types=1);

namespace App\Entity\Trait;

use Doctrine\ORM\Mapping as ORM;

trait Horodatage
{
    #[ORM\Column(name: 'created_at', type: 'datetimetz_immutable')]
    protected \DateTimeImmutable $createdAt;

    public function getCreatedAt(): \DateTimeImmutable
    {
        return $this->createdAt;
    }

    public function setCreatedAt(\DateTimeImmutable $createdAt): static
    {
        $this->createdAt = $createdAt;

        return $this;
    }
}
