<?php

// Généré par Ormeau depuis la base relations-forcees et réécrit à chaque génération : le code propre à Client va dans Client.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Commande;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class ClientBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    #[ORM\Column(name: 'nom', type: 'string', length: 80)]
    protected string $nom;

    /** @var Collection<int, Commande> */
    #[ORM\OneToMany(targetEntity: Commande::class, mappedBy: 'client')]
    protected Collection $commande;

    public function __construct()
    {
        $this->commande = new ArrayCollection();
    }

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getNom(): string
    {
        return $this->nom;
    }

    public function setNom(string $nom): static
    {
        $this->nom = $nom;

        return $this;
    }

    /** @return Collection<int, Commande> */
    public function getCommande(): Collection
    {
        return $this->commande;
    }
}
