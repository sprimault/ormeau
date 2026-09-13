<?php

// Généré par Ormeau et réécrit à chaque génération : le code propre à Commande va dans Commande.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Client;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class CommandeBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    /** Lecture seule : écrite par l'association client. */
    #[ORM\Column(name: 'client_id', type: 'integer', insertable: false, updatable: false)]
    protected int $clientId;

    #[ORM\ManyToOne(targetEntity: Client::class, inversedBy: 'commande')]
    #[ORM\JoinColumn(name: 'client_id', referencedColumnName: 'id', nullable: false)]
    protected Client $client;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getClientId(): int
    {
        return $this->clientId;
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
