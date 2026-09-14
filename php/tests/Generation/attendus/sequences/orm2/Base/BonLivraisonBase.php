<?php

// Généré par Ormeau depuis la base sequences et réécrit à chaque génération : le code propre à BonLivraison va dans BonLivraison.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class BonLivraisonBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'SEQUENCE')]
    #[ORM\SequenceGenerator(sequenceName: '"Compta"."Bon_Livraison_Id_seq"')]
    #[ORM\Column(name: '`Id`', type: 'bigint')]
    protected ?string $id = null;

    #[ORM\Column(name: 'numero', type: 'string', length: 20)]
    protected string $numero;

    public function getId(): ?string
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
}
