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

    #[ORM\Column(name: 'cmd_ref', type: 'guid')]
    protected string $cmdRef;

    #[ORM\Column(
        name: 'cmd_heure',
        type: 'time_immutable',
        nullable: true,
        options: ['default' => new \Doctrine\DBAL\Schema\DefaultExpression\CurrentTime()],
    )]
    protected ?\DateTimeImmutable $cmdHeure = null;

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

    public function getCmdRef(): string
    {
        return $this->cmdRef;
    }

    public function setCmdRef(string $cmdRef): static
    {
        $this->cmdRef = $cmdRef;

        return $this;
    }

    public function getCmdHeure(): ?\DateTimeImmutable
    {
        return $this->cmdHeure;
    }

    public function setCmdHeure(?\DateTimeImmutable $cmdHeure): static
    {
        $this->cmdHeure = $cmdHeure;

        return $this;
    }
}
