<?php

// Généré par Ormeau depuis la base reference-hors-identifiant et réécrit à chaque génération : le code propre à FicheClient va dans FicheClient.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class FicheClientBase
{
    #[ORM\Id]
    #[ORM\Column(name: 'client_code', type: 'string', length: 10)]
    protected string $clientCode;

    #[ORM\Column(name: 'note', type: 'text', nullable: true)]
    protected ?string $note = null;

    public function getClientCode(): string
    {
        return $this->clientCode;
    }

    public function setClientCode(string $clientCode): static
    {
        $this->clientCode = $clientCode;

        return $this;
    }

    public function getNote(): ?string
    {
        return $this->note;
    }

    public function setNote(?string $note): static
    {
        $this->note = $note;

        return $this;
    }
}
