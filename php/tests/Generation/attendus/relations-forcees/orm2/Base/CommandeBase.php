<?php

// Généré par Ormeau depuis la base relations-forcees et réécrit à chaque génération : le code propre à Commande va dans Commande.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Client;
use App\Entity\Utilisateur;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class CommandeBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    /** Lecture seule : écrite par l'association client. */
    #[ORM\Column(name: 'client_ref', type: 'integer', insertable: false, updatable: false)]
    protected int $clientRef;

    #[ORM\Column(name: 'fournisseur_id', type: 'integer', nullable: true)]
    protected ?int $fournisseurId = null;

    /** Lecture seule : écrite par l'association commercial. */
    #[ORM\Column(name: 'vendeur_id', type: 'integer', nullable: true, insertable: false, updatable: false)]
    protected ?int $vendeurId = null;

    #[ORM\ManyToOne(targetEntity: Client::class, inversedBy: 'commande')]
    #[ORM\JoinColumn(name: 'client_ref', referencedColumnName: 'id', nullable: false)]
    protected Client $client;

    #[ORM\ManyToOne(targetEntity: Utilisateur::class, inversedBy: 'commande')]
    #[ORM\JoinColumn(name: 'vendeur_id', referencedColumnName: 'id', onDelete: 'SET NULL')]
    protected ?Utilisateur $commercial = null;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getClientRef(): int
    {
        return $this->clientRef;
    }

    public function getFournisseurId(): ?int
    {
        return $this->fournisseurId;
    }

    public function setFournisseurId(?int $fournisseurId): static
    {
        $this->fournisseurId = $fournisseurId;

        return $this;
    }

    public function getVendeurId(): ?int
    {
        return $this->vendeurId;
    }

    public function getClient(): Client
    {
        return $this->client;
    }

    public function setClient(Client $client): static
    {
        $this->client = $client;

        return $this;
    }

    public function getCommercial(): ?Utilisateur
    {
        return $this->commercial;
    }

    public function setCommercial(?Utilisateur $commercial): static
    {
        $this->commercial = $commercial;

        return $this;
    }
}
