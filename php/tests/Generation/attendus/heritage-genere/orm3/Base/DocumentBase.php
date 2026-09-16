<?php

// Généré par Ormeau depuis la base heritage-genere et réécrit à chaque génération : le code propre à Document va dans Document.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

/** La colonne nature départage la hiérarchie : Doctrine l'écrit, elle n'a pas de propriété. */
#[ORM\MappedSuperclass]
abstract class DocumentBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    #[ORM\Column(name: 'quantite', type: 'integer', options: ['default' => 1])]
    protected int $quantite = 1;

    #[ORM\Column(name: 'prix_unitaire', type: 'decimal', precision: 12, scale: 4)]
    protected string $prixUnitaire;

    /**
     * Montant hors taxes, tenu par la base.
     *
     * Calculée par la base : ((quantite)::numeric * prix_unitaire).
     */
    #[ORM\Column(
        name: 'total_ht',
        type: 'decimal',
        precision: 14,
        scale: 4,
        insertable: false,
        updatable: false,
        generated: 'ALWAYS',
        options: ['comment' => 'Montant hors taxes, tenu par la base'],
    )]
    protected string $totalHt;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getQuantite(): int
    {
        return $this->quantite;
    }

    public function setQuantite(int $quantite): static
    {
        $this->quantite = $quantite;

        return $this;
    }

    public function getPrixUnitaire(): string
    {
        return $this->prixUnitaire;
    }

    public function setPrixUnitaire(string $prixUnitaire): static
    {
        $this->prixUnitaire = $prixUnitaire;

        return $this;
    }

    public function getTotalHt(): string
    {
        return $this->totalHt;
    }
}
