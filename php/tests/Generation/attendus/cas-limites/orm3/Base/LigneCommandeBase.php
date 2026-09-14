<?php

// Généré par Ormeau depuis la base cas-limites et réécrit à chaque génération : le code propre à LigneCommande va dans LigneCommande.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class LigneCommandeBase
{
    #[ORM\Id]
    #[ORM\Column(name: 'commande_id', type: 'integer')]
    protected int $commandeId;

    #[ORM\Id]
    #[ORM\Column(name: 'article_id', type: 'integer')]
    protected int $articleId;

    #[ORM\Column(name: 'quantite', type: 'integer', options: ['default' => 1])]
    protected int $quantite = 1;

    public function getCommandeId(): int
    {
        return $this->commandeId;
    }

    public function setCommandeId(int $commandeId): static
    {
        $this->commandeId = $commandeId;

        return $this;
    }

    public function getArticleId(): int
    {
        return $this->articleId;
    }

    public function setArticleId(int $articleId): static
    {
        $this->articleId = $articleId;

        return $this;
    }

    public function getQuantite(): int
    {
        return $this->quantite;
    }

    public function setQuantite(int $quantite): static
    {
        $this->quantite = $quantite;

        return $this;
    }
}
