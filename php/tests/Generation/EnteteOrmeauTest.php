<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\Generation;

use Ormeau\Doctrine\Generation\EnteteOrmeau;
use PHPUnit\Framework\Attributes\CoversClass;
use PHPUnit\Framework\TestCase;

/**
 * L'en-tête s'écrit et se relit par le même format : un fichier de la base
 * voisine qui ne serait plus reconnu se réécrirait en silence.
 */
#[CoversClass(EnteteOrmeau::class)]
final class EnteteOrmeauTest extends TestCase
{
    /**
     * Chaque en-tête écrit se relit en sa base, nom à point ou à tiret compris.
     */
    public function testChaqueEnteteSeRelitEnSaBase(): void
    {
        foreach (['gescom', 'gescom.v2', 'paie-2026', 'référentiel'] as $base) {
            foreach ([EnteteOrmeau::outil($base), EnteteOrmeau::classeBase($base, 'Client'), EnteteOrmeau::utilisateur($base)] as $entete) {
                self::assertSame($base, EnteteOrmeau::base("<?php\n\n// " . $entete . "\n\ndeclare(strict_types=1);\n"), $entete);
            }
        }
    }

    /**
     * Un fichier de la 0.5.0 ne nomme aucune base, et un fichier écrit sous
     * Windows se relit malgré ses fins de ligne.
     */
    public function testUnEnteteSansBaseRendNull(): void
    {
        self::assertNull(EnteteOrmeau::base("<?php\n\n// Généré par Ormeau et réécrit à chaque génération.\n"));
        self::assertNull(EnteteOrmeau::base("<?php\n\n// Créé par Ormeau, jamais réécrit : ce fichier appartient au projet.\n"));
        self::assertSame('gescom', EnteteOrmeau::base("<?php\r\n\r\n// " . EnteteOrmeau::outil('gescom') . "\r\n"));
    }

    /**
     * Seule la ligne d'en-tête compte : un commentaire venu de la base qui
     * reprendrait la formule plus bas ne fait pas croire à une autre base.
     */
    public function testSeuleLaLigneDEnteteEstLue(): void
    {
        $source = "<?php\n\n// Généré par Ormeau et réécrit à chaque génération.\n\n/** " . EnteteOrmeau::outil('paie') . " */\n";

        self::assertNull(EnteteOrmeau::base($source));
    }
}
