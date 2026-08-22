<?php

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

enum ModeRegeneration: string
{
    case ClasseDeBase = 'classe_de_base';
    case ReecritureAst = 'reecriture_ast';
}
