<?php

// Généré par Ormeau depuis la base index-filtre-sql-server et réécrit à chaque génération : le code propre à Commercial va dans Commercial.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
#[ORM\Index(name: 'ix_commercial_actifs', columns: ['nom'])]
#[ORM\Index(name: 'ix_commercial_matricule', columns: ['matricule'])]
#[ORM\UniqueConstraint(name: 'uq_commercial_email_actif', columns: ['email'])]
abstract class CommercialBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    #[ORM\Column(name: 'email', type: 'string', length: 100, nullable: true)]
    protected ?string $email = null;

    #[ORM\Column(name: 'nom', type: 'string', length: 50, nullable: true)]
    protected ?string $nom = null;

    #[ORM\Column(name: 'actif', type: 'boolean')]
    protected bool $actif;

    #[ORM\Column(name: 'matricule', type: 'string', length: 20, nullable: true)]
    protected ?string $matricule = null;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getEmail(): ?string
    {
        return $this->email;
    }

    public function setEmail(?string $email): static
    {
        $this->email = $email;

        return $this;
    }

    public function getNom(): ?string
    {
        return $this->nom;
    }

    public function setNom(?string $nom): static
    {
        $this->nom = $nom;

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

    public function getMatricule(): ?string
    {
        return $this->matricule;
    }

    public function setMatricule(?string $matricule): static
    {
        $this->matricule = $matricule;

        return $this;
    }
}
