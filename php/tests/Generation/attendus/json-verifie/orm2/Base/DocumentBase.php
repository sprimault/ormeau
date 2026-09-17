<?php

// Généré par Ormeau depuis la base json-verifie et réécrit à chaque génération : le code propre à Document va dans Document.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class DocumentBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    /** @var array<mixed>|null */
    #[ORM\Column(name: 'contenu', type: 'json', nullable: true)]
    protected ?array $contenu = null;

    /** @var array<mixed>|null */
    #[ORM\Column(name: 'objet', type: 'json', nullable: true)]
    protected ?array $objet = null;

    /** @var array<mixed>|null */
    #[ORM\Column(name: 'inverse', type: 'json', nullable: true)]
    protected ?array $inverse = null;

    /** @var array<mixed>|null */
    #[ORM\Column(name: 'positif', type: 'json', nullable: true)]
    protected ?array $positif = null;

    /** VALUE admet un scalaire, qui ne se relit pas en tableau : non reconnu. */
    #[ORM\Column(
        name: 'valeur',
        type: 'text',
        nullable: true,
        options: ['comment' => 'VALUE admet un scalaire, qui ne se relit pas en tableau : non reconnu'],
    )]
    protected ?string $valeur = null;

    /** Vérification qui dit autre chose en plus : non reconnue. */
    #[ORM\Column(
        name: 'borne',
        type: 'text',
        nullable: true,
        options: ['comment' => 'Vérification qui dit autre chose en plus : non reconnue'],
    )]
    protected ?string $borne = null;

    /** Longueur déclarée, que json recréerait en VARCHAR(MAX) : non reconnu. */
    #[ORM\Column(
        name: 'limite',
        type: 'string',
        length: 200,
        nullable: true,
        options: ['comment' => 'Longueur déclarée, que json recréerait en VARCHAR(MAX) : non reconnu'],
    )]
    protected ?string $limite = null;

    /** Texte Unicode : reste en chaîne, json se propose dans le fichier de décisions. */
    #[ORM\Column(
        name: 'unicode',
        type: 'string',
        nullable: true,
        options: ['comment' => 'Texte Unicode : reste en chaîne, json se propose dans le fichier de décisions'],
    )]
    protected ?string $unicode = null;

    /**
     * json forcé par décision : l'avertissement le suit.
     *
     * @var array<mixed>|null
     */
    #[ORM\Column(
        name: 'unicode_force',
        type: 'json',
        nullable: true,
        options: ['comment' => 'json forcé par décision : l\'avertissement le suit'],
    )]
    protected ?array $unicodeForce = null;

    /**
     * json forcé sans vérification : même avertissement.
     *
     * @var array<mixed>|null
     */
    #[ORM\Column(
        name: 'unicode_force_sans_verification',
        type: 'json',
        nullable: true,
        options: ['comment' => 'json forcé sans vérification : même avertissement'],
    )]
    protected ?array $unicodeForceSansVerification = null;

    public function getId(): ?int
    {
        return $this->id;
    }

    /** @return array<mixed>|null */
    public function getContenu(): ?array
    {
        return $this->contenu;
    }

    /** @param array<mixed>|null $contenu */
    public function setContenu(?array $contenu): static
    {
        $this->contenu = $contenu;

        return $this;
    }

    /** @return array<mixed>|null */
    public function getObjet(): ?array
    {
        return $this->objet;
    }

    /** @param array<mixed>|null $objet */
    public function setObjet(?array $objet): static
    {
        $this->objet = $objet;

        return $this;
    }

    /** @return array<mixed>|null */
    public function getInverse(): ?array
    {
        return $this->inverse;
    }

    /** @param array<mixed>|null $inverse */
    public function setInverse(?array $inverse): static
    {
        $this->inverse = $inverse;

        return $this;
    }

    /** @return array<mixed>|null */
    public function getPositif(): ?array
    {
        return $this->positif;
    }

    /** @param array<mixed>|null $positif */
    public function setPositif(?array $positif): static
    {
        $this->positif = $positif;

        return $this;
    }

    public function getValeur(): ?string
    {
        return $this->valeur;
    }

    public function setValeur(?string $valeur): static
    {
        $this->valeur = $valeur;

        return $this;
    }

    public function getBorne(): ?string
    {
        return $this->borne;
    }

    public function setBorne(?string $borne): static
    {
        $this->borne = $borne;

        return $this;
    }

    public function getLimite(): ?string
    {
        return $this->limite;
    }

    public function setLimite(?string $limite): static
    {
        $this->limite = $limite;

        return $this;
    }

    public function getUnicode(): ?string
    {
        return $this->unicode;
    }

    public function setUnicode(?string $unicode): static
    {
        $this->unicode = $unicode;

        return $this;
    }

    /** @return array<mixed>|null */
    public function getUnicodeForce(): ?array
    {
        return $this->unicodeForce;
    }

    /** @param array<mixed>|null $unicodeForce */
    public function setUnicodeForce(?array $unicodeForce): static
    {
        $this->unicodeForce = $unicodeForce;

        return $this;
    }

    /** @return array<mixed>|null */
    public function getUnicodeForceSansVerification(): ?array
    {
        return $this->unicodeForceSansVerification;
    }

    /** @param array<mixed>|null $unicodeForceSansVerification */
    public function setUnicodeForceSansVerification(?array $unicodeForceSansVerification): static
    {
        $this->unicodeForceSansVerification = $unicodeForceSansVerification;

        return $this;
    }
}
