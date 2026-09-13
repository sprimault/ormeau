<?php

// Généré par Ormeau et réécrit à chaque génération : le code propre à Client va dans Client.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Client;
use App\Entity\Commande;
use App\Entity\Enum\ClientCanal;
use App\Entity\Enum\OuiNon;
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

    #[ORM\Column(name: 'canal', type: 'string', length: 20, enumType: ClientCanal::class)]
    protected ClientCanal $canal;

    #[ORM\Column(name: 'actif', type: 'string', length: 1, enumType: OuiNon::class)]
    protected OuiNon $actif;

    /** Lecture seule : écrite par l'association client. */
    #[ORM\Column(name: 'parrain_ref', type: 'integer', nullable: true, insertable: false, updatable: false)]
    protected ?int $parrainRef = null;

    #[ORM\ManyToOne(targetEntity: Client::class, inversedBy: 'clientClient')]
    #[ORM\JoinColumn(name: 'parrain_ref', referencedColumnName: 'id')]
    protected ?Client $client = null;

    /** @var Collection<int, Client> */
    #[ORM\OneToMany(targetEntity: Client::class, mappedBy: 'client')]
    protected Collection $clientClient;

    /** @var Collection<int, Commande> */
    #[ORM\OneToMany(targetEntity: Commande::class, mappedBy: 'client')]
    protected Collection $commande;

    public function __construct()
    {
        $this->clientClient = new ArrayCollection();
        $this->commande = new ArrayCollection();
    }

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getCanal(): ClientCanal
    {
        return $this->canal;
    }

    public function setCanal(ClientCanal $canal): static
    {
        $this->canal = $canal;

        return $this;
    }

    public function getActif(): OuiNon
    {
        return $this->actif;
    }

    public function setActif(OuiNon $actif): static
    {
        $this->actif = $actif;

        return $this;
    }

    public function getParrainRef(): ?int
    {
        return $this->parrainRef;
    }

    public function getClient(): ?Client
    {
        return $this->client;
    }

    public function setClient(?Client $client): static
    {
        $this->client = $client;

        return $this;
    }

    /** @return Collection<int, Client> */
    public function getClientClient(): Collection
    {
        return $this->clientClient;
    }

    /** @return Collection<int, Commande> */
    public function getCommande(): Collection
    {
        return $this->commande;
    }
}
