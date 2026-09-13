<?php

// Généré par Ormeau et réécrit à chaque génération : le code propre à Eventail va dans Eventail.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class EventailBase
{
    #[ORM\Id]
    #[ORM\Column(name: 'cle', type: 'guid')]
    protected string $cle;

    #[ORM\Column(name: 'compteur_court', type: 'smallint')]
    protected int $compteurCourt;

    #[ORM\Column(name: 'compteur_long', type: 'bigint')]
    protected int $compteurLong;

    #[ORM\Column(name: 'note_libre', type: 'text', nullable: true)]
    protected ?string $noteLibre = null;

    #[ORM\Column(name: 'instant_local', type: 'datetime_immutable', nullable: true)]
    protected ?\DateTimeImmutable $instantLocal = null;

    #[ORM\Column(name: 'instant_absolu', type: 'datetimetz_immutable', nullable: true)]
    protected ?\DateTimeImmutable $instantAbsolu = null;

    #[ORM\Column(name: 'jour', type: 'date_immutable', nullable: true)]
    protected ?\DateTimeImmutable $jour = null;

    #[ORM\Column(name: 'duree', type: 'dateinterval', nullable: true)]
    protected ?\DateInterval $duree = null;

    /** @var array<mixed>|null */
    #[ORM\Column(name: 'charge_utile', type: 'json', nullable: true)]
    protected ?array $chargeUtile = null;

    #[ORM\Column(name: 'empreinte', type: 'blob', nullable: true)]
    protected mixed $empreinte = null;

    #[ORM\Column(name: 'adresse_ip', type: 'string', nullable: true)]
    protected ?string $adresseIp = null;

    #[ORM\Column(name: 'type_maison', type: 'string', nullable: true)]
    protected ?string $typeMaison = null;

    #[ORM\Column(name: 'total_ligne', type: 'decimal', precision: 14, scale: 4, insertable: false, updatable: false)]
    protected string $totalLigne;

    public function getCle(): string
    {
        return $this->cle;
    }

    public function setCle(string $cle): static
    {
        $this->cle = $cle;

        return $this;
    }

    public function getCompteurCourt(): int
    {
        return $this->compteurCourt;
    }

    public function setCompteurCourt(int $compteurCourt): static
    {
        $this->compteurCourt = $compteurCourt;

        return $this;
    }

    public function getCompteurLong(): int
    {
        return $this->compteurLong;
    }

    public function setCompteurLong(int $compteurLong): static
    {
        $this->compteurLong = $compteurLong;

        return $this;
    }

    public function getNoteLibre(): ?string
    {
        return $this->noteLibre;
    }

    public function setNoteLibre(?string $noteLibre): static
    {
        $this->noteLibre = $noteLibre;

        return $this;
    }

    public function getInstantLocal(): ?\DateTimeImmutable
    {
        return $this->instantLocal;
    }

    public function setInstantLocal(?\DateTimeImmutable $instantLocal): static
    {
        $this->instantLocal = $instantLocal;

        return $this;
    }

    public function getInstantAbsolu(): ?\DateTimeImmutable
    {
        return $this->instantAbsolu;
    }

    public function setInstantAbsolu(?\DateTimeImmutable $instantAbsolu): static
    {
        $this->instantAbsolu = $instantAbsolu;

        return $this;
    }

    public function getJour(): ?\DateTimeImmutable
    {
        return $this->jour;
    }

    public function setJour(?\DateTimeImmutable $jour): static
    {
        $this->jour = $jour;

        return $this;
    }

    public function getDuree(): ?\DateInterval
    {
        return $this->duree;
    }

    public function setDuree(?\DateInterval $duree): static
    {
        $this->duree = $duree;

        return $this;
    }

    /** @return array<mixed>|null */
    public function getChargeUtile(): ?array
    {
        return $this->chargeUtile;
    }

    /** @param array<mixed>|null $chargeUtile */
    public function setChargeUtile(?array $chargeUtile): static
    {
        $this->chargeUtile = $chargeUtile;

        return $this;
    }

    public function getEmpreinte(): mixed
    {
        return $this->empreinte;
    }

    public function setEmpreinte(mixed $empreinte): static
    {
        $this->empreinte = $empreinte;

        return $this;
    }

    public function getAdresseIp(): ?string
    {
        return $this->adresseIp;
    }

    public function setAdresseIp(?string $adresseIp): static
    {
        $this->adresseIp = $adresseIp;

        return $this;
    }

    public function getTypeMaison(): ?string
    {
        return $this->typeMaison;
    }

    public function setTypeMaison(?string $typeMaison): static
    {
        $this->typeMaison = $typeMaison;

        return $this;
    }

    public function getTotalLigne(): string
    {
        return $this->totalLigne;
    }
}
