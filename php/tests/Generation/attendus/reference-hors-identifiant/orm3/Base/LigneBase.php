<?php

// Généré par Ormeau et réécrit à chaque génération : le code propre à Ligne va dans Ligne.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Article;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class LigneBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    /** Lecture seule : écrite par l'association article. */
    #[ORM\Column(name: 'article_version', type: 'integer', insertable: false, updatable: false)]
    protected int $articleVersion;

    /** Lecture seule : écrite par l'association article. */
    #[ORM\Column(name: 'article_code', type: 'string', length: 20, insertable: false, updatable: false)]
    protected string $articleCode;

    #[ORM\ManyToOne(targetEntity: Article::class, inversedBy: 'ligne')]
    #[ORM\JoinColumn(name: 'article_version', referencedColumnName: 'version', nullable: false)]
    #[ORM\JoinColumn(name: 'article_code', referencedColumnName: 'code', nullable: false)]
    protected Article $article;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getArticleVersion(): int
    {
        return $this->articleVersion;
    }

    public function getArticleCode(): string
    {
        return $this->articleCode;
    }

    public function getArticle(): Article
    {
        return $this->article;
    }

    public function setArticle(Article $article): static
    {
        $this->article = $article;

        return $this;
    }
}
