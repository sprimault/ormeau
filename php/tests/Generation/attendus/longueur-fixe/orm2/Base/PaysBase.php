<?php

// Généré par Ormeau depuis la base longueur-fixe et réécrit à chaque génération : le code propre à Pays va dans Pays.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class PaysBase
{
    #[ORM\Id]
    #[ORM\Column(name: 'code', type: 'string', length: 2, options: ['fixed' => true])]
    protected string $code;

    /** Calque extrait avant le champ longueur_fixe : la forme character(n) reste lue. */
    #[ORM\Column(
        name: 'ancien_code',
        type: 'string',
        length: 3,
        nullable: true,
        options: ['comment' => 'Calque extrait avant le champ longueur_fixe : la forme character(n) reste lue', 'fixed' => true],
    )]
    protected ?string $ancienCode = null;

    #[ORM\Column(name: 'nom', type: 'string', length: 40)]
    protected string $nom;

    /** Sans longueur : blob. */
    #[ORM\Column(name: 'drapeau', type: 'blob', nullable: true, options: ['comment' => 'Sans longueur : blob'])]
    protected mixed $drapeau = null;

    public function getCode(): string
    {
        return $this->code;
    }

    public function setCode(string $code): static
    {
        $this->code = $code;

        return $this;
    }

    public function getAncienCode(): ?string
    {
        return $this->ancienCode;
    }

    public function setAncienCode(?string $ancienCode): static
    {
        $this->ancienCode = $ancienCode;

        return $this;
    }

    public function getNom(): string
    {
        return $this->nom;
    }

    public function setNom(string $nom): static
    {
        $this->nom = $nom;

        return $this;
    }

    public function getDrapeau(): mixed
    {
        return $this->drapeau;
    }

    public function setDrapeau(mixed $drapeau): static
    {
        $this->drapeau = $drapeau;

        return $this;
    }
}
