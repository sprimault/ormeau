<?php

// Généré par Ormeau depuis la base gescom et réécrit à chaque génération.

declare(strict_types=1);

namespace App\Entity\Enum;

enum Canal: string
{
    case Web = 'web';
    case Telephone = 'telephone';
    case Agence = 'agence';
}
