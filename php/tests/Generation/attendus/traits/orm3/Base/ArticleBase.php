<?php

// Généré par Ormeau et réécrit à chaque génération : le code propre à Article va dans Article.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class ArticleBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    #[ORM\Column(name: 'libelle', type: 'string', length: 120)]
    protected string $libelle;

    #[ORM\Column(name: 'cree_le', type: 'datetime_immutable')]
    protected \DateTimeImmutable $creeLe;

    #[ORM\Column(name: 'supprime_le', type: 'datetime_immutable', nullable: true)]
    protected ?\DateTimeImmutable $supprimeLe = null;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getLibelle(): string
    {
        return $this->libelle;
    }

    public function setLibelle(string $libelle): static
    {
        $this->libelle = $libelle;

        return $this;
    }

    public function getCreeLe(): \DateTimeImmutable
    {
        return $this->creeLe;
    }

    public function setCreeLe(\DateTimeImmutable $creeLe): static
    {
        $this->creeLe = $creeLe;

        return $this;
    }

    public function getSupprimeLe(): ?\DateTimeImmutable
    {
        return $this->supprimeLe;
    }

    public function setSupprimeLe(?\DateTimeImmutable $supprimeLe): static
    {
        $this->supprimeLe = $supprimeLe;

        return $this;
    }
}
