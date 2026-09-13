<?php

// Généré par Ormeau et réécrit à chaque génération : le code propre à Majuscule va dans Majuscule.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class MajusculeBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'SEQUENCE')]
    #[ORM\SequenceGenerator(sequenceName: '"Majuscule_Id_seq"')]
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
