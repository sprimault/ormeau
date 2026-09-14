<?php

// Généré par Ormeau et réécrit à chaque génération : le code propre à Commande va dans Commande.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class CommandeBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    #[ORM\Column(name: 'client_code', type: 'string', length: 10)]
    protected string $clientCode;

    #[ORM\Column(name: 'payeur_code', type: 'string', length: 10, nullable: true)]
    protected ?string $payeurCode = null;

    #[ORM\Column(name: 'livre_a', type: 'string', length: 80, nullable: true)]
    protected ?string $livreA = null;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getClientCode(): string
    {
        return $this->clientCode;
    }

    public function setClientCode(string $clientCode): static
    {
        $this->clientCode = $clientCode;

        return $this;
    }

    public function getPayeurCode(): ?string
    {
        return $this->payeurCode;
    }

    public function setPayeurCode(?string $payeurCode): static
    {
        $this->payeurCode = $payeurCode;

        return $this;
    }

    public function getLivreA(): ?string
    {
        return $this->livreA;
    }

    public function setLivreA(?string $livreA): static
    {
        $this->livreA = $livreA;

        return $this;
    }
}
