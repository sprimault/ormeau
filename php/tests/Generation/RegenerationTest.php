<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\Generation;

use Ormeau\Doctrine\Calque\CalqueLogique;
use Ormeau\Doctrine\Generation\Cible;
use Ormeau\Doctrine\Generation\ControleClasseUtilisateur;
use Ormeau\Doctrine\Generation\EtatFichier;
use Ormeau\Doctrine\Generation\GenerateurEntite;
use Ormeau\Doctrine\Generation\Rapport;
use PHPUnit\Framework\Attributes\CoversClass;
use PHPUnit\Framework\TestCase;

/**
 * Régénérer six mois plus tard sans écraser le travail fait entre-temps.
 *
 * C'est la raison d'être de l'outil : la classe de l'utilisateur n'est jamais
 * réécrite, la classe de base suit le calque, et ce qui a divergé se dit avec
 * le fichier et la ligne.
 */
#[CoversClass(GenerateurEntite::class)]
#[CoversClass(ControleClasseUtilisateur::class)]
final class RegenerationTest extends TestCase
{
    /** Répertoire de sortie du test en cours. */
    private string $sortie;

    /**
     * Prépare un répertoire de sortie vide.
     */
    protected function setUp(): void
    {
        $this->sortie = Repertoires::creer();
    }

    /**
     * Supprime ce que le test a écrit.
     */
    protected function tearDown(): void
    {
        Repertoires::supprimer($this->sortie);
    }

    /**
     * Le test qui justifie le mode classe de base : une méthode, un docblock
     * et une propriété non mappée ajoutés à la main survivent à une
     * régénération qui, elle, fait bien évoluer la classe de base.
     */
    public function testLeTravailALaMainSurvitALaRegeneration(): void
    {
        $this->generer(self::calque('client'));
        $client = $this->sortie . '/Client.php';
        $ajout = "    /** Écrit à la main. */\n    private array \$panier = [];\n\n    public function nomAffiche(): string\n    {\n        return 'métier';\n    }\n";
        file_put_contents($client, str_replace(' {}', "\n{\n" . $ajout . "}", (string) file_get_contents($client)));
        $avant = (string) file_get_contents($client);

        $rapport = $this->generer(self::calque('client', avecEmail: true));

        self::assertSame($avant, file_get_contents($client));
        self::assertStringContainsString('protected ?string $email = null;', (string) file_get_contents($this->sortie . '/Base/ClientBase.php'));
        self::assertSame([EtatFichier::Reecrit, EtatFichier::Conserve], array_map(static fn($f) => $f->etat, $rapport->fichiers));
        self::assertSame([], $rapport->divergences);
    }

    /**
     * Deux générations du même calque ne changent rien : la classe de base est
     * laissée intacte, et un outil qui surveille le répertoire ne voit rien
     * bouger.
     */
    public function testUneSecondeGenerationIdentiqueNeTouchePasLesFichiers(): void
    {
        $this->generer(self::calque('client'));
        $base = $this->sortie . '/Base/ClientBase.php';
        touch($base, 1_000_000_000);
        clearstatcache();

        $rapport = $this->generer(self::calque('client'));

        self::assertSame([EtatFichier::Inchange, EtatFichier::Conserve], array_map(static fn($f) => $f->etat, $rapport->fichiers));
        self::assertSame(1_000_000_000, filemtime($base));
    }

    /**
     * Une table renommée ne réécrit pas la classe de l'utilisateur : la
     * divergence nomme le fichier, la ligne, l'attribut tel qu'il est écrit,
     * et le nom attendu.
     */
    public function testUneTableRenommeeSeSignaleSansEcrire(): void
    {
        $this->generer(self::calque('t_clients'));
        $client = $this->sortie . '/Client.php';
        $avant = (string) file_get_contents($client);

        $rapport = $this->generer(self::calque('t_client'));

        self::assertSame($avant, file_get_contents($client));
        self::assertCount(1, $rapport->divergences);
        self::assertSame(
            $client . " ligne 11 : #[ORM\\Table(name: 't_clients')], la table s'appelle maintenant t_client",
            $rapport->divergences[0]->message(),
        );
    }

    /**
     * Une classe de l'utilisateur réécrite à la main sans l'héritage ni la
     * table se signale sur chaque point, sans rien écrire.
     */
    public function testUneClasseSansHeritageNiTableSeSignale(): void
    {
        $this->generer(self::calque('client'));
        file_put_contents($this->sortie . '/Client.php', "<?php\n\nnamespace App\\Entity;\n\nclass Client\n{\n}\n");

        $messages = array_map(static fn($d): string => $d->attendu, $this->generer(self::calque('client'))->divergences);

        self::assertSame([
            'Client n\'hérite pas de App\Entity\Base\ClientBase : la classe de base régénérée ne s\'applique pas',
            'attribut #[ORM\Entity] absent',
            'attribut #[ORM\Table] absent, la table s\'appelle client',
        ], $messages);
    }

    /**
     * Un alias différent, ou un attribut sur plusieurs lignes, se lit comme
     * celui que la génération écrit : le contrôle résout les noms, il ne
     * cherche pas une chaîne.
     */
    public function testLeControleResoutLesAlias(): void
    {
        $this->generer(self::calque('client'));
        file_put_contents($this->sortie . '/Client.php', <<<'PHP'
            <?php

            namespace App\Entity;

            use App\Entity\Base\ClientBase;
            use Doctrine\ORM\Mapping\Entity;
            use Doctrine\ORM\Mapping\Table as Tableau;

            #[Entity(repositoryClass: ClientRepository::class)]
            #[Tableau(
                name: 'client',
            )]
            class Client extends ClientBase
            {
            }
            PHP);

        self::assertSame([], $this->generer(self::calque('client'))->divergences);
    }

    /**
     * Génère un calque dans le répertoire du test, sous ORM 3.
     */
    private function generer(CalqueLogique $calque): Rapport
    {
        return (new GenerateurEntite())->generer($calque, $this->sortie, Cible::forcer(3));
    }

    /**
     * Un calque d'une entité Client sur la table donnée, avec ou sans colonne
     * email.
     */
    private static function calque(string $table, bool $avecEmail = false): CalqueLogique
    {
        $proprietes = [['nom' => 'id', 'colonne' => 'id', 'type_php' => 'int', 'type_doctrine' => 'integer', 'nullable' => false]];
        if ($avecEmail) {
            $proprietes[] = ['nom' => 'email', 'colonne' => 'email', 'type_php' => '?string', 'type_doctrine' => 'string', 'nullable' => true];
        }

        return CalqueLogique::depuisTableau([
            'version_ri' => 1,
            'empreinte_physique' => 'sha256:' . str_repeat('a', 64),
            'espace_de_noms' => 'App\\Entity',
            'entites' => [[
                'nom' => 'Client',
                'table' => ['nom' => $table, 'schema' => 'public'],
                'identifiant' => ['proprietes' => ['id'], 'strategie' => 'identite'],
                'proprietes' => $proprietes,
            ]],
        ]);
    }
}
