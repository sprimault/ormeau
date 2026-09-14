<?php

// Généré par Ormeau depuis la base reference-hors-identifiant et réécrit à chaque génération : le code propre à Client va dans Client.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
#[ORM\UniqueConstraint(name: 'client_code_key', columns: ['code'])]
abstract class ClientBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    #[ORM\Column(name: 'code', type: 'string', length: 10, unique: true)]
    protected string $code;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getCode(): string
    {
        return $this->code;
    }

    public function setCode(string $code): static
    {
        $this->code = $code;

        return $this;
    }
}
