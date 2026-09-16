<?php

// Généré par Ormeau depuis la base heritage-genere et réécrit à chaque génération : le code propre à Facture va dans Facture.php.

declare(strict_types=1);

namespace App\Entity\Base;

use App\Entity\Document;
use Doctrine\ORM\Mapping as ORM;

#[ORM\MappedSuperclass]
abstract class FactureBase extends Document
{
    #[ORM\Column(name: 'echeance', type: 'date_immutable')]
    protected \DateTimeImmutable $echeance;

    public function getEcheance(): \DateTimeImmutable
    {
        return $this->echeance;
    }

    public function setEcheance(\DateTimeImmutable $echeance): static
    {
        $this->echeance = $echeance;

        return $this;
    }
}
