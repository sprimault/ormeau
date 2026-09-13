<?php

// Généré par Ormeau et réécrit à chaque génération : le code propre à Commande va dans Commande.php.

declare(strict_types=1);

namespace App\Entity\Base;

use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class CommandeBase
{
    #[ORM\Id]
    #[ORM\GeneratedValue(strategy: 'IDENTITY')]
    #[ORM\Column(name: 'id', type: 'integer')]
    protected ?int $id = null;

    #[ORM\Column(name: 'audit_id', type: 'integer', nullable: true)]
    protected ?int $auditId = null;

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getAuditId(): ?int
    {
        return $this->auditId;
    }

    public function setAuditId(?int $auditId): static
    {
        $this->auditId = $auditId;

        return $this;
    }
}
