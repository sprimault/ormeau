<?php

// Généré par Ormeau depuis la base sequences et réécrit à chaque génération : le code propre à Bloc va dans Bloc.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class BlocBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'SEQUENCE')]
    #[ORM\SequenceGenerator(sequenceName: 'bloc_id_seq', initialValue: 5)]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    #[ORM\Column(name: 'numero', type: 'string', length: 20)]
    protected string $numero;

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
}
