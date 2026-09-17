<?php

// Créé par Ormeau depuis la base json-verifie, jamais réécrit : ce fichier appartient au projet.

declare(strict_types=1);

namespace App\Entity;

use App\Entity\Base\DocumentBase;
use Doctrine\ORM\Mapping as ORM;

/** Textes qu'une vérification déclare JSON, sous les formes que rend sys.check_constraints : seul un objet ou un tableau garanti devient un tableau PHP. */
#[ORM\Entity]
#[ORM\Table(
    name: 'document',
    options: ['comment' => 'Textes qu\'une vérification déclare JSON, sous les formes que rend sys.check_constraints : seul un objet ou un tableau garanti devient un tableau PHP'],
)]
class Document extends DocumentBase {}
