<?php

// Généré par Ormeau depuis la base gescom et réécrit à chaque génération : le code propre à TPays va dans TPays.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\TCommande;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class TPaysBase
{
    #[ORM\Id]
    #[ORM\Column(name: 'pay_code', type: 'string', length: 2, options: ['fixed' => true])]
    protected string $payCode;

    #[ORM\Column(name: 'pay_libelle', type: 'string', length: 60, options: ['collation' => 'fr-FR-x-icu'])]
    protected string $payLibelle;

    /** @var Collection<int, TCommande> */
    #[ORM\OneToMany(targetEntity: TCommande::class, mappedBy: 'cmdPay')]
    protected Collection $tcommande;

    public function __construct()
    {
        $this->tcommande = new ArrayCollection();
    }

    public function getPayCode(): string
    {
        return $this->payCode;
    }

    public function setPayCode(string $payCode): static
    {
        $this->payCode = $payCode;

        return $this;
    }

    public function getPayLibelle(): string
    {
        return $this->payLibelle;
    }

    public function setPayLibelle(string $payLibelle): static
    {
        $this->payLibelle = $payLibelle;

        return $this;
    }

    /** @return Collection<int, TCommande> */
    public function getTcommande(): Collection
    {
        return $this->tcommande;
    }
}
