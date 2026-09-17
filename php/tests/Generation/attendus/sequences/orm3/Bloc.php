<?php

// Créé par Ormeau depuis la base sequences, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\BlocBase;
use Doctrine\ORM\Mapping as ORM;

/** séquence incrémentée de 10 à partir de 5 : blocs réservés ou nœuds décalés, la base ne dit pas lequel. */
#[ORM\Entity]
#[ORM\Table(
    name: 'bloc',
    options: ['comment' => 'séquence incrémentée de 10 à partir de 5 : blocs réservés ou nœuds décalés, la base ne dit pas lequel'],
)]
class Bloc extends BlocBase {}
