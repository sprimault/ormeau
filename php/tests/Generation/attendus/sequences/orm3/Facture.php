<?php

// Créé par Ormeau depuis la base sequences, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\FactureBase;
use Doctrine\ORM\Mapping as ORM;

/** serial dans le search_path de l'extraction : nom nu. */
#[ORM\Entity]
#[ORM\Table(name: 'facture', options: ['comment' => 'serial dans le search_path de l\'extraction : nom nu'])]
class Facture extends FactureBase {}
