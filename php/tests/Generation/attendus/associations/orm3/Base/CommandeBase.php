<?php

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Article;
use App\Entity\Client;
use App\Entity\Commande;
use App\Entity\LigneCommande;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class CommandeBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    #[ORM\Column(name: 'client_id', type: 'integer', insertable: false, updatable: false)]
    protected int $clientId;

    #[ORM\Column(name: 'commande_remplacee_id', type: 'integer', nullable: true, insertable: false, updatable: false)]
    protected ?int $commandeRemplaceeId = null;

    #[ORM\Column(name: 'total_ht', type: 'decimal', precision: 12, scale: 2)]
    protected string $totalHt;

    #[ORM\ManyToOne(targetEntity: Client::class, inversedBy: 'commande')]
    #[ORM\JoinColumn(name: 'client_id', referencedColumnName: 'id', nullable: false, onDelete: 'RESTRICT')]
    protected Client $client;

    #[ORM\ManyToOne(targetEntity: Commande::class, inversedBy: 'commande')]
    #[ORM\JoinColumn(name: 'commande_remplacee_id', referencedColumnName: 'id', onDelete: 'SET NULL')]
    protected ?Commande $commandeRemplacee = null;

    /** @var Collection<int, Commande> */
    #[ORM\OneToMany(targetEntity: Commande::class, mappedBy: 'commandeRemplacee')]
    protected Collection $commande;

    /** @var Collection<int, LigneCommande> */
    #[ORM\OneToMany(targetEntity: LigneCommande::class, mappedBy: 'commande')]
    protected Collection $ligneCommande;

    /** @var Collection<int, Article> */
    #[ORM\ManyToMany(targetEntity: Article::class, inversedBy: 'commande')]
    #[ORM\JoinTable(name: 'commande_article')]
    #[ORM\JoinColumn(name: 'commande_id', referencedColumnName: 'id', onDelete: 'CASCADE')]
    #[ORM\InverseJoinColumn(name: 'article_id', referencedColumnName: 'id', onDelete: 'CASCADE')]
    protected Collection $article;

    public function __construct()
    {
        $this->commande = new ArrayCollection();
        $this->ligneCommande = new ArrayCollection();
        $this->article = new ArrayCollection();
    }

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getClientId(): int
    {
        return $this->clientId;
    }

    public function getCommandeRemplaceeId(): ?int
    {
        return $this->commandeRemplaceeId;
    }

    public function getTotalHt(): string
    {
        return $this->totalHt;
    }

    public function setTotalHt(string $totalHt): static
    {
        $this->totalHt = $totalHt;

        return $this;
    }

    public function getClient(): Client
    {
        return $this->client;
    }

    public function setClient(Client $client): static
    {
        $this->client = $client;

        return $this;
    }

    public function getCommandeRemplacee(): ?Commande
    {
        return $this->commandeRemplacee;
    }

    public function setCommandeRemplacee(?Commande $commandeRemplacee): static
    {
        $this->commandeRemplacee = $commandeRemplacee;

        return $this;
    }

    /** @return Collection<int, Commande> */
    public function getCommande(): Collection
    {
        return $this->commande;
    }

    /** @return Collection<int, LigneCommande> */
    public function getLigneCommande(): Collection
    {
        return $this->ligneCommande;
    }

    /** @return Collection<int, Article> */
    public function getArticle(): Collection
    {
        return $this->article;
    }

    public function addArticle(Article $article): static
    {
        if (!$this->article->contains($article)) {
            $this->article->add($article);
        }

        return $this;
    }

    public function removeArticle(Article $article): static
    {
        $this->article->removeElement($article);

        return $this;
    }
}
