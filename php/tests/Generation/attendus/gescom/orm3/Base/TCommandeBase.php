<?php

// Généré par Ormeau depuis la base gescom et réécrit à chaque génération : le code propre à TCommande va dans TCommande.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Enum\Canal;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class TCommandeBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'cmd_id', type: 'integer')]
    protected ?int $cmdId = null;

    #[ORM\Column(name: 'cmd_canal', type: 'string', enumType: Canal::class)]
    protected Canal $cmdCanal;

    #[ORM\Column(name: 'cmd_etiquettes', type: 'string', nullable: true)]
    protected ?string $cmdEtiquettes = null;

    public function getCmdId(): ?int
    {
        return $this->cmdId;
    }

    public function getCmdCanal(): Canal
    {
        return $this->cmdCanal;
    }

    public function setCmdCanal(Canal $cmdCanal): static
    {
        $this->cmdCanal = $cmdCanal;

        return $this;
    }

    public function getCmdEtiquettes(): ?string
    {
        return $this->cmdEtiquettes;
    }

    public function setCmdEtiquettes(?string $cmdEtiquettes): static
    {
        $this->cmdEtiquettes = $cmdEtiquettes;

        return $this;
    }
}
