<?php

// Généré par Ormeau depuis la base longueur-fixe et réécrit à chaque génération : le code propre à Piece va dans Piece.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class PieceBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    #[ORM\Column(name: 'code', type: 'string', length: 2, options: ['fixed' => true])]
    protected string $code;

    #[ORM\Column(name: 'code_ascii', type: 'string', length: 3, nullable: true, options: ['fixed' => true])]
    protected ?string $codeAscii = null;

    /** Longueur variable : pas de longueur fixe. */
    #[ORM\Column(
        name: 'libelle',
        type: 'string',
        length: 40,
        nullable: true,
        options: ['comment' => 'Longueur variable : pas de longueur fixe'],
    )]
    protected ?string $libelle = null;

    /** Binaire fixe : binary, longueur 16, fixe. */
    #[ORM\Column(
        name: 'empreinte',
        type: 'binary',
        length: 16,
        nullable: true,
        options: ['comment' => 'Binaire fixe : binary, longueur 16, fixe', 'fixed' => true],
    )]
    protected ?string $empreinte = null;

    /** Binaire de longueur déclarée : binary, longueur 16, variable. */
    #[ORM\Column(
        name: 'jeton',
        type: 'binary',
        length: 16,
        nullable: true,
        options: ['comment' => 'Binaire de longueur déclarée : binary, longueur 16, variable'],
    )]
    protected ?string $jeton = null;

    /** Binaire sans longueur : blob. */
    #[ORM\Column(
        name: 'document',
        type: 'blob',
        nullable: true,
        options: ['comment' => 'Binaire sans longueur : blob'],
    )]
    protected mixed $document = null;

    /** Sans type Doctrine : smallint, le plus proche. */
    #[ORM\Column(
        name: 'niveau',
        type: 'smallint',
        nullable: true,
        options: ['comment' => 'Sans type Doctrine : smallint, le plus proche'],
    )]
    protected ?int $niveau = null;

    public function getId(): ?int
    {
        return $this->id;
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

    public function getCodeAscii(): ?string
    {
        return $this->codeAscii;
    }

    public function setCodeAscii(?string $codeAscii): static
    {
        $this->codeAscii = $codeAscii;

        return $this;
    }

    public function getLibelle(): ?string
    {
        return $this->libelle;
    }

    public function setLibelle(?string $libelle): static
    {
        $this->libelle = $libelle;

        return $this;
    }

    public function getEmpreinte(): ?string
    {
        return $this->empreinte;
    }

    public function setEmpreinte(?string $empreinte): static
    {
        $this->empreinte = $empreinte;

        return $this;
    }

    public function getJeton(): ?string
    {
        return $this->jeton;
    }

    public function setJeton(?string $jeton): static
    {
        $this->jeton = $jeton;

        return $this;
    }

    public function getDocument(): mixed
    {
        return $this->document;
    }

    public function setDocument(mixed $document): static
    {
        $this->document = $document;

        return $this;
    }

    public function getNiveau(): ?int
    {
        return $this->niveau;
    }

    public function setNiveau(?int $niveau): static
    {
        $this->niveau = $niveau;

        return $this;
    }
}
