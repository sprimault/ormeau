<?php

// Généré par Ormeau depuis la base associations et réécrit à chaque génération : le code propre à Client va dans Client.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\AdresseFacturation;
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

    #[ORM\Column(name: 'raison_sociale', type: 'string', length: 120)]
    protected string $raisonSociale;

    #[ORM\OneToOne(targetEntity: AdresseFacturation::class, mappedBy: 'client')]
    protected ?AdresseFacturation $adresseFacturation = null;

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

    public function getRaisonSociale(): string
    {
        return $this->raisonSociale;
    }

    public function setRaisonSociale(string $raisonSociale): static
    {
        $this->raisonSociale = $raisonSociale;

        return $this;
    }

    public function getAdresseFacturation(): ?AdresseFacturation
    {
        return $this->adresseFacturation;
    }

    /** @return Collection<int, Commande> */
    public function getCommande(): Collection
    {
        return $this->commande;
    }
}
