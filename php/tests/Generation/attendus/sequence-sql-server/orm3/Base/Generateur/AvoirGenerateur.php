<?php

// Généré par Ormeau depuis la base sequence-sql-server et réécrit à chaque génération.

declare(strict_types=1);

namespace App\Entity\Base\Generateur;

use Doctrine\ORM\Id\SequenceGenerator;

/** Tire la séquence ventes.sq_avoir pour la clé de Avoir : ORM 3 ignore #[ORM\SequenceGenerator] sur une classe de base mappée. */
class AvoirGenerateur extends SequenceGenerator
{
    public function __construct()
    {
        parent::__construct('ventes.sq_avoir', 1);
    }
}
