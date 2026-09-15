<?php

// Généré par Ormeau depuis la base gescom et réécrit à chaque génération : le code propre à TTag va dans TTag.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\TClient;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class TTagBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'tag_id', type: 'integer')]
    protected ?int $tagId = null;

    #[ORM\Column(name: 'tag_libelle', type: 'string', length: 40, options: ['collation' => 'C'])]
    protected string $tagLibelle;

    /** @var Collection<int, TClient> */
    #[ORM\ManyToMany(targetEntity: TClient::class, mappedBy: 'ttag')]
    protected Collection $tclient;

    public function __construct()
    {
        $this->tclient = new ArrayCollection();
    }

    public function getTagId(): ?int
    {
        return $this->tagId;
    }

    public function getTagLibelle(): string
    {
        return $this->tagLibelle;
    }

    public function setTagLibelle(string $tagLibelle): static
    {
        $this->tagLibelle = $tagLibelle;

        return $this;
    }

    /** @return Collection<int, TClient> */
    public function getTclient(): Collection
    {
        return $this->tclient;
    }
}
