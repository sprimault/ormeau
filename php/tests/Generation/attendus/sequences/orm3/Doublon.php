<?php

// Créé par Ormeau depuis la base sequences, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\DoublonBase;
use Doctrine\ORM\Mapping as ORM;

/** calque ancien : nom nu d'une séquence présente dans deux schémas, rien n'est rattaché. */
#[ORM\Entity]
#[ORM\Table(
    name: 'doublon',
    schema: 'public',
    options: ['comment' => 'calque ancien : nom nu d\'une séquence présente dans deux schémas, rien n\'est rattaché'],
)]
class Doublon extends DoublonBase {}
