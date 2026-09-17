<?php

// Créé par Ormeau depuis la base gescom, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\TClientGrandCompteBase;
use Doctrine\ORM\Mapping as ORM;

#[ORM\Entity]
#[ORM\Table(name: 't_client_grand_compte', schema: 'gescom')]
class TClientGrandCompte extends TClientGrandCompteBase {}
