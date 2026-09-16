<?php

// Généré par Ormeau depuis la base fuseau et réécrit à chaque génération : le code propre à Evenement va dans Evenement.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class EvenementBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    /** Précision par défaut : Doctrine n'en lit que six. */
    #[ORM\Column(
        name: 'emis_le',
        type: 'datetimetz_immutable',
        options: ['comment' => 'Précision par défaut : Doctrine n\'en lit que six'],
    )]
    protected \DateTimeImmutable $emisLe;

    #[ORM\Column(name: 'valide_le', type: 'datetimetz_immutable', nullable: true)]
    protected ?\DateTimeImmutable $valideLe = null;

    /** Sans fuseau : Doctrine relit sept décimales. */
    #[ORM\Column(
        name: 'saisi_le',
        type: 'datetime_immutable',
        nullable: true,
        options: ['comment' => 'Sans fuseau : Doctrine relit sept décimales'],
    )]
    protected ?\DateTimeImmutable $saisiLe = null;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getEmisLe(): \DateTimeImmutable
    {
        return $this->emisLe;
    }

    public function setEmisLe(\DateTimeImmutable $emisLe): static
    {
        $this->emisLe = $emisLe;

        return $this;
    }

    public function getValideLe(): ?\DateTimeImmutable
    {
        return $this->valideLe;
    }

    public function setValideLe(?\DateTimeImmutable $valideLe): static
    {
        $this->valideLe = $valideLe;

        return $this;
    }

    public function getSaisiLe(): ?\DateTimeImmutable
    {
        return $this->saisiLe;
    }

    public function setSaisiLe(?\DateTimeImmutable $saisiLe): static
    {
        $this->saisiLe = $saisiLe;

        return $this;
    }
}
