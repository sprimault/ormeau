<?php

// Généré par Ormeau depuis la base gescom et réécrit à chaque génération : le code propre à TClientGrandCompte va dans TClientGrandCompte.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\TClient;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class TClientGrandCompteBase
{
    /** Fait partie de l'identifiant : à renseigner avant persist(). */
    #[ORM\Id]
    #[ORM\OneToOne(targetEntity: TClient::class, inversedBy: 'tclientGrandCompte')]
    #[ORM\JoinColumn(name: 'cli_id', referencedColumnName: 'cli_id')]
    protected TClient $cli;

    #[ORM\Column(name: 'remise_taux', type: 'decimal', precision: 4, scale: 2)]
    protected string $remiseTaux;

    public function getCli(): TClient
    {
        return $this->cli;
    }

    public function setCli(TClient $cli): static
    {
        $this->cli = $cli;

        return $this;
    }

    public function getRemiseTaux(): string
    {
        return $this->remiseTaux;
    }

    public function setRemiseTaux(string $remiseTaux): static
    {
        $this->remiseTaux = $remiseTaux;

        return $this;
    }
}
