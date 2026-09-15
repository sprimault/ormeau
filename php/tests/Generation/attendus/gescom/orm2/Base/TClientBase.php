<?php

// Généré par Ormeau depuis la base gescom et réécrit à chaque génération : le code propre à TClient va dans TClient.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Enum\TClientCliStatut;
use App\Entity\TClientAdresse;
use App\Entity\TClientContact;
use App\Entity\TClientGrandCompte;
use App\Entity\TCommercial;
use App\Entity\TTag;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
#[ORM\Index(
    name: 'ix_cli_actifs',
    columns: ['cli_com_id'],
    options: ['where' => '((cli_statut)::text = \'ACTIF\'::text)'],
)]
#[ORM\Index(name: 'ix_cli_nom', columns: ['cli_nom'])]
#[ORM\Index(name: 'ix_cli_nom_prefixe', columns: ['cli_nom'])]
#[ORM\UniqueConstraint(name: 'uq_cli_siret', columns: ['cli_siret'])]
abstract class TClientBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'cli_id', type: 'integer')]
    protected ?int $cliId = null;

    #[ORM\Column(name: 'cli_nom', type: 'string', length: 120)]
    protected string $cliNom;

    /** Nul tant que la fiche n'est pas validée. */
    #[ORM\Column(
        name: 'cli_siret',
        type: 'string',
        length: 14,
        unique: true,
        nullable: true,
        options: ['comment' => 'Nul tant que la fiche n\'est pas validée'],
    )]
    protected ?string $cliSiret = null;

    #[ORM\Column(
        name: 'cli_statut',
        type: 'string',
        length: 20,
        enumType: TClientCliStatut::class,
        options: ['default' => 'ACTIF'],
    )]
    protected TClientCliStatut $cliStatut = TClientCliStatut::Actif;

    /** Lecture seule : écrite par l'association cliCom. */
    #[ORM\Column(name: 'cli_com_id', type: 'integer', nullable: true, insertable: false, updatable: false)]
    protected ?int $cliComId = null;

    #[ORM\Column(name: 'cli_ca_ttc', type: 'decimal', precision: 12, scale: 2, nullable: true)]
    protected ?string $cliCaTtc = null;

    #[ORM\Column(
        name: 'cli_ca_ht',
        type: 'decimal',
        precision: 12,
        scale: 2,
        nullable: true,
        insertable: false,
        updatable: false,
    )]
    protected ?string $cliCaHt = null;

    #[ORM\Column(name: 'created_at', type: 'datetimetz_immutable', options: ['default' => 'CURRENT_TIMESTAMP'])]
    protected \DateTimeImmutable $createdAt;

    #[ORM\Column(name: 'updated_at', type: 'datetimetz_immutable', nullable: true)]
    protected ?\DateTimeImmutable $updatedAt = null;

    #[ORM\ManyToOne(targetEntity: TCommercial::class, inversedBy: 'tclient')]
    #[ORM\JoinColumn(name: 'cli_com_id', referencedColumnName: 'com_id', onDelete: 'SET NULL')]
    protected ?TCommercial $cliCom = null;

    #[ORM\OneToOne(targetEntity: TClientAdresse::class, mappedBy: 'cli')]
    protected ?TClientAdresse $tclientAdresse = null;

    /** @var Collection<int, TClientContact> */
    #[ORM\OneToMany(targetEntity: TClientContact::class, mappedBy: 'cli')]
    protected Collection $tclientContact;

    #[ORM\OneToOne(targetEntity: TClientGrandCompte::class, mappedBy: 'cli')]
    protected ?TClientGrandCompte $tclientGrandCompte = null;

    /** @var Collection<int, TTag> */
    #[ORM\ManyToMany(targetEntity: TTag::class, inversedBy: 'tclient')]
    #[ORM\JoinTable(name: 't_client_tag', options: ['comment' => 'Étiquettes posées sur un client'])]
    #[ORM\JoinColumn(name: 'cli_id', referencedColumnName: 'cli_id', onDelete: 'CASCADE')]
    #[ORM\InverseJoinColumn(name: 'tag_id', referencedColumnName: 'tag_id')]
    protected Collection $ttag;

    public function __construct()
    {
        $this->tclientContact = new ArrayCollection();
        $this->ttag = new ArrayCollection();
    }

    public function getCliId(): ?int
    {
        return $this->cliId;
    }

    public function getCliNom(): string
    {
        return $this->cliNom;
    }

    public function setCliNom(string $cliNom): static
    {
        $this->cliNom = $cliNom;

        return $this;
    }

    public function getCliSiret(): ?string
    {
        return $this->cliSiret;
    }

    public function setCliSiret(?string $cliSiret): static
    {
        $this->cliSiret = $cliSiret;

        return $this;
    }

    public function getCliStatut(): TClientCliStatut
    {
        return $this->cliStatut;
    }

    public function setCliStatut(TClientCliStatut $cliStatut): static
    {
        $this->cliStatut = $cliStatut;

        return $this;
    }

    public function getCliComId(): ?int
    {
        return $this->cliComId;
    }

    public function getCliCaTtc(): ?string
    {
        return $this->cliCaTtc;
    }

    public function setCliCaTtc(?string $cliCaTtc): static
    {
        $this->cliCaTtc = $cliCaTtc;

        return $this;
    }

    public function getCliCaHt(): ?string
    {
        return $this->cliCaHt;
    }

    public function getCreatedAt(): \DateTimeImmutable
    {
        return $this->createdAt;
    }

    public function setCreatedAt(\DateTimeImmutable $createdAt): static
    {
        $this->createdAt = $createdAt;

        return $this;
    }

    public function getUpdatedAt(): ?\DateTimeImmutable
    {
        return $this->updatedAt;
    }

    public function setUpdatedAt(?\DateTimeImmutable $updatedAt): static
    {
        $this->updatedAt = $updatedAt;

        return $this;
    }

    public function getCliCom(): ?TCommercial
    {
        return $this->cliCom;
    }

    public function setCliCom(?TCommercial $cliCom): static
    {
        $this->cliCom = $cliCom;

        return $this;
    }

    public function getTclientAdresse(): ?TClientAdresse
    {
        return $this->tclientAdresse;
    }

    /** @return Collection<int, TClientContact> */
    public function getTclientContact(): Collection
    {
        return $this->tclientContact;
    }

    public function getTclientGrandCompte(): ?TClientGrandCompte
    {
        return $this->tclientGrandCompte;
    }

    /** @return Collection<int, TTag> */
    public function getTtag(): Collection
    {
        return $this->ttag;
    }

    public function addTtag(TTag $tTag): static
    {
        if (!$this->ttag->contains($tTag)) {
            $this->ttag->add($tTag);
        }

        return $this;
    }

    public function removeTtag(TTag $tTag): static
    {
        $this->ttag->removeElement($tTag);

        return $this;
    }
}
