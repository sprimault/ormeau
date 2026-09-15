<?php

// Généré par Ormeau depuis la base defauts et réécrit à chaque génération : le code propre à Commande va dans Commande.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Trait\CreatedAt;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class CommandeBase
{
    use CreatedAt;

    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    public function getId(): ?int
    {
        return $this->id;
    }
}
