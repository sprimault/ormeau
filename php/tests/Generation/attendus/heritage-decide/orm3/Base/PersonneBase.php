<?php

// Généré par Ormeau et réécrit à chaque génération : le code propre à Personne va dans Personne.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Adresse;
use App\Entity\Prestataire;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\ORM\Mapping as ORM;

/** La colonne nature départage la hiérarchie : Doctrine l'écrit, elle n'a pas de propriété. */
#[ORM\MappedSuperclass]
abstract class PersonneBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    #[ORM\Column(name: 'nom', type: 'string', length: 80)]
    protected string $nom;

    #[ORM\OneToOne(targetEntity: Prestataire::class, mappedBy: 'personne')]
    protected ?Prestataire $prestataire = null;

    /** @var Collection<int, Adresse> */
    #[ORM\OneToMany(targetEntity: Adresse::class, mappedBy: 'personne')]
    protected Collection $adresse;

    public function __construct()
    {
        $this->adresse = new ArrayCollection();
    }

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getNom(): string
    {
        return $this->nom;
    }

    public function setNom(string $nom): static
    {
        $this->nom = $nom;

        return $this;
    }

    public function getPrestataire(): ?Prestataire
    {
        return $this->prestataire;
    }

    /** @return Collection<int, Adresse> */
    public function getAdresse(): Collection
    {
        return $this->adresse;
    }
}
