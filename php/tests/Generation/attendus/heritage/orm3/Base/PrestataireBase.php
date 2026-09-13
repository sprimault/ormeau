<?php

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Personne;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class PrestataireBase
{
    #[ORM\Column(name: 'societe', type: 'string', length: 120)]
    protected string $societe;

    #[ORM\Id]
    #[ORM\OneToOne(targetEntity: Personne::class, inversedBy: 'prestataire')]
    #[ORM\JoinColumn(name: 'id', referencedColumnName: 'id', nullable: false, onDelete: 'CASCADE')]
    protected Personne $personne;

    public function getSociete(): string
    {
        return $this->societe;
    }

    public function setSociete(string $societe): static
    {
        $this->societe = $societe;

        return $this;
    }

    public function getPersonne(): Personne
    {
        return $this->personne;
    }

    public function setPersonne(Personne $personne): static
    {
        $this->personne = $personne;

        return $this;
    }
}
