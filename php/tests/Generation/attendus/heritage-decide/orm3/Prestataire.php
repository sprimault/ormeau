<?php

// Créé par Ormeau, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\PrestataireBase;
use Doctrine\ORM\Mapping as ORM;

/** Absent des valeurs décidées : garde son un-vers-un. */
#[ORM\Entity]
#[ORM\Table(name: 'prestataire', options: ['comment' => 'Absent des valeurs décidées : garde son un-vers-un'])]
class Prestataire extends PrestataireBase {}
