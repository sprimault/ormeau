<?php

// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

declare(strict_types=1);

namespace Ormeau\Doctrine\Tests\Calque;

use Ormeau\Doctrine\Calque\ActionSuppression;
use Ormeau\Doctrine\Calque\CalqueInvalide;
use Ormeau\Doctrine\Calque\CalqueLogique;
use Ormeau\Doctrine\Calque\Lecture;
use Ormeau\Doctrine\Calque\Origine;
use Ormeau\Doctrine\Calque\StrategieIdentifiant;
use PHPUnit\Framework\Attributes\CoversClass;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;

/**
 * Ce qui se joue ici, ce sont les refus du modèle et ce qu'ils disent.
 *
 * Un calque accepté à tort produit des entités fausses plus loin, sans rien
 * pour le signaler ; un calque refusé sans chemin laisse l'utilisateur chercher
 * le champ fautif dans plusieurs milliers de lignes. La lecture des calques de
 * référence, elle, est dans ReferenceTest.
 */
#[CoversClass(CalqueLogique::class)]
#[CoversClass(Lecture::class)]
final class CalqueLogiqueTest extends TestCase
{
    /**
     * Les listes optionnelles absentes valent la liste vide, pas null : la
     * génération itère dessus sans avoir à s'en méfier.
     */
    public function testLesListesOptionnellesValentLaListeVide(): void
    {
        $donnees = self::calque();
        unset($donnees['avertissements']);
        $calque = CalqueLogique::depuisTableau($donnees);

        self::assertSame([], $calque->enumerations);
        self::assertSame([], $calque->traits);
        self::assertSame([], $calque->avertissements);

        $entite = $calque->entites[0];
        self::assertSame([], $entite->traits);
        self::assertSame([], $entite->associations);
        self::assertSame([], $entite->index);
        self::assertNull($entite->heritage);
    }

    /**
     * Les défauts du format sont appliqués quand le binaire omet le champ :
     * une propriété est insérable et modifiable sauf mention contraire.
     */
    public function testAppliqueLesDefautsDuFormat(): void
    {
        $propriete = CalqueLogique::depuisTableau(self::calque())->entites[0]->proprietes[0];

        self::assertTrue($propriete->insertable);
        self::assertTrue($propriete->modifiable);
        self::assertFalse($propriete->unique);
        self::assertNull($propriete->longueur);
        self::assertNull($propriete->defaut);
    }

    /**
     * La chaîne vide est un défaut, pas une absence : DEFAULT '' doit
     * survivre jusqu'à l'attribut Doctrine.
     */
    public function testLeDefautVideResteUnDefaut(): void
    {
        $donnees = self::calque();
        $donnees['entites'][0]['proprietes'][0]['defaut'] = '';

        self::assertSame('', CalqueLogique::depuisTableau($donnees)->entites[0]->proprietes[0]->defaut);
    }

    /**
     * Les vocabulaires fermés deviennent des énumérations PHP, et les objets
     * imbriqués sont construits jusqu'au bout.
     */
    public function testTypeLesObjetsImbriques(): void
    {
        $donnees = self::calque();
        $donnees['entites'][0]['associations'] = [[
            'nom' => 'pays',
            'genre' => 'plusieurs_vers_un',
            'cible' => 'Pays',
            'proprietaire' => true,
            'jointure' => [[
                'colonne' => 'pays_id',
                'colonne_referencee' => 'id',
                'nullable' => true,
                'a_la_suppression' => 'set_null',
            ]],
            'origine' => 'contrainte',
        ]];

        $entite = CalqueLogique::depuisTableau($donnees)->entites[0];

        self::assertSame(StrategieIdentifiant::Identite, $entite->identifiant?->strategie);
        self::assertSame(Origine::Contrainte, $entite->associations[0]->origine);
        self::assertSame(ActionSuppression::SetNull, $entite->associations[0]->jointure[0]->aLaSuppression);
    }

    /**
     * Un champ requis manquant est nommé par son chemin complet.
     *
     * @param list<string|int> $chemin clés à parcourir jusqu'au champ à retirer
     */
    #[DataProvider('champsRequis')]
    public function testRefuseUnChampRequisAbsentEnNommantSonChemin(array $chemin, string $attendu): void
    {
        $donnees = self::sans(self::calque(), $chemin);

        $this->expectException(CalqueInvalide::class);
        $this->expectExceptionMessage(sprintf('Champ %s absent du calque', $attendu));

        CalqueLogique::depuisTableau($donnees);
    }

    /**
     * Des champs requis à chaque profondeur, jusqu'aux avertissements dont
     * la confiance sert de filtre en CI.
     *
     * @return iterable<string, array{list<string|int>, string}>
     */
    public static function champsRequis(): iterable
    {
        yield 'racine' => [['empreinte_physique'], 'empreinte_physique'];
        yield 'entite' => [['entites', 0, 'table'], 'entites[0].table'];
        yield 'table' => [['entites', 0, 'table', 'schema'], 'entites[0].table.schema'];
        yield 'propriete' => [['entites', 0, 'proprietes', 1, 'type_doctrine'], 'entites[0].proprietes[1].type_doctrine'];
        yield 'confiance' => [['avertissements', 0, 'confiance'], 'avertissements[0].confiance'];
    }

