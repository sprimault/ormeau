<?php

// Généré par Ormeau depuis la base gescom et réécrit à chaque génération : le code propre à TFacture va dans TFacture.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class TFactureBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'fac_id', type: 'integer')]
    protected ?int $facId = null;

    #[ORM\Column(name: 'fac_cli_id', type: 'integer')]
    protected int $facCliId;

    #[ORM\Column(name: 'fac_total', type: 'decimal', precision: 12, scale: 2)]
    protected string $facTotal;

    public function getFacId(): ?int
    {
        return $this->facId;
    }

    public function getFacCliId(): int
    {
        return $this->facCliId;
    }

    public function setFacCliId(int $facCliId): static
    {
        $this->facCliId = $facCliId;

        return $this;
    }

    public function getFacTotal(): string
    {
        return $this->facTotal;
    }

    public function setFacTotal(string $facTotal): static
    {
        $this->facTotal = $facTotal;

        return $this;
    }
}
