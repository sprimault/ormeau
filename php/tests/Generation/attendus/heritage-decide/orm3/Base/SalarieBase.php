<?php

// Généré par Ormeau depuis la base heritage-decide et réécrit à chaque génération : le code propre à Salarie va dans Salarie.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Affectation;
use App\Entity\Personne;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class SalarieBase extends Personne
{
    #[ORM\Column(name: 'matricule', type: 'string', length: 20)]
    protected string $matricule;

    /** @var Collection<int, Affectation> */
    #[ORM\OneToMany(targetEntity: Affectation::class, mappedBy: 'salarie')]
    protected Collection $affectation;

    public function __construct()
    {
        parent::__construct();
        $this->affectation = new ArrayCollection();
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

    /** @return Collection<int, Affectation> */
    public function getAffectation(): Collection
    {
        return $this->affectation;
    }
}