    /**
     * Un champ du mauvais type est refusé, pas converti : "80" pour une
     * longueur ne vient pas du binaire.
     */
    public function testRefuseUnChampDuMauvaisType(): void
    {
        $donnees = self::calque();
        $donnees['entites'][0]['proprietes'][1]['longueur'] = '80';

        $this->expectException(CalqueInvalide::class);
        $this->expectExceptionMessage('Champ entites[0].proprietes[1].longueur : entier attendu');

        CalqueLogique::depuisTableau($donnees);
    }

    /**
     * Une valeur hors vocabulaire est refusée, et le message liste ce qui est
     * admis : c'est un calque abîmé ou d'une version plus récente.
     */
    public function testRefuseUneValeurHorsVocabulaire(): void
    {
        $donnees = self::calque();
        $donnees['entites'][0]['identifiant']['strategie'] = 'uuid';

        $this->expectException(CalqueInvalide::class);
        $this->expectExceptionMessage(
            'Champ entites[0].identifiant.strategie : « uuid » hors du vocabulaire (identite, sequence, aucune, assignee)',
        );

        CalqueLogique::depuisTableau($donnees);
    }

    /**
     * Une liste que le format veut non vide est refusée vide : un identifiant
     * sans propriété ne désignerait rien.
     */
    public function testRefuseUnIdentifiantSansPropriete(): void
    {
        $donnees = self::calque();
        $donnees['entites'][0]['identifiant']['proprietes'] = [];

        $this->expectException(CalqueInvalide::class);
        $this->expectExceptionMessage('Champ entites[0].identifiant.proprietes : au moins un élément attendu');

        CalqueLogique::depuisTableau($donnees);
    }

    /**
     * Un objet attendu qui arrive en liste est refusé : json_decode rend les
     * deux en tableau, seul le contrôle les distingue.
     */
    public function testRefuseUneListeALaPlaceDUnObjet(): void
    {
        $donnees = self::calque();
        $donnees['entites'][0]['table'] = ['clients', 'public'];

        $this->expectException(CalqueInvalide::class);
        $this->expectExceptionMessage('Champ entites[0].table : objet attendu');

        CalqueLogique::depuisTableau($donnees);
    }

    /**
     * La valeur d'un cas d'énumération admet deux types, et seulement ceux-là.
     */
    public function testLaValeurDUnCasEstUneChaineOuUnEntier(): void
    {
        $donnees = self::calque();
        $donnees['enumerations'] = [[
            'nom' => 'Statut',
            'type_support' => 'int',
            'cas' => [['nom' => 'Actif', 'valeur' => 1], ['nom' => 'Inactif', 'valeur' => 1.5]],
            'origine' => 'verification',
        ]];

        $this->expectException(CalqueInvalide::class);
        $this->expectExceptionMessage('Champ enumerations[0].cas[1].valeur : chaîne ou entier attendu');

        CalqueLogique::depuisTableau($donnees);
    }

    /**
     * Le filtre sur la confiance sert à la CI : elle échoue sur les
     * avertissements sûrs et laisse passer les incertains.
     */
    public function testFiltreLesAvertissementsParConfiance(): void
    {
        $donnees = self::calque();
        $donnees['avertissements'] = [
            self::avertissement('type_non_reconnu', 0.3),
            self::avertissement('table_sans_cle_primaire', 1),
            self::avertissement('fk_implicite_probable', 0.6),
        ];

        $retenus = CalqueLogique::depuisTableau($donnees)->avertissementsAuDessusDe(0.6);

        self::assertSame(
            ['table_sans_cle_primaire', 'fk_implicite_probable'],
            array_map(static fn($a): string => $a->code, $retenus),
        );
        self::assertSame([0, 1], array_keys($retenus));
    }

    /**
     * Le plus petit calque valide qui porte une entité complète.
     *
     * @return array<string, mixed>
     */
    private static function calque(): array
    {
        return [
            'version_ri' => 1,
            'empreinte_physique' => 'sha256:' . str_repeat('c', 64),
            'espace_de_noms' => 'App\\Entity',
            'entites' => [[
                'nom' => 'Client',
                'table' => ['nom' => 'clients', 'schema' => 'public'],
                'identifiant' => ['proprietes' => ['id'], 'strategie' => 'identite'],
                'proprietes' => [
                    ['nom' => 'id', 'colonne' => 'id', 'type_php' => 'int', 'type_doctrine' => 'integer', 'nullable' => false],
                    ['nom' => 'nom', 'colonne' => 'nom', 'type_php' => 'string', 'type_doctrine' => 'string', 'nullable' => false, 'longueur' => 80],
                ],
            ]],
            'avertissements' => [self::avertissement('table_sans_cle_primaire', 1)],
        ];
    }

    /**
     * Fabrique un avertissement valide.
     *
     * @return array<string, mixed>
     */
    private static function avertissement(string $code, float|int $confiance): array
    {
        return ['code' => $code, 'cible' => 'public.clients', 'message' => '', 'resolution' => 'aucune', 'confiance' => $confiance];
    }

    /**
     * Rend une copie des données sans le champ désigné par un chemin de clés.
     *
     * @param array<mixed>     $donnees
     * @param list<string|int> $chemin
     *
     * @return array<mixed>
     */
    private static function sans(array $donnees, array $chemin): array
    {
        $cle = array_shift($chemin);
        if ($chemin === []) {
            unset($donnees[$cle]);

            return $donnees;
        }

        $sous = $donnees[$cle];
        self::assertIsArray($sous);
        $donnees[$cle] = self::sans($sous, $chemin);

        return $donnees;
    }
}
