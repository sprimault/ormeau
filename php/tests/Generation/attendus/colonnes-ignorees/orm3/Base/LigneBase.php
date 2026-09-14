<?php

// Généré par Ormeau depuis la base colonnes-ignorees et réécrit à chaque génération : le code propre à Ligne va dans Ligne.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Commande;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class LigneBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    /** Lecture seule : écrite par l'association commande. */
    #[ORM\Column(name: 'commande_id', type: 'integer', insertable: false, updatable: false)]
    protected int $commandeId;

    #[ORM\Column(name: 'article_code', type: 'string', length: 20)]
    protected string $articleCode;

    #[ORM\ManyToOne(targetEntity: Commande::class, inversedBy: 'ligne')]
    #[ORM\JoinColumn(name: 'commande_id', referencedColumnName: 'id', nullable: false, onDelete: 'CASCADE')]
    protected Commande $commande;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getCommandeId(): int
    {
        return $this->commandeId;
    }

    public function getArticleCode(): string
    {
        return $this->articleCode;
    }

    public function setArticleCode(string $articleCode): static
    {
        $this->articleCode = $articleCode;

        return $this;
    }

    public function getCommande(): Commande
    {
        return $this->commande;
    }

    public function setCommande(Commande $commande): static
    {
        $this->commande = $commande;

        return $this;
    }
}
