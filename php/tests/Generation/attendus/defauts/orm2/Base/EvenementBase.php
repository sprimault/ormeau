<?php

// Généré par Ormeau depuis la base defauts et réécrit à chaque génération : le code propre à Evenement va dans Evenement.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class EvenementBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    #[ORM\Column(name: 'horodate', type: 'datetimetz_immutable', options: ['default' => 'CURRENT_TIMESTAMP'])]
    protected \DateTimeImmutable $horodate;

    #[ORM\Column(
        name: 'horodate_sql',
        type: 'datetimetz_immutable',
        nullable: true,
        options: ['default' => 'CURRENT_TIMESTAMP'],
    )]
    protected ?\DateTimeImmutable $horodateSql = null;

    #[ORM\Column(
        name: 'horodate_transaction',
        type: 'datetimetz_immutable',
        nullable: true,
        options: ['default' => 'CURRENT_TIMESTAMP'],
    )]
    protected ?\DateTimeImmutable $horodateTransaction = null;

    #[ORM\Column(
        name: 'horodate_local',
        type: 'datetime_immutable',
        nullable: true,
        options: ['default' => 'CURRENT_TIMESTAMP'],
    )]
    protected ?\DateTimeImmutable $horodateLocal = null;

    /** Heure locale convertie selon le fuseau de la session : non reconnu. */
    #[ORM\Column(
        name: 'horodate_local_fuseau',
        type: 'datetimetz_immutable',
        nullable: true,
        options: ['comment' => 'Heure locale convertie selon le fuseau de la session : non reconnu'],
    )]
    protected ?\DateTimeImmutable $horodateLocalFuseau = null;

    /** Instant de l'horloge, pas de la transaction : non reconnu. */
    #[ORM\Column(
        name: 'horloge',
        type: 'datetimetz_immutable',
        nullable: true,
        options: ['comment' => 'Instant de l\'horloge, pas de la transaction : non reconnu'],
    )]
    protected ?\DateTimeImmutable $horloge = null;

    #[ORM\Column(name: 'horodate_arrondi', type: 'datetimetz_immutable', nullable: true)]
    protected ?\DateTimeImmutable $horodateArrondi = null;

    /** Minuit du jour : non reconnu. */
    #[ORM\Column(
        name: 'horodate_jour',
        type: 'datetime_immutable',
        nullable: true,
        options: ['comment' => 'Minuit du jour : non reconnu'],
    )]
    protected ?\DateTimeImmutable $horodateJour = null;

    #[ORM\Column(name: 'jour', type: 'date_immutable', options: ['default' => 'CURRENT_DATE'])]
    protected \DateTimeImmutable $jour;

    #[ORM\Column(
        name: 'jour_maintenant',
        type: 'date_immutable',
        nullable: true,
        options: ['default' => 'CURRENT_DATE'],
    )]
    protected ?\DateTimeImmutable $jourMaintenant = null;

    #[ORM\Column(name: 'jour_local', type: 'date_immutable', nullable: true, options: ['default' => 'CURRENT_DATE'])]
    protected ?\DateTimeImmutable $jourLocal = null;

    /** Type forcé par décision : le défaut tombe. */
    #[ORM\Column(
        name: 'jour_force',
        type: 'string',
        nullable: true,
        options: ['comment' => 'Type forcé par décision : le défaut tombe'],
    )]
    protected ?string $jourForce = null;

    #[ORM\Column(name: 'heure', type: 'time_immutable', nullable: true, options: ['default' => 'CURRENT_TIME'])]
    protected ?\DateTimeImmutable $heure = null;

    #[ORM\Column(name: 'heure_locale', type: 'time_immutable', nullable: true, options: ['default' => 'CURRENT_TIME'])]
    protected ?\DateTimeImmutable $heureLocale = null;

    #[ORM\Column(name: 'heure_maintenant', type: 'time_immutable', nullable: true)]
    protected ?\DateTimeImmutable $heureMaintenant = null;

    /** DEFAULT NULL : aucun défaut, et rien à signaler. */
    #[ORM\Column(
        name: 'libelle',
        type: 'string',
        length: 20,
        nullable: true,
        options: ['comment' => 'DEFAULT NULL : aucun défaut, et rien à signaler'],
    )]
    protected ?string $libelle = null;

    #[ORM\Column(name: 'reference', type: 'guid')]
    protected string $reference;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getHorodate(): \DateTimeImmutable
    {
        return $this->horodate;
    }

    public function setHorodate(\DateTimeImmutable $horodate): static
    {
        $this->horodate = $horodate;

        return $this;
    }

    public function getHorodateSql(): ?\DateTimeImmutable
    {
        return $this->horodateSql;
    }

    public function setHorodateSql(?\DateTimeImmutable $horodateSql): static
    {
        $this->horodateSql = $horodateSql;

        return $this;
    }

    public function getHorodateTransaction(): ?\DateTimeImmutable
    {
        return $this->horodateTransaction;
    }

    public function setHorodateTransaction(?\DateTimeImmutable $horodateTransaction): static
    {
        $this->horodateTransaction = $horodateTransaction;

        return $this;
    }

    public function getHorodateLocal(): ?\DateTimeImmutable
    {
        return $this->horodateLocal;
    }

    public function setHorodateLocal(?\DateTimeImmutable $horodateLocal): static
    {
        $this->horodateLocal = $horodateLocal;

        return $this;
    }

    public function getHorodateLocalFuseau(): ?\DateTimeImmutable
    {
        return $this->horodateLocalFuseau;
    }

    public function setHorodateLocalFuseau(?\DateTimeImmutable $horodateLocalFuseau): static
    {
        $this->horodateLocalFuseau = $horodateLocalFuseau;

        return $this;
    }

    public function getHorloge(): ?\DateTimeImmutable
    {
        return $this->horloge;
    }

    public function setHorloge(?\DateTimeImmutable $horloge): static
    {
        $this->horloge = $horloge;

        return $this;
    }

    public function getHorodateArrondi(): ?\DateTimeImmutable
    {
        return $this->horodateArrondi;
    }

    public function setHorodateArrondi(?\DateTimeImmutable $horodateArrondi): static
    {
        $this->horodateArrondi = $horodateArrondi;

        return $this;
    }

    public function getHorodateJour(): ?\DateTimeImmutable
    {
        return $this->horodateJour;
    }

    public function setHorodateJour(?\DateTimeImmutable $horodateJour): static
    {
        $this->horodateJour = $horodateJour;

        return $this;
    }

    public function getJour(): \DateTimeImmutable
    {
        return $this->jour;
    }

    public function setJour(\DateTimeImmutable $jour): static
    {
        $this->jour = $jour;

        return $this;
    }

    public function getJourMaintenant(): ?\DateTimeImmutable
    {
        return $this->jourMaintenant;
    }

    public function setJourMaintenant(?\DateTimeImmutable $jourMaintenant): static
    {
        $this->jourMaintenant = $jourMaintenant;

        return $this;
    }

    public function getJourLocal(): ?\DateTimeImmutable
    {
        return $this->jourLocal;
    }

    public function setJourLocal(?\DateTimeImmutable $jourLocal): static
    {
        $this->jourLocal = $jourLocal;

        return $this;
    }

    public function getJourForce(): ?string
    {
        return $this->jourForce;
    }

    public function setJourForce(?string $jourForce): static
    {
        $this->jourForce = $jourForce;

        return $this;
    }

    public function getHeure(): ?\DateTimeImmutable
    {
        return $this->heure;
    }

    public function setHeure(?\DateTimeImmutable $heure): static
    {
        $this->heure = $heure;

        return $this;
    }

    public function getHeureLocale(): ?\DateTimeImmutable
    {
        return $this->heureLocale;
    }

    public function setHeureLocale(?\DateTimeImmutable $heureLocale): static
    {
        $this->heureLocale = $heureLocale;

        return $this;
    }

    public function getHeureMaintenant(): ?\DateTimeImmutable
    {
        return $this->heureMaintenant;
    }

    public function setHeureMaintenant(?\DateTimeImmutable $heureMaintenant): static
    {
        $this->heureMaintenant = $heureMaintenant;

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

    public function getReference(): string
    {
        return $this->reference;
    }

    public function setReference(string $reference): static
    {
        $this->reference = $reference;

        return $this;
    }
}
