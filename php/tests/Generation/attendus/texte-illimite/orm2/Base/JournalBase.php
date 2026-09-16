<?php

// Généré par Ormeau depuis la base texte-illimite et réécrit à chaque génération : le code propre à Journal va dans Journal.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class JournalBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    #[ORM\Column(name: 'message', type: 'text', nullable: true)]
    protected ?string $message = null;

    /** varchar sans longueur : illimité, donc text. */
    #[ORM\Column(
        name: 'detail',
        type: 'text',
        nullable: true,
        options: ['comment' => 'varchar sans longueur : illimité, donc text'],
    )]
    protected ?string $detail = null;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getMessage(): ?string
    {
        return $this->message;
    }

    public function setMessage(?string $message): static
    {
        $this->message = $message;

        return $this;
    }

    public function getDetail(): ?string
    {
        return $this->detail;
    }

    public function setDetail(?string $detail): static
    {
        $this->detail = $detail;

        return $this;
    }
}
