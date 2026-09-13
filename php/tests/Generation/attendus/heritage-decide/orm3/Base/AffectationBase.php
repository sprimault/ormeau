<?php

// Généré par Ormeau et réécrit à chaque génération : le code propre à Affectation va dans Affectation.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Salarie;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class AffectationBase
{
    #[ORM\Id]
    #[ORM\Column(name: 'debut', type: 'date_immutable')]
    protected \DateTimeImmutable $debut;

    #[ORM\Column(name: 'service', type: 'string', length: 60)]
    protected string $service;

    /** Fait partie de l'identifiant : à renseigner avant persist(). */
    #[ORM\Id]
    #[ORM\ManyToOne(targetEntity: Salarie::class, inversedBy: 'affectation')]
    #[ORM\JoinColumn(name: 'salarie_id', referencedColumnName: 'id', nullable: false)]
    protected Salarie $salarie;

    public function getDebut(): \DateTimeImmutable
    {
        return $this->debut;
    }

    public function setDebut(\DateTimeImmutable $debut): static
    {
        $this->debut = $debut;

        return $this;
    }

    public function getService(): string
    {
        return $this->service;
    }

    public function setService(string $service): static
    {
        $this->service = $service;

        return $this;
    }

    public function getSalarie(): Salarie
    {
        return $this->salarie;
    }

    public function setSalarie(Salarie $salarie): static
    {
        $this->salarie = $salarie;

        return $this;
    }
}
