<?php

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Article;
use App\Entity\Commande;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class LigneCommandeBase
{
    #[ORM\Column(name: 'quantite', type: 'integer')]
    protected int $quantite;

    #[ORM\Id]
    #[ORM\ManyToOne(targetEntity: Commande::class, inversedBy: 'ligneCommande')]
    #[ORM\JoinColumn(name: 'commande_id', referencedColumnName: 'id', nullable: false)]
    protected Commande $commande;

    #[ORM\Id]
    #[ORM\ManyToOne(targetEntity: Article::class, inversedBy: 'ligneCommande')]
    #[ORM\JoinColumn(name: 'article_id', referencedColumnName: 'id', nullable: false)]
    protected Article $article;

    public function getQuantite(): int
    {
        return $this->quantite;
    }

    public function setQuantite(int $quantite): static
    {
        $this->quantite = $quantite;

        return $this;
    }

    public function getCommande(): Commande
    {
        return $this->commande;
    }

    public function setCommande(Commande $commande): static
    {
        $this->commande = $commande;

        return $this;
    }

    public function getArticle(): Article
    {
        return $this->article;
    }

    public function setArticle(Article $article): static
    {
        $this->article = $article;

        return $this;
    }
}
