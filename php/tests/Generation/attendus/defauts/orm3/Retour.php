<?php

// Créé par Ormeau depuis la base defauts, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\RetourBase;
use Doctrine\ORM\Mapping as ORM;

/** Même horodatage sans défaut, comme avoir : un autre trait, sans quoi le défaut passerait d'une table à l'autre. */
#[ORM\Entity]
#[ORM\Table(
    name: 'retour',
    options: ['comment' => 'Même horodatage sans défaut, comme avoir : un autre trait, sans quoi le défaut passerait d\'une table à l\'autre'],
)]
class Retour extends RetourBase {}
