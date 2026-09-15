<?php

// Généré par Ormeau depuis la base heritage et réécrit à chaque génération : le code propre à Salarie va dans Salarie.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Personne;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class SalarieBase
{
    /** Fait partie de l'identifiant : à renseigner avant persist(). */
    #[ORM\Id]
    #[ORM\OneToOne(targetEntity: Personne::class, inversedBy: 'salarie')]
    #[ORM\JoinColumn(name: 'id', referencedColumnName: 'id', nullable: false, onDelete: 'CASCADE')]
    protected Personne $personne;

    #[ORM\Column(name: 'matricule', type: 'string', length: 20)]
    protected string $matricule;

    public function getPersonne(): Personne
    {
        return $this->personne;
    }

    public function setPersonne(Personne $personne): static
    {
        $this->personne = $personne;

        return $this;
    }

    public function getMatricule(): string
    {
        return $this->matricule;
    }

    public function setMatricule(string $matricule): static
    {
        $this->matricule = $matricule;

        return $this;
    }
}
