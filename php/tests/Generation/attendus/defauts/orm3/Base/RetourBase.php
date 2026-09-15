<?php

// Généré par Ormeau depuis la base defauts et réécrit à chaque génération : le code propre à Retour va dans Retour.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Trait\Horodatage;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class RetourBase
{
    use Horodatage;

    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    public function getId(): ?int
    {
        return $this->id;
    }
}
