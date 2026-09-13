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

    #[ORM\Column(name: 'etat', type: 'string', length: 20)]
    protected string $etat;

    #[ORM\Column(name: 'motif', type: 'string', length: 60, nullable: true)]
    protected ?string $motif = null;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getEtat(): string
    {
        return $this->etat;
    }

    public function setEtat(string $etat): static
    {
        $this->etat = $etat;

        return $this;
    }

    public function getMotif(): ?string
    {
        return $this->motif;
    }

    public function setMotif(?string $motif): static
    {
        $this->motif = $motif;

        return $this;
    }
}
