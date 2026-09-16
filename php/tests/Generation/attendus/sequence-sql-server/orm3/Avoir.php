<?php

// Créé par Ormeau depuis la base sequence-sql-server, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\AvoirBase;
use Doctrine\ORM\Mapping as ORM;

/** Clé par DEFAULT NEXT VALUE FOR : la séquence se lit dans le champ sequence du défaut, pas dans l'expression. */
#[ORM\Entity]
#[ORM\Table(
    name: 'avoir',
    schema: 'ventes',
    options: ['comment' => 'Clé par DEFAULT NEXT VALUE FOR : la séquence se lit dans le champ sequence du défaut, pas dans l\'expression'],
)]
class Avoir extends AvoirBase {}
