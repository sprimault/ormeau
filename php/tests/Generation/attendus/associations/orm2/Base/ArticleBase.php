<?php

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Commande;
use App\Entity\LigneCommande;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class ArticleBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    #[ORM\Column(name: 'libelle', type: 'string', length: 120)]
    protected string $libelle;

    /** @var Collection<int, LigneCommande> */
    #[ORM\OneToMany(targetEntity: LigneCommande::class, mappedBy: 'article')]
    protected Collection $ligneCommande;

    /** @var Collection<int, Commande> */
    #[ORM\ManyToMany(targetEntity: Commande::class, mappedBy: 'article')]
    protected Collection $commande;

    public function __construct()
    {
        $this->ligneCommande = new ArrayCollection();
        $this->commande = new ArrayCollection();
    }

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getLibelle(): string
    {
        return $this->libelle;
    }

    public function setLibelle(string $libelle): static
    {
        $this->libelle = $libelle;

        return $this;
    }

    /** @return Collection<int, LigneCommande> */
    public function getLigneCommande(): Collection
    {
        return $this->ligneCommande;
    }

    /** @return Collection<int, Commande> */
    public function getCommande(): Collection
    {
        return $this->commande;
    }
}
