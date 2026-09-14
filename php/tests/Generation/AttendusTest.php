<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\Generation;

use Ormeau\Doctrine\Calque\LecteurCalque;
use Ormeau\Doctrine\Generation\Cible;
use Ormeau\Doctrine\Generation\ControleClasseUtilisateur;
use Ormeau\Doctrine\Generation\Emetteur;
use Ormeau\Doctrine\Generation\GenerateurEntite;
use Ormeau\Doctrine\Generation\Rapport;
use Ormeau\Doctrine\Generation\RenduEntite;
use Ormeau\Doctrine\Generation\TypesPhp;
use PHPUnit\Framework\Attributes\CoversClass;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;

/**
 * Le code produit pour chaque cas de référence, sous chaque cible, comparé
 * octet pour octet à un attendu versionné.
 *
 * Ces attendus sont le vrai contrat du générateur : c'est eux qu'un
 * intégrateur retrouve dans son dépôt, et une modification du rendu doit se
 * relire dans leur diff. Ils ne se régénèrent que derrière
 * ORMEAU_MAJ_ATTENDUS=1, jamais d'eux-mêmes : un attendu réécrit sans être
 * relu ne teste plus rien.
 */
#[CoversClass(GenerateurEntite::class)]
#[CoversClass(RenduEntite::class)]
#[CoversClass(Emetteur::class)]
#[CoversClass(TypesPhp::class)]
#[CoversClass(ControleClasseUtilisateur::class)]
final class AttendusTest extends TestCase
{
    /** Attendus, un répertoire par cas puis par majeure d'ORM. */
    private const ATTENDUS = __DIR__ . '/attendus';

    /**
     * Le rendu d'un cas sous une cible est celui de l'attendu, fichier pour
     * fichier, rapport compris.
     */
    #[DataProvider('cas')]
    public function testLeRenduEstCeluiDeLAttendu(string $cas, int $orm): void
    {
        $sortie = Repertoires::creer();
        try {
            $calque = (new LecteurCalque())->lire(Repertoires::REFERENCES . '/' . $cas . '/logique.json');
            $rapport = (new GenerateurEntite())->generer($calque, $sortie, Cible::forcer($orm), $cas);
            file_put_contents($sortie . '/rapport.txt', self::rapport($rapport, $sortie));

            $obtenu = Repertoires::lire($sortie);
            $attendu = self::ATTENDUS . '/' . $cas . '/orm' . $orm;

            if (getenv('ORMEAU_MAJ_ATTENDUS') === '1') {
                Repertoires::supprimer($attendu);
                foreach ($obtenu as $relatif => $contenu) {
                    @mkdir(dirname($attendu . '/' . $relatif), 0o777, true);
                    file_put_contents($attendu . '/' . $relatif, $contenu);
                }
            }

            self::assertSame(Repertoires::lire($attendu), $obtenu, 'relancer avec ORMEAU_MAJ_ATTENDUS=1, puis relire le diff');
        } finally {
            Repertoires::supprimer($sortie);
        }
    }

    /**
     * Chaque cas sous chaque majeure d'ORM : les deux rendus sont versionnés
     * côte à côte, et leur diff montre exactement ce qu'un changement de cible
     * change.
     *
     * @return iterable<string, array{string, int}>
     */
    public static function cas(): iterable
    {
        foreach (Repertoires::CAS as $cas) {
            foreach (Cible::MAJEURES_ORM as $orm) {
                yield $cas . ' orm' . $orm => [$cas, $orm];
            }
        }
    }

    /**
     * Met le rapport en texte, chemins relatifs à la sortie.
     */
    private static function rapport(Rapport $rapport, string $sortie): string
    {
        $lignes = [];
        foreach ($rapport->fichiers as $fichier) {
            $lignes[] = $fichier->etat->value . ' ' . substr($fichier->chemin, strlen($sortie) + 1);
        }
        foreach ($rapport->ecartees as $ecartee) {
            $lignes[] = 'écartée ' . $ecartee->nom . ' : ' . $ecartee->raison;
        }
        foreach ($rapport->omises as $omise) {
            $lignes[] = 'omise ' . $omise->entite . '::' . $omise->association . ' : ' . $omise->raison;
        }
        foreach ($rapport->divergences as $divergence) {
            $lignes[] = 'à reprendre ' . $divergence->message();
        }

        return implode("\n", $lignes) . "\n";
    }
}
