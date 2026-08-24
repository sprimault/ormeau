<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\Calque;

use JsonException;
use Ormeau\Doctrine\Calque\CalqueInvalide;
use Ormeau\Doctrine\Calque\LecteurCalque;
use PHPUnit\Framework\Attributes\CoversClass;
use PHPUnit\Framework\TestCase;

/**
 * Le lecteur est la frontière entre les deux moitiés du projet : tout ce que
 * PHP sait du calque passe par lui. Ce qu'on vérifie ici, ce sont ses refus —
 * un calque qu'il accepte à tort produit des entités fausses plus loin, sans
 * rien pour le signaler.
 */
#[CoversClass(LecteurCalque::class)]
final class LecteurCalqueTest extends TestCase
{
    /** @var list<string> */
    private array $temporaires = [];

    /**
     * Supprime les fichiers écrits par le test.
     */
    protected function tearDown(): void
    {
        foreach ($this->temporaires as $chemin) {
            @unlink($chemin);
        }
        $this->temporaires = [];
    }

    /**
     * Un calque complet et à la bonne version se lit sans perte.
     */
    public function testLitUnCalqueValide(): void
    {
        $chemin = $this->fichier([
            'version_ri' => 1,
            'empreinte_physique' => 'sha256:' . str_repeat('a', 64),
            'espace_de_noms' => 'App\\Entity',
            'entites' => [['nom' => 'Client']],
        ]);

        $calque = (new LecteurCalque())->lire($chemin);

        self::assertSame(1, $calque->versionRi);
        self::assertSame('App\\Entity', $calque->espaceDeNoms);
        self::assertCount(1, $calque->entites);
        self::assertSame([], $calque->avertissements);
    }

    /**
     * Une version supérieure est refusée : le paquet ne sait pas ce qu'elle
     * ajoute, et générer au mieux produirait des entités fausses en silence.
     */
    public function testRefuseUneVersionPlusRecente(): void
    {
        $chemin = $this->fichier([
            'version_ri' => LecteurCalque::VERSION_CONNUE + 1,
            'empreinte_physique' => 'sha256:' . str_repeat('a', 64),
            'espace_de_noms' => 'App\\Entity',
            'entites' => [],
        ]);

        $this->expectException(CalqueInvalide::class);
        $this->expectExceptionMessageMatches('/version 2/');

        (new LecteurCalque())->lire($chemin);
    }

    /**
     * Une version antérieure passe : elle ne contient rien d'inconnu.
     */
    public function testAccepteUneVersionAnterieure(): void
    {
        $chemin = $this->fichier([
            'version_ri' => 1,
            'empreinte_physique' => 'sha256:' . str_repeat('b', 64),
            'espace_de_noms' => 'App\\Entity',
            'entites' => [],
        ]);

        self::assertSame(1, (new LecteurCalque())->lire($chemin)->versionRi);
    }

    /**
     * version_ri absente ou non entière : le document n'est pas un calque.
     */
    public function testRefuseUneVersionAbsenteOuNonEntiere(): void
    {
        foreach ([[], ['version_ri' => '1'], ['version_ri' => 1.5]] as $donnees) {
            $chemin = $this->fichier($donnees);

            try {
                (new LecteurCalque())->lire($chemin);
                self::fail('calque accepte alors que version_ri est invalide');
            } catch (CalqueInvalide $e) {
                self::assertStringContainsString('version_ri', $e->getMessage());
            }
        }
    }

    /**
     * Un chemin qui ne mène à rien donne une exception du paquet, pas un
     * avertissement PHP suivi d'une erreur de type incompréhensible.
     */
    public function testRefuseUnFichierAbsent(): void
    {
        $this->expectException(CalqueInvalide::class);
        $this->expectExceptionMessageMatches('/illisible/');

        (new LecteurCalque())->lire(sys_get_temp_dir() . '/ormeau-absent-' . uniqid() . '.json');
    }

    /**
     * Un fichier tronqué remonte en JsonException : le lecteur ne l'habille pas,
     * le message du décodeur est plus précis que ce qu'on écrirait.
     */
    public function testRemonteUnJsonMalForme(): void
    {
        $chemin = tempnam(sys_get_temp_dir(), 'ormeau');
        self::assertNotFalse($chemin);
        $this->temporaires[] = $chemin;
        file_put_contents($chemin, '{"version_ri": 1');

        $this->expectException(JsonException::class);

        (new LecteurCalque())->lire($chemin);
    }

    /**
     * Écrit un calque temporaire et rend son chemin.
     *
     * @param array<string, mixed> $donnees
     */
    private function fichier(array $donnees): string
    {
        $chemin = tempnam(sys_get_temp_dir(), 'ormeau');
        self::assertNotFalse($chemin);
        $this->temporaires[] = $chemin;

        file_put_contents($chemin, json_encode($donnees, JSON_THROW_ON_ERROR));

        return $chemin;
    }
}
