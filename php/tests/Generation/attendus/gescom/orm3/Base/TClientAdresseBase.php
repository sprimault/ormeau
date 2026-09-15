<?php

// Généré par Ormeau depuis la base gescom et réécrit à chaque génération : le code propre à TClientAdresse va dans TClientAdresse.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\TClient;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
#[ORM\UniqueConstraint(name: 'uq_adresse_client', columns: ['cli_id'])]
abstract class TClientAdresseBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'adr_id', type: 'integer')]
    protected ?int $adrId = null;

    /** Lecture seule : écrite par l'association cli. */
    #[ORM\Column(name: 'cli_id', type: 'integer', insertable: false, updatable: false)]
    protected int $cliId;

    #[ORM\Column(name: 'adr_rue', type: 'string', length: 200)]
    protected string $adrRue;

    #[ORM\OneToOne(targetEntity: TClient::class, inversedBy: 'tclientAdresse')]
    #[ORM\JoinColumn(name: 'cli_id', referencedColumnName: 'cli_id', nullable: false, onDelete: 'CASCADE')]
    protected TClient $cli;

    public function getAdrId(): ?int
    {
        return $this->adrId;
    }

    public function getCliId(): int
    {
        return $this->cliId;
    }

    public function getAdrRue(): string
    {
        return $this->adrRue;
    }

    public function setAdrRue(string $adrRue): static
    {
        $this->adrRue = $adrRue;

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
