<?php

// Généré par Ormeau depuis la base identifiants-penibles et réécrit à chaque génération : le code propre à LignesDeCommande va dans LignesDeCommande.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class LignesDeCommandeBase
{
    #[ORM\Id]
    #[ORM\Column(name: '`N° Commande`', type: 'integer')]
    protected int $nCommande;

    #[ORM\Id]
    #[ORM\Column(name: '`Quantité`', type: 'integer')]
    protected int $quantité;

    public function getNCommande(): int
    {
        return $this->nCommande;
    }

    public function setNCommande(int $nCommande): static
    {
        $this->nCommande = $nCommande;

        return $this;
    }

    public function getQuantité(): int
    {
        return $this->quantité;
    }

    public function setQuantité(int $quantité): static
    {
        $this->quantité = $quantité;

        return $this;
    }
}
