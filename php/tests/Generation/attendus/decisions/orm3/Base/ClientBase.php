<?php

// Généré par Ormeau depuis la base decisions et réécrit à chaque génération : le code propre à Client va dans Client.php.

declare(strict_types=1);

namespace Gescom\Domaine\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class ClientBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: '`CLI_ID`', type: 'integer')]
    protected ?int $cliId = null;

    #[ORM\Column(name: '`CLI_NOM`', type: 'string', length: 80)]
    protected string $cliNom;

    #[ORM\Column(name: '`CLI_ACTIF`', type: 'boolean')]
    protected bool $cliActif;

    public function getCliId(): ?int
    {
        return $this->cliId;
    }

    public function getCliNom(): string
    {
        return $this->cliNom;
    }

    public function setCliNom(string $cliNom): static
    {
        $this->cliNom = $cliNom;

        return $this;
    }

    public function isCliActif(): bool
    {
        return $this->cliActif;
    }

    public function setCliActif(bool $cliActif): static
    {
        $this->cliActif = $cliActif;

        return $this;
    }
}
