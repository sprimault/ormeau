<?php

// Créé par Ormeau depuis la base sequences, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\EcritureBase;
use Doctrine\ORM\Mapping as ORM;

/** serial dans un schéma Public distinct de public : le préfixe entre guillemets reste. */
#[ORM\Entity]
#[ORM\Table(
    name: '`ecriture`',
    schema: 'Public',
    options: ['comment' => 'serial dans un schéma Public distinct de public : le préfixe entre guillemets reste'],
)]
class Ecriture extends EcritureBase {}
