<?php

// Généré par Ormeau et réécrit à chaque génération : le code propre à Commande va dans Commande.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Ligne;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class CommandeBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    #[ORM\Column(name: 'numero', type: 'string', length: 20)]
    protected string $numero;

    /** @var Collection<int, Ligne> */
    #[ORM\OneToMany(targetEntity: Ligne::class, mappedBy: 'commande')]
    protected Collection $ligne;

    public function __construct()
    {
        $this->ligne = new ArrayCollection();
    }

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getNumero(): string
    {
        return $this->numero;
    }

    public function setNumero(string $numero): static
    {
        $this->numero = $numero;

        return $this;
    }

    /** @return Collection<int, Ligne> */
    public function getLigne(): Collection
    {
        return $this->ligne;
    }
}
