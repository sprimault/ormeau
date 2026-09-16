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

    #[ORM\Column(name: 'fac_date', type: 'date_immutable', options: ['default' => 'CURRENT_DATE'])]
    protected \DateTimeImmutable $facDate;

    #[ORM\Column(name: 'fac_saisie', type: 'date_immutable', nullable: true, options: ['default' => 'CURRENT_DATE'])]
    protected ?\DateTimeImmutable $facSaisie = null;

    #[ORM\Column(name: 'fac_taux', type: 'float', nullable: true)]
    protected ?float $facTaux = null;

    #[ORM\Column(name: 'fac_jeton', type: 'blob', nullable: true)]
    protected mixed $facJeton = null;

    #[ORM\Column(name: 'fac_niveau', type: 'smallint', nullable: true)]
    protected ?int $facNiveau = null;

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

    public function getFacDate(): \DateTimeImmutable
    {
        return $this->facDate;
    }

    public function setFacDate(\DateTimeImmutable $facDate): static
    {
        $this->facDate = $facDate;

        return $this;
    }

    public function getFacSaisie(): ?\DateTimeImmutable
    {
        return $this->facSaisie;
    }

    public function setFacSaisie(?\DateTimeImmutable $facSaisie): static
    {
        $this->facSaisie = $facSaisie;

        return $this;
    }

    public function getFacTaux(): ?float
    {
        return $this->facTaux;
    }

    public function setFacTaux(?float $facTaux): static
    {
        $this->facTaux = $facTaux;

        return $this;
    }

    public function getFacJeton(): mixed
    {
        return $this->facJeton;
    }

    public function setFacJeton(mixed $facJeton): static
    {
        $this->facJeton = $facJeton;

        return $this;
    }

    public function getFacNiveau(): ?int
    {
        return $this->facNiveau;
    }

    public function setFacNiveau(?int $facNiveau): static
    {
        $this->facNiveau = $facNiveau;

        return $this;
    }
}
