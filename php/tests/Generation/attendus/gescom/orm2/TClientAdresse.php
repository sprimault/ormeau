<?php

// Créé par Ormeau depuis la base gescom, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\TClientAdresseBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 't_client_adresse', schema: 'gescom')]
class TClientAdresse extends TClientAdresseBase {}
