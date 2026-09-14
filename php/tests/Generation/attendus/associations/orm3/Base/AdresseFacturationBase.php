<?php

// Généré par Ormeau depuis la base associations et réécrit à chaque génération : le code propre à AdresseFacturation va dans AdresseFacturation.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Client;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class AdresseFacturationBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    /** Lecture seule : écrite par l'association client. */
    #[ORM\Column(name: 'client_id', type: 'integer', insertable: false, updatable: false)]
    protected int $clientId;

    #[ORM\Column(name: 'rue', type: 'string', length: 200)]
    protected string $rue;

    #[ORM\OneToOne(targetEntity: Client::class, inversedBy: 'adresseFacturation')]
    #[ORM\JoinColumn(name: 'client_id', referencedColumnName: 'id', nullable: false, onDelete: 'CASCADE')]
    protected Client $client;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getClientId(): int
    {
        return $this->clientId;
    }

    public function getRue(): string
    {
        return $this->rue;
    }

    public function setRue(string $rue): static
    {
        $this->rue = $rue;

        return $this;
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
}
