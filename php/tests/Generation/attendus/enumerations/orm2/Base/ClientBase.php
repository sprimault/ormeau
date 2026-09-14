<?php

// Généré par Ormeau depuis la base enumerations et réécrit à chaque génération : le code propre à Client va dans Client.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Enum\ClientActif;
use App\Entity\Enum\ClientCanal;
use App\Entity\Enum\StatutClient;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class ClientBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    /** Type énuméré natif : le cas le plus sûr. */
    #[ORM\Column(
        name: 'statut',
        type: 'string',
        enumType: StatutClient::class,
        options: ['comment' => 'Type énuméré natif : le cas le plus sûr'],
    )]
    protected StatutClient $statut;

    /** Valeurs fermées par un CHECK. */
    #[ORM\Column(
        name: 'canal',
        type: 'string',
        length: 20,
        enumType: ClientCanal::class,
        options: ['comment' => 'Valeurs fermées par un CHECK'],
    )]
    protected ClientCanal $canal;

    /** O/N : des cas nommés O et N n'apprendraient rien. */
    #[ORM\Column(
        name: 'actif',
        type: 'string',
        length: 1,
        enumType: ClientActif::class,
        options: ['comment' => 'O/N : des cas nommés O et N n\'apprendraient rien'],
    )]
    protected ClientActif $actif;

    #[ORM\Column(name: 'solde', type: 'decimal', precision: 12, scale: 2)]
    protected string $solde;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getStatut(): StatutClient
    {
        return $this->statut;
    }

    public function setStatut(StatutClient $statut): static
    {
        $this->statut = $statut;

        return $this;
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

    public function getActif(): ClientActif
    {
        return $this->actif;
    }

    public function setActif(ClientActif $actif): static
    {
        $this->actif = $actif;

        return $this;
    }

    public function getSolde(): string
    {
        return $this->solde;
    }

    public function setSolde(string $solde): static
    {
        $this->solde = $solde;

        return $this;
    }
}
