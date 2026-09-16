<?php

// Généré par Ormeau depuis la base identifiants-penibles et réécrit à chaque génération : le code propre à CatégoriesDesNoms va dans CatégoriesDesNoms.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
#[ORM\Index(name: 'IX_libelle_simple', columns: ['`Libellé`'])]
abstract class CatégoriesDesNomsBase
{
    #[ORM\Id]
    #[ORM\Column(name: '`Libellé`', type: 'string', length: 120)]
    protected string $libellé;

    public function getLibellé(): string
    {
        return $this->libellé;
    }

    public function setLibellé(string $libellé): static
    {
        $this->libellé = $libellé;

        return $this;
    }
}
