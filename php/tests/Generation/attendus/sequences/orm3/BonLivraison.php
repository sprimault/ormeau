<?php

// Créé par Ormeau depuis la base sequences, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\BonLivraisonBase;
use Doctrine\ORM\Mapping as ORM;

/** bigserial dans un schéma entre guillemets : PostgreSQL cite le nom qualifié. */
#[ORM\Entity]
#[ORM\Table(
    name: '`Bon_Livraison`',
    schema: 'Compta',
    options: ['comment' => 'bigserial dans un schéma entre guillemets : PostgreSQL cite le nom qualifié'],
)]
class BonLivraison extends BonLivraisonBase {}
