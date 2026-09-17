<?php

// Généré par Ormeau depuis la base defauts-sql-server et réécrit à chaque génération : le code propre à Evenement va dans Evenement.php.

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

    /** CURRENT_TIMESTAMP s'écrit ainsi au catalogue. */
    #[ORM\Column(
        name: 'horodate',
        type: 'datetime_immutable',
        options: ['default' => new \Doctrine\DBAL\Schema\DefaultExpression\CurrentTimestamp(), 'comment' => 'CURRENT_TIMESTAMP s\'écrit ainsi au catalogue'],
    )]
    protected \DateTimeImmutable $horodate;

    #[ORM\Column(
        name: 'horodate_precis',
        type: 'datetime_immutable',
        nullable: true,
        options: ['default' => new \Doctrine\DBAL\Schema\DefaultExpression\CurrentTimestamp()],
    )]
    protected ?\DateTimeImmutable $horodatePrecis = null;

    /** Instant exact, mais aucune expression Doctrine ne l'écrit : non reconnu. */
    #[ORM\Column(
        name: 'horodate_fuseau',
        type: 'datetimetz_immutable',
        nullable: true,
        options: ['comment' => 'Instant exact, mais aucune expression Doctrine ne l\'écrit : non reconnu'],
    )]
    protected ?\DateTimeImmutable $horodateFuseau = null;

    /** Heure locale étiquetée +00:00 : non reconnu. */
    #[ORM\Column(
        name: 'horodate_local_fuseau',
        type: 'datetimetz_immutable',
        nullable: true,
        options: ['comment' => 'Heure locale étiquetée +00:00 : non reconnu'],
    )]
    protected ?\DateTimeImmutable $horodateLocalFuseau = null;

    /** Précision plus fine que getdate() : non reconnu. */
    #[ORM\Column(
        name: 'horodate_systeme',
        type: 'datetime_immutable',
        nullable: true,
        options: ['comment' => 'Précision plus fine que getdate() : non reconnu'],
    )]
    protected ?\DateTimeImmutable $horodateSysteme = null;

    /** Heure UTC, pas l'heure locale : non reconnu. */
    #[ORM\Column(
        name: 'horodate_utc',
        type: 'datetime_immutable',
        nullable: true,
        options: ['comment' => 'Heure UTC, pas l\'heure locale : non reconnu'],
    )]
    protected ?\DateTimeImmutable $horodateUtc = null;

    #[ORM\Column(
        name: 'jour',
        type: 'date_immutable',
        options: ['default' => new \Doctrine\DBAL\Schema\DefaultExpression\CurrentTimestamp()],
    )]
    protected \DateTimeImmutable $jour;

    /** Forme qu'écrit Doctrine sous DBAL 3. */
    #[ORM\Column(
        name: 'jour_converti',
        type: 'date_immutable',
        nullable: true,
        options: ['default' => new \Doctrine\DBAL\Schema\DefaultExpression\CurrentTimestamp(), 'comment' => 'Forme qu\'écrit Doctrine sous DBAL 3'],
    )]
    protected ?\DateTimeImmutable $jourConverti = null;

    #[ORM\Column(
        name: 'heure',
        type: 'time_immutable',
        nullable: true,
        options: ['default' => new \Doctrine\DBAL\Schema\DefaultExpression\CurrentTimestamp()],
    )]
    protected ?\DateTimeImmutable $heure = null;

    #[ORM\Column(
        name: 'heure_convertie',
        type: 'time_immutable',
        nullable: true,
        options: ['default' => new \Doctrine\DBAL\Schema\DefaultExpression\CurrentTimestamp()],
    )]
    protected ?\DateTimeImmutable $heureConvertie = null;

    /** Forme d'un autre dialecte : non reconnue sous SQL Server. */
    #[ORM\Column(
        name: 'forme_postgres',
        type: 'datetime_immutable',
        nullable: true,
        options: ['comment' => 'Forme d\'un autre dialecte : non reconnue sous SQL Server'],
    )]
    protected ?\DateTimeImmutable $formePostgres = null;

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

    public function getHorodatePrecis(): ?\DateTimeImmutable
    {
        return $this->horodatePrecis;
    }

    public function setHorodatePrecis(?\DateTimeImmutable $horodatePrecis): static
    {
        $this->horodatePrecis = $horodatePrecis;

        return $this;
    }

    public function getHorodateFuseau(): ?\DateTimeImmutable
    {
        return $this->horodateFuseau;
    }

    public function setHorodateFuseau(?\DateTimeImmutable $horodateFuseau): static
    {
        $this->horodateFuseau = $horodateFuseau;

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

    public function getHorodateSysteme(): ?\DateTimeImmutable
    {
        return $this->horodateSysteme;
    }

    public function setHorodateSysteme(?\DateTimeImmutable $horodateSysteme): static
    {
        $this->horodateSysteme = $horodateSysteme;

        return $this;
    }

    public function getHorodateUtc(): ?\DateTimeImmutable
    {
        return $this->horodateUtc;
    }

    public function setHorodateUtc(?\DateTimeImmutable $horodateUtc): static
    {
        $this->horodateUtc = $horodateUtc;

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

    public function getJourConverti(): ?\DateTimeImmutable
    {
        return $this->jourConverti;
    }

    public function setJourConverti(?\DateTimeImmutable $jourConverti): static
    {
        $this->jourConverti = $jourConverti;

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

    public function getHeureConvertie(): ?\DateTimeImmutable
    {
        return $this->heureConvertie;
    }

    public function setHeureConvertie(?\DateTimeImmutable $heureConvertie): static
    {
        $this->heureConvertie = $heureConvertie;

        return $this;
    }

    public function getFormePostgres(): ?\DateTimeImmutable
    {
        return $this->formePostgres;
    }

    public function setFormePostgres(?\DateTimeImmutable $formePostgres): static
    {
        $this->formePostgres = $formePostgres;

        return $this;
    }
}
