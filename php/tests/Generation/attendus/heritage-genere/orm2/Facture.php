<?php

// Créé par Ormeau depuis la base heritage-genere, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\FactureBase;
use Doctrine\ORM\Mapping as ORM;

/** Déclarée enfant de document par le fichier de décisions. */
#[ORM\Entity]
#[ORM\Table(name: 'facture', options: ['comment' => 'Déclarée enfant de document par le fichier de décisions'])]
class Facture extends FactureBase {}
