<?php

// Généré par Ormeau depuis la base gescom et réécrit à chaque génération : le code propre à TCommercial va dans TCommercial.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\TClient;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
#[ORM\UniqueConstraint(name: 'uq_com_email_actif', columns: ['com_email'], options: ['where' => 'com_actif'])]
abstract class TCommercialBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'com_id', type: 'integer')]
    protected ?int $comId = null;

    #[ORM\Column(name: 'com_nom', type: 'string', length: 80)]
    protected string $comNom;

    #[ORM\Column(name: 'com_actif', type: 'boolean', options: ['default' => true])]
    protected bool $comActif = true;

    #[ORM\Column(name: 'com_email', type: 'string', length: 120, nullable: true)]
    protected ?string $comEmail = null;

    #[ORM\Column(name: 'com_empreinte', type: 'blob', nullable: true)]
    protected mixed $comEmpreinte = null;

    /** @var Collection<int, TClient> */
    #[ORM\OneToMany(targetEntity: TClient::class, mappedBy: 'cliCom')]
    protected Collection $tclient;

    public function __construct()
    {
        $this->tclient = new ArrayCollection();
    }

    public function getComId(): ?int
    {
        return $this->comId;
    }

    public function getComNom(): string
    {
        return $this->comNom;
    }

    public function setComNom(string $comNom): static
    {
        $this->comNom = $comNom;

        return $this;
    }

    public function isComActif(): bool
    {
        return $this->comActif;
    }

    public function setComActif(bool $comActif): static
    {
        $this->comActif = $comActif;

        return $this;
    }

    public function getComEmail(): ?string
    {
        return $this->comEmail;
    }

    public function setComEmail(?string $comEmail): static
    {
        $this->comEmail = $comEmail;

        return $this;
    }

    public function getComEmpreinte(): mixed
    {
        return $this->comEmpreinte;
    }

    public function setComEmpreinte(mixed $comEmpreinte): static
    {
        $this->comEmpreinte = $comEmpreinte;

        return $this;
    }

    /** @return Collection<int, TClient> */
    public function getTclient(): Collection
    {
        return $this->tclient;
    }
}
