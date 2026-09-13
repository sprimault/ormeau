<?php

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Commande;
use App\Entity\Profil;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class UtilisateurBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    #[ORM\Column(name: 'email', type: 'string', length: 120)]
    protected string $email;

    /** @var Collection<int, Commande> */
    #[ORM\OneToMany(targetEntity: Commande::class, mappedBy: 'commercial')]
    protected Collection $commande;

    #[ORM\OneToOne(targetEntity: Profil::class, mappedBy: 'utilisateur')]
    protected ?Profil $profil = null;

    public function __construct()
    {
        $this->commande = new ArrayCollection();
    }

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getEmail(): string
    {
        return $this->email;
    }

    public function setEmail(string $email): static
    {
        $this->email = $email;

        return $this;
    }

    /** @return Collection<int, Commande> */
    public function getCommande(): Collection
    {
        return $this->commande;
    }

    public function getProfil(): ?Profil
    {
        return $this->profil;
    }
}
