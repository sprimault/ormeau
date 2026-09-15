<?php

// Généré par Ormeau depuis la base gescom et réécrit à chaque génération : le code propre à TCategorie va dans TCategorie.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\TCategorie;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class TCategorieBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'cat_id', type: 'integer')]
    protected ?int $catId = null;

    #[ORM\Column(name: 'cat_libelle', type: 'string', length: 60)]
    protected string $catLibelle;

    /** Lecture seule : écrite par l'association tcategorie. */
    #[ORM\Column(name: 'cat_parent', type: 'integer', nullable: true, insertable: false, updatable: false)]
    protected ?int $catParent = null;

    #[ORM\ManyToOne(targetEntity: TCategorie::class, inversedBy: 'tcategorieTcategorie')]
    #[ORM\JoinColumn(name: 'cat_parent', referencedColumnName: 'cat_id')]
    protected ?TCategorie $tcategorie = null;

    /** @var Collection<int, TCategorie> */
    #[ORM\OneToMany(targetEntity: TCategorie::class, mappedBy: 'tcategorie')]
    protected Collection $tcategorieTcategorie;

    public function __construct()
    {
        $this->tcategorieTcategorie = new ArrayCollection();
    }

    public function getCatId(): ?int
    {
        return $this->catId;
    }

    public function getCatLibelle(): string
    {
        return $this->catLibelle;
    }

    public function setCatLibelle(string $catLibelle): static
    {
        $this->catLibelle = $catLibelle;

        return $this;
    }

    public function getCatParent(): ?int
    {
        return $this->catParent;
    }

    public function getTcategorie(): ?TCategorie
    {
        return $this->tcategorie;
    }

    public function setTcategorie(?TCategorie $tcategorie): static
    {
        $this->tcategorie = $tcategorie;

        return $this;
    }

    /** @return Collection<int, TCategorie> */
    public function getTcategorieTcategorie(): Collection
    {
        return $this->tcategorieTcategorie;
    }
}
