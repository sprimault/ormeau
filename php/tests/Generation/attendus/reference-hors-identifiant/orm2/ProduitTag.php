<?php

// Créé par Ormeau depuis la base reference-hors-identifiant, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\ProduitTagBase;
use Doctrine\ORM\Mapping as ORM;

/** forme d'une table de jointure, mais une clé vise tag.libelle : pas de plusieurs-vers-plusieurs. */
#[ORM\Entity]
#[ORM\Table(
    name: 'produit_tag',
    options: ['comment' => 'forme d\'une table de jointure, mais une clé vise tag.libelle : pas de plusieurs-vers-plusieurs'],
)]
class ProduitTag extends ProduitTagBase {}
