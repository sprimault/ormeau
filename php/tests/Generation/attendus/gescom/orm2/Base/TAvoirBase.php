<?php

// Généré par Ormeau depuis la base gescom et réécrit à chaque génération : le code propre à TAvoir va dans TAvoir.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class TAvoirBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'SEQUENCE')]
    #[ORM\SequenceGenerator(sequenceName: 'gescom.t_avoir_avo_id_seq')]
    #[ORM\Column(name: 'avo_id', type: 'integer')]
    protected ?int $avoId = null;

    #[ORM\Column(name: 'avo_fac_id', type: 'integer')]
    protected int $avoFacId;

    public function getAvoId(): ?int
    {
        return $this->avoId;
    }

    public function getAvoFacId(): int
    {
        return $this->avoFacId;
    }

    public function setAvoFacId(int $avoFacId): static
    {
        $this->avoFacId = $avoFacId;

        return $this;
    }
}
