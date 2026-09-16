<?php

// Généré par Ormeau depuis la base sequence-sql-server et réécrit à chaque génération.

declare(strict_types=1);

namespace App\Entity\Base\Generateur;

use Doctrine\ORM\Id\SequenceGenerator;

/** Tire la séquence sq_note pour la clé de Note : ORM 3 ignore #[ORM\SequenceGenerator] sur une classe de base mappée. */
class NoteGenerateur extends SequenceGenerator
{
    public function __construct()
    {
        parent::__construct('sq_note', 1);
    }
}
