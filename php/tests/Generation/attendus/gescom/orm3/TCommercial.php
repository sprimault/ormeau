<?php

// Créé par Ormeau depuis la base gescom, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\TCommercialBase;
use Doctrine\ORM\Mapping as ORM;

/** Force de vente. */
#[ORM\Entity]
#[ORM\Table(name: 't_commercial', schema: 'gescom', options: ['comment' => 'Force de vente'])]
class TCommercial extends TCommercialBase {}
