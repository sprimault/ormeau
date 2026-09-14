<?php

// Généré par Ormeau depuis la base enumerations et réécrit à chaque génération.

declare(strict_types=1);

namespace App\Entity\Enum;

enum ClientCanal: string
{
    case Direct = 'DIRECT';
    case Partenaire = 'PARTENAIRE';
    case Web = 'WEB';
}
