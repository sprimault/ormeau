<?php

// Créé par Ormeau depuis la base texte-illimite, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\FicheBase;
use Doctrine\ORM\Mapping as ORM;

/** Textes sans longueur déclarée : text quand Doctrine sait le recréer, chaîne et avertissement quand il le recréerait sans Unicode. */
#[ORM\Entity]
#[ORM\Table(
    name: 'fiche',
    options: ['comment' => 'Textes sans longueur déclarée : text quand Doctrine sait le recréer, chaîne et avertissement quand il le recréerait sans Unicode'],
)]
class Fiche extends FicheBase {}
