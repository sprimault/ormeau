<?php

// Généré par Ormeau depuis la base heritage-decide et réécrit à chaque génération : le code propre à Adresse va dans Adresse.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Personne;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class AdresseBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    /** Lecture seule : écrite par l'association personne. */
    #[ORM\Column(name: 'personne_id', type: 'integer', insertable: false, updatable: false)]
    protected int $personneId;

    #[ORM\Column(name: 'ville', type: 'string', length: 60)]
    protected string $ville;

    #[ORM\ManyToOne(targetEntity: Personne::class, inversedBy: 'adresse')]
    #[ORM\JoinColumn(name: 'personne_id', referencedColumnName: 'id', nullable: false)]
    protected Personne $personne;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getPersonneId(): int
    {
        return $this->personneId;
    }

    public function getVille(): string
    {
        return $this->ville;
    }

    public function setVille(string $ville): static
    {
        $this->ville = $ville;

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
