<?php

// Créé par Ormeau depuis la base heritage-genere, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\DocumentBase;
use Doctrine\ORM\Mapping as ORM;

/** Racine de la hiérarchie, et porte une colonne que la base calcule. */
#[ORM\Entity]
#[ORM\Table(
    name: 'document',
    options: ['comment' => 'Racine de la hiérarchie, et porte une colonne que la base calcule'],
)]
#[ORM\InheritanceType('JOINED')]
#[ORM\DiscriminatorColumn(
    name: 'nature',
    type: 'string',
    length: 1,
    options: ['default' => 'D', 'comment' => 'Nature du document', 'fixed' => true],
)]
#[ORM\DiscriminatorMap(['D' => Document::class, 'F' => Facture::class])]
class Document extends DocumentBase {}
