<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\Generation;

use LogicException;
use Ormeau\Doctrine\Calque\CalqueLogique;
use Ormeau\Doctrine\Generation\Cible;
use Ormeau\Doctrine\Generation\GenerateurEntite;
use Ormeau\Doctrine\Generation\ModeRegeneration;
use PHPUnit\Framework\Attributes\CoversClass;
use PHPUnit\Framework\TestCase;

/**
 * Le mode de régénération, et ce qui fait écarter une entité plutôt que la
 * générer à moitié. Le rendu lui-même est figé par AttendusTest, et validé par
 * Doctrine dans MappingTest.
 */
#[CoversClass(GenerateurEntite::class)]
final class GenerateurEntiteTest extends TestCase
{
    /**
     * La classe de base est le mode par défaut : le changer serait une rupture
     * pour tout projet déjà généré, et rien d'autre ne le signalerait.
     */
    public function testLeModeParDefautEstLaClasseDeBase(): void
    {
        self::assertSame(ModeRegeneration::ClasseDeBase, (new GenerateurEntite())->mode());
    }

    /**
     * Les valeurs de l'énumération sont un contrat : elles viennent d'un drapeau
     * de commande et d'une configuration Symfony, les renommer casse les deux.
     */
    public function testLesValeursDuModeSontStables(): void
    {
        self::assertSame('classe_de_base', ModeRegeneration::ClasseDeBase->value);
        self::assertSame('reecriture_ast', ModeRegeneration::ReecritureAst->value);
    }

    /**
     * La réécriture par AST n'est pas écrite : elle échoue franchement, avant
     * d'écrire quoi que ce soit, plutôt que de retomber sur l'autre mode.
     */
    public function testLaReecritureParAstEchoueSansEcrire(): void
    {
        $sortie = Repertoires::creer();
        try {
            (new GenerateurEntite(ModeRegeneration::ReecritureAst))->generer(self::calque([]), $sortie, Cible::forcer(3));
            self::fail('la réécriture par AST a généré');
        } catch (LogicException $e) {
            self::assertStringContainsString('reecriture_ast', $e->getMessage());
            self::assertSame([], Repertoires::lire($sortie));
        } finally {
            Repertoires::supprimer($sortie);
        }
    }

    /**
     * Une entité que Doctrine refuserait, ou dont une association vise une
     * entité absente ou écartée, est écartée avec sa raison ; l'écart se
     * propage de Commande à Livraison, et les autres s'écrivent.
     */
    public function testEcarteCeQuiNeSeGenerePasEntierEtEcritLeReste(): void
    {
        $sortie = Repertoires::creer();
        try {
            $rapport = (new GenerateurEntite())->generer(self::calque([
                self::entite('Client'),
                self::entite('Journal', ['identifiant' => null]),
                self::entite('List'),
                self::entite('Commande', ['associations' => [[
                    'nom' => 'fournisseur', 'genre' => 'plusieurs_vers_un', 'cible' => 'Fournisseur', 'proprietaire' => true, 'origine' => 'contrainte',
                ]]]),
                self::entite('Livraison', ['associations' => [[
                    'nom' => 'commande', 'genre' => 'plusieurs_vers_un', 'cible' => 'Commande', 'proprietaire' => true, 'origine' => 'contrainte',
                ]]]),
                self::entite('Etiquette', ['table' => ['nom' => 'etiquette', 'schema' => 'public']]),
                self::entite('etiquette', ['table' => ['nom' => 'etiquette', 'schema' => 'archive']]),
            ]), $sortie, Cible::forcer(3));

            self::assertSame([
                'Journal' => 'la table n\'a pas de clé primaire, et Doctrine exige un identifiant',
                'List' => 'List est un mot réservé de PHP, à renommer dans renommages',
                'Commande' => 'l\'association fournisseur vise Fournisseur, absente du calque',
                'Livraison' => 'l\'association commande vise Commande (public.commande), écartée',
                'Etiquette' => 'le nom Etiquette est porté par plusieurs entités, à départager dans renommages',
                'etiquette' => 'le nom etiquette est porté par plusieurs entités, à départager dans renommages',
            ], array_column(array_map(static fn($e): array => [$e->nom, $e->raison], $rapport->ecartees), 1, 0));
            self::assertSame(['Base/ClientBase.php', 'Client.php'], array_keys(Repertoires::lire($sortie)));
        } finally {
            Repertoires::supprimer($sortie);
        }
    }

