<?php

// Créé par Ormeau depuis la base relations-forcees, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\ProfilBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 'profil')]
class Profil extends ProfilBase {}
