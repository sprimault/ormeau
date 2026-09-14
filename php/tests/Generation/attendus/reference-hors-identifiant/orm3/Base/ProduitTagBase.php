<?php

// Généré par Ormeau depuis la base reference-hors-identifiant et réécrit à chaque génération : le code propre à ProduitTag va dans ProduitTag.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Produit;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class ProduitTagBase
{
    #[ORM\Id]
    #[ORM\Column(name: 'tag_libelle', type: 'string', length: 30)]
    protected string $tagLibelle;

    /** Fait partie de l'identifiant : à renseigner avant persist(). */
    #[ORM\Id]
    #[ORM\ManyToOne(targetEntity: Produit::class, inversedBy: 'produitTag')]
    #[ORM\JoinColumn(name: 'produit_id', referencedColumnName: 'id', nullable: false)]
    protected Produit $produit;

    public function getTagLibelle(): string
    {
        return $this->tagLibelle;
    }

    public function setTagLibelle(string $tagLibelle): static
    {
        $this->tagLibelle = $tagLibelle;

        return $this;
    }

    public function getProduit(): Produit
    {
        return $this->produit;
    }

    public function setProduit(Produit $produit): static
    {
        $this->produit = $produit;

        return $this;
    }
}
