<?php

// Généré par Ormeau et réécrit à chaque génération : le code propre à Profil va dans Profil.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Utilisateur;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class ProfilBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    /** Lecture seule : écrite par l'association utilisateur. */
    #[ORM\Column(name: 'utilisateur_ref', type: 'integer', insertable: false, updatable: false)]
    protected int $utilisateurRef;

    #[ORM\Column(name: 'parrain_id', type: 'integer', nullable: true)]
    protected ?int $parrainId = null;

    #[ORM\OneToOne(targetEntity: Utilisateur::class, inversedBy: 'profil')]
    #[ORM\JoinColumn(name: 'utilisateur_ref', referencedColumnName: 'id', nullable: false)]
    protected Utilisateur $utilisateur;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getUtilisateurRef(): int
    {
        return $this->utilisateurRef;
    }

    public function getParrainId(): ?int
    {
        return $this->parrainId;
    }

    public function setParrainId(?int $parrainId): static
    {
        $this->parrainId = $parrainId;

        return $this;
    }

    public function getUtilisateur(): Utilisateur
    {
        return $this->utilisateur;
    }

    public function setUtilisateur(Utilisateur $utilisateur): static
    {
        $this->utilisateur = $utilisateur;

        return $this;
    }
}
