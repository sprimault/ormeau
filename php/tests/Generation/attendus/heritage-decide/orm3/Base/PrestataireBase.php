<?php

// Généré par Ormeau depuis la base heritage-decide et réécrit à chaque génération : le code propre à Prestataire va dans Prestataire.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Personne;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class PrestataireBase
{
    /** Fait partie de l'identifiant : à renseigner avant persist(). */
    #[ORM\Id]
    #[ORM\OneToOne(targetEntity: Personne::class, inversedBy: 'prestataire')]
    #[ORM\JoinColumn(name: 'id', referencedColumnName: 'id', onDelete: 'CASCADE')]
    protected Personne $personne;

    #[ORM\Column(name: 'societe', type: 'string', length: 120)]
    protected string $societe;

    public function getPersonne(): Personne
    {
        return $this->personne;
    }

    public function setPersonne(Personne $personne): static
    {
        $this->personne = $personne;

        return $this;
    }

    public function getSociete(): string
    {
        return $this->societe;
    }

    public function setSociete(string $societe): static
    {
        $this->societe = $societe;

        return $this;
    }
}
