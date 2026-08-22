<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Generation;

enum ModeRegeneration: string
{
    case ClasseDeBase = 'classe_de_base';
    case ReecritureAst = 'reecriture_ast';
}