    /**
     * Un défaut sur une colonne énumérée initialise la propriété avec le cas
     * qui porte cette valeur, y compris pour une énumération adossée à des
     * entiers, où le calque écrit le défaut en texte. La colonne garde le
     * défaut dans le type stocké.
     */
    public function testUnDefautEnumereInitialiseAvecLeCas(): void
    {
        $sortie = Repertoires::creer();
        try {
            $client = self::entite('Client');
            $client['proprietes'][] = self::propriete('statut', 'string', ['enumeration' => 'Statut', 'defaut' => 'A']);
            $client['proprietes'][] = self::propriete('niveau', 'integer', ['enumeration' => 'Niveau', 'defaut' => '2']);
            $client['proprietes'][] = self::propriete('canal', 'string', ['enumeration' => 'Statut', 'defaut' => 'Z', 'nullable' => true]);

            (new GenerateurEntite())->generer(self::calque([$client], [
                ['nom' => 'Statut', 'type_support' => 'string', 'cas' => [['nom' => 'Actif', 'valeur' => 'A']], 'origine' => 'verification'],
                ['nom' => 'Niveau', 'type_support' => 'int', 'cas' => [['nom' => 'Un', 'valeur' => 1], ['nom' => 'Deux', 'valeur' => 2]], 'origine' => 'verification'],
            ]), $sortie, Cible::forcer(3));

            $base = Repertoires::lire($sortie)['Base/ClientBase.php'];
            self::assertStringContainsString("enumType: Statut::class, options: ['default' => 'A'])]\n    protected Statut \$statut = Statut::Actif;", $base);
            self::assertStringContainsString("enumType: Niveau::class, options: ['default' => 2])]\n    protected Niveau \$niveau = Niveau::Deux;", $base);
            self::assertStringContainsString('protected ?Statut $canal = null;', $base);
            self::assertStringContainsString("enum Niveau: int\n{\n    case Un = 1;\n    case Deux = 2;\n}", Repertoires::lire($sortie)['Enum/Niveau.php']);
        } finally {
            Repertoires::supprimer($sortie);
        }
    }

    /**
     * Une entité qui nomme un trait ou une énumération absents du calque, ou
     * que PHP refuse d'écrire, est écartée avec le nom en cause ; le fichier
     * refusé n'est pas écrit.
     */
    public function testEcarteUneEntiteDontLeTraitOuLEnumerationNeSEcritPas(): void
    {
        $sortie = Repertoires::creer();
        try {
            $sansTrait = self::entite('Client', ['traits' => ['Horodatage']]);
            $enumReservee = self::entite('Commande');
            $enumReservee['proprietes'][] = self::propriete('etat', 'string', ['enumeration' => 'Match']);
            $casReserve = self::entite('Facture');
            $casReserve['proprietes'][] = self::propriete('mode', 'string', ['enumeration' => 'Mode']);

            $rapport = (new GenerateurEntite())->generer(self::calque([$sansTrait, $enumReservee, $casReserve], [
                ['nom' => 'Match', 'type_support' => 'string', 'cas' => [['nom' => 'Oui', 'valeur' => 'O']], 'origine' => 'decision'],
                ['nom' => 'Mode', 'type_support' => 'string', 'cas' => [['nom' => 'Class', 'valeur' => 'C']], 'origine' => 'decision'],
            ]), $sortie, Cible::forcer(3));

            self::assertSame([
                'Client' => 'le trait Horodatage est absent du calque',
                'Commande' => 'l\'énumération Match n\'est pas générée : Match est un mot réservé de PHP',
                'Facture' => 'l\'énumération Mode n\'est pas générée : un cas ne peut pas s\'appeler class',
            ], array_column(array_map(static fn($e): array => [$e->nom, $e->raison], $rapport->ecartees), 1, 0));
            self::assertSame([], Repertoires::lire($sortie));
        } finally {
            Repertoires::supprimer($sortie);
        }
    }

