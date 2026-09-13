<?php

// Généré par Ormeau et réécrit à chaque génération.

declare(strict_types=1);

namespace App\Entity\Enum;

enum StatutClient: string
{
    case Prospect = 'PROSPECT';
    case Actif = 'ACTIF';
    case Suspendu = 'SUSPENDU';
    case Radie = 'RADIE';
}
