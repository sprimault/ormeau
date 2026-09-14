<?php

// Généré par Ormeau depuis la base heritage et réécrit à chaque génération : le code propre à Personne va dans Personne.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Prestataire;
use App\Entity\Salarie;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class PersonneBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    #[ORM\Column(name: 'nom', type: 'string', length: 80)]
    protected string $nom;

    #[ORM\OneToOne(targetEntity: Salarie::class, mappedBy: 'personne')]
    protected ?Salarie $salarie = null;

    #[ORM\OneToOne(targetEntity: Prestataire::class, mappedBy: 'personne')]
    protected ?Prestataire $prestataire = null;

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

    public function getSalarie(): ?Salarie
    {
        return $this->salarie;
    }

    public function getPrestataire(): ?Prestataire
    {
        return $this->prestataire;
    }
}
