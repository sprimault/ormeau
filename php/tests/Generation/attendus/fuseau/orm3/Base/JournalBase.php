<?php

// Généré par Ormeau depuis la base fuseau et réécrit à chaque génération : le code propre à Journal va dans Journal.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class JournalBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    #[ORM\Column(name: 'consigne_le', type: 'datetimetz_immutable')]
    protected \DateTimeImmutable $consigneLe;

    #[ORM\Column(name: 'heure', type: 'time_immutable', nullable: true)]
    protected ?\DateTimeImmutable $heure = null;

    /** Calque extrait avant le champ fuseau : la forme de PostgreSQL reste lue. */
    #[ORM\Column(
        name: 'ancien',
        type: 'datetimetz_immutable',
        nullable: true,
        options: ['comment' => 'Calque extrait avant le champ fuseau : la forme de PostgreSQL reste lue'],
    )]
    protected ?\DateTimeImmutable $ancien = null;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getConsigneLe(): \DateTimeImmutable
    {
        return $this->consigneLe;
    }

    public function setConsigneLe(\DateTimeImmutable $consigneLe): static
    {
        $this->consigneLe = $consigneLe;

        return $this;
    }

    public function getHeure(): ?\DateTimeImmutable
    {
        return $this->heure;
    }

    public function setHeure(?\DateTimeImmutable $heure): static
    {
        $this->heure = $heure;

        return $this;
    }

    public function getAncien(): ?\DateTimeImmutable
    {
        return $this->ancien;
    }

    public function setAncien(?\DateTimeImmutable $ancien): static
    {
        $this->ancien = $ancien;

        return $this;
    }
}