    /**
     * Une entité identifiée par une association vers une entité elle-même
     * identifiée par une association est écartée : Doctrine refuse cette
     * identité en chaîne. La raison cite les deux tables physiques et la
     * sortie, et une entité qui vise l'entité écartée a sa propre raison.
     */
    public function testEcarteUneIdentiteDeriveeEnChaine(): void
    {
        $sortie = Repertoires::creer();
        try {
            $personne = self::entite('Personne');
            $salarie = self::entite('Salarie', [
                'identifiant' => ['proprietes' => ['id'], 'strategie' => 'assignee'],
                'associations' => [self::jointure('personne', 'un_vers_un', 'Personne', 'id')],
            ]);
            $affectation = self::entite('Affectation', [
                'identifiant' => ['proprietes' => ['salarieId', 'debut'], 'strategie' => 'assignee'],
                'proprietes' => [
                    self::propriete('salarieId', 'integer', ['colonne' => 'salarie_id']),
                    self::propriete('debut', 'string'),
                ],
                'associations' => [self::jointure('salarie', 'plusieurs_vers_un', 'Salarie', 'salarie_id')],
            ]);
            $note = self::entite('Note', ['associations' => [self::jointure('affectation', 'plusieurs_vers_un', 'Affectation', 'affectation_id')]]);

            $rapport = (new GenerateurEntite())->generer(self::calque([$personne, $salarie, $affectation, $note]), $sortie, Cible::forcer(3));

            self::assertSame([
                'Affectation' => 'Doctrine ne sait pas identifier Affectation (public.affectation) par salarie : '
                    . 'Salarie (public.salarie) est elle-même identifiée par une association. '
                    . 'Déclarer public.salarie dans un héritage (heritages), ou écarter l\'une des deux tables (tables_ignorees)',
                'Note' => 'l\'association affectation vise Affectation (public.affectation), écartée',
            ], array_column(array_map(static fn($e): array => [$e->nom, $e->raison], $rapport->ecartees), 1, 0));
            self::assertArrayHasKey('Base/SalarieBase.php', Repertoires::lire($sortie));
            self::assertSame([], $rapport->omises, 'Salarie ne porte pas de côté inverse ici');
        } finally {
            Repertoires::supprimer($sortie);
        }
    }

    /**
     * Une table fille dont la clé ne porte pas le nom de celle de la racine
     * écarte toute la hiérarchie : Doctrine joint par les noms de la racine,
     * et la carte de la racine citerait une classe absente. La raison nomme
     * les colonnes et la sortie ; l'entité qui vise la classe fille a la sienne.
     */
    public function testEcarteUneHierarchieDontLaCleDUneFillePorteUnAutreNom(): void
    {
        $sortie = Repertoires::creer();
        try {
            $rapport = (new GenerateurEntite())->generer(self::calque([
                self::entite('Personne', ['valeur_discriminante' => 'P', 'proprietes' => [self::propriete('id', 'integer'), self::propriete('nature', 'string')]]),
                self::entite('Salarie', [
                    'heritage' => self::heritage('Personne'),
                    'valeur_discriminante' => 'S',
                    'identifiant' => ['proprietes' => ['personneId'], 'strategie' => 'assignee'],
                    'proprietes' => [self::propriete('personneId', 'integer', ['colonne' => 'personne_id'])],
                ]),
                self::entite('Contrat', ['associations' => [self::jointure('salarie', 'plusieurs_vers_un', 'Salarie', 'salarie_id')]]),
            ]), $sortie, Cible::forcer(3));

            self::assertSame([
                'Personne' => 'même hiérarchie que Salarie (public.salarie), écartée',
                'Salarie' => 'Doctrine joint public.salarie à public.personne par les colonnes de clé de la racine (id), '
                    . 'que public.salarie nomme personne_id. Retirer public.salarie des valeurs de heritages',
                'Contrat' => 'l\'association salarie vise Salarie (public.salarie), écartée',
            ], array_column(array_map(static fn($e): array => [$e->nom, $e->raison], $rapport->ecartees), 1, 0));
            self::assertSame([], Repertoires::lire($sortie));
        } finally {
            Repertoires::supprimer($sortie);
        }
    }

    /**
     * Ce qu'aucune inférence ne produit mais qu'un calque écrit à la main peut
     * contenir — table unique, valeur manquante, parent absent — écarte la
     * hiérarchie plutôt que d'écrire une carte fausse.
     */
    public function testEcarteUnHeritageQueLeCalqueNeDecritPasEntier(): void
    {
        $sortie = Repertoires::creer();
        try {
            $rapport = (new GenerateurEntite())->generer(self::calque([
                self::entite('Vehicule', ['valeur_discriminante' => 'V']),
                self::entite('Voiture', ['heritage' => ['strategie' => 'table_unique', 'parent' => 'Vehicule', 'colonne_discriminante' => 'genre'], 'valeur_discriminante' => 'C']),
                self::entite('Animal', ['valeur_discriminante' => 'A']),
                self::entite('Chien', ['heritage' => self::heritage('Animal')]),
                self::entite('Rose', ['heritage' => self::heritage('Plante'), 'valeur_discriminante' => 'R']),
            ]), $sortie, Cible::forcer(3));

            self::assertSame([
                'Vehicule' => 'même hiérarchie que Voiture (public.voiture), écartée',
                'Voiture' => 'héritage en table unique, que ni l\'inférence ni les décisions ne produisent : ce paquet ne le génère pas',
                'Animal' => 'même hiérarchie que Chien (public.chien), écartée',
                'Chien' => 'la hiérarchie de Animal est incomplète : colonne discriminante ou valeur de Chien absente du calque',
                'Rose' => 'le parent Plante est absent du calque',
            ], array_column(array_map(static fn($e): array => [$e->nom, $e->raison], $rapport->ecartees), 1, 0));
        } finally {
            Repertoires::supprimer($sortie);
        }
    }

