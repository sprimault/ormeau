<?php

// Créé par Ormeau depuis la base sequences, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\ApostropheBase;
use Doctrine\ORM\Mapping as ORM;

/** séquence libre dont le nom porte une apostrophe, doublée dans le littéral. */
#[ORM\Entity]
#[ORM\Table(
    name: 'apostrophe',
    schema: 'public',
    options: ['comment' => 'séquence libre dont le nom porte une apostrophe, doublée dans le littéral'],
)]
class Apostrophe extends ApostropheBase {}
