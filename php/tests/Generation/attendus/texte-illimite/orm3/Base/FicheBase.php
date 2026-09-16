<?php

// Généré par Ormeau depuis la base texte-illimite et réécrit à chaque génération : le code propre à Fiche va dans Fiche.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class FicheBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    /** Recréé en VARCHAR(MAX), à l'identique : text. */
    #[ORM\Column(
        name: 'note',
        type: 'text',
        nullable: true,
        options: ['comment' => 'Recréé en VARCHAR(MAX), à l\'identique : text'],
    )]
    protected ?string $note = null;

    /** Recréé en VARCHAR(MAX) il perdrait l'Unicode sans erreur : chaîne et avertissement. */
    #[ORM\Column(
        name: 'note_unicode',
        type: 'string',
        nullable: true,
        options: ['comment' => 'Recréé en VARCHAR(MAX) il perdrait l\'Unicode sans erreur : chaîne et avertissement'],
    )]
    protected ?string $noteUnicode = null;

    /** Même cas que nvarchar(max). */
    #[ORM\Column(
        name: 'ancienne_note',
        type: 'string',
        nullable: true,
        options: ['comment' => 'Même cas que nvarchar(max)'],
    )]
    protected ?string $ancienneNote = null;

    /** Longueur déclarée : chaîne, sans avertissement. */
    #[ORM\Column(
        name: 'libelle',
        type: 'string',
        length: 80,
        options: ['comment' => 'Longueur déclarée : chaîne, sans avertissement'],
    )]
    protected string $libelle;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getNote(): ?string
    {
        return $this->note;
    }

    public function setNote(?string $note): static
    {
        $this->note = $note;

        return $this;
    }

    public function getNoteUnicode(): ?string
    {
        return $this->noteUnicode;
    }

    public function setNoteUnicode(?string $noteUnicode): static
    {
        $this->noteUnicode = $noteUnicode;

        return $this;
    }

    public function getAncienneNote(): ?string
    {
        return $this->ancienneNote;
    }

    public function setAncienneNote(?string $ancienneNote): static
    {
        $this->ancienneNote = $ancienneNote;

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
