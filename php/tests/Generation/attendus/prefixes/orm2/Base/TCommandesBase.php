<?php

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class TCommandesBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: '`ID`', type: 'integer')]
    protected ?int $id = null;

    #[ORM\Column(name: '`TOTAL_HT`', type: 'decimal', precision: 12, scale: 2)]
    protected string $totalHt;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getTotalHt(): string
    {
        return $this->totalHt;
    }

    public function setTotalHt(string $totalHt): static
    {
        $this->totalHt = $totalHt;

        return $this;
    }
}
