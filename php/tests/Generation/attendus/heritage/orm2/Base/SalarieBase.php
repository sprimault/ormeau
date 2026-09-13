<?php

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Personne;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class SalarieBase
{
    #[ORM\Column(name: 'matricule', type: 'string', length: 20)]
    protected string $matricule;

    #[ORM\Id]
    #[ORM\OneToOne(targetEntity: Personne::class, inversedBy: 'salarie')]
    #[ORM\JoinColumn(name: 'id', referencedColumnName: 'id', nullable: false, onDelete: 'CASCADE')]
    protected Personne $personne;

    public function getMatricule(): string
    {
        return $this->matricule;
    }

    public function setMatricule(string $matricule): static
    {
        $this->matricule = $matricule;

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
