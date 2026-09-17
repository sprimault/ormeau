<?php

// Généré par Ormeau depuis la base nommage-court et réécrit à chaque génération : le code propre à Fournisseur va dans Fournisseur.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
#[ORM\UniqueConstraint(name: 'uq_fournisseur_pays_libelle', columns: ['code_pays', 'libelle'])]
#[ORM\UniqueConstraint(name: 'uq_fournisseur_libelle', columns: ['libelle'])]
abstract class FournisseurBase
{
    #[ORM\Id]
    #[ORM\Column(name: 'code_pays', type: 'string', length: 2)]
    protected string $codePays;

    #[ORM\Id]
    #[ORM\Column(name: 'code_interne', type: 'string', length: 20)]
    protected string $codeInterne;

    #[ORM\Column(name: 'libelle', type: 'string', length: 120, unique: true)]
    protected string $libelle;

    #[ORM\Column(name: 'heure_livraison', type: 'time_immutable', nullable: true)]
    protected ?\DateTimeImmutable $heureLivraison = null;

    public function getCodePays(): string
    {
        return $this->codePays;
    }

    public function setCodePays(string $codePays): static
    {
        $this->codePays = $codePays;

        return $this;
    }

    public function getCodeInterne(): string
    {
        return $this->codeInterne;
    }

    public function setCodeInterne(string $codeInterne): static
    {
        $this->codeInterne = $codeInterne;

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

    public function getHeureLivraison(): ?\DateTimeImmutable
    {
        return $this->heureLivraison;
    }

    public function setHeureLivraison(?\DateTimeImmutable $heureLivraison): static
    {
        $this->heureLivraison = $heureLivraison;

        return $this;
    }
}
