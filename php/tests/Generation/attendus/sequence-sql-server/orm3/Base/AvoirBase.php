<?php

// Généré par Ormeau depuis la base sequence-sql-server et réécrit à chaque génération : le code propre à Avoir va dans Avoir.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class AvoirBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'CUSTOM')]
    #[ORM\CustomIdGenerator(class: Generateur\AvoirGenerateur::class)]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    #[ORM\Column(name: 'montant', type: 'decimal', precision: 12, scale: 2)]
    protected string $montant;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getMontant(): string
    {
        return $this->montant;
    }

    public function setMontant(string $montant): static
    {
        $this->montant = $montant;

        return $this;
    }
}
