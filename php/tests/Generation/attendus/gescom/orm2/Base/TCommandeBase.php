<?php

// Généré par Ormeau depuis la base gescom et réécrit à chaque génération : le code propre à TCommande va dans TCommande.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Enum\Canal;
use App\Entity\TPays;
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

    #[ORM\Column(name: 'cmd_heure', type: 'time_immutable', nullable: true, options: ['default' => 'CURRENT_TIME'])]
    protected ?\DateTimeImmutable $cmdHeure = null;

    /** Lecture seule : écrite par l'association cmdPay. */
    #[ORM\Column(
        name: 'cmd_pay_code',
        type: 'string',
        length: 2,
        nullable: true,
        insertable: false,
        updatable: false,
        options: ['fixed' => true],
    )]
    protected ?string $cmdPayCode = null;

    /** @var array<mixed>|null */
    #[ORM\Column(name: 'cmd_options', type: 'json', nullable: true, options: ['jsonb' => true])]
    protected ?array $cmdOptions = null;

    #[ORM\Column(name: 'cmd_notes', type: 'text', nullable: true)]
    protected ?string $cmdNotes = null;

    #[ORM\ManyToOne(targetEntity: TPays::class, inversedBy: 'tcommande')]
    #[ORM\JoinColumn(name: 'cmd_pay_code', referencedColumnName: 'pay_code')]
    protected ?TPays $cmdPay = null;

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

    public function getCmdPayCode(): ?string
    {
        return $this->cmdPayCode;
    }

    /** @return array<mixed>|null */
    public function getCmdOptions(): ?array
    {
        return $this->cmdOptions;
    }

    /** @param array<mixed>|null $cmdOptions */
    public function setCmdOptions(?array $cmdOptions): static
    {
        $this->cmdOptions = $cmdOptions;

        return $this;
    }

    public function getCmdNotes(): ?string
    {
        return $this->cmdNotes;
    }

    public function setCmdNotes(?string $cmdNotes): static
    {
        $this->cmdNotes = $cmdNotes;

        return $this;
    }

    public function getCmdPay(): ?TPays
    {
        return $this->cmdPay;
    }

    public function setCmdPay(?TPays $cmdPay): static
    {
        $this->cmdPay = $cmdPay;

        return $this;
    }
}
