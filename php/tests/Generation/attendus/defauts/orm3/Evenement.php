<?php

// Créé par Ormeau depuis la base defauts, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\EvenementBase;
use Doctrine\ORM\Mapping as ORM;

/** Défauts calculés, sous les formes que rend pg_get_expr : le sens dépend de l'expression et du type de la colonne. */
#[ORM\Entity]
#[ORM\Table(
    name: 'evenement',
    options: ['comment' => 'Défauts calculés, sous les formes que rend pg_get_expr : le sens dépend de l\'expression et du type de la colonne'],
)]
class Evenement extends EvenementBase {}
