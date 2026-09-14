<?php

// Généré par Ormeau depuis la base reference-hors-identifiant et réécrit à chaque génération : le code propre à Produit va dans Produit.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\ProduitTag;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class ProduitBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    /** @var Collection<int, ProduitTag> */
    #[ORM\OneToMany(targetEntity: ProduitTag::class, mappedBy: 'produit')]
    protected Collection $produitTag;

    public function __construct()
    {
        $this->produitTag = new ArrayCollection();
    }

    public function getId(): ?int
    {
        return $this->id;
    }

    /** @return Collection<int, ProduitTag> */
    public function getProduitTag(): Collection
    {
        return $this->produitTag;
    }
}