    /**
     * Une colonne discriminante retirée des propriétés par colonnes_ignorees
     * reste déclarée sur la racine, sans le type ni la longueur qu'aucune
     * propriété ne donne plus : Doctrine prend alors ses défauts.
     */
    public function testUneColonneDiscriminanteSansProprieteResteDeclaree(): void
    {
        $sortie = Repertoires::creer();
        try {
            (new GenerateurEntite())->generer(self::calque([
                self::entite('Personne', ['valeur_discriminante' => 'P']),
                self::entite('Salarie', ['heritage' => self::heritage('Personne'), 'valeur_discriminante' => 'S', 'identifiant' => ['proprietes' => ['id'], 'strategie' => 'assignee']]),
            ]), $sortie, Cible::forcer(3));

            self::assertStringContainsString(
                "#[ORM\\DiscriminatorColumn(name: 'nature')]\n",
                Repertoires::lire($sortie)['Personne.php'] ?? '',
            );
        } finally {
            Repertoires::supprimer($sortie);
        }
    }

    /**
     * Un héritage joint décidé, sur la colonne discriminante nature.
     *
     * @return array<string, string>
     */
    private static function heritage(string $parent): array
    {
        return ['strategie' => 'jointe', 'parent' => $parent, 'colonne_discriminante' => 'nature', 'origine' => 'decision'];
    }

    /**
     * Une association propriétaire sur une colonne de jointure vers l'id de sa
     * cible.
     *
     * @return array<string, mixed>
     */
    private static function jointure(string $nom, string $genre, string $cible, string $colonne): array
    {
        return [
            'nom' => $nom,
            'genre' => $genre,
            'cible' => $cible,
            'proprietaire' => true,
            'jointure' => [['colonne' => $colonne, 'colonne_referencee' => 'id', 'nullable' => false]],
            'origine' => 'contrainte',
        ];
    }

    /**
     * Un calque réduit aux entités et énumérations données.
     *
     * @param list<array<string, mixed>> $entites      entités du calque
     * @param list<array<string, mixed>> $enumerations énumérations du calque
     */
    private static function calque(array $entites, array $enumerations = []): CalqueLogique
    {
        return CalqueLogique::depuisTableau([
            'version_ri' => 1,
            'empreinte_physique' => 'sha256:' . str_repeat('b', 64),
            'espace_de_noms' => 'App\\Entity',
            'entites' => $entites,
            'enumerations' => $enumerations,
        ]);
    }

    /**
     * Une propriété non nullable du type Doctrine donné, modifiable par les
     * champs donnés.
     *
     * @param array<string, mixed> $modifications champs à ajouter ou remplacer
     *
     * @return array<string, mixed>
     */
    private static function propriete(string $nom, string $typeDoctrine, array $modifications = []): array
    {
        return array_merge(
            ['nom' => $nom, 'colonne' => $nom, 'type_php' => 'string', 'type_doctrine' => $typeDoctrine, 'nullable' => false],
            $modifications,
        );
    }

    /**
     * Une entité à clé IDENTITY, modifiable par les champs donnés ; un champ
     * à null est retiré.
     *
     * @param array<string, mixed> $modifications champs à ajouter ou remplacer ; null en retire un
     *
     * @return array<string, mixed>
     */
    private static function entite(string $nom, array $modifications = []): array
    {
        $entite = array_merge([
            'nom' => $nom,
            'table' => ['nom' => strtolower($nom), 'schema' => 'public'],
            'identifiant' => ['proprietes' => ['id'], 'strategie' => 'identite'],
            'proprietes' => [['nom' => 'id', 'colonne' => 'id', 'type_php' => 'int', 'type_doctrine' => 'integer', 'nullable' => false]],
        ], $modifications);

        return array_filter($entite, static fn($valeur): bool => $valeur !== null);
    }
}
