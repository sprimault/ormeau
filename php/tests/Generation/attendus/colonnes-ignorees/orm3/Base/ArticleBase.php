<?php

// Généré par Ormeau et réécrit à chaque génération : le code propre à Article va dans Article.php.

declare(strict_types=1);

namespace App\Entity\Base;

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

    #[ORM\Column(name: 'libelle', type: 'string', length: 80)]
    protected string $libelle;

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

    public function getLibelle(): string
    {
        return $this->libelle;
    }

    public function setLibelle(string $libelle): static
    {
        $this->libelle = $libelle;

        return $this;
    }
}
