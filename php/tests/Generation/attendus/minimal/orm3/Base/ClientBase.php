<?php

// Généré par Ormeau depuis la base minimal et réécrit à chaque génération : le code propre à Client va dans Client.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
#[ORM\Index(name: 'client_raison_sociale_idx', columns: ['raison_sociale'])]
#[ORM\UniqueConstraint(name: 'client_siret_key', columns: ['siret'])]
abstract class ClientBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    /** Dénomination légale. */
    #[ORM\Column(name: 'raison_sociale', type: 'string', length: 120, options: ['comment' => 'Dénomination légale'])]
    protected string $raisonSociale;

    #[ORM\Column(name: 'siret', type: 'string', length: 14, unique: true, nullable: true)]
    protected ?string $siret = null;

    #[ORM\Column(name: 'chiffre_affaires', type: 'decimal', precision: 12, scale: 2, nullable: true)]
    protected ?string $chiffreAffaires = null;

    #[ORM\Column(name: 'actif', type: 'boolean', options: ['default' => true])]
    protected bool $actif = true;

    #[ORM\Column(name: 'cree_le', type: 'datetimetz_immutable')]
    protected \DateTimeImmutable $creeLe;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getRaisonSociale(): string
    {
        return $this->raisonSociale;
    }

    public function setRaisonSociale(string $raisonSociale): static
    {
        $this->raisonSociale = $raisonSociale;

        return $this;
    }

    public function getSiret(): ?string
    {
        return $this->siret;
    }

    public function setSiret(?string $siret): static
    {
        $this->siret = $siret;

        return $this;
    }

    public function getChiffreAffaires(): ?string
    {
        return $this->chiffreAffaires;
    }

    public function setChiffreAffaires(?string $chiffreAffaires): static
    {
        $this->chiffreAffaires = $chiffreAffaires;

        return $this;
    }

    public function isActif(): bool
    {
        return $this->actif;
    }

    public function setActif(bool $actif): static
    {
        $this->actif = $actif;

        return $this;
    }

    public function getCreeLe(): \DateTimeImmutable
    {
        return $this->creeLe;
    }

    public function setCreeLe(\DateTimeImmutable $creeLe): static
    {
        $this->creeLe = $creeLe;

        return $this;
    }
}
