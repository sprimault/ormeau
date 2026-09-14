<?php

// Généré par Ormeau depuis la base schemas-cites et réécrit à chaque génération : le code propre à Tag va dans Tag.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class TagBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    public function getId(): ?int
    {
        return $this->id;
    }
}
