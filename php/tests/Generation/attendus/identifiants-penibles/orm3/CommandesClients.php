<?php

// Créé par Ormeau depuis la base identifiants-penibles, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\CommandesClientsBase;
use Doctrine\ORM\Mapping as ORM;

/** Nom avec espaces, tel que SQL Server l'accepte entre crochets. */
#[ORM\Entity]
#[ORM\Table(
    name: '`Commandes Clients`',
    options: ['comment' => 'Nom avec espaces, tel que SQL Server l\'accepte entre crochets'],
)]
class CommandesClients extends CommandesClientsBase {}
