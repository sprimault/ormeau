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
            $client . " ligne 13 : #[ORM\\Table(name: 't_clients')], la table s'appelle maintenant t_client",
            $rapport->divergences[0]->message(),
        );
    }

    /**
     * Un commentaire de table ajouté, changé ou retiré en base se signale sur
     * #[ORM\Table], qui le porte pour schema:create ; le docblock de la classe
     * appartient à l'utilisateur et n'est pas relu.
     */
    public function testUnCommentaireDeTableChangeSeSignale(): void
    {
        $this->generer(self::calque('client', commentaire: 'Fiche client'));
        $client = $this->sortie . '/Client.php';
        $actuel = "#[ORM\\Table(name: 'client', options: ['comment' => 'Fiche client'])]";

        self::assertSame(
            [$client . ' ligne 14 : ' . $actuel . ", le commentaire de la table est maintenant 'Fiche d\\'un client'"],
            array_map(static fn($d): string => $d->message(), $this->generer(self::calque('client', commentaire: "Fiche d'un client"))->divergences),
        );
        self::assertSame(
            [$client . ' ligne 14 : ' . $actuel . ', la table n\'a plus de commentaire'],
            array_map(static fn($d): string => $d->message(), $this->generer(self::calque('client'))->divergences),
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
     * Un héritage décidé après une première génération ne touche pas la
     * classe de la racine, déjà écrite : chaque attribut qui lui manque se
     * signale en entier, prêt à recopier. La classe fille n'a rien à
     * reprendre, sa classe de base régénérée hérite désormais de la racine.
     */
    public function testUnHeritageDecideApresCoupSeSignaleSurLaRacine(): void
    {
        $this->generer(self::hierarchie([]));
        $personne = $this->sortie . '/Personne.php';
        $avant = (string) file_get_contents($personne);

        $rapport = $this->generer(self::hierarchie(['Salarie' => 'S']));

        self::assertSame($avant, file_get_contents($personne));
        self::assertStringContainsString('abstract class SalarieBase extends Personne', (string) file_get_contents($this->sortie . '/Base/SalarieBase.php'));
        self::assertSame([
            $personne . " ligne 12 : attribut #[ORM\\InheritanceType('JOINED')] absent, Personne est la racine d'un héritage",
            $personne . " ligne 12 : attribut #[ORM\\DiscriminatorColumn(name: 'nature', type: 'string', length: 1)] absent, Personne est la racine d'un héritage",
            $personne . " ligne 12 : attribut #[ORM\\DiscriminatorMap(['P' => Personne::class, 'S' => Salarie::class])] absent, Personne est la racine d'un héritage",
        ], array_map(static fn($d): string => $d->message(), $rapport->divergences));
    }

    /**
     * Une classe fille déclarée après la création de la racine change sa
     * carte, et seule la carte se signale ; un héritage retiré laisse des
     * attributs qui ne désignent plus rien.
     */
    public function testUneCarteQuiNeCorrespondPlusSeSignale(): void
    {
        $this->generer(self::hierarchie(['Salarie' => 'S']));
        $personne = $this->sortie . '/Personne.php';
        $carte = "#[ORM\\DiscriminatorMap(['P' => Personne::class, 'S' => Salarie::class])]";

        self::assertSame(
            [$personne . ' ligne 16 : ' . $carte . ", attendu #[ORM\\DiscriminatorMap(['P' => Personne::class, 'S' => Salarie::class, 'C' => Cadre::class])]"],
            array_map(static fn($d): string => $d->message(), $this->generer(self::hierarchie(['Salarie' => 'S', 'Cadre' => 'C']))->divergences),
        );
        self::assertSame(
            array_fill(0, 3, 'Personne n\'est la racine d\'aucun héritage déclaré, attribut à retirer'),
            array_map(static fn($d): string => $d->attendu, $this->generer(self::hierarchie([]))->divergences),
        );
    }

    /**
     * Une classe de l'utilisateur rangée dans un sous-répertoire est retrouvée
     * par sa classe de base : rien n'est recréé à la racine, et la classe de
     * base qui la vise importe son nom qualifié réel.
     */
    public function testUneClasseRangeeEstRetrouvee(): void
    {
        $this->generer(self::ventes());
        $client = Repertoires::ranger($this->sortie, 'Client', 'Ventes');

        $rapport = $this->generer(self::ventes());

        self::assertFileDoesNotExist($this->sortie . '/Client.php');
        self::assertSame([], $rapport->divergences);
        self::assertContains([$client, EtatFichier::Conserve], array_map(static fn($f): array => [$f->chemin, $f->etat], $rapport->fichiers));
        self::assertStringContainsString("use App\\Entity\\Ventes\\Client;\n", (string) file_get_contents($this->sortie . '/Base/CommandeBase.php'));
    }

    /**
     * Une hiérarchie rangée d'un bloc garde sa carte, écrite en noms courts ;
     * une classe fille rangée seule ne correspond plus à la carte de la
     * racine, qui se signale avec la classe qualifiée attendue.
     */
    public function testUneHierarchieRangeeSuitSesClasses(): void
    {
        $this->generer(self::hierarchie(['Salarie' => 'S']));
        Repertoires::ranger($this->sortie, 'Salarie', 'Rh');

        self::assertSame(
            ["attendu #[ORM\\DiscriminatorMap(['P' => Personne::class, 'S' => \\App\\Entity\\Rh\\Salarie::class])]"],
            array_map(static fn($d): string => $d->attendu, $this->generer(self::hierarchie(['Salarie' => 'S']))->divergences),
        );

        Repertoires::ranger($this->sortie, 'Personne', 'Rh');
        $rapport = $this->generer(self::hierarchie(['Salarie' => 'S']));

        self::assertSame([], $rapport->divergences);
        self::assertStringContainsString(
            "use App\\Entity\\Rh\\Personne;\n",
            (string) file_get_contents($this->sortie . '/Base/SalarieBase.php'),
        );
    }

    /**
     * Deux classes pour la même entité écartent l'entité, avec les deux
     * fichiers : choisir serait deviner laquelle le projet utilise. L'entité
     * qui la vise est écartée à son tour.
     */
    public function testDeuxClassesPourUneEntiteLEcartent(): void
    {
        $this->generer(self::ventes());
        $copie = Repertoires::ranger($this->sortie, 'Client', 'Ventes', copier: true);

        $rapport = $this->generer(self::ventes());

        self::assertSame([
            'Client' => sprintf('plusieurs classes Client héritent de ClientBase, n\'en garder qu\'une : %s/Client.php, %s', $this->sortie, $copie),
            'Commande' => 'l\'association client vise Client (public.client), écartée',
        ], array_column(array_map(static fn($e): array => [$e->nom, $e->raison], $rapport->ecartees), 1, 0));
    }

    /**
     * Génère un calque dans le répertoire du test, sous ORM 3.
     */
    private function generer(CalqueLogique $calque): Rapport
    {
        return (new GenerateurEntite())->generer($calque, $this->sortie, Cible::forcer(3));
    }

    /**
     * Un calque d'un Client et d'une Commande qui le vise.
     */
    private static function ventes(): CalqueLogique
    {
        $id = ['nom' => 'id', 'colonne' => 'id', 'type_php' => 'int', 'type_doctrine' => 'integer', 'nullable' => false];

        return CalqueLogique::depuisTableau([
            'version_ri' => 1,
            'empreinte_physique' => 'sha256:' . str_repeat('a', 64),
            'espace_de_noms' => 'App\\Entity',
            'entites' => [
                [
                    'nom' => 'Client',
                    'table' => ['nom' => 'client', 'schema' => 'public'],
                    'identifiant' => ['proprietes' => ['id'], 'strategie' => 'identite'],
                    'proprietes' => [$id],
                ],
                [
                    'nom' => 'Commande',
                    'table' => ['nom' => 'commande', 'schema' => 'public'],
                    'identifiant' => ['proprietes' => ['id'], 'strategie' => 'identite'],
                    'proprietes' => [$id, ['nom' => 'clientId', 'colonne' => 'client_id', 'type_php' => 'int', 'type_doctrine' => 'integer', 'nullable' => false]],
                    'associations' => [[
                        'nom' => 'client', 'genre' => 'plusieurs_vers_un', 'cible' => 'Client', 'proprietaire' => true, 'origine' => 'contrainte',
                        'jointure' => [['colonne' => 'client_id', 'colonne_referencee' => 'id', 'nullable' => false]],
                    ]],
                ],
            ],
        ]);
    }

    /**
     * Un calque d'une Personne et de classes filles directes, déclarées
     * héritières avec leur valeur discriminante ; sans classe fille, aucun
     * héritage.
     *
     * @param array<string, string> $filles valeur discriminante de chaque classe fille, par nom
     */
    private static function hierarchie(array $filles): CalqueLogique
    {
        $id = ['nom' => 'id', 'colonne' => 'id', 'type_php' => 'int', 'type_doctrine' => 'integer', 'nullable' => false];
        $entites = [[
            'nom' => 'Personne',
            'table' => ['nom' => 'personne', 'schema' => 'public'],
            'identifiant' => ['proprietes' => ['id'], 'strategie' => 'identite'],
            'proprietes' => [$id, ['nom' => 'nature', 'colonne' => 'nature', 'type_php' => 'string', 'type_doctrine' => 'string', 'nullable' => false, 'longueur' => 1]],
        ]];
        if ($filles !== []) {
            $entites[0]['valeur_discriminante'] = 'P';
        }
        foreach ($filles as $nom => $valeur) {
            $entites[] = [
                'nom' => $nom,
                'table' => ['nom' => strtolower($nom), 'schema' => 'public'],
                'heritage' => ['strategie' => 'jointe', 'parent' => 'Personne', 'colonne_discriminante' => 'nature', 'origine' => 'decision'],
                'valeur_discriminante' => $valeur,
                'identifiant' => ['proprietes' => ['id'], 'strategie' => 'assignee'],
                'proprietes' => [$id],
            ];
        }

        return CalqueLogique::depuisTableau([
            'version_ri' => 1,
            'empreinte_physique' => 'sha256:' . str_repeat('a', 64),
            'espace_de_noms' => 'App\\Entity',
            'entites' => $entites,
        ]);
    }

    /**
     * Un calque d'une entité Client sur la table donnée, avec ou sans colonne
     * email.
     */
    private static function calque(string $table, bool $avecEmail = false, ?string $commentaire = null): CalqueLogique
    {
        $proprietes = [['nom' => 'id', 'colonne' => 'id', 'type_php' => 'int', 'type_doctrine' => 'integer', 'nullable' => false]];
        if ($avecEmail) {
            $proprietes[] = ['nom' => 'email', 'colonne' => 'email', 'type_php' => '?string', 'type_doctrine' => 'string', 'nullable' => true];
        }

        return CalqueLogique::depuisTableau([
            'version_ri' => 1,
            'empreinte_physique' => 'sha256:' . str_repeat('a', 64),
            'espace_de_noms' => 'App\\Entity',
            'entites' => [array_filter([
                'nom' => 'Client',
                'table' => ['nom' => $table, 'schema' => 'public'],
                'identifiant' => ['proprietes' => ['id'], 'strategie' => 'identite'],
                'proprietes' => $proprietes,
                'commentaire' => $commentaire,
            ], static fn($valeur): bool => $valeur !== null)],
        ]);
    }
}
