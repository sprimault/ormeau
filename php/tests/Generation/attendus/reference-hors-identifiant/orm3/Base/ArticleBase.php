<?php

// Généré par Ormeau depuis la base reference-hors-identifiant et réécrit à chaque génération : le code propre à Article va dans Article.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Ligne;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class ArticleBase
{
    #[ORM\Id]
    #[ORM\Column(name: 'code', type: 'string', length: 20)]
    protected string $code;

    #[ORM\Id]
    #[ORM\Column(name: 'version', type: 'integer')]
    protected int $version;

    /** @var Collection<int, Ligne> */
    #[ORM\OneToMany(targetEntity: Ligne::class, mappedBy: 'article')]
    protected Collection $ligne;

    public function __construct()
    {
        $this->ligne = new ArrayCollection();
    }

    public function getCode(): string
    {
        return $this->code;
    }

    public function setCode(string $code): static
    {
        $this->code = $code;

        return $this;
    }

    public function getVersion(): int
    {
        return $this->version;
    }

    public function setVersion(int $version): static
    {
        $this->version = $version;

        return $this;
    }

    /** @return Collection<int, Ligne> */
    public function getLigne(): Collection
    {
        return $this->ligne;
    }
}
