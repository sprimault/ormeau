<?php

// Généré par Ormeau depuis la base identifiants-penibles et réécrit à chaque génération : le code propre à CommandesClients va dans CommandesClients.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class CommandesClientsBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: '`N° Commande`', type: 'integer')]
    protected ?int $nCommande = null;

    #[ORM\Column(name: '`Date de création`', type: 'datetime_immutable')]
    protected \DateTimeImmutable $dateDeCréation;

    #[ORM\Column(name: '`CA TTC`', type: 'decimal', precision: 12, scale: 2, nullable: true)]
    protected ?string $caTtc = null;

    public function getNCommande(): ?int
    {
        return $this->nCommande;
    }

    public function getDateDeCréation(): \DateTimeImmutable
    {
        return $this->dateDeCréation;
    }

    public function setDateDeCréation(\DateTimeImmutable $dateDeCréation): static
    {
        $this->dateDeCréation = $dateDeCréation;

        return $this;
    }

    public function getCaTtc(): ?string
    {
        return $this->caTtc;
    }

    public function setCaTtc(?string $caTtc): static
    {
        $this->caTtc = $caTtc;

        return $this;
    }
}
