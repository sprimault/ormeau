<?php

// Généré par Ormeau depuis la base gescom et réécrit à chaque génération.

declare(strict_types=1);

namespace App\Entity\Enum;

enum TClientCliStatut: string
{
    case Actif = 'ACTIF';
    case Suspendu = 'SUSPENDU';
    case Archive = 'ARCHIVE';
}
