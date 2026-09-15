<?php

// Généré par Ormeau depuis la base gescom et réécrit à chaque génération : le code propre à TClientContact va dans TClientContact.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\TClient;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class TClientContactBase
{
    #[ORM\Id]
    #[ORM\Column(name: 'ctc_id', type: 'integer')]
    protected int $ctcId;

    #[ORM\Column(name: 'role', type: 'string', length: 30)]
    protected string $role;

    /** Fait partie de l'identifiant : à renseigner avant persist(). */
    #[ORM\Id]
    #[ORM\ManyToOne(targetEntity: TClient::class, inversedBy: 'tclientContact')]
    #[ORM\JoinColumn(name: 'cli_id', referencedColumnName: 'cli_id', nullable: false)]
    protected TClient $cli;

    public function getCtcId(): int
    {
        return $this->ctcId;
    }

    public function setCtcId(int $ctcId): static
    {
        $this->ctcId = $ctcId;

        return $this;
    }

    public function getRole(): string
    {
        return $this->role;
    }

    public function setRole(string $role): static
    {
        $this->role = $role;

        return $this;
    }

    public function getCli(): TClient
    {
        return $this->cli;
    }

    public function setCli(TClient $cli): static
    {
        $this->cli = $cli;

        return $this;
    }
}
