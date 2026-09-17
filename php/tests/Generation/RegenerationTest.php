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
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;
use RuntimeException;

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
     * Deux bases dans le même répertoire : la seconde ne réécrit pas la classe
     * de base de la première, et le refus nomme le fichier et les deux bases.
     */
    public function testUneAutreBaseNEcrasePasLesFichiersDeLaPremiere(): void
    {
        $this->generer(self::calque('client'));
        $base = $this->sortie . '/Base/ClientBase.php';
        $avant = (string) file_get_contents($base);

        $rapport = $this->generer(self::calque('client', avecEmail: true), 'paie');

        self::assertSame($avant, file_get_contents($base));
        self::assertSame([[$base, 'gescom', 'paie']], array_map(static fn($r): array => [$r->fichier, $r->baseExistante, $r->baseDemandee], $rapport->refus));
        self::assertSame(['Client' => $base . ' vient de la base gescom'], array_column(array_map(static fn($e): array => [$e->nom, $e->raison], $rapport->ecartees), 1, 0));
        self::assertSame([], $rapport->fichiers);
    }

    /**
     * Remplacer une base nommée réécrit ses fichiers, et l'en-tête prend la
     * nouvelle base : la génération suivante passe sans l'option, la classe de
     * base faisant foi et non la classe de l'utilisateur, écrite une fois.
     */
    public function testRemplacerUneBaseNommeeReecritEtRenommeLEntete(): void
    {
        $this->generer(self::calque('client'));

        self::assertSame([], $this->generer(self::calque('client', avecEmail: true), 'gescom-v2', ['gescom'])->refus);
        self::assertStringStartsWith(
            "<?php\n\n// Généré par Ormeau depuis la base gescom-v2 et réécrit",
            (string) file_get_contents($this->sortie . '/Base/ClientBase.php'),
        );
        self::assertStringContainsString('depuis la base gescom,', (string) file_get_contents($this->sortie . '/Client.php'));
        self::assertSame([], $this->generer(self::calque('client', avecEmail: true), 'gescom-v2')->refus);
    }

    /**
     * Un fichier de la 0.5.0 ne nomme aucune base : il est repris, sans quoi
     * tout projet existant serait bloqué, et reçoit la base qui l'écrit.
     */
    public function testUnFichierSansBaseEstRepris(): void
    {
        $this->generer(self::calque('client'));
        $base = $this->sortie . '/Base/ClientBase.php';
        file_put_contents($base, str_replace(
            '// Généré par Ormeau depuis la base gescom et réécrit à chaque génération',
            '// Généré par Ormeau et réécrit à chaque génération',
            (string) file_get_contents($base),
        ));

        $rapport = $this->generer(self::calque('client'), 'paie');

        self::assertSame([], $rapport->refus);
        self::assertStringContainsString('depuis la base paie et réécrit', (string) file_get_contents($base));
    }

    /**
     * Sans classe de base, la classe de l'utilisateur d'une autre base fait
     * foi : recréer la classe de base sous elle la ferait hériter d'une classe
     * qui n'est pas la sienne.
     */
    public function testUneClasseDeLUtilisateurDUneAutreBaseSansClasseDeBaseSeRefuse(): void
    {
        $this->generer(self::calque('client'));
        unlink($this->sortie . '/Base/ClientBase.php');

        $rapport = $this->generer(self::calque('client'), 'paie');

        self::assertSame([$this->sortie . '/Client.php'], array_map(static fn($r): string => $r->fichier, $rapport->refus));
        self::assertFileDoesNotExist($this->sortie . '/Base/ClientBase.php');
    }

    /**
     * Une énumération d'une autre base n'est pas réécrite, et l'entité qui
     * s'en sert est écartée par le chemin habituel.
     */
    public function testUneEnumerationDUneAutreBaseSeRefuse(): void
    {
        $calque = static fn(): CalqueLogique => CalqueLogique::depuisTableau([
            'version_ri' => 2,
            'empreinte_physique' => 'sha256:' . str_repeat('a', 64),
            'espace_de_noms' => 'App\\Entity',
            'entites' => [[
                'nom' => 'Commande',
                'table' => ['nom' => 'commande', 'schema' => 'public'],
                'identifiant' => ['proprietes' => ['id'], 'strategie' => 'identite'],
                'proprietes' => [
                    ['nom' => 'id', 'colonne' => 'id', 'type_doctrine' => 'integer', 'nullable' => false],
                    ['nom' => 'etat', 'colonne' => 'etat', 'type_doctrine' => 'string', 'nullable' => false, 'enumeration' => 'Etat'],
                ],
            ]],
            'enumerations' => [['nom' => 'Etat', 'type_support' => 'string', 'cas' => [['nom' => 'Ouvert', 'valeur' => 'O']], 'origine' => 'verification']],
        ]);
        $this->generer($calque());
        $enumeration = $this->sortie . '/Enum/Etat.php';

        $rapport = $this->generer($calque(), 'paie');

        self::assertSame([$enumeration], array_map(static fn($r): string => $r->fichier, $rapport->refus));
        self::assertSame(
            ['Commande' => 'l\'énumération Etat n\'est pas générée : ' . $enumeration . ' vient de la base gescom'],
            array_column(array_map(static fn($e): array => [$e->nom, $e->raison], $rapport->ecartees), 1, 0),
        );
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
     * Une classe créée en 0.5.0 dans un schéma à citer porte des backticks sur
     * le schéma, forme que PostgreSQL refusait : la divergence le dit, pour
     * que la régénération ne passe pas pour la cause. Même chose pour un mot
     * réservé écrit nu sans schéma.
     */
    public function testUneFormeCasseeEn050SeSignaleCommeInutilisable(): void
    {
        $this->generer(self::schemas());
        $avoir = $this->sortie . '/Avoir.php';
        file_put_contents($avoir, str_replace(
            "#[ORM\\Table(name: '`avoir`', schema: 'Compta')]",
            "#[ORM\\Table(name: 'avoir', schema: '`Compta`')]",
            (string) file_get_contents($avoir),
        ));
        $avant = (string) file_get_contents($avoir);

        $messages = array_map(static fn($d): string => $d->message(), $this->generer(self::schemas())->divergences);

        self::assertSame($avant, file_get_contents($avoir));
        $actuel = $avoir . " ligne 13 : #[ORM\\Table(name: 'avoir', schema: '`Compta`')], ";
        $cassee = ' ; forme écrite par la 0.5.0, que PostgreSQL refuse : cette table était inutilisable';
        self::assertSame([
            $actuel . 'la table s\'appelle maintenant `avoir`' . $cassee,
            $actuel . 'la table est dans le schéma Compta' . $cassee,
        ], $messages);

        Repertoires::supprimer($this->sortie);
        $this->sortie = Repertoires::creer();
        $this->generer(self::calque('user'));
        $client = $this->sortie . '/Client.php';
        file_put_contents($client, str_replace("name: '`user`'", "name: 'user'", (string) file_get_contents($client)));

        self::assertSame(
            [$client . " ligne 13 : #[ORM\\Table(name: 'user')], la table s'appelle maintenant `user`" . $cassee],
            array_map(static fn($d): string => $d->message(), $this->generer(self::calque('user'))->divergences),
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
     * Un fichier PHP illisible sous le répertoire des entités arrête la
     * génération avant la première écriture, où qu'il soit : à la racine, rangé,
     * ou à côté de la vraie classe rangée ailleurs. Écarter la seule entité
     * réécrirait les classes de base qui la visent sans leur côté inverse, et
     * son nom qualifié ne se lit pas dans un fichier qu'on ne sait pas lire.
     *
     * @return iterable<string, array{callable(string): string}>
     */
    public static function fichiersIllisibles(): iterable
    {
        $casser = static function (string $chemin): void {
            file_put_contents($chemin, "<?php\n\nnamespace App\\Entity;\n\nclass Client extends {\n");
        };

        yield 'à la racine' => [static function (string $sortie) use ($casser): string {
            $casser($sortie . '/Client.php');

            return 'Client.php';
        }];
        yield 'rangé en sous-répertoire' => [static function (string $sortie) use ($casser): string {
            $casser(Repertoires::ranger($sortie, 'Client', 'Ventes'));

            return 'Ventes/Client.php';
        }];
        yield 'à côté de la classe rangée' => [static function (string $sortie) use ($casser): string {
            Repertoires::ranger($sortie, 'Client', 'Ventes');
            $casser($sortie . '/Client.php');

            return 'Client.php';
        }];
    }

    /**
     * Le refus nomme le fichier relatif au répertoire des entités, sa ligne, et
     * pourquoi ce fichier-là compte ; rien n'est écrit, pas même une classe de
     * base.
     *
     * @param callable(string): string $preparer casse un fichier et rend son chemin relatif
     */
    #[DataProvider('fichiersIllisibles')]
    public function testUnFichierIllisibleArreteLaGenerationSansRienEcrire(callable $preparer): void
    {
        $this->generer(self::ventes());
        $relatif = $preparer($this->sortie);
        $avant = Repertoires::lire($this->sortie);

        try {
            $this->generer(self::ventes());
            self::fail('la génération est passée malgré un fichier illisible');
        } catch (RuntimeException $e) {
            self::assertStringContainsString($relatif . ' ligne 5', $e->getMessage());
            self::assertStringContainsString('rien n\'est écrit', $e->getMessage());
            self::assertStringContainsString('répertoire des entités', $e->getMessage());
        }
        self::assertSame($avant, Repertoires::lire($this->sortie));
    }

    /**
     * Une classe renommée dans son fichier n'est plus retrouvée : le fichier
     * reste à sa place, et le contrôle dit qu'il n'y trouve pas la classe
     * attendue au lieu d'en créer une seconde.
     */
    public function testUneClasseRenommeeDansSonFichierSeSignale(): void
    {
        $this->generer(self::calque('t_client'));
        $client = $this->sortie . '/Client.php';
        file_put_contents($client, str_replace('class Client extends', 'class Cliente extends', (string) file_get_contents($client)));
        $avant = (string) file_get_contents($client);

        $rapport = $this->generer(self::calque('t_client'));

        self::assertSame($avant, file_get_contents($client));
        self::assertSame(
            [$client . ' ligne 1 : classe Client introuvable, la génération ne peut pas s\'y appuyer'],
            array_map(static fn($d): string => $d->message(), $rapport->divergences),
        );
    }

    /**
     * Sur un calque à plusieurs schémas, une seconde génération identique ne
     * signale rien ; une table passée dans un autre schéma se signale sur
     * #[ORM\Table], qui porte l'argument schema.
     */
    public function testUnSchemaChangeSeSignale(): void
    {
        $this->generer(self::schemas());
        self::assertSame([], $this->generer(self::schemas())->divergences);

        $avoir = $this->sortie . '/Avoir.php';
        self::assertSame(
            [$avoir . " ligne 13 : #[ORM\\Table(name: '`avoir`', schema: 'Compta')], la table est dans le schéma Ventes"],
            array_map(static fn($d): string => $d->message(), $this->generer(self::schemas('Ventes'))->divergences),
        );
    }

    /**
     * Une colonne discriminante dont le nom ou la longueur ont changé dans la
     * classe de la racine se signale, avec l'attribut attendu en entier.
     */
    public function testUneColonneDiscriminanteChangeeSeSignale(): void
    {
        $this->generer(self::hierarchie(['Salarie' => 'S']));
        $personne = $this->sortie . '/Personne.php';
        $attendu = "attendu #[ORM\\DiscriminatorColumn(name: 'nature', type: 'string', length: 1)]";

        foreach (["name: 'genre', type: 'string', length: 1", "name: 'nature', type: 'string', length: 2"] as $ecrit) {
            file_put_contents($personne, (string) preg_replace(
                "/#\\[ORM\\\\DiscriminatorColumn\\([^)]*\\)\\]/",
                '#[ORM\\DiscriminatorColumn(' . $ecrit . ')]',
                (string) file_get_contents($personne),
            ));

            self::assertSame(
                [$attendu],
                array_map(static fn($d): string => $d->attendu, $this->generer(self::hierarchie(['Salarie' => 'S']))->divergences),
                $ecrit,
            );
        }
    }

    /**
     * Génère un calque dans le répertoire du test, sous ORM 3, depuis la base
     * donnée, gescom à défaut.
     *
     * @param list<string> $remplacables bases dont les fichiers peuvent être réécrits
     */
    private function generer(CalqueLogique $calque, string $base = 'gescom', array $remplacables = []): Rapport
    {
        return (new GenerateurEntite())->generer($calque, $this->sortie, Cible::forcer(3), $base, $remplacables);
    }

    /**
     * Un calque d'un Client et d'une Commande qui le vise.
     */
    private static function ventes(): CalqueLogique
    {
        $id = ['nom' => 'id', 'colonne' => 'id', 'type_doctrine' => 'integer', 'nullable' => false];

        return CalqueLogique::depuisTableau([
            'version_ri' => 2,
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
                    'proprietes' => [$id, ['nom' => 'clientId', 'colonne' => 'client_id', 'type_doctrine' => 'integer', 'nullable' => false]],
                    'associations' => [[
                        'nom' => 'client', 'genre' => 'plusieurs_vers_un', 'cible' => 'Client', 'proprietaire' => true, 'origine' => 'contrainte',
                        'jointure' => [['colonne' => 'client_id', 'colonne_referencee' => 'id', 'nullable' => false]],
                    ]],
                ],
            ],
        ]);
    }

    /**
     * Un calque sur deux schémas : un Avoir dans Compta à défaut, schéma à
     * citer, et un Client dans public.
     */
    private static function schemas(string $schemaAvoir = 'Compta'): CalqueLogique
    {
        $id = ['nom' => 'id', 'colonne' => 'id', 'type_doctrine' => 'integer', 'nullable' => false];

        return CalqueLogique::depuisTableau([
            'version_ri' => 2,
            'empreinte_physique' => 'sha256:' . str_repeat('a', 64),
            'espace_de_noms' => 'App\\Entity',
            'entites' => [
                [
                    'nom' => 'Avoir',
                    'table' => ['nom' => 'avoir', 'schema' => $schemaAvoir],
                    'identifiant' => ['proprietes' => ['id'], 'strategie' => 'identite'],
                    'proprietes' => [$id],
                ],
                [
                    'nom' => 'Client',
                    'table' => ['nom' => 'client', 'schema' => 'public'],
                    'identifiant' => ['proprietes' => ['id'], 'strategie' => 'identite'],
                    'proprietes' => [$id],
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
        $id = ['nom' => 'id', 'colonne' => 'id', 'type_doctrine' => 'integer', 'nullable' => false];
        $entites = [[
            'nom' => 'Personne',
            'table' => ['nom' => 'personne', 'schema' => 'public'],
            'identifiant' => ['proprietes' => ['id'], 'strategie' => 'identite'],
            'proprietes' => [$id, ['nom' => 'nature', 'colonne' => 'nature', 'type_doctrine' => 'string', 'nullable' => false, 'longueur' => 1]],
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
            'version_ri' => 2,
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
        $proprietes = [['nom' => 'id', 'colonne' => 'id', 'type_doctrine' => 'integer', 'nullable' => false]];
        if ($avecEmail) {
            $proprietes[] = ['nom' => 'email', 'colonne' => 'email', 'type_doctrine' => 'string', 'nullable' => true];
        }

        return CalqueLogique::depuisTableau([
            'version_ri' => 2,
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
